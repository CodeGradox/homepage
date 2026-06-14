package importmap

import (
	"strings"
	"testing"
)

// stubResolver returns a fixed URL per logical name and records misses.
type stubResolver struct{ paths map[string]string }

func (s stubResolver) Path(logical string) string {
	if p, ok := s.paths[logical]; ok {
		return p
	}
	return "/assets/MISSING/" + logical
}

func newTestMap() *Map {
	return New(stubResolver{paths: map[string]string{
		"javascript/application.js":                                      "/assets/application-aaa.js",
		"javascript/stimulus.min.js":                                     "/assets/stimulus-bbb.js",
		"javascript/controllers/index.js":                                "/assets/index-ccc.js",
		"javascript/controllers/application.js":                          "/assets/c-app-ddd.js",
		"javascript/controllers/speed_reader_controller.js":              "/assets/speed-eee.js",
		"javascript/controllers/scrollable_table_patterns_controller.js": "/assets/scroll-fff.js",
	}})
}

func TestTagsContainImportMapAndEntrypoint(t *testing.T) {
	html := string(newTestMap().Tags())

	if !strings.Contains(html, `<script type="importmap">`) {
		t.Error("missing importmap script tag")
	}
	if !strings.Contains(html, `<script type="module">import "application"</script>`) {
		t.Error("missing module entrypoint")
	}
	// Bare specifiers used across the JS modules must all be pinned.
	for _, spec := range []string{
		`"@hotwired/stimulus"`,
		`"controllers"`,
		`"controllers/speed_reader_controller"`,
	} {
		if !strings.Contains(html, spec) {
			t.Errorf("import map missing pin %s", spec)
		}
	}
}

func TestDigestChangesWhenAssetURLChanges(t *testing.T) {
	before := newTestMap().Digest()

	changed := New(stubResolver{paths: map[string]string{
		"javascript/application.js": "/assets/application-ZZZ.js", // different digest
	}})
	after := changed.Digest()

	if before == after {
		t.Fatal("digest should change when a pinned module URL changes")
	}
}
