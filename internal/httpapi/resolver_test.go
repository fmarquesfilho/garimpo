package httpapi

import "testing"

func TestExtrairDadosURL(t *testing.T) {
	casos := []struct {
		url    string
		nome   string
		shopID string
		itemID string
	}{
		{
			url:    "https://shopee.com.br/Sérum-Vitamina-C-30ml-Hidratante-i.123456.789012?sp_atk=abc",
			nome:   "Sérum Vitamina C 30ml Hidratante",
			shopID: "123456",
			itemID: "789012",
		},
		{
			url:    "https://shopee.com.br/Perfume-Kenzo-100ml-i.99.88",
			nome:   "Perfume Kenzo 100ml",
			shopID: "99",
			itemID: "88",
		},
		{
			url:    "https://shopee.com.br/product/123/456",
			nome:   "",
			shopID: "123",
			itemID: "456",
		},
		{
			url:    "https://shopee.com.br/alguma-pagina-sem-produto",
			nome:   "",
			shopID: "",
			itemID: "",
		},
	}

	for _, c := range casos {
		nome, shopID, itemID := extrairDadosURL(c.url)
		if nome != c.nome {
			t.Errorf("URL=%s\n  nome: esperava %q, veio %q", c.url, c.nome, nome)
		}
		if shopID != c.shopID {
			t.Errorf("URL=%s\n  shopID: esperava %q, veio %q", c.url, c.shopID, shopID)
		}
		if itemID != c.itemID {
			t.Errorf("URL=%s\n  itemID: esperava %q, veio %q", c.url, c.itemID, itemID)
		}
	}
}
