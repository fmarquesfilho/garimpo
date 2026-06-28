//go:build gcp

package store

import (
	"context"
	"fmt"
	"os"
)

// Novo (gcp) cria o BigQueryStore a partir do ambiente:
//
//	GOOGLE_CLOUD_PROJECT  projeto GCP
//	BQ_DATASET            dataset (ex.: garimpo)
//	BQ_TABELA             tabela de eventos (ex.: eventos)
func Novo(ctx context.Context) (EventoStore, error) {
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
	return NovoBigQueryStore(ctx, projeto, dataset, tabela, tabelaSnap)
}

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
