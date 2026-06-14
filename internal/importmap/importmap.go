// Package importmap renders an HTML import map: the browser-native mechanism
// for mapping bare module specifiers ("controllers/index") to URLs. It is the
// Go counterpart of config/importmap.rb. There is no bundler and no build step
// — the map is JSON injected into the page, and the browser loads ES modules
// directly. Each pin resolves to a fingerprinted asset URL, so a changed module
// produces a changed map, which is the basis for cache busting.
package importmap

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"html/template"
)

// resolver turns a logical asset name into its public, fingerprinted URL.
// assets.Pipeline satisfies this.
type resolver interface {
	Path(logical string) string
}

type pin struct {
	specifier string // the bare specifier used in `import "..."`
	asset     string // logical asset name to resolve to a URL
}

// Map is an ordered set of pins plus the module entrypoint.
type Map struct {
	resolver   resolver
	pins       []pin
	entrypoint string
}

// New builds the import map. The pins mirror config/importmap.rb: the
// application entrypoint, the Stimulus vendor file, and one entry per Stimulus
// controller. Order is preserved purely for stable, readable HTML output.
func New(r resolver) *Map {
	return &Map{
		resolver:   r,
		entrypoint: "application",
		pins: []pin{
			{"application", "javascript/application.js"},
			{"@hotwired/stimulus", "javascript/stimulus.min.js"},
			{"controllers", "javascript/controllers/index.js"},
			{"controllers/application", "javascript/controllers/application.js"},
			{"controllers/index", "javascript/controllers/index.js"},
			{"controllers/speed_reader_controller", "javascript/controllers/speed_reader_controller.js"},
			{"controllers/scrollable_table_patterns_controller", "javascript/controllers/scrollable_table_patterns_controller.js"},
		},
	}
}

// Digest is a short, content-derived fingerprint of the whole map. Changing any
// pinned module changes its URL and therefore this value, which the layout uses
// as an ETag — the equivalent of Rails' stale_when_importmap_changes.
func (m *Map) Digest() string {
	imports := m.imports()
	// Render to canonical JSON; encoding/json sorts map keys, so the digest is
	// stable regardless of pin order.
	b, _ := json.Marshal(imports)
	return shortHash(b)
}

// Tags renders the two <script> tags for the document head: the import map
// itself, followed by the module entrypoint. The output is the analogue of
// Rails' javascript_importmap_tags. Modern browsers (which this site requires)
// support import maps natively, so no es-module-shims polyfill is emitted.
func (m *Map) Tags() template.HTML {
	imports := m.imports()
	doc := struct {
		Imports map[string]string `json:"imports"`
	}{Imports: imports}

	body, _ := json.MarshalIndent(doc, "", "  ")

	var b []byte
	b = append(b, `<script type="importmap">`...)
	b = append(b, body...)
	b = append(b, "</script>\n"...)
	b = append(b, `<script type="module">import "`...)
	b = append(b, template.JSEscapeString(m.entrypoint)...)
	b = append(b, `"</script>`...)
	// #nosec G203 -- the map is built entirely from server-controlled asset
	// paths and a constant entrypoint; there is no user input to escape.
	return template.HTML(b)
}

func shortHash(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])[:16]
}

func (m *Map) imports() map[string]string {
	imports := make(map[string]string, len(m.pins))
	for _, p := range m.pins {
		imports[p.specifier] = m.resolver.Path(p.asset)
	}
	return imports
}
