// Package logging reproduces the Rails/lograge log line in Go on top of the
// standard library's log/slog. Each line reads:
//
//	2026-06-14 12:00:00.123 [INFO] method=GET path=/ status=200 duration=1.23
//
// The level is derived from the HTTP status the same way the Rails
// CustomLogFormatter did it: 2xx/3xx INFO, 4xx WARN, 5xx ERROR.
package logging

import (
	"context"
	"io"
	"log/slog"
	"slices"
	"strconv"
	"strings"
)

// New returns a slog.Logger that writes lograge-style lines to w at or above
// the given level.
func New(w io.Writer, level slog.Level) *slog.Logger {
	return slog.New(&handler{w: w, level: level})
}

// LevelForStatus maps an HTTP status to a log level, matching the Rails
// formatter: success/redirect INFO, client error WARN, server error ERROR.
func LevelForStatus(status int) slog.Level {
	switch {
	case status >= 500:
		return slog.LevelError
	case status >= 400:
		return slog.LevelWarn
	default:
		return slog.LevelInfo
	}
}

// handler renders records as `TIMESTAMP [LEVEL] key=value ...` lines. It is
// intentionally minimal: no groups, no JSON, just the flat key=value shape
// lograge produced.
type handler struct {
	w     io.Writer
	level slog.Level
	attrs []slog.Attr
}

func (h *handler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *handler) Handle(_ context.Context, r slog.Record) error {
	var b strings.Builder
	b.WriteString(r.Time.Format("2006-01-02 15:04:05.000"))
	b.WriteString(" [")
	b.WriteString(levelLabel(r.Level))
	b.WriteString("] ")

	if r.Message != "" {
		writePair(&b, "msg", r.Message)
		b.WriteByte(' ')
	}
	for _, a := range h.attrs {
		writeAttr(&b, a)
	}
	r.Attrs(func(a slog.Attr) bool {
		writeAttr(&b, a)
		return true
	})

	out := strings.TrimRight(b.String(), " ") + "\n"
	_, err := io.WriteString(h.w, out)
	return err
}

func (h *handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	next := *h
	next.attrs = slices.Concat(h.attrs, attrs)
	return &next
}

// WithGroup is unused by this app; grouping has no place in a flat key=value
// line, so we ignore it and return the handler unchanged.
func (h *handler) WithGroup(string) slog.Handler { return h }

func writeAttr(b *strings.Builder, a slog.Attr) {
	writePair(b, a.Key, a.Value.String())
	b.WriteByte(' ')
}

func writePair(b *strings.Builder, key, value string) {
	b.WriteString(key)
	b.WriteByte('=')
	if value == "" || strings.ContainsAny(value, " \t\"") {
		b.WriteString(strconv.Quote(value))
	} else {
		b.WriteString(value)
	}
}

func levelLabel(level slog.Level) string {
	switch {
	case level >= slog.LevelError:
		return "ERROR"
	case level >= slog.LevelWarn:
		return "WARN"
	case level >= slog.LevelInfo:
		return "INFO"
	default:
		return "DEBUG"
	}
}
