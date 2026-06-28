package store

import "context"

// Repository é o ponto de acesso unificado à camada de persistência.
// Compõe sub-interfaces segregadas por domínio (Interface Segregation).
//
// Consumidores declaram dependência apenas na sub-interface que precisam:
//
//	func meuHandler(buscas store.BuscaRepo) { ... }
//
// O Server do httpapi recebe o Repository inteiro e distribui as partes.
type Repository interface {
	Eventos() EventoRepo
	Snapshots() SnapshotRepo
	Buscas() BuscaRepo
	Publicacoes() PublicacaoRepo
	Destinos() DestinoRepo
	Templates() TemplateRepo
	Favoritos() FavoritoRepo
	Tenants() TenantRepo
	EnsureSchema(ctx context.Context) error
	Nome() string
}

// ── Sub-interfaces segregadas ────────────────────────────────────────────────

// EventoRepo registra eventos de curadoria (seleção, publicação).
type EventoRepo interface {
	Registrar(ctx context.Context, e Evento) error
}

// SnapshotRepo persiste e consulta snapshots de mercado.
type SnapshotRepo interface {
	RegistrarSnapshot(ctx context.Context, s Snapshot) error
	Estatisticas(ctx context.Context, dias int) (Estatisticas, error)
	HistoricoColetas(ctx context.Context, dias int) ([]ColetaResumo, error)
	Novidades(ctx context.Context, buscaID string, dias int) (NovidadesLojas, error)
	EvolucaoLojas(ctx context.Context, dias int) (EvolucaoLojasResult, error)
}

// BuscaRepo persiste perfis de busca (coleta agendada).
type BuscaRepo interface {
	SalvarBusca(ctx context.Context, b Busca) error
	ListarBuscas(ctx context.Context) ([]Busca, error)
}

// PublicacaoRepo gerencia publicações agendadas e conversões.
type PublicacaoRepo interface {
	SalvarPublicacao(ctx context.Context, p Publicacao) error
	ListarPublicacoes(ctx context.Context, status string) ([]Publicacao, error)
	AtualizarPublicacao(ctx context.Context, id, status, detalhe string) error
	Conversoes(ctx context.Context, dias int) ([]ConversaoResumo, error)
}

// DestinoRepo persiste destinos de publicação (Telegram, WhatsApp, etc.).
type DestinoRepo interface {
	ListarDestinos(ctx context.Context) ([]Destino, error)
	BuscarDestino(ctx context.Context, id string) (Destino, error)
	SalvarDestino(ctx context.Context, d Destino) error
	DeletarDestino(ctx context.Context, id string) error
}

// TemplateRepo persiste templates de mensagem.
type TemplateRepo interface {
	ListarTemplates(ctx context.Context) ([]Template, error)
	BuscarTemplate(ctx context.Context, id string) (Template, error)
	SalvarTemplate(ctx context.Context, t Template) error
	DeletarTemplate(ctx context.Context, id string) error
}

// FavoritoRepo persiste produtos favoritos do usuário.
type FavoritoRepo interface {
	SalvarFavorito(ctx context.Context, f Favorito) error
	ListarFavoritos(ctx context.Context, ownerUID string) ([]Favorito, error)
	RemoverFavorito(ctx context.Context, ownerUID, produtoID string) error
}

// TenantRepo persiste configurações por tenant (multi-tenancy).
type TenantRepo interface {
	BuscarTenant(ctx context.Context, uid string) (*TenantConfig, error)
	SalvarTenant(ctx context.Context, cfg TenantConfig) error
	ExcluirTenant(ctx context.Context, uid string) error
}
