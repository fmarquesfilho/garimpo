package source

// CategoriaShopee mapeia IDs de categoria nível 1 da Shopee Brasil para nomes legíveis.
// Fonte: mapeamento empírico via productCatIds retornados pela API de afiliados.
// Atualizar se novos IDs aparecerem nos dados coletados.
var CategoriaShopee = map[int]string{
	100001: "Alimentos",
	100009: "Celulares",
	100011: "Roupas Femininas",
	100012: "Calçados",
	100013: "Acessórios Celular",
	100017: "Roupas Masculinas",
	100535: "Celulares & Tablets",
	100630: "Beleza",
	100631: "Saúde & Bem-estar",
	100632: "Brinquedos & Bebês",
	100633: "Acessórios & Bolsas",
	100636: "Casa & Decoração",
	100637: "Moda",
	100640: "Perfumaria",
	100643: "Papelaria & Livros",
	100644: "Áudio & Eletrônicos",
	100658: "Manicure & Pedicure",
	100659: "Cuidados com o Cabelo",
	100663: "Maquiagem",
	100664: "Cuidados com a Pele",
}

// NomeCategoria retorna o nome legível para um ID de categoria, ou "" se desconhecido.
func NomeCategoria(catID int) string {
	return CategoriaShopee[catID]
}

// NomeCategoriaPrincipal retorna o nome da primeira categoria (nível 1) da lista.
func NomeCategoriaPrincipal(catIDs []int) string {
	for _, id := range catIDs {
		if nome := CategoriaShopee[id]; nome != "" {
			return nome
		}
	}
	return ""
}
