// Comando shopee-introspect: descobre campos disponíveis na API GraphQL de Afiliados da Shopee.
// Uso:
//
//	export SHOPEE_APP_ID=... SHOPEE_SECRET=...
//	go run ./cmd/shopee-introspect
//
// Envia uma query de introspecção (__type) para descobrir os campos do tipo
// retornado por productOfferV2. Útil para encontrar campos de origem (origin,
// country, shopType, etc.) que não estão sendo usados no adaptador atual.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const endpoint = "https://open-api.affiliate.shopee.com.br/graphql"

func main() {
	appID := os.Getenv("SHOPEE_APP_ID")
	secret := os.Getenv("SHOPEE_SECRET")
	if appID == "" || secret == "" {
		log.Fatal("SHOPEE_APP_ID e SHOPEE_SECRET devem estar definidos no ambiente")
	}

	// Termos relacionados a origem que nos interessam
	termosOrigem := []string{
		"origin", "country", "shop_type", "shoptype", "location",
		"brand", "seller", "warehouse", "domestic", "local",
		"imported", "cross_border", "crossborder",
	}

	// Query de introspecção: tenta descobrir o tipo do nó productOfferV2
	// A API de afiliados pode não suportar introspecção completa (__schema),
	// mas podemos testar com __type ou pedir campos extras diretamente.
	queries := []struct {
		nome  string
		query string
	}{
		{
			nome: "Introspection __schema (types)",
			query: `{ __schema { types { name kind fields { name type { name kind } } } } }`,
		},
		{
			nome: "Introspection __type(ProductOfferNode)",
			query: `{ __type(name: "ProductOfferNode") { name fields { name type { name kind ofType { name kind } } } } }`,
		},
		{
			nome: "Introspection __type(ProductOffer)",
			query: `{ __type(name: "ProductOffer") { name fields { name type { name kind ofType { name kind } } } } }`,
		},
		{
			// Tenta buscar um produto pedindo campos extras que podem existir
			nome: "productOfferV2 com campos extras",
			query: `{ productOfferV2(listType: 1, sortType: 5, limit: 1) { nodes { itemId productName shopName shopType shopId brandName sellerLocation originCountry productOrigin imageUrl priceMin sales ratingStar commissionRate productCatIds offerLink } pageInfo { page hasNextPage } } }`,
		},
	}

	fmt.Println("=== Shopee Affiliate API — Introspecção de Campos ===")
	fmt.Printf("AppID: %s\n", appID)
	fmt.Printf("Endpoint: %s\n\n", endpoint)

	for _, q := range queries {
		fmt.Printf("--- %s ---\n", q.nome)
		resultado, err := executarQuery(appID, secret, q.query)
		if err != nil {
			fmt.Printf("ERRO: %v\n\n", err)
			continue
		}

		// Pretty-print
		var pretty bytes.Buffer
		json.Indent(&pretty, resultado, "", "  ")
		fmt.Println(pretty.String())

		// Se for introspecção de tipo, lista campos relacionados a origem
		var resposta map[string]any
		json.Unmarshal(resultado, &resposta)
		camposEncontrados := extrairCamposOrigem(resultado, termosOrigem)
		if len(camposEncontrados) > 0 {
			fmt.Println("\n🎯 CAMPOS RELACIONADOS A ORIGEM ENCONTRADOS:")
			for _, c := range camposEncontrados {
				fmt.Printf("  → %s\n", c)
			}
		}
		fmt.Println()
	}
}

func executarQuery(appID, secret, query string) ([]byte, error) {
	body, _ := json.Marshal(map[string]string{"query": query})

	ts := strconv.FormatInt(time.Now().Unix(), 10)
	sum := sha256.Sum256([]byte(appID + ts + string(body) + secret))
	sig := hex.EncodeToString(sum[:])

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
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
		return nil, err
	}
	return raw, nil
}

func extrairCamposOrigem(dados []byte, termos []string) []string {
	// Busca no JSON bruto por campos cujo nome contenha termos de origem
	var encontrados []string
	text := strings.ToLower(string(dados))
	for _, t := range termos {
		if strings.Contains(text, t) {
			encontrados = append(encontrados, t)
		}
	}
	return encontrados
}
