// Package domain contém o núcleo do problema: o produto candidato e seu
// resultado pontuado. Não depende de nada externo (nem fonte, nem estratégia).
package domain

// Product é um candidato a ser anunciado. É a unidade que flui pelo sistema.
type Product struct {
	ID         string
	Name       string
	Category   string
	Price      float64 // em BRL
	Commission float64 // taxa de comissão, 0..1 (ex.: 0.12 = 12%)
	Sales30d   int     // proxy de demanda (vendas; ver nota no adaptador Shopee)
	Rating     float64 // avaliação média, 0..5
	Link       string  // link de afiliado (offerLink), quando a fonte fornece
	Image      string  // URL da imagem principal do produto (quando disponível)
	ShopName   string  // nome da loja do produto (quando disponível)
	ShopID     string  // ID da loja na Shopee (para enriquecimento de origem)
	CatIDs     []int   // IDs de categorias Shopee (hierárquicos: nível 1, 2, 3)
	Origin     string  // país de origem do produto (ex.: "Coreia", "Japão") — preenchido via API ou fallback por loja
}

// Scored é um produto já pontuado por uma estratégia.
// Reasons guarda a contribuição de cada componente do score — isso dá
// EXPLICABILIDADE: dá para ver POR QUE um produto subiu no ranking.
type Scored struct {
	Product Product
	Score   float64
	Reasons map[string]float64

	// Suspeito sinaliza "produto-fantasma": comissão alta no pool mas sem
	// tração (zero vendas) ou sem credibilidade (nota zero). Não é eliminado —
	// é marcado, para ela decidir com a informação à vista.
	Suspeito bool

	// Exploracao marca um candidato sorteado fora do topo (hold-out de
	// exploração), para gerar dados não-enviesados sobre o que mais converte.
	Exploracao bool
}
