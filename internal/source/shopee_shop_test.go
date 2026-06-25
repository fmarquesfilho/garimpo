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
						"productLink": "https://shopee.com.br/Serum-Coreano-i.99999.111",
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
						"productLink": "https://shopee.com.br/Tonico-Facial-i.99999.222",
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

// --- Testes de amostragem rotativa ─────────────────────────────────────────

func TestShopeeShopSourceStartPageRotation(t *testing.T) {
	// Servidor retorna hasNextPage=true sempre para simular catálogo grande
	paginasRecebidas := []int{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := make([]byte, r.ContentLength)
		r.Body.Read(body)
		var req map[string]string
		json.Unmarshal(body, &req)
		query := req["query"]
		// Extrai o número da página da query
		for _, part := range strings.Split(query, ",") {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, "page:") {
				pageStr := strings.TrimSpace(strings.TrimPrefix(part, "page:"))
				page := 0
				for _, c := range pageStr {
					if c >= '0' && c <= '9' {
						page = page*10 + int(c-'0')
					}
				}
				paginasRecebidas = append(paginasRecebidas, page)
			}
		}
		w.Write([]byte(`{"data":{"shopOfferV2":{"nodes":[{"itemId":"1","productName":"P","offerLink":"","priceMin":10,"sales":1,"ratingStar":4,"commissionRate":0.1,"shopName":"S","imageUrl":""}],"pageInfo":{"page":1,"hasNextPage":true}}}}`))
	}))
	defer srv.Close()

	source := NewShopeeShopSource("app", "secret", []int64{12345})
	source.Endpoint = srv.URL
	source.StartPage = 3 // Simula rotação: começa na página 3
	source.MaxPages = 2  // Busca 2 páginas

	_, err := source.Fetch()
	if err != nil {
		t.Fatal(err)
	}

	// Deveria ter buscado páginas 3 e 4
	if len(paginasRecebidas) != 2 {
		t.Fatalf("esperava 2 requests, veio %d", len(paginasRecebidas))
	}
	if paginasRecebidas[0] != 3 {
		t.Errorf("primeira página deveria ser 3, veio %d", paginasRecebidas[0])
	}
	if paginasRecebidas[1] != 4 {
		t.Errorf("segunda página deveria ser 4, veio %d", paginasRecebidas[1])
	}

	// LastPageInfo deve indicar próxima página = 5
	info, ok := source.LastPageInfo[12345]
	if !ok {
		t.Fatal("LastPageInfo deveria ter entrada para shop 12345")
	}
	if info.NextPage != 5 {
		t.Errorf("NextPage deveria ser 5, veio %d", info.NextPage)
	}
	if !info.HasMore {
		t.Error("HasMore deveria ser true (catálogo não acabou)")
	}
	if info.PagesFetched != 2 {
		t.Errorf("PagesFetched deveria ser 2, veio %d", info.PagesFetched)
	}
}

func TestShopeeShopSourceRotationResetsOnEndOfCatalog(t *testing.T) {
	// Servidor retorna hasNextPage=false para simular fim do catálogo
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"data":{"shopOfferV2":{"nodes":[{"itemId":"1","productName":"P","offerLink":"","priceMin":10,"sales":1,"ratingStar":4,"commissionRate":0.1,"shopName":"S","imageUrl":""}],"pageInfo":{"page":5,"hasNextPage":false}}}}`))
	}))
	defer srv.Close()

	source := NewShopeeShopSource("app", "secret", []int64{99999})
	source.Endpoint = srv.URL
	source.StartPage = 5
	source.MaxPages = 3

	_, err := source.Fetch()
	if err != nil {
		t.Fatal(err)
	}

	info := source.LastPageInfo[99999]
	if info.NextPage != 1 {
		t.Errorf("NextPage deveria resetar para 1 (fim do catálogo), veio %d", info.NextPage)
	}
	if info.HasMore {
		t.Error("HasMore deveria ser false (catálogo acabou)")
	}
}

func TestShopeeShopSourceDefaultStartPage(t *testing.T) {
	// StartPage = 0 deve ser tratado como 1
	paginasRecebidas := []int{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := make([]byte, r.ContentLength)
		r.Body.Read(body)
		var req map[string]string
		json.Unmarshal(body, &req)
		query := req["query"]
		for _, part := range strings.Split(query, ",") {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, "page:") {
				pageStr := strings.TrimSpace(strings.TrimPrefix(part, "page:"))
				page := 0
				for _, c := range pageStr {
					if c >= '0' && c <= '9' {
						page = page*10 + int(c-'0')
					}
				}
				paginasRecebidas = append(paginasRecebidas, page)
			}
		}
		w.Write([]byte(`{"data":{"shopOfferV2":{"nodes":[],"pageInfo":{"page":1,"hasNextPage":false}}}}`))
	}))
	defer srv.Close()

	source := NewShopeeShopSource("app", "secret", []int64{11111})
	source.Endpoint = srv.URL
	// StartPage não definido (0) → deve buscar a partir de 1

	_, err := source.Fetch()
	if err != nil {
		t.Fatal(err)
	}
	if len(paginasRecebidas) == 0 {
		t.Fatal("deveria ter feito pelo menos 1 request")
	}
	if paginasRecebidas[0] != 1 {
		t.Errorf("sem StartPage, deveria começar da 1, veio %d", paginasRecebidas[0])
	}
}

func TestShopeeShopSourceMultipleShopsHaveIndependentPageInfo(t *testing.T) {
	requestCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		body := make([]byte, r.ContentLength)
		r.Body.Read(body)
		// Segunda loja retorna hasNextPage=false
		hasNext := "true"
		if requestCount > 2 {
			hasNext = "false"
		}
		w.Write([]byte(`{"data":{"shopOfferV2":{"nodes":[{"itemId":"1","productName":"P","offerLink":"","priceMin":10,"sales":1,"ratingStar":4,"commissionRate":0.1,"shopName":"S","imageUrl":""}],"pageInfo":{"page":1,"hasNextPage":` + hasNext + `}}}}`))
	}))
	defer srv.Close()

	source := NewShopeeShopSource("app", "secret", []int64{111, 222})
	source.Endpoint = srv.URL
	source.StartPage = 1
	source.MaxPages = 2

	_, err := source.Fetch()
	if err != nil {
		t.Fatal(err)
	}

	// Loja 111: buscou 2 páginas (ambas hasNextPage=true) → NextPage=3
	info111 := source.LastPageInfo[111]
	if info111.NextPage != 3 {
		t.Errorf("loja 111: NextPage deveria ser 3, veio %d", info111.NextPage)
	}
	if !info111.HasMore {
		t.Error("loja 111: HasMore deveria ser true")
	}

	// Loja 222: primeira request retorna hasNextPage=false → NextPage=1 (reset)
	info222 := source.LastPageInfo[222]
	if info222.NextPage != 1 {
		t.Errorf("loja 222: NextPage deveria ser 1 (reset), veio %d", info222.NextPage)
	}
	if info222.HasMore {
		t.Error("loja 222: HasMore deveria ser false")
	}
}

func TestShopeeShopSourcePageDelay(t *testing.T) {
	// Testa que PageDelay funciona sem deadlock (não verifica timing exato)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"data":{"shopOfferV2":{"nodes":[{"itemId":"1","productName":"P","offerLink":"","priceMin":10,"sales":1,"ratingStar":4,"commissionRate":0.1,"shopName":"S","imageUrl":""}],"pageInfo":{"page":1,"hasNextPage":true}}}}`))
	}))
	defer srv.Close()

	source := NewShopeeShopSource("app", "secret", []int64{12345})
	source.Endpoint = srv.URL
	source.StartPage = 1
	source.MaxPages = 3
	source.PageDelay = 1 // 1ns — não trava mas exercita o código

	produtos, err := source.Fetch()
	if err != nil {
		t.Fatal(err)
	}
	if len(produtos) != 3 {
		t.Errorf("esperava 3 produtos (1 por página × 3 páginas), veio %d", len(produtos))
	}
}
