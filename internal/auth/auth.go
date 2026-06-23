// Package auth valida tokens de autenticação Firebase (ID tokens JWT).
// Atrás de build tag gcp; sem ela, aceita tudo (dev local).
package auth

import "context"

// User é o usuário autenticado extraído do token.
type User struct {
	UID   string
	Email string
	Admin bool // true se o email está na lista ADMIN_EMAILS
}

// Verifier valida tokens e extrai o usuário.
type Verifier interface {
	// Verify valida o token Bearer e retorna o usuário.
	// Se o token for vazio ou inválido, retorna nil (anônimo).
	Verify(ctx context.Context, idToken string) *User
}

// NopVerifier aceita tudo — usado em dev/local.
type NopVerifier struct{}

func (NopVerifier) Verify(_ context.Context, _ string) *User {
	// Em dev, todos os requests são "anônimos" — sem restrição.
	return nil
}
