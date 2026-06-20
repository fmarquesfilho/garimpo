//go:build gcp

package store

import (
	"context"
	"fmt"
	"os"
)

// Novo (gcp) cria o BigQueryStore a partir do ambiente:
//   GOOGLE_CLOUD_PROJECT  projeto GCP
//   BQ_DATASET            dataset (ex.: garimpo)
//   BQ_TABELA             tabela de eventos (ex.: eventos)
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
