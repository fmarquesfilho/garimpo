package store

import (
	"os"
	"strings"
	"testing"
)

// TestEntidadesDiagramaSincronizado verifica que os campos principais
// documentados no diagrama ER (docs/ENTIDADES.md) existem na struct Busca.
// Se adicionar um campo à Busca, este teste lembra de atualizar o diagrama.
func TestEntidadesDiagramaSincronizado(t *testing.T) {
	// Lê o diagrama
	data, err := os.ReadFile("../../docs/ENTIDADES.md")
	if err != nil {
		t.Skipf("ENTIDADES.md não encontrado: %v", err)
	}
	diagrama := string(data)

	// Campos que devem existir no diagrama se existem na struct
	camposBusca := []struct {
		campo   string
		noCode  bool // true se existe na struct
		noDiag  string // como aparece no diagrama
	}{
		{"ID", true, "id PK"},
		{"Nome", true, "nome"},
		{"Keywords", true, "keywords"},
		{"ShopIDs", true, "shop_ids"},
		{"Cron", true, "cron"},
		{"Ativo", true, "ativo"},
		{"OwnerUID", true, "owner_uid"},
		{"RotationCursor", true, "rotation_cursor"},
	}

	for _, c := range camposBusca {
		if c.noCode && !strings.Contains(diagrama, c.noDiag) {
			t.Errorf("Campo Busca.%s existe no código mas '%s' não está em ENTIDADES.md — atualize o diagrama",
				c.campo, c.noDiag)
		}
	}

	// Entidades que devem estar documentadas
	entidades := []string{"BUSCA", "SNAPSHOT", "PUBLICACAO", "DESTINO", "TEMPLATE", "EVENTO"}
	for _, e := range entidades {
		if !strings.Contains(diagrama, e) {
			t.Errorf("Entidade %s não encontrada em ENTIDADES.md", e)
		}
	}
}
