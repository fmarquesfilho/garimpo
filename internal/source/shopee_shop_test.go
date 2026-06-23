package source

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestShopeeShopSourceFetchOK(t *testing.T) {
	respBody := `{
		"data": {
			"shopOfferV2": {
				"nodes": [
					{
						"itemId": "111",
						"productName": "Sérum Coreano",
						"productLink": "https://shopee.com.br/serum",
						"offerLink": "https://shope.ee/aff111",
						"priceMin": 59.90,
						"sales": 120,
						"ratingStar": 4.7,
						"commissionRate": 0.08,
						"shopName": "Glory of Seoul",
						"imageUrl": "https://img.shopee.com/111.jpg"
					},
					{
						"itemId": "222",
						"productName": "Tônico Facial",
						"productLink": "https://shopee.com.br/tonico",
						"offerLink": "https://shope.ee/aff222",
						"priceMin": 35.00,
						"sales": 50,
						"ratingStar": 4.3,
						"commissionRate": 0.05,
						"shopName": "Glory of Seoul",
						"imageUrl": ""
					}
				],
				"pageInfo": {"page": 1, "hasNextPage": false}
			}
		}
	}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verifica que a query contém shopId
		body := make([]byte, r.ContentLength)
		r.Body.Read(body)
		if !strings.Contains(string(body), "shopOfferV2") {
			t.Errorf("query deveria conter shopOfferV2: %s", string(body))
		}
		w.Write([]byte(respBody))
	}))
	defer srv.Close()

	source := NewShopeeShopSource("appid", "secret", []int64{99999})
	source.Endpoint = srv.URL

	produtos, err := source.Fetch()
	if err != nil {
		t.Fatal(err)
	}
	if len(produtos) != 2 {
		t.Fatalf("esperava 2 produtos, veio %d", len(produtos))
	}
	if produtos[0].ID != "111" || produtos[0].Name != "Sérum Coreano" {
		t.Errorf("primeiro produto errado: %+v", produtos[0])
	}
	if produtos[0].Image != "https://img.shopee.com/111.jpg" {
		t.Errorf("imagem errada: %q", produtos[0].Image)
	}
	if produtos[1].Commission != 0.05 {
		t.Errorf("comissão do segundo produto errada: %f", produtos[1].Commission)
	}
}

func TestShopeeShopSourceComKeyword(t *testing.T) {
	var queryRecebida string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := make([]byte, r.ContentLength)
		r.Body.Read(body)
		var req map[string]string
		json.Unmarshal(body, &req)
		queryRecebida = req["query"]
		w.Write([]byte(`{"data":{"shopOfferV2":{"nodes":[],"pageInfo":{"page":1,"hasNextPage":false}}}}`))
	}))
	defer srv.Close()

	source := NewShopeeShopSource("appid", "secret", []int64{12345})
	source.Endpoint = srv.URL
	source.Keyword = "sérum"

	_, err := source.Fetch()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(queryRecebida, `keyword: "sérum"`) {
		t.Errorf("query deveria conter keyword: %q", queryRecebida)
	}
	if !strings.Contains(queryRecebida, "shopId: 12345") {
		t.Errorf("query deveria conter shopId: %q", queryRecebida)
	}
}

func TestShopeeShopSourceSemCredenciais(t *testing.T) {
	source := NewShopeeShopSource("", "", []int64{123})
	_, err := source.Fetch()
	if err == nil {
		t.Error("deveria retornar erro sem credenciais")
	}
}

func TestShopeeShopSourceSemShopIDs(t *testing.T) {
	source := NewShopeeShopSource("app", "secret", nil)
	_, err := source.Fetch()
	if err == nil {
		t.Error("deveria retornar erro sem shopIDs")
	}
}

func TestShopeeShopSourceErroAPI(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"errors":[{"message":"rate limit exceeded","extensions":{"code":429}}]}`))
	}))
	defer srv.Close()

	source := NewShopeeShopSource("app", "secret", []int64{123})
	source.Endpoint = srv.URL

	_, err := source.Fetch()
	if err == nil || !strings.Contains(err.Error(), "rate limit") {
		t.Errorf("esperava erro de rate limit, veio: %v", err)
	}
}

func TestShopeeShopSourceMultiplasLojas(t *testing.T) {
	chamadas := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		chamadas++
		w.Write([]byte(`{"data":{"shopOfferV2":{"nodes":[{"itemId":"` +
			string(rune('0'+chamadas)) + `","productName":"P","offerLink":"","priceMin":10,"sales":1,"ratingStar":4,"commissionRate":0.1,"shopName":"S","imageUrl":""}],` +
			`"pageInfo":{"page":1,"hasNextPage":false}}}}`))
	}))
	defer srv.Close()

	source := NewShopeeShopSource("app", "secret", []int64{111, 222, 333})
	source.Endpoint = srv.URL

	produtos, err := source.Fetch()
	if err != nil {
		t.Fatal(err)
	}
	if chamadas != 3 {
		t.Errorf("deveria fazer 3 chamadas (uma por loja), fez %d", chamadas)
	}
	if len(produtos) != 3 {
		t.Errorf("deveria retornar 3 produtos, veio %d", len(produtos))
	}
}

func TestShopeeAPISourceBuildQueryComItemID(t *testing.T) {
	var queryRecebida string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := make([]byte, r.ContentLength)
		r.Body.Read(body)
		var req map[string]string
		json.Unmarshal(body, &req)
		queryRecebida = req["query"]
		w.Write([]byte(`{"data":{"productOfferV2":{"nodes":[],"pageInfo":{"page":1,"hasNextPage":false}}}}`))
	}))
	defer srv.Close()

	src := NewShopeeAPISource("app", "secret")
	src.ItemID = "10301156503"
	src.Limit = 1
	src.MaxPages = 1
	src.Endpoint = srv.URL

	src.Fetch()

	// A query deve conter itemId como parâmetro nomeado, NÃO como keyword
	if !strings.Contains(queryRecebida, "itemId: 10301156503") {
		t.Errorf("query deveria conter 'itemId: 10301156503', veio:\n%s", queryRecebida)
	}
	if strings.Contains(queryRecebida, `keyword: "10301156503"`) {
		t.Errorf("query NÃO deveria usar itemId como keyword, veio:\n%s", queryRecebida)
	}
}

func TestShopeeAPISourceBuildQueryComKeyword(t *testing.T) {
	var queryRecebida string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := make([]byte, r.ContentLength)
		r.Body.Read(body)
		var req map[string]string
		json.Unmarshal(body, &req)
		queryRecebida = req["query"]
		w.Write([]byte(`{"data":{"productOfferV2":{"nodes":[],"pageInfo":{"page":1,"hasNextPage":false}}}}`))
	}))
	defer srv.Close()

	src := NewShopeeAPISource("app", "secret")
	src.Keyword = "sérum vitamina c"
	src.Endpoint = srv.URL

	src.Fetch()

	if !strings.Contains(queryRecebida, `keyword: "sérum vitamina c"`) {
		t.Errorf("query deveria conter keyword entre aspas, veio:\n%s", queryRecebida)
	}
	if strings.Contains(queryRecebida, "itemId:") {
		t.Errorf("query não deveria ter itemId quando só keyword é usado:\n%s", queryRecebida)
	}
}
