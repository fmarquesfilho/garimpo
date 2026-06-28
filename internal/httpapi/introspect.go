package httpapi

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// adminShopeeIntrospect faz introspecção do schema GraphQL da API de afiliados
// e retorna os campos disponíveis. Restrito a admin.
// GET /api/admin/shopee-introspect
func (srv *Server) adminShopeeIntrospect(w http.ResponseWriter, r *http.Request) {

	appID := os.Getenv("SHOPEE_APP_ID")
	secret := os.Getenv("SHOPEE_SECRET")
	if appID == "" || secret == "" {
		writeErr(w, http.StatusServiceUnavailable, "SHOPEE_APP_ID/SHOPEE_SECRET não configurados no ambiente")
		return
	}

	// Termos que indicam campos de origem
	termosOrigem := []string{
		"origin", "country", "shop_type", "shoptype", "location",
		"brand", "seller", "warehouse", "domestic", "local",
		"imported", "cross_border", "crossborder",
	}

	// Queries de introspecção
	queries := []struct {
		Nome  string `json:"nome"`
		Query string `json:"query"`
	}{
		{
			Nome:  "Introspection __schema (todos os tipos)",
			Query: `{ __schema { types { name kind fields { name type { name kind } } } } }`,
		},
		{
			Nome:  "Introspection __type(ProductOfferNode)",
			Query: `{ __type(name: "ProductOfferNode") { name fields { name type { name kind ofType { name kind } } } } }`,
		},
		{
			Nome:  "Introspection __type(ProductOffer)",
			Query: `{ __type(name: "ProductOffer") { name fields { name type { name kind ofType { name kind } } } } }`,
		},
		{
			Nome:  "productOfferV2 com campos extras (teste)",
			Query: `{ productOfferV2(listType: 1, sortType: 5, limit: 1) { nodes { itemId productName shopName shopType brandName sellerLocation originCountry productOrigin imageUrl priceMin sales ratingStar commissionRate productCatIds offerLink } pageInfo { page hasNextPage } } }`,
		},
	}

	type resultado struct {
		Nome         string   `json:"nome"`
		Sucesso      bool     `json:"sucesso"`
		Erro         string   `json:"erro,omitempty"`
		Resposta     any      `json:"resposta,omitempty"`
		CamposOrigem []string `json:"campos_origem,omitempty"`
	}

	var resultados []resultado

	for _, q := range queries {
		raw, err := shopeeGraphQL(appID, secret, q.Query)
		r := resultado{Nome: q.Nome}

		if err != nil {
			r.Erro = err.Error()
			resultados = append(resultados, r)
			continue
		}

		var parsed any
		_ = json.Unmarshal(raw, &parsed)
		r.Resposta = parsed
		r.Sucesso = true

		// Procura termos de origem no response
		text := strings.ToLower(string(raw))
		for _, t := range termosOrigem {
			if strings.Contains(text, t) {
				r.CamposOrigem = append(r.CamposOrigem, t)
			}
		}

		resultados = append(resultados, r)
	}

	// Resumo: algum campo de origem foi encontrado?
	var todosOrigem []string
	for _, r := range resultados {
		todosOrigem = append(todosOrigem, r.CamposOrigem...)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"app_id":               appID,
		"endpoint":             "https://open-api.affiliate.shopee.com.br/graphql",
		"resultados":           resultados,
		"campos_origem_global": todosOrigem,
		"conclusao": func() string {
			if len(todosOrigem) > 0 {
				return fmt.Sprintf("Encontrados %d termos relacionados a origem. Verifique se são campos reais no schema.", len(todosOrigem))
			}
			return "Nenhum campo de origem encontrado. Use o fallback (origem_padrao por loja)."
		}(),
	})
}

// shopeeGraphQL executa uma query GraphQL autenticada na API de afiliados.
func shopeeGraphQL(appID, secret, query string) ([]byte, error) {
	body, _ := json.Marshal(map[string]string{"query": query})

	ts := strconv.FormatInt(time.Now().Unix(), 10)
	sum := sha256.Sum256([]byte(appID + ts + string(body) + secret))
	sig := hex.EncodeToString(sum[:])

	endpoint := "https://open-api.affiliate.shopee.com.br/graphql"
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("introspect criar request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization",
		fmt.Sprintf("SHA256 Credential=%s, Timestamp=%s, Signature=%s", appID, ts, sig))

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("requisição falhou: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("introspect ler resposta: %w", err)
	}
	return raw, nil
}
