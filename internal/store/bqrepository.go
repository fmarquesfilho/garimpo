//go:build gcp

package store

import (
	"context"
)

// BQRepository implementa Repository sobre BigQuery.
// Compõe o BigQueryStore existente e expõe as sub-interfaces.
type BQRepository struct {
	bq       *BigQueryStore
	destinos *BQDestinoStore
	tmpl     *BQTemplateStore
	tenants  TenantRepo // injetado externamente (Firestore ou memória)
}

// NovoBQRepository cria o Repository completo a partir de um BigQueryStore.
// O TenantRepo é injetado porque pode vir de Firestore ou memória.
func NovoBQRepository(bq *BigQueryStore, tenants TenantRepo) *BQRepository {
	return &BQRepository{
		bq:       bq,
		destinos: NovoBQDestinoStore(bq.Client(), bq.Dataset()),
		tmpl:     NovoBQTemplateStore(bq.Client(), bq.Dataset()),
		tenants:  tenants,
	}
}

func (r *BQRepository) Eventos() EventoRepo                    { return r.bq }
func (r *BQRepository) Snapshots() SnapshotRepo                { return r.bq }
func (r *BQRepository) Buscas() BuscaRepo                      { return r.bq }
func (r *BQRepository) Publicacoes() PublicacaoRepo            { return r.bq }
func (r *BQRepository) Destinos() DestinoRepo                  { return r.destinos }
func (r *BQRepository) Templates() TemplateRepo                { return r.tmpl }
func (r *BQRepository) Favoritos() FavoritoRepo                { return r.bq }
func (r *BQRepository) Tenants() TenantRepo                    { return r.tenants }
func (r *BQRepository) EnsureSchema(ctx context.Context) error { return r.bq.EnsureSchema(ctx) }
func (r *BQRepository) Nome() string                           { return "bigquery" }
