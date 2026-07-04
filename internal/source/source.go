// Package source define a PORTA de entrada de candidatos (ports & adapters).
// O motor de scoring não sabe se os dados vêm da Shopee, Amazon ou de um mock
// de teste — ele só conhece a interface ProductSource.
package source

import "github.com/fmarquesfilho/garimpo/internal/domain"

// SearchQuery encapsula os parâmetros de busca de forma genérica.
// Cada marketplace usa os campos que fazem sentido para sua API.
type SearchQuery struct {
	Keyword string // termo de busca
	Limit   int    // quantidade máxima de resultados
	SortBy  string // critério de ordenação (relevance, sales, price)
	ShopID  string // ID da loja (quando aplicável)
}

// ProductSource é a porta: qualquer fonte de candidatos a implementa.
// Trocar de fonte (Shopee -> Amazon) NÃO altera nenhuma outra parte do sistema.
type ProductSource interface {
	// Search busca produtos por keyword com os parâmetros fornecidos.
	Search(q SearchQuery) ([]domain.Product, error)

	// FetchShop busca produtos de uma loja específica.
	// Retorna ErrNotSupported se o marketplace não suporta busca por loja.
	FetchShop(shopID string, limit int) ([]domain.Product, error)

	// Marketplace retorna o identificador do marketplace desta fonte.
	Marketplace() string

	// Name retorna o nome descritivo da fonte (para logs/debug).
	Name() string
}

// SourceFactory cria uma instância de ProductSource com as credenciais fornecidas.
// Cada marketplace registra sua factory no Registry.
type SourceFactory func(cfg SourceConfig) ProductSource

// SourceConfig agrupa credenciais e configuração necessárias para criar uma fonte.
// Campos opcionais — cada marketplace usa o que precisa.
type SourceConfig struct {
	// Shopee affiliate API credentials
	AppID  string
	Secret string

	// Amazon Product Advertising API credentials
	AccessKey  string
	SecretKey  string
	PartnerTag string

	// Mercado Livre (futuro)
	ClientID     string
	ClientSecret string
	AccessToken  string
}
