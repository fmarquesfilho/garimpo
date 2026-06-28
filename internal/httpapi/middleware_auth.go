package httpapi

import (
	"context"
	"net/http"
	"os"

	"github.com/fmarquesfilho/garimpo/internal/auth"
)

// ctxKey é o tipo para chaves de contexto (evita colisão).
type ctxKey string

const ctxUsuario ctxKey = "usuario"

// usuarioDoCtx extrai o usuário autenticado do contexto (setado pelo middleware).
func usuarioDoCtx(r *http.Request) *auth.User {
	u, _ := r.Context().Value(ctxUsuario).(*auth.User)
	return u
}

// requireAuth é um middleware que exige autenticação Firebase.
// Retorna 401 se não autenticado. Seta o usuário no contexto.
func (srv *Server) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := srv.usuarioDoRequest(r)
		if user == nil {
			writeErr(w, http.StatusUnauthorized, "autenticação necessária")
			return
		}
		ctx := context.WithValue(r.Context(), ctxUsuario, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// requireAdmin é um middleware que exige autenticação + role admin.
// Retorna 403 se não admin.
func (srv *Server) requireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := usuarioDoCtx(r)
		if user == nil || !user.Admin {
			writeErr(w, http.StatusForbidden, "acesso restrito a administradores")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// requireColetaToken é um middleware que exige o X-Garimpo-Token.
// Retorna 401 se token inválido.
func (srv *Server) requireColetaToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tok := os.Getenv("COLETA_TOKEN")
		if tok != "" && r.Header.Get("X-Garimpo-Token") != tok {
			writeErr(w, http.StatusUnauthorized, "token de coleta inválido")
			return
		}
		next.ServeHTTP(w, r)
	})
}
