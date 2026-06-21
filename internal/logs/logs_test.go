package logs

import (
	"log/slog"
	"testing"
)

func TestNivel(t *testing.T) {
	casos := map[string]slog.Level{
		"debug":   slog.LevelDebug,
		"DEBUG":   slog.LevelDebug,
		"info":    slog.LevelInfo,
		"":        slog.LevelInfo,
		"warn":    slog.LevelWarn,
		"warning": slog.LevelWarn,
		"error":   slog.LevelError,
		"xpto":    slog.LevelInfo, // desconhecido -> info
	}
	for in, quer := range casos {
		if got := nivel(in); got != quer {
			t.Errorf("nivel(%q)=%v, quer %v", in, got, quer)
		}
	}
}

func TestInitNaoPanica(t *testing.T) {
	if Init() == nil {
		t.Error("Init devolveu nil")
	}
}
