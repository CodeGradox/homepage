package web

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"homepage/internal/logging"
)

// statusRecorder captures the response status code so the logger can report it
// and choose a log level from it.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// requestLogger emits one lograge-style line per request: method, path, status
// and duration in milliseconds. The /up health check is silenced, matching the
// Rails config.silence_healthcheck_path setting.
func requestLogger(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/up" {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		elapsed := float64(time.Since(start).Microseconds()) / 1000.0
		level := logging.LevelForStatus(rec.status)
		logger.LogAttrs(r.Context(), level, "",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", rec.status),
			slog.String("duration", strconv.FormatFloat(elapsed, 'f', 2, 64)),
		)
	})
}

// recoverer turns a panic in a handler into a 500 and an ERROR log line instead
// of crashing the server.
func recoverer(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				logger.LogAttrs(context.Background(), slog.LevelError, "",
					slog.String("panic", toString(v)),
					slog.String("path", r.URL.Path),
				)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func toString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	if err, ok := v.(error); ok {
		return err.Error()
	}
	return "panic"
}
