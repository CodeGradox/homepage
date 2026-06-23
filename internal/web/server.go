package web

import (
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"

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
