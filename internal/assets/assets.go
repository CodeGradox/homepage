// Package assets is a small content-addressed asset pipeline — a Propshaft in
// miniature. At startup it walks the embedded assets tree, fingerprints every
// file with a digest of its contents, and serves the digested copies with a
// far-future cache header. Logical names ("stylesheets/style.css") resolve to
// digested URLs ("/assets/stylesheets/style-1a2b3c4d.css") so a changed file
// always gets a fresh URL and an immutable cache entry.
package assets

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"mime"
	"net/http"
	"path"
	"strings"
	"time"
)

// URLPrefix is the path under which all digested assets are served.
const URLPrefix = "/assets/"

type asset struct {
	digestedPath string // logical path with the digest spliced in
	body         []byte
	contentType  string
}

// Pipeline holds the fingerprinted assets and resolves logical names to URLs.
type Pipeline struct {
	byLogical  map[string]*asset // "stylesheets/style.css"        -> asset
	byDigested map[string]*asset // "stylesheets/style-abc123.css" -> asset
}

// Load reads every file in fsys (an asset tree rooted at the assets directory),
// fingerprints it, and returns a ready pipeline. It fails fast: a broken asset
// tree is a programming error, not a runtime condition.
func Load(fsys fs.FS) (*Pipeline, error) {
	p := &Pipeline{
		byLogical:  map[string]*asset{},
		byDigested: map[string]*asset{},
	}

	err := fs.WalkDir(fsys, ".", func(name string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}

		body, err := fs.ReadFile(fsys, name)
		if err != nil {
			return err
		}

		logical := name
		a := &asset{
			digestedPath: digestedName(logical, body),
			body:         body,
			contentType:  contentType(logical),
		}
		p.byLogical[logical] = a
		p.byDigested[a.digestedPath] = a
		return nil
	})
	if err != nil {
		return nil, err
	}
	return p, nil
}

// Path returns the public URL for a logical asset name, e.g.
// Path("stylesheets/style.css") -> "/assets/stylesheets/style-abc123.css".
// It panics on an unknown name: assets are known at build time, so a miss is a
// bug in a template, surfaced loudly rather than served as a 404.
func (p *Pipeline) Path(logical string) string {
	a, ok := p.byLogical[logical]
	if !ok {
		panic(fmt.Sprintf("assets: no such asset %q", logical))
	}
	return URLPrefix + a.digestedPath
}

// Handler serves digested assets. Because the URL changes whenever the content
// changes, responses are immutable and cached for a year.
func (p *Pipeline) Handler() http.Handler {
	return http.StripPrefix(URLPrefix, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a, ok := p.byDigested[r.URL.Path]
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", a.contentType)
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		http.ServeContent(w, r, a.digestedPath, time.Time{}, strings.NewReader(string(a.body)))
	}))
}

// digestedName splices an 8-char content digest before the extension:
// "stylesheets/style.css" -> "stylesheets/style-1a2b3c4d.css".
func digestedName(logical string, body []byte) string {
	sum := sha256.Sum256(body)
	digest := hex.EncodeToString(sum[:])[:8]

	dir, file := path.Split(logical)
	ext := path.Ext(file)
	base := strings.TrimSuffix(file, ext)
	return dir + base + "-" + digest + ext
}

func contentType(logical string) string {
	if ct := mime.TypeByExtension(path.Ext(logical)); ct != "" {
		return ct
	}
	return "application/octet-stream"
}
