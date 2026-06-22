package store

import "testing"

func TestNormalizarBusca_KeywordLegadoViraKeywords(t *testing.T) {
	b := NormalizarBusca(Busca{KeywordLegado: "perfume"})
	if len(b.Keywords) != 1 || b.Keywords[0] != "perfume" {
		t.Errorf("esperava Keywords=[perfume], veio %v", b.Keywords)
	}
	if b.ID != "perfume" {
		t.Errorf("esperava ID=perfume, veio %q", b.ID)
	}
}

func TestNormalizarBusca_NomeLegadoViraID(t *testing.T) {
	b := NormalizarBusca(Busca{NomeLegado: "Perfumaria Diária", KeywordLegado: "perfume"})
	if b.ID != "perfumaria-diaria" {
		t.Errorf("esperava ID=perfumaria-diaria, veio %q", b.ID)
	}
}

func TestNormalizarBusca_KeywordsPreservadas(t *testing.T) {
	b := NormalizarBusca(Busca{ID: "meu-id", Keywords: []string{"kenzo", "shiseido"}})
	if len(b.Keywords) != 2 {
		t.Errorf("esperava 2 keywords, veio %d", len(b.Keywords))
	}
	if b.ID != "meu-id" {
		t.Errorf("esperava ID=meu-id, veio %q", b.ID)
	}
}

func TestNormalizarBusca_EstrategiaPadrao(t *testing.T) {
	b := NormalizarBusca(Busca{Keywords: []string{"x"}})
	if b.Estrategia != "nicho" {
		t.Errorf("estrategia padrão deveria ser nicho, veio %q", b.Estrategia)
	}
}

func TestNormalizarBusca_LimpaLegados(t *testing.T) {
	b := NormalizarBusca(Busca{NomeLegado: "teste", KeywordLegado: "x"})
	if b.NomeLegado != "" || b.KeywordLegado != "" {
		t.Errorf("legados deveriam ser limpos: nome=%q keyword=%q", b.NomeLegado, b.KeywordLegado)
	}
}

func TestNormalizarBusca_SemDadosRetornaBusca(t *testing.T) {
	b := NormalizarBusca(Busca{})
	if b.ID != "" {
		t.Errorf("sem dados deveria ter ID vazio, veio %q", b.ID)
	}
}

func TestSlugificar(t *testing.T) {
	testes := []struct {
		entrada string
		espera  string
	}{
		{"Perfumaria Diária", "perfumaria-diaria"},
		{"cosméticos", "cosmeticos"},
		{"  bem-estar  ", "bem-estar"},
		{"Ação & Aventura", "acao-aventura"},
		{"kenzo", "kenzo"},
		{"", "busca"},
		{"   ", "busca"},
		{"12345", "12345"},
		{"São Paulo", "sao-paulo"},
	}
	for _, tt := range testes {
		got := slugificar(tt.entrada)
		if got != tt.espera {
			t.Errorf("slugificar(%q) = %q, esperava %q", tt.entrada, got, tt.espera)
		}
	}
}
