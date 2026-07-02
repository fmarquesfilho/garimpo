// Package couponsource define a porta de coleta de cupons (ports & adapters).
// Segue o mesmo padrão do internal/source — interface + factory + registry.
package couponsource

import "github.com/fmarquesfilho/garimpo/internal/domain"

// CouponSource is the port for coupon collection. Each marketplace implements this.
type CouponSource interface {
	// FetchCoupons retrieves available coupons for the given tenant.
	FetchCoupons(cfg FetchConfig) ([]domain.Coupon, error)

	// Marketplace returns the marketplace identifier.
	Marketplace() string

	// Name returns a descriptive name for logging.
	Name() string
}

// FetchConfig holds per-request parameters for coupon fetching.
type FetchConfig struct {
	OwnerUID string
	PageSize int // max per page (500 for Shopee, 100 for Amazon)
}

// CouponSourceFactory creates a CouponSource with the given credentials.
type CouponSourceFactory func(cfg SourceConfig) CouponSource

// SourceConfig agrupa credenciais necessárias para criar uma fonte de cupons.
type SourceConfig struct {
	// Shopee
	AppID  string
	Secret string

	// Amazon (OAuth 2.0)
	AccessKey  string
	SecretKey  string
	PartnerTag string

	// Mercado Livre (OAuth 2.0)
	ClientID     string
	ClientSecret string
	AccessToken  string
	RefreshToken string
}
