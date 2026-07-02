package couponsource

import (
	"fmt"
	"sync"

	"github.com/fmarquesfilho/garimpo/internal/domain"
)

// ErrUnsupportedMarketplace indicates no adapter is registered for the marketplace.
var ErrUnsupportedMarketplace = fmt.Errorf("marketplace não suportado para cupons")

// Registry maps marketplace → CouponSourceFactory. Thread-safe for concurrent reads.
type Registry struct {
	mu        sync.RWMutex
	factories map[string]CouponSourceFactory
}

// DefaultRegistry is the global registry. Adapters register via Register().
var DefaultRegistry = NewRegistry()

func NewRegistry() *Registry {
	return &Registry{factories: make(map[string]CouponSourceFactory)}
}

// Register associates a marketplace with a factory.
func (r *Registry) Register(marketplace string, factory CouponSourceFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.factories[marketplace] = factory
}

// Get returns the factory for the given marketplace.
func (r *Registry) Get(marketplace string) (CouponSourceFactory, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	f, ok := r.factories[marketplace]
	if !ok {
		return nil, fmt.Errorf("marketplace %q: %w", marketplace, ErrUnsupportedMarketplace)
	}
	return f, nil
}

// Create is a shortcut: get factory and create source.
func (r *Registry) Create(marketplace string, cfg SourceConfig) (CouponSource, error) {
	factory, err := r.Get(marketplace)
	if err != nil {
		return nil, err
	}
	return factory(cfg), nil
}

// Supported returns registered marketplace identifiers.
func (r *Registry) Supported() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	keys := make([]string, 0, len(r.factories))
	for k := range r.factories {
		keys = append(keys, k)
	}
	return keys
}

func init() {
	DefaultRegistry.Register(domain.MarketplaceShopee, func(cfg SourceConfig) CouponSource {
		return NewShopeeCouponAdapter(cfg.AppID, cfg.Secret)
	})
	DefaultRegistry.Register(domain.MarketplaceAmazon, func(cfg SourceConfig) CouponSource {
		return NewAmazonCouponAdapter(cfg.AccessKey, cfg.SecretKey, cfg.PartnerTag)
	})
}
