package store_test

import (
	"testing"

	"github.com/fmarquesfilho/garimpo/internal/store"
)

// ── Compile-time interface conformance checks ────────────────────────────────
// Se alguma implementação deixar de satisfazer a interface, o build quebra aqui.

// NopRepository satisfaz Repository.
var _ store.Repository = (*store.NopRepository)(nil)

// DestinoRepo — verifies MemDestinoRepo implements the DestinoRepo interface
var _ store.DestinoRepo = (*store.MemDestinoRepo)(nil)

// TemplateRepo — verifies MemTemplateRepo implements the TemplateRepo interface
var _ store.TemplateRepo = (*store.MemTemplateRepo)(nil)

// FavoritoRepo — verifies MemFavoritoRepo implements the FavoritoRepo interface
var _ store.FavoritoRepo = (*store.MemFavoritoRepo)(nil)

// TenantRepo — verifies MemTenantRepo implements the TenantRepo interface
var _ store.TenantRepo = (*store.MemTenantRepo)(nil)

func TestNopRepositoryConformance(t *testing.T) {
	repo := store.NovoNopRepository()

	// Verifica que cada accessor retorna non-nil
	if repo.Eventos() == nil {
		t.Fatal("Eventos() retornou nil")
	}
	if repo.Snapshots() == nil {
		t.Fatal("Snapshots() retornou nil")
	}
	if repo.Buscas() == nil {
		t.Fatal("Buscas() retornou nil")
	}
	if repo.Publicacoes() == nil {
		t.Fatal("Publicacoes() retornou nil")
	}
	if repo.Destinos() == nil {
		t.Fatal("Destinos() retornou nil")
	}
	if repo.Templates() == nil {
		t.Fatal("Templates() retornou nil")
	}
	if repo.Favoritos() == nil {
		t.Fatal("Favoritos() retornou nil")
	}
	if repo.Tenants() == nil {
		t.Fatal("Tenants() retornou nil")
	}
	if repo.Nome() == "" {
		t.Fatal("Nome() retornou string vazia")
	}
}
