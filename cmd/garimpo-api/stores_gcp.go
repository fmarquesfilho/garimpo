//go:build gcp

package main

import (
	"github.com/fmarquesfilho/garimpo/internal/publish"
	"github.com/fmarquesfilho/garimpo/internal/store"
)

// criarStoresAuxiliares tenta usar BigQuery para destinos e templates.
// Se o EventoStore for BigQueryStore, reutiliza o mesmo client.
func criarStoresAuxiliares(eventos store.EventoStore) (publish.DestinoStore, publish.TemplateStore) {
	if bq, ok := eventos.(*store.BigQueryStore); ok {
		return store.NovoBQDestinoStore(bq.Client(), bq.Dataset()),
			store.NovoBQTemplateStore(bq.Client(), bq.Dataset())
	}
	// Fallback: memória
	return publish.NovoMemDestinoStore(), publish.NovoMemTemplateStore()
}
