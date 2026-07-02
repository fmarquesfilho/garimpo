package source

import (
	"fmt"
	"sync"

	"github.com/fmarquesfilho/garimpo/internal/domain"
)

// Registry mantém o mapeamento marketplace → factory.
// É seguro para acesso concorrente (leituras após init são lock-free).
// O padrão idiomático em Go é registrar na init() de cada adapter e consumir no main.
type Registry struct {
	mu        sync.RWMutex
	factories map[string]SourceFactory
}

// DefaultRegistry é o registry global usado pelo collector.
// Adapters se registram via Register() em seus init() ou no setup do serviço.
var DefaultRegistry = NewRegistry()

func NewRegistry() *Registry {
	return &Registry{
		factories: make(map[string]SourceFactory),
	}
}

// Register associa um marketplace a uma factory.
func (r *Registry) Register(marketplace string, factory SourceFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.factories[marketplace] = factory
}

// Get retorna a factory para o marketplace dado, ou erro se não registrado.
func (r *Registry) Get(marketplace string) (SourceFactory, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	f, ok := r.factories[marketplace]
	if !ok {
		return nil, fmt.Errorf("marketplace %q não registrado: %w", marketplace, ErrUnsupportedMarketplace)
	}
	return f, nil
}

// Create é um atalho: busca a factory e já cria a source.
func (r *Registry) Create(marketplace string, cfg SourceConfig) (ProductSource, error) {
	factory, err := r.Get(marketplace)
	if err != nil {
		return nil, err
	}
	return factory(cfg), nil
}

// Supported retorna a lista de marketplaces registrados.
func (r *Registry) Supported() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	keys := make([]string, 0, len(r.factories))
	for k := range r.factories {
		keys = append(keys, k)
	}
	return keys
}

// ErrUnsupportedMarketplace indica que não há adapter registrado para o marketplace.
var ErrUnsupportedMarketplace = fmt.Errorf("marketplace não suportado")

func init() {
	// Registra adapters conhecidos no registry global.
	DefaultRegistry.Register(domain.MarketplaceShopee, func(cfg SourceConfig) ProductSource {
		return NewShopeeAdapter(cfg.AppID, cfg.Secret)
	})
	DefaultRegistry.Register(domain.MarketplaceAmazon, func(cfg SourceConfig) ProductSource {
		return NewAmazonAdapter(cfg.AccessKey, cfg.SecretKey, cfg.PartnerTag)
	})
}
