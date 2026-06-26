package source

import "strings"

// inferirOrigem tenta derivar o país de origem a partir dos campos disponíveis
// na API de afiliados da Shopee. Retorna string vazia se não for possível.
//
// Campos candidatos (dependem de o que a API efetivamente retornar):
//   - sellerLocation: localização informada pelo seller (ex.: "KR", "JP", "CN")
//   - shopType: tipo de loja ("mall", "preferred", "overseas")
//
// Se a introspecção revelar campos adicionais no futuro, este é o ponto de extensão.
func inferirOrigem(sellerLocation, shopType string) string {
	loc := strings.TrimSpace(strings.ToLower(sellerLocation))
	if loc == "" {
		return ""
	}

	// Mapeamento de códigos ISO ou nomes conhecidos para nomes legíveis em PT-BR
	mapa := map[string]string{
		"kr":          "Coreia",
		"kor":         "Coreia",
		"korea":       "Coreia",
		"south korea": "Coreia",
		"coreia":      "Coreia",
		"coréia":      "Coreia",
		"jp":          "Japão",
		"jpn":         "Japão",
		"japan":       "Japão",
		"japão":       "Japão",
		"japao":       "Japão",
		"cn":          "China",
		"chn":         "China",
		"china":       "China",
		"br":          "Brasil",
		"bra":         "Brasil",
		"brazil":      "Brasil",
		"brasil":      "Brasil",
		"us":          "EUA",
		"usa":         "EUA",
		"tw":          "Taiwan",
		"taiwan":      "Taiwan",
	}

	if nome, ok := mapa[loc]; ok {
		return nome
	}

	// Se não está no mapa mas tem conteúdo, retorna capitalizado
	if len(loc) > 0 {
		return strings.ToUpper(loc[:1]) + loc[1:]
	}
	return ""
}

// NormalizarOrigem converte variações de nome de país para o formato padrão
// usado nos badges. Usado pelo frontend e pelo filtro.
func NormalizarOrigem(origin string) string {
	if origin == "" {
		return ""
	}
	return inferirOrigem(origin, "")
}
