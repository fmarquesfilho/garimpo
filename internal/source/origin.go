package source

import "strings"

// inferirOrigemDeShopType tenta inferir se a loja é cross-border a partir do
// campo shopType da API. O shopType é um enum com valores como "mall",
// "preferred", "overseas". Se contiver "overseas" ou "cb" (cross-border),
// retorna "Importado" — mas NÃO especifica o país. Para país específico,
// o fallback origem_padrao da Busca é necessário.
//
// Retorna string vazia se shopType não indicar nada útil sobre origem.
func inferirOrigemDeShopType(shopTypes []string) string {
	for _, st := range shopTypes {
		lower := strings.ToLower(strings.TrimSpace(st))
		if lower == "overseas" || lower == "cb" || lower == "cross_border" || lower == "crossborder" {
			return "Importado"
		}
	}
	return ""
}

// NormalizarOrigem converte variações de nome de país para o formato padrão
// usado nos badges do frontend.
func NormalizarOrigem(origin string) string {
	if origin == "" {
		return ""
	}
	loc := strings.TrimSpace(strings.ToLower(origin))

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
		"importado":   "Importado",
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
