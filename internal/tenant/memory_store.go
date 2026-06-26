package tenant

import (
	"context"
	"sync"
)

// MemoryStore mantém configs em memória — útil para testes e dev local.
type MemoryStore struct {
	mu      sync.RWMutex
	configs map[string]Config
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{configs: make(map[string]Config)}
}

func (m *MemoryStore) Buscar(_ context.Context, uid string) (*Config, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	cfg, ok := m.configs[uid]
	if !ok {
		return nil, nil
	}
	return &cfg, nil
}

func (m *MemoryStore) Salvar(_ context.Context, cfg Config) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.configs[cfg.UID] = cfg
	return nil
}

func (m *MemoryStore) Excluir(_ context.Context, uid string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.configs, uid)
	return nil
}
