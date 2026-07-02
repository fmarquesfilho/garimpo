package source

import (
	"fmt"

	"github.com/fmarquesfilho/garimpo/internal/apperr"
	"github.com/fmarquesfilho/garimpo/internal/domain"
)

// ShopeeAdapter implementa ProductSource para o marketplace Shopee.
// Encapsula ShopeeAPISource e ShopeeShopSource, expondo a interface uniforme.
type ShopeeAdapter struct {
	appID  string
	secret string
}

func NewShopeeAdapter(appID, secret string) *ShopeeAdapter {
	return &ShopeeAdapter{appID: appID, secret: secret}
}

func (a *ShopeeAdapter) Marketplace() string { return domain.MarketplaceShopee }
func (a *ShopeeAdapter) Name() string        { return "shopee-adapter" }

func (a *ShopeeAdapter) Search(q SearchQuery) ([]domain.Product, error) {
	src := NewShopeeAPISource(a.appID, a.secret)
	src.Keyword = q.Keyword
	if q.Limit > 0 {
		src.Limit = q.Limit
	}
	return src.Fetch()
}

func (a *ShopeeAdapter) FetchShop(shopID string, limit int) ([]domain.Product, error) {
	id := parseShopID(shopID)
	if id == 0 {
		return nil, fmt.Errorf("shopID %q: %w", shopID, apperr.ErrInvalidInput)
	}
	src := NewShopeeShopSource(a.appID, a.secret, []int64{id})
	if limit > 0 {
		src.Limit = limit
	}
	return src.Fetch()
}

func parseShopID(s string) int64 {
	var id int64
	for _, c := range s {
		if c >= '0' && c <= '9' {
			id = id*10 + int64(c-'0')
		}
	}
	return id
}
