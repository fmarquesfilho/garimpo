//go:build !gcp

package main

import (
	"github.com/fmarquesfilho/garimpo/internal/publish"
	"github.com/fmarquesfilho/garimpo/internal/store"
)

// criarStoresAuxiliares sem GCP: usa memória.
func criarStoresAuxiliares(_ store.EventoStore) (publish.DestinoStore, publish.TemplateStore) {
	return publish.NovoMemDestinoStore(), publish.NovoMemTemplateStore()
}
