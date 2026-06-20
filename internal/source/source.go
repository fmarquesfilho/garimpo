// Package source define a PORTA de entrada de candidatos (ports & adapters).
// O motor de scoring não sabe se os dados vêm de CSV, da API da Shopee ou de
// um mock de teste — ele só conhece a interface ProductSource.
package source

import "github.com/fmarquesfilho/garimpo/internal/domain"

// ProductSource é a porta: qualquer fonte de candidatos a implementa.
// Trocar de fonte (CSV -> API) NÃO altera nenhuma outra parte do sistema.
type ProductSource interface {
	Fetch() ([]domain.Product, error)
	Name() string
}
