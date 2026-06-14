package logging

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
)

func TestLineFormatMatchesLograge(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, slog.LevelInfo)

	logger.LogAttrs(context.Background(), slog.LevelInfo, "",
		slog.String("method", "GET"),
		slog.String("path", "/"),
		slog.Int("status", 200),
	)

	line := buf.String()
	if !strings.Contains(line, "[INFO]") {
		t.Errorf("missing level label: %q", line)
	}
	if !strings.Contains(line, "method=GET path=/ status=200") {
		t.Errorf("unexpected key=value rendering: %q", line)
	}
	// Timestamp prefix: "2006-01-02 15:04:05.000".
	if line[4] != '-' || line[7] != '-' || line[10] != ' ' {
		t.Errorf("unexpected timestamp format: %q", line)
	}
}

func TestLevelForStatus(t *testing.T) {
	cases := map[int]slog.Level{
		200: slog.LevelInfo,
		301: slog.LevelInfo,
		404: slog.LevelWarn,
		422: slog.LevelWarn,
		500: slog.LevelError,
	}
	for status, want := range cases {
		if got := LevelForStatus(status); got != want {
			t.Errorf("LevelForStatus(%d) = %v, want %v", status, got, want)
		}
	}
}

func TestValuesWithSpacesAreQuoted(t *testing.T) {
	var buf bytes.Buffer
	New(&buf, slog.LevelInfo).LogAttrs(context.Background(), slog.LevelError, "",
		slog.String("panic", "something broke"),
	)
	if !strings.Contains(buf.String(), `panic="something broke"`) {
		t.Errorf("expected quoted value, got %q", buf.String())
	}
}

func TestLevelBelowThresholdIsDropped(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, slog.LevelWarn)
	logger.LogAttrs(context.Background(), slog.LevelInfo, "", slog.String("k", "v"))
	if buf.Len() != 0 {
		t.Errorf("expected info to be dropped at warn threshold, got %q", buf.String())
	}
}
