//go:build gcp

package main

import (
	"context"
	"log/slog"

	"github.com/fmarquesfilho/garimpo/internal/store"
)

// initSnapshots creates a BigQueryStore for snapshot persistence from YAML config.
// Returns NopSnapshots if config is incomplete or connection fails (graceful degradation).
func initSnapshots(ctx context.Context, cfg BigQueryExporterConfig, logger *slog.Logger) store.SnapshotRepo {
	project := ResolveCredentialEnv(cfg.ProjectEnv)
	if project == "" || cfg.Dataset == "" || cfg.ProductsTable == "" {
		logger.Warn("BigQuery export disabled (missing exporters.bigquery config)")
		return store.NopSnapshots()
	}
	bq, err := store.NovoBigQueryStore(ctx, project, cfg.Dataset, "eventos", cfg.ProductsTable)
	if err != nil {
		logger.Warn("BigQuery init failed, snapshot export disabled",
			slog.String("error", err.Error()))
		return store.NopSnapshots()
	}
	logger.Info("BigQuery snapshot export enabled",
		slog.String("project", project),
		slog.String("dataset", cfg.Dataset),
		slog.String("table", cfg.ProductsTable))
	return bq
}
