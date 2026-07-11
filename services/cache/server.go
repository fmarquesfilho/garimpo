package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"sort"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	cachepb "github.com/fmarquesfilho/garimpo/gen/go/cache/v1"
	collectorpb "github.com/fmarquesfilho/garimpo/gen/go/collector/v1"
)

const schemaVersion = "1.0.0"

// CacheServer implements cache.v1.CacheService.
type CacheServer struct {
	cachepb.UnimplementedCacheServiceServer

	cache        *LRUCache
	reverseIndex map[string][]string // busca_id → []cache_key (with owner prefix)
	riMu         sync.RWMutex

	collector collectorpb.CollectorServiceClient
	sfGroup   singleflight.Group
	logger    *slog.Logger
	ttl       time.Duration
}

// NewCacheServer creates a new CacheServer.
func NewCacheServer(cache *LRUCache, collector collectorpb.CollectorServiceClient, ttl time.Duration, logger *slog.Logger) *CacheServer {
	return &CacheServer{
		cache:        cache,
		reverseIndex: make(map[string][]string),
		collector:    collector,
		logger:       logger,
		ttl:          ttl,
	}
}

// cacheKey builds a tenant-isolated cache key: "{owner_uid}:{collection_key}".
func cacheKey(ownerUID, collectionKey string) string {
	return ownerUID + ":" + collectionKey
}

// reverseKey builds a tenant-isolated reverse index key: "{owner_uid}:{busca_id}".
func reverseKey(ownerUID, buscaID string) string {
	return ownerUID + ":" + buscaID
}

// Get busca produtos por collection_keys.
func (s *CacheServer) Get(ctx context.Context, req *cachepb.GetRequest) (*cachepb.GetResponse, error) {
	if len(req.GetCollectionKeys()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "collection_keys é obrigatório")
	}
	if req.GetBuscaId() == "" {
		return nil, status.Error(codes.InvalidArgument, "busca_id é obrigatório")
	}
	if req.GetOwnerUid() == "" {
		return nil, status.Error(codes.InvalidArgument, "owner_uid é obrigatório")
	}

	ownerUID := req.GetOwnerUid()
	buscaID := req.GetBuscaId()

	// Register reverse index for future invalidation
	s.registerReverseIndex(ownerUID, buscaID, req.GetCollectionKeys())

	// Try cache hit for ALL keys
	allHit := true
	var products []*collectorpb.Product
	var fetchedAt time.Time

	for _, key := range req.GetCollectionKeys() {
		ck := cacheKey(ownerUID, key)
		entry := s.cache.Get(ck)
		if entry == nil {
			allHit = false
			break
		}
		products = append(products, entry.Products...)
		fetchedAt = entry.FetchedAt
	}

	if allHit {
		return &cachepb.GetResponse{
			Products:      products,
			CacheHit:      true,
			FetchedAt:     fetchedAt.Format(time.RFC3339),
			SchemaVersion: schemaVersion,
		}, nil
	}

	// Cache miss — fetch from Collector via singleflight
	sfKey := fmt.Sprintf("%s:%s", ownerUID, hashKeys(req.GetCollectionKeys()))
	result, err, _ := s.sfGroup.Do(sfKey, func() (interface{}, error) {
		return s.fetchFromCollector(ctx, req)
	})
	if err != nil {
		return nil, fmt.Errorf("fetch via singleflight: %w", err)
	}

	fetchedProducts := result.([]*collectorpb.Product)

	// Store in cache
	now := time.Now()
	hash := hashProducts(fetchedProducts)
	for _, key := range req.GetCollectionKeys() {
		ck := cacheKey(ownerUID, key)
		// Filter products per key (for simplicity, store all in each key)
		entry := &CacheEntry{
			CollectionKey: ck,
			Products:      fetchedProducts,
			FetchedAt:     now,
			ExpiresAt:     now.Add(s.ttl),
			Hash:          hash,
			SizeBytes:     estimateSize(fetchedProducts),
		}
		s.cache.Put(ck, entry)
	}

	return &cachepb.GetResponse{
		Products:      fetchedProducts,
		CacheHit:      false,
		FetchedAt:     now.Format(time.RFC3339),
		SchemaVersion: schemaVersion,
	}, nil
}

// Invalidate remove entries do cache por busca_id e/ou collection_keys.
func (s *CacheServer) Invalidate(_ context.Context, req *cachepb.InvalidateRequest) (*cachepb.InvalidateResponse, error) {
	if req.GetBuscaId() == "" && len(req.GetCollectionKeys()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "busca_id ou collection_keys é obrigatório")
	}

	ownerUID := req.GetOwnerUid()
	keysToRemove := make(map[string]struct{})

	// Gather keys from reverse index
	if req.GetBuscaId() != "" {
		rk := reverseKey(ownerUID, req.GetBuscaId())
		s.riMu.Lock()
		if keys, ok := s.reverseIndex[rk]; ok {
			for _, k := range keys {
				keysToRemove[k] = struct{}{}
			}
			delete(s.reverseIndex, rk)
		}
		s.riMu.Unlock()
	}

	// Add explicit collection_keys
	for _, key := range req.GetCollectionKeys() {
		keysToRemove[cacheKey(ownerUID, key)] = struct{}{}
	}

	// Remove from cache
	var removed int32
	for k := range keysToRemove {
		if s.cache.Delete(k) {
			removed++
		}
	}

	s.logger.Info("cache invalidated",
		slog.String("busca_id", req.GetBuscaId()),
		slog.Int("keys_removed", int(removed)))

	return &cachepb.InvalidateResponse{
		KeysRemoved: removed,
		Success:     true,
	}, nil
}

// Healthz retorna status e métricas do cache.
func (s *CacheServer) Healthz(_ context.Context, _ *cachepb.HealthzRequest) (*cachepb.HealthzResponse, error) {
	sizeBytes, hits, misses, entries := s.cache.Stats()
	return &cachepb.HealthzResponse{
		Ready:          true,
		CacheSizeBytes: sizeBytes,
		HitsTotal:      hits,
		MissesTotal:    misses,
		EntriesCount:   entries,
	}, nil
}

// registerReverseIndex registers the mapping busca_id → cache_keys for invalidation.
func (s *CacheServer) registerReverseIndex(ownerUID, buscaID string, collectionKeys []string) {
	rk := reverseKey(ownerUID, buscaID)
	cacheKeys := make([]string, len(collectionKeys))
	for i, k := range collectionKeys {
		cacheKeys[i] = cacheKey(ownerUID, k)
	}

	s.riMu.Lock()
	defer s.riMu.Unlock()

	existing := s.reverseIndex[rk]
	merged := mergeUnique(existing, cacheKeys)
	s.reverseIndex[rk] = merged
}

// fetchFromCollector calls Collector.Fetch or FetchShop based on marketplace/keys.
func (s *CacheServer) fetchFromCollector(ctx context.Context, req *cachepb.GetRequest) ([]*collectorpb.Product, error) {
	var allProducts []*collectorpb.Product

	for _, key := range req.GetCollectionKeys() {
		// Try to parse as shop_id (numeric = shop, otherwise keyword)
		var resp *collectorpb.FetchResponse
		var shopResp *collectorpb.FetchShopResponse
		var err error

		if isNumeric(key) {
			shopID := parseInt64(key)
			if shopID > 0 {
				shopResp, err = s.collector.FetchShop(ctx, &collectorpb.FetchShopRequest{
					ShopId:      shopID,
					Limit:       50,
					OwnerUid:    req.GetOwnerUid(),
					Marketplace: req.GetMarketplace(),
				})
				if err != nil {
					s.logger.Error("FetchShop failed",
						slog.String("key", key),
						slog.String("error", err.Error()))
					return nil, status.Errorf(codes.Internal, "FetchShop falhou: %v", err)
				}
				allProducts = append(allProducts, shopResp.GetProducts()...)
				continue
			}
		}

		resp, err = s.collector.Fetch(ctx, &collectorpb.FetchRequest{
			Keyword:     key,
			Limit:       50,
			OwnerUid:    req.GetOwnerUid(),
			Marketplace: req.GetMarketplace(),
		})
		if err != nil {
			s.logger.Error("Fetch failed",
				slog.String("key", key),
				slog.String("error", err.Error()))
			return nil, status.Errorf(codes.Internal, "Fetch falhou: %v", err)
		}
		allProducts = append(allProducts, resp.GetProducts()...)
	}

	return allProducts, nil
}

// hashProducts computes SHA-256 of the canonical serialization of products.
func hashProducts(products []*collectorpb.Product) string {
	// Sort by item_id for deterministic hash
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

// hashKeys creates a short hash of the sorted collection keys for singleflight.
func hashKeys(keys []string) string {
	sorted := make([]string, len(keys))
	copy(sorted, keys)
	sort.Strings(sorted)

	h := sha256.New()
	for _, k := range sorted {
		h.Write([]byte(k))
	}
	return hex.EncodeToString(h.Sum(nil))[:16]
}

// estimateSize estimates the memory size of a product list (rough).
func estimateSize(products []*collectorpb.Product) int64 {
	if len(products) == 0 {
		return 64 // base overhead
	}
	// ~100 bytes per product is a conservative estimate
	return int64(len(products)) * 100
}

// mergeUnique merges two slices, removing duplicates.
func mergeUnique(a, b []string) []string {
	seen := make(map[string]struct{}, len(a)+len(b))
	for _, v := range a {
		seen[v] = struct{}{}
	}
	for _, v := range b {
		seen[v] = struct{}{}
	}
	result := make([]string, 0, len(seen))
	for v := range seen {
		result = append(result, v)
	}
	return result
}

// isNumeric checks if a string is a valid positive integer.
func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// parseInt64 parses a string to int64, returns 0 on failure.
func parseInt64(s string) int64 {
	var n int64
	for _, r := range s {
		n = n*10 + int64(r-'0')
	}
	return n
}
