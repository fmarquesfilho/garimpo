//go:build !gcp

package store

import "context"

// Novo (padrão) não persiste nada. Compile com -tags gcp para usar o BigQuery.
func Novo(ctx context.Context) (EventoStore, error) {
	_ = ctx
	return NopStore{}, nil
}

// NovoRepository (padrão) cria o NopRepository em memória.
func NovoRepository(_ context.Context, _ TenantRepo) (Repository, error) {
	return NovoNopRepository(), nil
}
