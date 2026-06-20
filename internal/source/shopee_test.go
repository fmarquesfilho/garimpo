package source

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBuildQueryIncluiParametros(t *testing.T) {
	s := NewShopeeAPISource("app", "sec")
	s.ListType = 1
	s.SortType = 5
	s.ProductCatID = 100017
	s.Keyword = "perfume"
	q := s.buildQuery(1)

	for _, frag := range []string{"listType: 1", "sortType: 5", "productCatId: 100017", `keyword: "perfume"`, "productOfferV2", "commissionRate"} {
		if !strings.Contains(q, frag) {
			t.Errorf("query deveria conter %q:\n%s", frag, q)
		}
	}
}

func TestShopeeFetchMapeiaCampos(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// confirma que vai assinado
		if !strings.HasPrefix(r.Header.Get("Authorization"), "SHA256 Credential=") {
			t.Errorf("Authorization ausente/errado: %q", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		// preços/notas como string, vendas como número — o caso real misturado
		_, _ = w.Write([]byte(`{"data":{"productOfferV2":{"nodes":[
			{"itemId":123,"productName":"Perfume Floral","priceMin":"99.90","sales":50,"ratingStar":"4.5","commissionRate":"0.085","offerLink":"https://s.shopee/x"}
		],"pageInfo":{"hasNextPage":false}}}}`))
	}))
	defer srv.Close()

	s := NewShopeeAPISource("app", "sec")
	s.Endpoint = srv.URL
	s.CategoryLabel = "perfumaria"

	produtos, err := s.Fetch()
	if err != nil {
		t.Fatal(err)
	}
	if len(produtos) != 1 {
		t.Fatalf("esperava 1 produto, veio %d", len(produtos))
	}
	p := produtos[0]
	if p.ID != "123" || p.Name != "Perfume Floral" || p.Category != "perfumaria" {
		t.Errorf("texto/categoria errados: %+v", p)
	}
	if p.Price != 99.90 || p.Commission != 0.085 || p.Sales30d != 50 || p.Rating != 4.5 {
		t.Errorf("numéricos (flex) errados: %+v", p)
	}
	if p.Link != "https://s.shopee/x" {
		t.Errorf("offerLink não mapeado: %q", p.Link)
	}
}

func TestShopeeFetchPropagaErroDaAPI(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"errors":[{"message":"Invalid Signature","extensions":{"code":10020}}]}`))
	}))
	defer srv.Close()

	s := NewShopeeAPISource("app", "sec")
	s.Endpoint = srv.URL
	_, err := s.Fetch()
	if err == nil || !strings.Contains(err.Error(), "10020") {
		t.Errorf("esperava erro com código 10020, veio: %v", err)
	}
}

func TestShopeeFetchSemCredenciais(t *testing.T) {
	if _, err := (&ShopeeAPISource{}).Fetch(); err == nil {
		t.Error("sem AppID/Secret deveria falhar")
	}
}
