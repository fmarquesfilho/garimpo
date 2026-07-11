//go:build !gcp

package main

import (
	"context"
	"log/slog"

	"github.com/fmarquesfilho/garimpo/internal/store"
)

// initSnapshots (non-GCP build) returns a no-op snapshot repo.
// Used in local development and tests without BigQuery.
func initSnapshots(_ context.Context, _ BigQueryExporterConfig, logger *slog.Logger) store.SnapshotRepo {
	logger.Info("snapshot export disabled (build without -tags gcp)")
	return store.NopSnapshots()
}
