package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	collectorpb "github.com/fmarquesfilho/garimpo/gen/go/collector/v1"
	couponpb "github.com/fmarquesfilho/garimpo/gen/go/coupon/v1"
	"github.com/fmarquesfilho/garimpo/internal/couponsource"
	"github.com/fmarquesfilho/garimpo/internal/domain"
	"github.com/fmarquesfilho/garimpo/internal/source"
)

// CouponCollectorServer implements coupon.v1.CouponCollectorService.
type CouponCollectorServer struct {
	couponpb.UnimplementedCouponCollectorServiceServer
	src    couponsource.CouponSource
	logger *slog.Logger
}

func NewCouponCollectorServer(src couponsource.CouponSource, logger *slog.Logger) *CouponCollectorServer {
	return &CouponCollectorServer{src: src, logger: logger}
}

func (s *CouponCollectorServer) FetchCoupons(ctx context.Context, req *couponpb.FetchCouponsRequest) (*couponpb.FetchCouponsResponse, error) {
	mkt := source.ProtoToMarketplace(req.GetMarketplace())
	if mkt != s.src.Marketplace() {
		return nil, status.Errorf(codes.Unimplemented,
			"este coupon-collector serve %s, não %s", s.src.Marketplace(), mkt)
	}

	pageSize := int(req.GetPageSize())
	if pageSize <= 0 {
		pageSize = 500
	}

	coupons, err := s.src.FetchCoupons(couponsource.FetchConfig{
		OwnerUID: req.GetOwnerUid(),
		PageSize: pageSize,
	})
	if err != nil {
		s.logger.Error("coupon fetch failed",
			slog.String("marketplace", s.src.Marketplace()),
			slog.String("owner_uid", req.GetOwnerUid()),
			slog.String("error", err.Error()))
		return nil, status.Errorf(codes.Internal, "falha ao buscar cupons: %v", err)
	}

	s.logger.Info("coupons collected",
		slog.String("marketplace", s.src.Marketplace()),
		slog.Int("count", len(coupons)))

	return &couponpb.FetchCouponsResponse{
		Coupons:    toProtoCoupons(coupons),
		TotalFound: source.SafeInt32(len(coupons)),
		FetchedAt:  time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func toProtoCoupons(coupons []domain.Coupon) []*couponpb.CouponProto {
	result := make([]*couponpb.CouponProto, 0, len(coupons))
	for _, c := range coupons {
		var dt couponpb.DiscountType
		switch c.DiscountType {
		case domain.DiscountTypePercentage:
			dt = couponpb.DiscountType_DISCOUNT_TYPE_PERCENTAGE
		case domain.DiscountTypeFixed:
			dt = couponpb.DiscountType_DISCOUNT_TYPE_FIXED
		}

		var mkt collectorpb.Marketplace
		switch c.Marketplace {
		case domain.MarketplaceShopee:
			mkt = collectorpb.Marketplace_MARKETPLACE_SHOPEE
		case domain.MarketplaceAmazon:
			mkt = collectorpb.Marketplace_MARKETPLACE_AMAZON
		case domain.MarketplaceMercadoLivre:
			mkt = collectorpb.Marketplace_MARKETPLACE_MERCADOLIVRE
		}

		cats, _ := json.Marshal(c.ApplicableCategories)

		result = append(result, &couponpb.CouponProto{
			CouponId:             c.ID,
			Marketplace:          mkt,
			Code:                 c.Code,
			DiscountType:         dt,
			DiscountValue:        c.DiscountValue,
			MinSpend:             c.MinSpend,
			StartTime:            formatUnix(c.StartTime),
			EndTime:              formatUnix(c.EndTime),
			ApplicableCategories: c.ApplicableCategories,
			Status:               c.Status,
		})
		_ = cats // categories also available as JSON if needed
	}
	return result
}

func formatUnix(ts int64) string {
	if ts == 0 {
		return ""
	}
	return time.Unix(ts, 0).UTC().Format(time.RFC3339)
}
