//go:build gcp

package auth

import "context"

// Novo (gcp) cria o FirebaseVerifier.
func Novo(ctx context.Context) (Verifier, error) {
	v, err := NovoFirebaseVerifier(ctx)
	if err != nil {
		return nil, err
	}
	return v, nil
}
