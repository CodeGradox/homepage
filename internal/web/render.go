package web

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// renderer holds one fully-parsed template per page. Each page is parsed on top
// of the shared layout so that its {{define "title"}} / {{define "content"}}
// blocks override the layout's defaults — the html/template equivalent of a
// Rails layout with content_for.
type renderer struct {
	pages map[string]*template.Template
}

// newRenderer parses layout.html.tmpl together with every other *.html.tmpl in
// tmplFS, wiring in the template helpers (asset, importmapTags).
func newRenderer(tmplFS fs.FS, funcs template.FuncMap) (*renderer, error) {
	layout, err := fs.ReadFile(tmplFS, "layout.html.tmpl")
	if err != nil {
		return nil, fmt.Errorf("read layout: %w", err)
	}

	names, err := fs.Glob(tmplFS, "*.html.tmpl")
	if err != nil {
		return nil, err
	}

	r := &renderer{pages: map[string]*template.Template{}}
	for _, name := range names {
		if name == "layout.html.tmpl" {
			continue
		}
		page, err := fs.ReadFile(tmplFS, name)
		if err != nil {
			return nil, err
		}

		t := template.New("layout").Funcs(funcs)
		if _, err := t.Parse(string(layout)); err != nil {
			return nil, fmt.Errorf("parse layout for %s: %w", name, err)
		}
		if _, err := t.Parse(string(page)); err != nil {
			return nil, fmt.Errorf("parse %s: %w", name, err)
		}

		key := strings.TrimSuffix(path.Base(name), ".html.tmpl")
		r.pages[key] = t
	}
	return r, nil
}

// render writes the named page. It buffers first so that a template error
// surfaces as a clean 500 instead of a half-written response. The body is
// fingerprinted into an ETag for conditional GETs; because the import map JSON
// is inlined into the page, the ETag changes whenever a pinned module changes —
// the equivalent of Rails' stale_when_importmap_changes.
func (r *renderer) render(w http.ResponseWriter, req *http.Request, page string, data any) {
	t, ok := r.pages[page]
	if !ok {
		http.Error(w, "template not found: "+page, http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, "layout", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sum := sha256.Sum256(buf.Bytes())
	etag := `"` + hex.EncodeToString(sum[:])[:16] + `"`
	w.Header().Set("ETag", etag)
	if match := req.Header.Get("If-None-Match"); match == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = buf.WriteTo(w)
}
