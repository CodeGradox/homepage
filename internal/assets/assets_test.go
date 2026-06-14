package assets

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
)

func testPipeline(t *testing.T) *Pipeline {
	t.Helper()
	fsys := fstest.MapFS{
		"stylesheets/style.css":     {Data: []byte("body{color:red}")},
		"javascript/application.js": {Data: []byte("export default 1")},
		"images/svg/go-logo.svg":    {Data: []byte("<svg></svg>")},
	}
	p, err := Load(fsys)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	return p
}

func TestPathInsertsDigestBeforeExtension(t *testing.T) {
	p := testPipeline(t)
	got := p.Path("stylesheets/style.css")

	const wantPrefix = "/assets/stylesheets/style-"
	if len(got) <= len(wantPrefix) || got[:len(wantPrefix)] != wantPrefix {
		t.Fatalf("Path = %q, want prefix %q", got, wantPrefix)
	}
	if got[len(got)-4:] != ".css" {
		t.Fatalf("Path = %q, want .css suffix", got)
	}
}

func TestPathIsStableForSameContent(t *testing.T) {
	if testPipeline(t).Path("stylesheets/style.css") != testPipeline(t).Path("stylesheets/style.css") {
		t.Fatal("digest is not deterministic for identical content")
	}
}

func TestPathPanicsOnUnknownAsset(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for unknown asset")
		}
	}()
	testPipeline(t).Path("does/not/exist.css")
}

func TestHandlerServesDigestedAssetWithImmutableCache(t *testing.T) {
	p := testPipeline(t)
	url := p.Path("stylesheets/style.css")

	req := httptest.NewRequest(http.MethodGet, url, nil)
	rec := httptest.NewRecorder()
	p.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/css; charset=utf-8" {
		t.Fatalf("content-type = %q", ct)
	}
	if cc := rec.Header().Get("Cache-Control"); cc != "public, max-age=31536000, immutable" {
		t.Fatalf("cache-control = %q", cc)
	}
	if rec.Body.String() != "body{color:red}" {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

func TestHandlerReturns404ForUndigestedPath(t *testing.T) {
	p := testPipeline(t)
	req := httptest.NewRequest(http.MethodGet, "/assets/stylesheets/style.css", nil)
	rec := httptest.NewRecorder()
	p.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}
