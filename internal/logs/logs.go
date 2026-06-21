// Package logs centraliza o logging estruturado por criticidade (slog, stdlib —
// sem dependência). Níveis: DEBUG < INFO < WARN < ERROR.
//
//   - Em produção (Cloud Run), emite JSON — o Cloud Logging indexa os campos e
//     deixa filtrar por severity, rota, categoria etc.
//   - Em dev, emite texto legível.
//
// Configuração por ambiente:
//
//	LOG_LEVEL=debug|info|warn|error   (padrão: info)
//	LOG_FORMAT=json|text              (padrão: json em Cloud Run, text fora)
package logs

import (
	"log/slog"
	"os"
	"strings"
)

// Init configura o logger global a partir do ambiente e o devolve.
func Init() *slog.Logger {
	lvl := nivel(os.Getenv("LOG_LEVEL"))
	opts := &slog.HandlerOptions{Level: lvl}

	var h slog.Handler
	if formatoJSON() {
		h = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		h = slog.NewTextHandler(os.Stdout, opts)
	}
	l := slog.New(h)
	slog.SetDefault(l)
	return l
}

// nivel traduz a string de ambiente para slog.Level (padrão INFO).
func nivel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// formatoJSON decide o formato. JSON quando LOG_FORMAT=json OU quando rodando no
// Cloud Run (variável K_SERVICE presente). Texto caso contrário.
func formatoJSON() bool {
	switch strings.ToLower(os.Getenv("LOG_FORMAT")) {
	case "json":
		return true
	case "text":
		return false
	}
	return os.Getenv("K_SERVICE") != "" // heurística Cloud Run
}
