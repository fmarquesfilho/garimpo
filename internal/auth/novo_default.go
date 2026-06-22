//go:build !gcp

package auth

import "context"

// Novo (padrão) retorna NopVerifier — sem validação em dev.
func Novo(ctx context.Context) (Verifier, error) {
	_ = ctx
	return NopVerifier{}, nil
}
