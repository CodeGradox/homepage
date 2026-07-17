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
		{"/wcag_contrast", "<title>WCAG Contrast Checker</title>"},
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

func TestThemeToggleRendered(t *testing.T) {
	h := newTestHandler(t)
	for _, path := range []string{"/", "/speed_reader", "/scrollable_table_patterns", "/wcag_contrast"} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if !strings.Contains(rec.Body.String(), `<form class="theme-toggle" method="post" action="/theme">`) {
			t.Errorf("%s: sidebar is missing the theme toggle", path)
		}
	}
}

func TestThemeCookieControlsDataThemeAttribute(t *testing.T) {
	h := newTestHandler(t)

	cases := []struct {
		cookie   string // empty means no cookie sent
		contains string
		excludes string
	}{
		{"", "", `data-theme`},
		{"light", `<html lang="en" data-theme="light">`, ""},
		{"dark", `<html lang="en" data-theme="dark">`, ""},
		{"bogus", "", `data-theme`}, // unknown values fall back to system
	}
	for _, c := range cases {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		if c.cookie != "" {
			req.AddCookie(&http.Cookie{Name: "theme", Value: c.cookie})
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		body := rec.Body.String()
		if c.contains != "" && !strings.Contains(body, c.contains) {
			t.Errorf("cookie %q: body missing %q", c.cookie, c.contains)
		}
		if c.excludes != "" && strings.Contains(body, c.excludes) {
			t.Errorf("cookie %q: body unexpectedly contains %q", c.cookie, c.excludes)
		}
	}
}

func TestSetThemeStoresCookieAndRedirects(t *testing.T) {
	h := newTestHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/theme", strings.NewReader("theme=dark"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", "http://example.com/speed_reader")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "/speed_reader" {
		t.Errorf("Location = %q, want /speed_reader", loc)
	}

	cookies := rec.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != "theme" || cookies[0].Value != "dark" {
		t.Fatalf("cookies = %v, want one theme=dark cookie", cookies)
	}
	if cookies[0].MaxAge <= 0 {
		t.Errorf("theme cookie MaxAge = %d, want a positive lifetime", cookies[0].MaxAge)
	}
}

func TestSetThemeSystemClearsCookie(t *testing.T) {
	h := newTestHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/theme", strings.NewReader("theme=system"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "/" {
		t.Errorf("Location = %q, want / when Referer is absent", loc)
	}

	cookies := rec.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != "theme" || cookies[0].MaxAge >= 0 {
		t.Fatalf("cookies = %v, want an expired theme cookie", cookies)
	}
}

func TestSetThemeRejectsUnknownValues(t *testing.T) {
	h := newTestHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/theme", strings.NewReader("theme=hotdog"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
	if len(rec.Result().Cookies()) != 0 {
		t.Error("invalid theme should not set a cookie")
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
