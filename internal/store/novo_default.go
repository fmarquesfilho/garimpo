//go:build !gcp

package store

import "context"

// NovoRepository (padrão) cria o NopRepository em memória.
func NovoRepository(_ context.Context, _ TenantRepo) (Repository, error) {
	return NovoNopRepository(), nil
}
