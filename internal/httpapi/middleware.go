package httpapi

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/fmarquesfilho/garimpo/internal/logs"
)

// ── Middleware ─────────────────────────────────────────────────────────────

type respCapturado struct {
	http.ResponseWriter
	status int
}

func (r *respCapturado) WriteHeader(c int) {
	r.status = c
	r.ResponseWriter.WriteHeader(c)
}

func (srv *Server) logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inicio := time.Now()
		rc := &respCapturado{ResponseWriter: w, status: 200}
		next.ServeHTTP(rc, r)

		dur := time.Since(inicio)
		attrs := []any{
			slog.String("metodo", r.Method),
			slog.String("rota", r.URL.Path),
			slog.Int("status", rc.status),
			slog.Duration("dur", dur),
		}

		nivel := "info"
		switch {
		case rc.status >= 500:
			srv.Logger.Error("requisição", attrs...)
			nivel = "error"
		case r.URL.Path == "/api/health":
			srv.Logger.Debug("requisição", attrs...)
			nivel = "debug"
		default:
			srv.Logger.Info("requisição", attrs...)
		}

		if srv.LogBuffer != nil {
			srv.LogBuffer.Push(logs.Entry{
				Nivel: nivel, Msg: "requisição", Metodo: r.Method,
				Rota: r.URL.Path, Status: rc.status,
				DurMs: float64(dur.Microseconds()) / 1000.0, Em: inicio,
			})
		}
	})
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
