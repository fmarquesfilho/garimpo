//go:build !gcp

package store

import "context"

// Novo (padrão) não persiste nada. Compile com -tags gcp para usar o BigQuery.
func Novo(ctx context.Context) (EventoStore, error) {
	_ = ctx
	return NopStore{}, nil
}
