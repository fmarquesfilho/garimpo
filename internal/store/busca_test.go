package store

import (
	"testing"
)

func TestNormalizarBusca_GeraID(t *testing.T) {
	tests := []struct {
		nome   string
		busca  Busca
		wantID string
	}{
		{"keyword única", Busca{Keywords: []string{"sérum"}}, "serum"},
		{"múltiplas keywords", Busca{Keywords: []string{"skin1004", "cosrx"}}, "skin1004"},
		{"shop_id sem keyword", Busca{ShopIDs: []int64{258316442}}, "loja-258316442"},
		{"categoria sem keyword ou loja", Busca{Categorias: []string{"cosméticos"}}, "cosmeticos"},
		{"nome explícito", Busca{Nome: "Skincare coreana"}, "skincare-coreana"},
		{"keyword legada", Busca{KeywordLegado: "perfume"}, "perfume"},
		{"ID já definido não muda", Busca{ID: "minha-busca", Keywords: []string{"x"}}, "minha-busca"},
	}
	for _, tt := range tests {
		t.Run(tt.nome, func(t *testing.T) {
			got := NormalizarBusca(tt.busca)
			if got.ID != tt.wantID {
				t.Errorf("ID = %q, want %q", got.ID, tt.wantID)
			}
		})
	}
}

func TestNormalizarBusca_InfereFontes(t *testing.T) {
	tests := []struct {
		nome       string
		busca      Busca
		wantFontes []string
	}{
		{"keyword → curadoria", Busca{Keywords: []string{"sérum"}}, []string{"curadoria"}},
		{"shop_id → quedas+novos", Busca{ShopIDs: []int64{123}}, []string{"quedas", "novos"}},
		{"keyword + shop → curadoria+quedas+novos", Busca{Keywords: []string{"x"}, ShopIDs: []int64{123}}, []string{"curadoria", "quedas", "novos"}},
		{"categorias → curadoria", Busca{Categorias: []string{"cosméticos"}}, []string{"curadoria"}},
		{"fontes explícitas não são sobrescritas", Busca{Keywords: []string{"x"}, Fontes: []string{"novos"}}, []string{"novos"}},
		{"sem nada → vazio", Busca{}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.nome, func(t *testing.T) {
			got := NormalizarBusca(tt.busca)
			if len(got.Fontes) != len(tt.wantFontes) {
				t.Fatalf("Fontes = %v, want %v", got.Fontes, tt.wantFontes)
			}
			for i, f := range got.Fontes {
				if f != tt.wantFontes[i] {
					t.Errorf("Fontes[%d] = %q, want %q", i, f, tt.wantFontes[i])
				}
			}
		})
	}
}

func TestNormalizarBusca_MigraCategoria(t *testing.T) {
	b := NormalizarBusca(Busca{Categoria: "cosméticos", Keywords: []string{"x"}})
	if len(b.Categorias) != 1 || b.Categorias[0] != "cosméticos" {
		t.Errorf("Categorias = %v, want [cosméticos]", b.Categorias)
	}
}

func TestNormalizarBusca_DiasJanelaDefault(t *testing.T) {
	b := NormalizarBusca(Busca{Keywords: []string{"x"}})
	if b.DiasJanela != 7 {
		t.Errorf("DiasJanela = %d, want 7", b.DiasJanela)
	}

	b2 := NormalizarBusca(Busca{Keywords: []string{"x"}, DiasJanela: 3})
	if b2.DiasJanela != 3 {
		t.Errorf("DiasJanela = %d, want 3 (user-defined)", b2.DiasJanela)
	}
}

func TestNormalizarBusca_EstrategiaDefault(t *testing.T) {
	b := NormalizarBusca(Busca{Keywords: []string{"x"}})
	if b.Estrategia != "nicho" {
		t.Errorf("Estrategia = %q, want 'nicho'", b.Estrategia)
	}
}

func TestNormalizarBusca_Cenarios(t *testing.T) {
	// Verifica que os 17 cenários documentados produzem buscas válidas
	cenarios := []struct {
		nome  string
		busca Busca
	}{
		{"1: keyword única", Busca{Keywords: []string{"sérum"}}},
		{"2: múltiplas keywords", Busca{Keywords: []string{"skin1004", "cosrx", "innisfree"}}},
		{"3: loja sem keyword", Busca{ShopIDs: []int64{258316442}}},
		{"4: múltiplas lojas", Busca{ShopIDs: []int64{258316442, 123456789}}},
		{"5: keyword + loja", Busca{Keywords: []string{"sérum"}, ShopIDs: []int64{258316442}}},
		{"6: multi keyword + multi loja", Busca{Keywords: []string{"a", "b"}, ShopIDs: []int64{1, 2}}},
		{"7: sem nada, só quedas", Busca{Fontes: []string{"quedas"}}},
		{"8: keyword manual (sem cron)", Busca{Keywords: []string{"maquiagem"}}},
		{"9: favoritos", Busca{Fontes: []string{"favoritos"}}},
		{"10: novos sem keyword/loja", Busca{Fontes: []string{"novos"}}},
		{"11: novos em loja", Busca{ShopIDs: []int64{123}, Fontes: []string{"novos"}}},
		{"12: novos com keyword", Busca{Keywords: []string{"retinol"}, Fontes: []string{"novos"}}},
		{"13: novos em loja + keyword", Busca{Keywords: []string{"sérum"}, ShopIDs: []int64{123}, Fontes: []string{"novos"}}},
		{"14: categoria sem keyword/loja", Busca{Categorias: []string{"cosméticos", "skincare"}}},
		{"15: loja + categoria", Busca{ShopIDs: []int64{123}, Categorias: []string{"perfumaria"}}},
		{"16: keyword + categoria", Busca{Keywords: []string{"sérum"}, Categorias: []string{"skincare"}}},
		{"17: keyword + loja + categoria", Busca{Keywords: []string{"retinol"}, ShopIDs: []int64{123}, Categorias: []string{"skincare"}}},
	}

	for _, c := range cenarios {
		t.Run(c.nome, func(t *testing.T) {
			b := NormalizarBusca(c.busca)
			// Toda busca normalizada deve ter:
			// - Um ID não-vazio (gerado ou explícito)
			// - Estrategia preenchida
			// - DiasJanela > 0
			if b.ID == "" {
				t.Error("ID vazio após normalização")
			}
			if b.Estrategia == "" {
				t.Error("Estrategia vazia")
			}
			if b.DiasJanela <= 0 {
				t.Error("DiasJanela <= 0")
			}
		})
	}
}
