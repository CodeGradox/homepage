package web

import (
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"homepage/internal/assets"
	"homepage/internal/importmap"
)

// Server holds everything needed to serve the site: the rendered pages, the
// asset pipeline, and the precomputed scrollable-tables dataset.
type Server struct {
	logger   *slog.Logger
	renderer *renderer
	assets   *assets.Pipeline
	tables   tableData
}

// New builds a Server from the embedded template and asset trees.
func New(logger *slog.Logger, tmplFS fs.FS, pipeline *assets.Pipeline) (*Server, error) {
	imap := importmap.New(pipeline)

	funcs := template.FuncMap{
		"asset":         pipeline.Path,
		"importmapTags": imap.Tags,
	}

	r, err := newRenderer(tmplFS, funcs)
	if err != nil {
		return nil, err
	}

	return &Server{
		logger:   logger,
		renderer: r,
		assets:   pipeline,
		tables:   buildTableData(),
	}, nil
}

// Handler returns the fully-wired HTTP handler: routes plus the logging and
// recovery middleware.
func (s *Server) Handler(publicFS fs.FS) http.Handler {
	mux := http.NewServeMux()

	// Pages.
	mux.HandleFunc("GET /{$}", s.home)
	mux.HandleFunc("GET /speed_reader", s.speedReader)
	mux.HandleFunc("GET /scrollable_table_patterns", s.scrollableTablePatterns)
	mux.HandleFunc("GET /wcag_contrast", s.wcagContrast)

	// Theme preference — the sidebar toggle posts here; the choice is stored in
	// a cookie so it survives across visits without any client-side JavaScript.
	mux.HandleFunc("POST /theme", s.setTheme)

	// Health check — returns 200 if the app is up. Matches Rails' /up.
	mux.HandleFunc("GET /up", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Fingerprinted assets and unfingerprinted public files (robots.txt, icons).
	mux.Handle("GET "+assets.URLPrefix, s.assets.Handler())
	mux.Handle("GET /", http.FileServerFS(publicFS))

	return recoverer(s.logger, requestLogger(s.logger, mux))
}

func (s *Server) home(w http.ResponseWriter, r *http.Request) {
	s.renderer.render(w, r, "homepage", nil)
}

func (s *Server) speedReader(w http.ResponseWriter, r *http.Request) {
	s.renderer.render(w, r, "speed_reader", nil)
}

func (s *Server) scrollableTablePatterns(w http.ResponseWriter, r *http.Request) {
	s.renderer.render(w, r, "scrollable_table_patterns", s.tables)
}

func (s *Server) wcagContrast(w http.ResponseWriter, r *http.Request) {
	s.renderer.render(w, r, "wcag_contrast", nil)
}

// setTheme stores the visitor's theme choice in a cookie and sends them back to
// the page they came from. Choosing "system" clears the cookie instead, so the
// site falls back to the OS preference.
func (s *Server) setTheme(w http.ResponseWriter, r *http.Request) {
	cookie := &http.Cookie{
		Name:     themeCookie,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}

	switch theme := r.FormValue("theme"); theme {
	case "light", "dark":
		cookie.Value = theme
		cookie.MaxAge = int((365 * 24 * time.Hour).Seconds())
	case "system":
		cookie.MaxAge = -1
	default:
		http.Error(w, "invalid theme", http.StatusBadRequest)
		return
	}
	http.SetCookie(w, cookie)

	http.Redirect(w, r, returnPath(r.Referer()), http.StatusSeeOther)
}

// pagePaths is the allowlist for the theme toggle's return redirect — exactly
// the site's page routes.
var pagePaths = []string{"/", "/speed_reader", "/scrollable_table_patterns", "/wcag_contrast"}

// returnPath turns the Referer header into a safe redirect target, falling
// back to the homepage unless it matches a known page. Returning the allowlist
// constant rather than the parsed value means no attacker-controlled bytes can
// reach the redirect.
func returnPath(referer string) string {
	u, err := url.Parse(referer)
	if err != nil {
		return "/"
	}
	for _, p := range pagePaths {
		if u.Path == p {
			return p
		}
	}
	return "/"
}
