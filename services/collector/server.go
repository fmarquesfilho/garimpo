package main

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	collectorpb "github.com/fmarquesfilho/garimpo/gen/go/collector/v1"
	"github.com/fmarquesfilho/garimpo/internal/source"
)

// CollectorServer implementa collector.v1.CollectorService.
type CollectorServer struct {
	collectorpb.UnimplementedCollectorServiceServer
	appID  string
	secret string
}

func NewCollectorServer(appID, secret string) *CollectorServer {
	return &CollectorServer{appID: appID, secret: secret}
}

func (s *CollectorServer) Fetch(ctx context.Context, req *collectorpb.FetchRequest) (*collectorpb.FetchResponse, error) {
	if req.GetKeyword() == "" {
		return nil, status.Error(codes.InvalidArgument, "keyword é obrigatório") //nolint:wrapcheck // gRPC status
	}

	src := source.NewShopeeAPISource(s.appID, s.secret)
	src.Keyword = req.GetKeyword()
	if req.GetLimit() > 0 {
		src.Limit = int(req.GetLimit())
	}

	produtos, err := src.Fetch()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "falha ao buscar: %v", err)
	}

	resp := &collectorpb.FetchResponse{
		TotalFound: int32(min(len(produtos), int(^uint32(0)>>1))), //nolint:gosec // bounded by Shopee API limits
		FetchedAt:  time.Now().UTC().Format(time.RFC3339),
	}

	for _, p := range produtos {
		resp.Products = append(resp.Products, &collectorpb.Product{
			ItemId:          parseItemID(p.ID),
			ShopId:          parseItemID(p.ShopID),
			Name:            p.Name,
			Price:           p.Price,
			OriginalPrice:   p.PriceMax,
			Sold:            int32(min(p.Sales30d, int(^uint32(0)>>1))), //nolint:gosec // bounded
			Rating:          p.Rating,
			ImageUrl:        p.Image,
			ProductUrl:      p.Link,
			ShopName:        p.ShopName,
			DiscountPercent: p.DiscountRate,
		})
	}

	return resp, nil
}

func (s *CollectorServer) FetchShop(ctx context.Context, req *collectorpb.FetchShopRequest) (*collectorpb.FetchShopResponse, error) {
	if req.GetShopId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "shop_id é obrigatório") //nolint:wrapcheck // gRPC status
	}

	src := source.NewShopeeShopSource(s.appID, s.secret, []int64{req.GetShopId()})
	if req.GetLimit() > 0 {
		src.Limit = int(req.GetLimit())
	}
	src.PageDelay = 200 * time.Millisecond

	produtos, err := src.Fetch()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "falha ao buscar shop: %v", err)
	}

	resp := &collectorpb.FetchShopResponse{
		TotalFound: int32(min(len(produtos), int(^uint32(0)>>1))), //nolint:gosec // bounded
		FetchedAt:  time.Now().UTC().Format(time.RFC3339),
	}

	for _, p := range produtos {
		resp.Products = append(resp.Products, &collectorpb.Product{
			ItemId:          parseItemID(p.ID),
			ShopId:          parseItemID(p.ShopID),
			Name:            p.Name,
			Price:           p.Price,
			OriginalPrice:   p.PriceMax,
			Sold:            int32(min(p.Sales30d, int(^uint32(0)>>1))), //nolint:gosec // bounded
			Rating:          p.Rating,
			ImageUrl:        p.Image,
			ProductUrl:      p.Link,
			ShopName:        p.ShopName,
			DiscountPercent: p.DiscountRate,
		})
	}

	return resp, nil
}

func parseItemID(s string) int64 {
	var id int64
	for _, c := range s {
		if c >= '0' && c <= '9' {
			id = id*10 + int64(c-'0')
		}
	}
	return id
}
