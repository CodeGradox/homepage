package main

import (
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"homepage/internal/assets"
	"homepage/internal/logging"
	"homepage/internal/web"
)

// newTestHandler builds the real server from the embedded templates and assets,
// exercising the actual import map pins and template helpers.
func newTestHandler(t *testing.T) http.Handler {
	t.Helper()

	templateFS := mustSub(t, templatesDir, "templates")
	assetFS := mustSub(t, assetsDir, "assets")
	publicFS := mustSub(t, publicDir, "public")

	pipeline, err := assets.Load(assetFS)
	if err != nil {
		t.Fatalf("assets.Load: %v", err)
	}
	server, err := web.New(logging.New(io.Discard, slog.LevelError), templateFS, pipeline)
	if err != nil {
		t.Fatalf("web.New: %v", err)
	}
	return server.Handler(publicFS)
}

func mustSub(t *testing.T, fsys fs.FS, dir string) fs.FS {
	t.Helper()
	sub, err := fs.Sub(fsys, dir)
	if err != nil {
		t.Fatalf("fs.Sub(%s): %v", dir, err)
	}
	return sub
}

func TestPagesRender(t *testing.T) {
	h := newTestHandler(t)

	cases := []struct {
		path     string
		contains string
	}{
		{"/", "<title>Magnus Åsrud</title>"},
		{"/speed_reader", "<title>Speed Reader</title>"},
		{"/scrollable_table_patterns", "AC-001"},
	}
	for _, c := range cases {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, c.path, nil))

		if rec.Code != http.StatusOK {
			t.Errorf("%s: status = %d, want 200", c.path, rec.Code)
		}
		if !strings.Contains(rec.Body.String(), c.contains) {
			t.Errorf("%s: body missing %q", c.path, c.contains)
		}
	}
}

func TestImportMapAndFingerprintedAssetsAreWired(t *testing.T) {
	h := newTestHandler(t)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	body := rec.Body.String()

	if !strings.Contains(body, `<script type="importmap">`) {
		t.Error("homepage is missing the import map")
	}
	// Stylesheet should be referenced by its digested URL.
	if !strings.Contains(body, "/assets/stylesheets/style-") {
		t.Error("stylesheet is not fingerprinted")
	}
}

func TestHealthCheck(t *testing.T) {
	rec := httptest.NewRecorder()
	newTestHandler(t).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/up", nil))
	if rec.Code != http.StatusOK {
		t.Errorf("/up status = %d, want 200", rec.Code)
	}
}

func TestETagConditionalGet(t *testing.T) {
	h := newTestHandler(t)

	first := httptest.NewRecorder()
	h.ServeHTTP(first, httptest.NewRequest(http.MethodGet, "/", nil))
	etag := first.Header().Get("ETag")
	if etag == "" {
		t.Fatal("no ETag on homepage response")
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("If-None-Match", etag)
	second := httptest.NewRecorder()
	h.ServeHTTP(second, req)

	if second.Code != http.StatusNotModified {
		t.Errorf("conditional GET status = %d, want 304", second.Code)
	}
}

func TestPublicFilesServed(t *testing.T) {
	rec := httptest.NewRecorder()
	newTestHandler(t).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/robots.txt", nil))
	if rec.Code != http.StatusOK {
		t.Errorf("/robots.txt status = %d, want 200", rec.Code)
	}
}
