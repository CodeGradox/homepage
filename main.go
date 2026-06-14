// Command homepage serves Magnus Åsrud's personal site. Templates, stylesheets,
// JavaScript and images are all embedded into the binary, so the whole site
// ships as a single static executable with no runtime file dependencies.
package main

import (
	"context"
	"embed"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"homepage/internal/assets"
	"homepage/internal/logging"
	"homepage/internal/web"
)

//go:embed templates
var templatesDir embed.FS

//go:embed all:assets
var assetsDir embed.FS

//go:embed all:public
var publicDir embed.FS

func main() {
	logger := logging.New(os.Stdout, logLevel())

	templateFS, err := fs.Sub(templatesDir, "templates")
	if err != nil {
		log.Fatal(err)
	}
	assetFS, err := fs.Sub(assetsDir, "assets")
	if err != nil {
		log.Fatal(err)
	}
	publicFS, err := fs.Sub(publicDir, "public")
	if err != nil {
		log.Fatal(err)
	}

	pipeline, err := assets.Load(assetFS)
	if err != nil {
		log.Fatal(err)
	}

	server, err := web.New(logger, templateFS, pipeline)
	if err != nil {
		log.Fatal(err)
	}

	addr := ":" + port()
	httpServer := &http.Server{
		Addr:              addr,
		Handler:           server.Handler(publicFS),
		ReadHeaderTimeout: 5 * time.Second,
	}

	logger.LogAttrs(context.Background(), slog.LevelInfo, "", slog.String("event", "boot"), slog.String("addr", addr))
	if err := httpServer.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func port() string {
	if p := os.Getenv("PORT"); p != "" {
		return p
	}
	return "8080"
}

// logLevel mirrors the Rails RAILS_LOG_LEVEL knob, defaulting to info.
func logLevel() slog.Level {
	switch os.Getenv("LOG_LEVEL") {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
