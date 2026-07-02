package source

import (
	collectorpb "github.com/fmarquesfilho/garimpo/gen/go/collector/v1"
	"github.com/fmarquesfilho/garimpo/internal/domain"
)

// ToProtoProduct converte um domain.Product em collectorpb.Product.
// Centraliza o mapeamento que antes era duplicado em cada collector server.
func ToProtoProduct(p domain.Product) *collectorpb.Product {
	return &collectorpb.Product{
		ItemId:          ParseNumericID(p.ID),
		ShopId:          ParseNumericID(p.ShopID),
		Name:            p.Name,
		Price:           p.Price,
		OriginalPrice:   p.PriceMax,
		Sold:            int32(min(p.Sales30d, int(^uint32(0)>>1))), //nolint:gosec // bounded
		Rating:          p.Rating,
		ImageUrl:        p.Image,
		ProductUrl:      p.ProductLink,
		ShopName:        p.ShopName,
		DiscountPercent: p.DiscountRate,
		Commission:      p.Commission,
		Category:        p.Category,
		Link:            p.Link,
		Marketplace:     MarketplaceToProto(p.Marketplace),
	}
}

// ToProtoProducts converte uma slice de domain.Product em slice de proto Product.
func ToProtoProducts(products []domain.Product) []*collectorpb.Product {
	result := make([]*collectorpb.Product, 0, len(products))
	for _, p := range products {
		result = append(result, ToProtoProduct(p))
	}
	return result
}

// ParseNumericID extrai um int64 de uma string numérica.
// Seguro para IDs da Shopee; para ASINs (alfanuméricos), retorna 0.
func ParseNumericID(s string) int64 {
	var id int64
	for _, c := range s {
		if c >= '0' && c <= '9' {
			id = id*10 + int64(c-'0')
		}
	}
	return id
}

// MarketplaceToProto converte a string de marketplace para o enum proto.
func MarketplaceToProto(m string) collectorpb.Marketplace {
	switch m {
	case domain.MarketplaceShopee:
		return collectorpb.Marketplace_MARKETPLACE_SHOPEE
	case domain.MarketplaceAmazon:
		return collectorpb.Marketplace_MARKETPLACE_AMAZON
	case domain.MarketplaceMercadoLivre:
		return collectorpb.Marketplace_MARKETPLACE_MERCADOLIVRE
	default:
		return collectorpb.Marketplace_MARKETPLACE_UNSPECIFIED
	}
}

// ProtoToMarketplace converte o enum proto para a string de marketplace.
func ProtoToMarketplace(m collectorpb.Marketplace) string {
	switch m {
	case collectorpb.Marketplace_MARKETPLACE_SHOPEE:
		return domain.MarketplaceShopee
	case collectorpb.Marketplace_MARKETPLACE_AMAZON:
		return domain.MarketplaceAmazon
	case collectorpb.Marketplace_MARKETPLACE_MERCADOLIVRE:
		return domain.MarketplaceMercadoLivre
	default:
		return domain.MarketplaceShopee // default retrocompat
	}
}
