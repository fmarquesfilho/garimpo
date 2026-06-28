package publish

import (
	"context"
	"fmt"
	"sync"

	"github.com/fmarquesfilho/garimpo/internal/apperr"
)

// MemDestinoStore é uma implementação em memória do DestinoStore (dev/testes).
type MemDestinoStore struct {
	mu       sync.RWMutex
	destinos map[string]Destino
}

func NovoMemDestinoStore() *MemDestinoStore {
	return &MemDestinoStore{destinos: make(map[string]Destino)}
}

func (m *MemDestinoStore) Listar(_ context.Context) ([]Destino, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	lista := make([]Destino, 0, len(m.destinos))
	for _, d := range m.destinos {
		if d.Ativo {
			lista = append(lista, d)
		}
	}
	return lista, nil
}

func (m *MemDestinoStore) Buscar(_ context.Context, id string) (Destino, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	d, ok := m.destinos[id]
	if !ok {
		return Destino{}, fmt.Errorf("destino %q: %w", id, apperr.ErrNotFound)
	}
	return d, nil
}

func (m *MemDestinoStore) Salvar(_ context.Context, d Destino) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.destinos[d.ID] = d
	return nil
}

func (m *MemDestinoStore) Deletar(_ context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.destinos, id)
	return nil
}
