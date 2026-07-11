package busca_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/fmarquesfilho/garimpo/internal/busca"
)

type fixture struct {
	ID             string            `json:"id"`
	Tipo           string            `json:"tipo"`
	Keywords       []string          `json:"keywords"`
	ShopIDs        []int64           `json:"shop_ids"`
	ShopNames      map[string]string `json:"shop_names"`
	Categorias     []string          `json:"categorias"`
	CollectionKeys []string          `json:"collection_keys"`
	Cron           *string           `json:"cron"`
	Marketplaces   []string          `json:"marketplaces"`
	OwnerUID       string            `json:"owner_uid"`
	ComissaoMin    *float64          `json:"comissao_min"`
	VendasMin      *int              `json:"vendas_min"`
	Fontes         []string          `json:"fontes"`
}

func loadFixtures(t *testing.T) []fixture {
	t.Helper()
	data, err := os.ReadFile("../../fixtures/buscas.json")
	if err != nil {
		t.Fatalf("failed to read fixtures: %v", err)
	}
	var fixtures []fixture
	if err := json.Unmarshal(data, &fixtures); err != nil {
		t.Fatalf("failed to parse fixtures: %v", err)
	}
	return fixtures
}

func TestDeriveCollectionKeys_Fixtures(t *testing.T) {
	fixtures := loadFixtures(t)

	for _, fx := range fixtures {
		t.Run(fx.ID, func(t *testing.T) {
			got := busca.DeriveCollectionKeys(fx.ShopIDs, fx.Keywords, fx.Categorias...)

			if len(got) != len(fx.CollectionKeys) {
				t.Fatalf("DeriveCollectionKeys(%v, %v, %v) = %v (len %d), want %v (len %d)",
					fx.ShopIDs, fx.Keywords, fx.Categorias, got, len(got), fx.CollectionKeys, len(fx.CollectionKeys))
			}

			for i := range got {
				if got[i] != fx.CollectionKeys[i] {
					t.Errorf("index %d: got %q, want %q", i, got[i], fx.CollectionKeys[i])
				}
			}
		})
	}
}

func TestDeriveCollectionKeys_Properties(t *testing.T) {
	// Property: sorted, non-empty, no duplicates
	cases := []struct {
		name     string
		shopIDs  []int64
		keywords []string
	}{
		{"shop only", []int64{999, 111, 555}, nil},
		{"keywords only", nil, []string{"Banana", "  apple  ", "CHERRY"}},
		{"mixed", []int64{42}, []string{"foo", "42"}}, // dedupe: "42" appears in both
		{"whitespace keywords", nil, []string{"  ", "", "valid"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var kws []string
			if tc.keywords != nil {
				kws = tc.keywords
			}
			got := busca.DeriveCollectionKeys(tc.shopIDs, kws)

			// Non-empty (at least for valid inputs)
			if len(tc.shopIDs) > 0 || hasNonEmpty(kws) {
				if len(got) == 0 {
					t.Fatal("expected non-empty result")
				}
			}

			// Sorted
			for i := 1; i < len(got); i++ {
				if got[i] < got[i-1] {
					t.Errorf("not sorted: %v", got)
					break
				}
			}

			// No duplicates
			seen := make(map[string]bool)
			for _, k := range got {
				if seen[k] {
					t.Errorf("duplicate: %q in %v", k, got)
				}
				seen[k] = true
			}

			// No empty strings
			for _, k := range got {
				if k == "" {
					t.Error("empty string in result")
				}
			}
		})
	}
}

func hasNonEmpty(ss []string) bool {
	for _, s := range ss {
		if len(s) > 0 && len(s) != len("  ") {
			return true
		}
	}
	return false
}
