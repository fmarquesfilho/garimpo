package source

import (
	"fmt"

	"github.com/fmarquesfilho/garimpo/internal/domain"
)

// AmazonAdapter implementa ProductSource para o marketplace Amazon.
// Encapsula AmazonCreatorsSource, expondo a interface uniforme.
type AmazonAdapter struct {
	accessKey  string
	secretKey  string
	partnerTag string
}

func NewAmazonAdapter(accessKey, secretKey, partnerTag string) *AmazonAdapter {
	return &AmazonAdapter{
		accessKey:  accessKey,
		secretKey:  secretKey,
		partnerTag: partnerTag,
	}
}

func (a *AmazonAdapter) Marketplace() string { return domain.MarketplaceAmazon }
func (a *AmazonAdapter) Name() string        { return "amazon-adapter" }

func (a *AmazonAdapter) Search(q SearchQuery) ([]domain.Product, error) {
	src := NewAmazonCreatorsSource(a.accessKey, a.secretKey, a.partnerTag)
	src.Keyword = q.Keyword
	if q.Limit > 0 {
		src.Limit = q.Limit
	}
	return src.Fetch()
}

func (a *AmazonAdapter) FetchShop(_ string, _ int) ([]domain.Product, error) {
	// Amazon Creators API não suporta busca por loja.
	return nil, fmt.Errorf("FetchShop não suportado para Amazon: %w", ErrUnsupportedMarketplace)
}
