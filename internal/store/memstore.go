package store

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/fmarquesfilho/garimpo/internal/apperr"
)

// ── NopRepository (dev/local) ────────────────────────────────────────────────

// NopRepository é a implementação em memória de Repository.
// Útil para dev local e testes. Não persiste nada entre restarts.
type NopRepository struct {
	destinos  *MemDestinoRepo
	templates *MemTemplateRepo
	favoritos *MemFavoritoRepo
	tenants   *MemTenantRepo
}

// NovoNopRepository cria um Repository em memória com templates padrão.
func NovoNopRepository() *NopRepository {
	r := &NopRepository{
		destinos:  &MemDestinoRepo{items: make(map[string]Destino)},
		templates: novoMemTemplateRepo(),
		favoritos: &MemFavoritoRepo{items: make(map[string][]Favorito)},
		tenants:   &MemTenantRepo{items: make(map[string]TenantConfig)},
	}
	return r
}

func (r *NopRepository) Eventos() EventoRepo                { return nopEventoRepo{} }
func (r *NopRepository) Snapshots() SnapshotRepo            { return nopSnapshotRepo{} }
func (r *NopRepository) Buscas() BuscaRepo                  { return nopBuscaRepo{} }
func (r *NopRepository) Publicacoes() PublicacaoRepo        { return nopPublicacaoRepo{} }
func (r *NopRepository) Destinos() DestinoRepo              { return r.destinos }
func (r *NopRepository) Templates() TemplateRepo            { return r.templates }
func (r *NopRepository) Favoritos() FavoritoRepo            { return r.favoritos }
func (r *NopRepository) Tenants() TenantRepo                { return r.tenants }
func (r *NopRepository) EnsureSchema(context.Context) error { return nil }
func (r *NopRepository) Nome() string                       { return "nop" }

// ── Nop sub-repos (descartam dados) ──────────────────────────────────────────

type nopEventoRepo struct{}

func (nopEventoRepo) Registrar(context.Context, Evento) error { return nil }

type nopSnapshotRepo struct{}

// NopSnapshots returns a no-op SnapshotRepo (used when BigQuery is unavailable).
func NopSnapshots() SnapshotRepo { return nopSnapshotRepo{} }

func (nopSnapshotRepo) RegistrarSnapshot(context.Context, Snapshot) error { return nil }
func (nopSnapshotRepo) Estatisticas(_ context.Context, dias int) (Estatisticas, error) {
	return Estatisticas{Fonte: "nop", DiasJanela: dias, GeradoEm: time.Now().UTC()}, nil
}
func (nopSnapshotRepo) HistoricoColetas(context.Context, int) ([]ColetaResumo, error) {
	return nil, nil
}
func (nopSnapshotRepo) Novidades(_ context.Context, buscaID string, dias int) (NovidadesLojas, error) {
	return NovidadesLojas{BuscaID: buscaID, DiasJanela: dias}, nil
}
func (nopSnapshotRepo) EvolucaoLojas(_ context.Context, dias int) (EvolucaoLojasResult, error) {
	return EvolucaoLojasResult{DiasJanela: dias}, nil
}

type nopBuscaRepo struct{}

func (nopBuscaRepo) SalvarBusca(context.Context, Busca) error      { return nil }
func (nopBuscaRepo) ListarBuscas(context.Context) ([]Busca, error) { return nil, nil }

type nopPublicacaoRepo struct{}

func (nopPublicacaoRepo) SalvarPublicacao(context.Context, Publicacao) error { return nil }
func (nopPublicacaoRepo) ListarPublicacoes(context.Context, string) ([]Publicacao, error) {
	return nil, nil
}
func (nopPublicacaoRepo) AtualizarPublicacao(context.Context, string, string, string) error {
	return nil
}
func (nopPublicacaoRepo) Conversoes(context.Context, int) ([]ConversaoResumo, error) { return nil, nil }

// ── MemDestinoRepo ───────────────────────────────────────────────────────────

type MemDestinoRepo struct {
	mu    sync.RWMutex
	items map[string]Destino
}

func (m *MemDestinoRepo) ListarDestinos(_ context.Context) ([]Destino, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Destino, 0, len(m.items))
	for _, d := range m.items {
		if d.Ativo {
			out = append(out, d)
		}
	}
	return out, nil
}

func (m *MemDestinoRepo) BuscarDestino(_ context.Context, id string) (Destino, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	d, ok := m.items[id]
	if !ok {
		return Destino{}, fmt.Errorf("destino %q: %w", id, apperr.ErrNotFound)
	}
	return d, nil
}

func (m *MemDestinoRepo) SalvarDestino(_ context.Context, d Destino) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items[d.ID] = d
	return nil
}

func (m *MemDestinoRepo) DeletarDestino(_ context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.items, id)
	return nil
}

// ── MemTemplateRepo ──────────────────────────────────────────────────────────

type MemTemplateRepo struct {
	mu    sync.RWMutex
	items map[string]Template
}

func novoMemTemplateRepo() *MemTemplateRepo {
	m := &MemTemplateRepo{items: make(map[string]Template)}
	m.items["padrao"] = Template{
		ID: "padrao", Nome: "Padrão",
		Corpo:   "✨ <b>{{nome}}</b>\n📂 <i>{{categoria}}</i>\n💸 <b>{{preco}}</b>\n🎯 {{estrategia}}",
		ComFoto: false, Ativo: true,
		CriadoEm: time.Now().UTC().Format(time.RFC3339),
	}
	m.items["foto"] = Template{
		ID: "foto", Nome: "Com foto",
		Corpo:   "✨ <b>{{nome}}</b>\n💸 <b>{{preco}}</b>",
		ComFoto: true, Ativo: true,
		CriadoEm: time.Now().UTC().Format(time.RFC3339),
	}
	return m
}

func (m *MemTemplateRepo) ListarTemplates(_ context.Context) ([]Template, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Template, 0, len(m.items))
	for _, t := range m.items {
		if t.Ativo {
			out = append(out, t)
		}
	}
	return out, nil
}

func (m *MemTemplateRepo) BuscarTemplate(_ context.Context, id string) (Template, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	t, ok := m.items[id]
	if !ok {
		return Template{}, fmt.Errorf("template %q: %w", id, apperr.ErrNotFound)
	}
	return t, nil
}

func (m *MemTemplateRepo) SalvarTemplate(_ context.Context, t Template) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items[t.ID] = t
	return nil
}

func (m *MemTemplateRepo) DeletarTemplate(_ context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.items, id)
	return nil
}

// ── MemFavoritoRepo ──────────────────────────────────────────────────────────

type MemFavoritoRepo struct {
	mu    sync.RWMutex
	items map[string][]Favorito // key = ownerUID
}

func (m *MemFavoritoRepo) SalvarFavorito(_ context.Context, f Favorito) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items[f.OwnerUID] = append(m.items[f.OwnerUID], f)
	return nil
}

func (m *MemFavoritoRepo) ListarFavoritos(_ context.Context, ownerUID string) ([]Favorito, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.items[ownerUID], nil
}

func (m *MemFavoritoRepo) RemoverFavorito(_ context.Context, ownerUID, produtoID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	favs := m.items[ownerUID]
	for i, f := range favs {
		if f.ProdutoID == produtoID {
			m.items[ownerUID] = append(favs[:i], favs[i+1:]...)
			return nil
		}
	}
	return nil
}

// ── MemTenantRepo ────────────────────────────────────────────────────────────

type MemTenantRepo struct {
	mu    sync.RWMutex
	items map[string]TenantConfig
}

// NovoMemTenantRepo cria um MemTenantRepo inicializado.
func NovoMemTenantRepo() *MemTenantRepo {
	return &MemTenantRepo{items: make(map[string]TenantConfig)}
}

func (m *MemTenantRepo) BuscarTenant(_ context.Context, uid string) (*TenantConfig, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	cfg, ok := m.items[uid]
	if !ok {
		return nil, nil
	}
	return &cfg, nil
}

func (m *MemTenantRepo) SalvarTenant(_ context.Context, cfg TenantConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items[cfg.UID] = cfg
	return nil
}

func (m *MemTenantRepo) ExcluirTenant(_ context.Context, uid string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.items, uid)
	return nil
}
