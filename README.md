# homepage

Magnus Åsrud's personal site. A small server-rendered Go application — a port of
the original Rails app. Templates, stylesheets, JavaScript and images are all
embedded into a single static binary.

## Running locally

```sh
bin/dev          # serves on http://localhost:3000
```

Or directly:

```sh
PORT=8080 go run .
```

## Layout

```
main.go                  Wiring: embeds, logger, server, ListenAndServe
templates/               html/template views (layout + one file per page)
assets/                  Source assets (css, js, images), embedded and fingerprinted
public/                  Unfingerprinted files served as-is (robots.txt, icons)
internal/
  assets/                Mini-Propshaft: content-addressed asset pipeline
  importmap/             Renders the <script type="importmap"> tags
  logging/               slog handler producing the lograge-style log line
  web/                   Renderer, handlers, middleware, server
```

### How the Rails pieces map to Go

| Rails                                | Go                                                      |
| ------------------------------------ | ------------------------------------------------------ |
| Slim views + `content_for`           | `html/template` with `{{block}}` overrides             |
| Propshaft fingerprinting             | `internal/assets` (digest spliced into the filename)   |
| importmap-rails                      | `internal/importmap` (renders the JSON map; no bundler)|
| lograge structured logs              | `internal/logging` (custom `slog.Handler`)             |
| `stale_when_importmap_changes`       | ETag over the rendered body (the map is inlined)       |

Import maps are a browser-native feature: there is no build step. The map is
JSON injected into the page head, and the browser loads the ES modules — and the
pinned Stimulus controllers — directly. Each pin resolves to a fingerprinted
asset URL.

## Tests and static analysis

```sh
go test ./...    # unit + end-to-end handler tests
bin/ci           # gofmt, vet, staticcheck, gosec, govulncheck, build, test
```

The static-analysis suite is the Go counterpart of the old Rails tooling:
`go vet` + `staticcheck` (~ rubocop), `gosec` (~ brakeman, security scanning),
and `govulncheck` (~ bundler-audit, known CVEs). `gopls` is the LSP server for
editor integration.

## Deployment

The Dockerfile builds a static binary onto a distroless non-root base, listening
on `:8080` (override with `PORT`). It exposes the same port as the previous
image, so a Kamal/registry deploy can point at it unchanged.

```sh
docker build -t homepage .
docker run -p 8080:8080 homepage
```
