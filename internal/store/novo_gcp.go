//go:build gcp

package store

import (
	"context"
	"fmt"
	"os"
)

// NovoRepository (gcp) cria o Repository completo com BigQuery.
// O TenantRepo é injetado externamente (Firestore, memória, etc.).
func NovoRepository(ctx context.Context, tenants TenantRepo) (Repository, error) {
	projeto := os.Getenv("GOOGLE_CLOUD_PROJECT")
	dataset := os.Getenv("BQ_DATASET")
	tabela := os.Getenv("BQ_TABELA")
	if projeto == "" || dataset == "" || tabela == "" {
		return nil, fmt.Errorf("defina GOOGLE_CLOUD_PROJECT, BQ_DATASET e BQ_TABELA")
	}
	tabelaSnap := os.Getenv("BQ_TABELA_SNAP")
	if tabelaSnap == "" {
		tabelaSnap = "snapshots"
	}
	bq, err := NovoBigQueryStore(ctx, projeto, dataset, tabela, tabelaSnap)
	if err != nil {
		return nil, err
	}
	return NovoBQRepository(bq, tenants), nil
}
