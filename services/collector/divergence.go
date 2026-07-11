package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sort"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"

	cachepb "github.com/fmarquesfilho/garimpo/gen/go/cache/v1"
	collectorpb "github.com/fmarquesfilho/garimpo/gen/go/collector/v1"
)

// ErrPurgeAPIFailed is returned when the Cloudflare Purge API returns an error status.
var ErrPurgeAPIFailed = errors.New("purge API failed")

// DivergenceDetector detects data changes and invalidates cache layers.
type DivergenceDetector struct {
	cacheClient  cachepb.CacheServiceClient
	httpClient   *http.Client
	cfZoneID     string
	cfPurgeToken string
	logger       *slog.Logger
}

// NewDivergenceDetector creates a DivergenceDetector.
// Returns nil if CACHE_GRPC_ADDRESS is not configured (cache disabled).
func NewDivergenceDetector(logger *slog.Logger) *DivergenceDetector {
	cacheAddr := os.Getenv("CACHE_GRPC_ADDRESS")
	if cacheAddr == "" {
		logger.Info("CACHE_GRPC_ADDRESS not set, divergence detection disabled")
		return nil
	}

	conn, err := grpc.NewClient(cacheAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Warn("failed to connect to cache sidecar",
			slog.String("address", cacheAddr),
			slog.String("error", err.Error()))
		return nil
	}

	return &DivergenceDetector{
		cacheClient:  cachepb.NewCacheServiceClient(conn),
		httpClient:   &http.Client{Timeout: 5 * time.Second},
		cfZoneID:     os.Getenv("CF_ZONE_ID"),
		cfPurgeToken: os.Getenv("CF_PURGE_TOKEN"),
		logger:       logger,
	}
}

// DetectAndInvalidate compares collected products with cached data and
// invalidates both L2 (gRPC) and L1 (Cloudflare Purge API) if divergence
// is detected. Non-blocking: errors are logged but never propagated.
func (d *DivergenceDetector) DetectAndInvalidate(ctx context.Context, buscaID, ownerUID string, collectionKeys []string, products []*collectorpb.Product) {
	if d == nil {
		return
	}

	newHash := hashProductsForDivergence(products)

	// Check current cache state
	cachedHash := d.getCachedHash(ctx, ownerUID, collectionKeys)
	if cachedHash == "" {
		// Nothing in cache (cold start) — no divergence
		return
	}

	if newHash == cachedHash {
		// Data is the same — no action needed
		return
	}

	d.logger.Info("divergence detected",
		slog.String("busca_id", buscaID),
		slog.String("old_hash", cachedHash[:8]),
		slog.String("new_hash", newHash[:8]))

	// Invalidate L2 (local gRPC — fast)
	d.invalidateL2(ctx, buscaID, ownerUID)

	// Purge L1 (Cloudflare API — with 1 retry)
	d.purgeL1(buscaID)
}

// getCachedHash retrieves the hash of the first cached entry for the given keys.
func (d *DivergenceDetector) getCachedHash(ctx context.Context, ownerUID string, collectionKeys []string) string {
	resp, err := d.cacheClient.Get(ctx, &cachepb.GetRequest{
		CollectionKeys: collectionKeys[:1], // Check first key only for hash
		BuscaId:        "divergence-check",
		OwnerUid:       ownerUID,
		Marketplace:    collectorpb.Marketplace_MARKETPLACE_SHOPEE,
	})
	if err != nil {
		// Cache not available — skip
		return ""
	}
	if !resp.GetCacheHit() {
		return ""
	}
	// Compute hash of what we got back to compare
	return hashProductsForDivergence(resp.GetProducts())
}

// invalidateL2 calls CacheService.Invalidate via local gRPC.
func (d *DivergenceDetector) invalidateL2(ctx context.Context, buscaID, ownerUID string) {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	resp, err := d.cacheClient.Invalidate(ctx, &cachepb.InvalidateRequest{
		BuscaId:  buscaID,
		OwnerUid: ownerUID,
	})
	if err != nil {
		d.logger.Warn("L2 invalidation failed",
			slog.String("busca_id", buscaID),
			slog.String("error", err.Error()))
		return
	}
	d.logger.Info("L2 invalidated",
		slog.String("busca_id", buscaID),
		slog.Int("keys_removed", int(resp.GetKeysRemoved())))
}

// purgeL1 calls Cloudflare Purge API to invalidate L1 cache by tag.
// Retries once on failure.
func (d *DivergenceDetector) purgeL1(buscaID string) {
	if d.cfZoneID == "" || d.cfPurgeToken == "" {
		d.logger.Debug("L1 purge skipped (no CF_ZONE_ID or CF_PURGE_TOKEN)")
		return
	}

	tag := "busca:" + buscaID

	for attempt := 0; attempt < 2; attempt++ {
		err := d.doPurge(tag)
		if err == nil {
			d.logger.Info("L1 purged",
				slog.String("tag", tag),
				slog.Int("attempt", attempt+1))
			return
		}

		d.logger.Warn("L1 purge attempt failed",
			slog.String("tag", tag),
			slog.Int("attempt", attempt+1),
			slog.String("error", err.Error()))

		if attempt == 0 {
			time.Sleep(1 * time.Second) // Wait before retry
		}
	}

	d.logger.Warn("L1 purge failed after 2 attempts, skipping (TTL will expire)",
		slog.String("busca_id", buscaID))
}

// doPurge makes the actual HTTP call to Cloudflare Purge API.
func (d *DivergenceDetector) doPurge(tag string) error {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/purge_cache", d.cfZoneID)

	body, _ := json.Marshal(map[string]interface{}{
		"tags": []string{tag},
	})

	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+d.cfPurgeToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("%w: status %d", ErrPurgeAPIFailed, resp.StatusCode)
	}

	return nil
}

// hashProductsForDivergence computes SHA-256 of canonical product serialization.
func hashProductsForDivergence(products []*collectorpb.Product) string {
	sorted := make([]*collectorpb.Product, len(products))
	copy(sorted, products)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].GetItemId() < sorted[j].GetItemId()
	})

	h := sha256.New()
	for _, p := range sorted {
		data, _ := proto.Marshal(p)
		h.Write(data)
	}
	return hex.EncodeToString(h.Sum(nil))
}
