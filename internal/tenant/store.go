package tenant

import "context"

// Store persiste e recupera configurações de tenants.
type Store interface {
	// Buscar retorna a config do tenant. Retorna nil se não existe.
	Buscar(ctx context.Context, uid string) (*Config, error)
	// Salvar cria ou atualiza a config do tenant.
	Salvar(ctx context.Context, cfg Config) error
	// Excluir remove config e todos os dados do tenant (LGPD).
	Excluir(ctx context.Context, uid string) error
}

// NopStore é um store que não persiste nada — usado em dev local.
type NopStore struct{}

func (NopStore) Buscar(_ context.Context, _ string) (*Config, error) { return nil, nil }
func (NopStore) Salvar(_ context.Context, _ Config) error            { return nil }
func (NopStore) Excluir(_ context.Context, _ string) error           { return nil }
