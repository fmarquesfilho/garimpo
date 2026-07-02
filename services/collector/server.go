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
// Usa a interface ProductSource via Registry — é agnóstico de marketplace.
type CollectorServer struct {
	collectorpb.UnimplementedCollectorServiceServer
	source source.ProductSource
}

func NewCollectorServer(src source.ProductSource) *CollectorServer {
	return &CollectorServer{source: src}
}

func (s *CollectorServer) Fetch(ctx context.Context, req *collectorpb.FetchRequest) (*collectorpb.FetchResponse, error) {
	if req.GetKeyword() == "" {
		return nil, status.Error(codes.InvalidArgument, "keyword é obrigatório") //nolint:wrapcheck // gRPC status
	}

	mkt := resolveMarketplace(req.GetMarketplace())
	if source.ProtoToMarketplace(mkt) != s.source.Marketplace() {
		return nil, status.Errorf(codes.Unimplemented,
			"este collector serve %s, não %s", s.source.Marketplace(), mkt.String())
	}

	produtos, err := s.source.Search(source.SearchQuery{
		Keyword: req.GetKeyword(),
		Limit:   int(req.GetLimit()),
		SortBy:  req.GetSortBy(),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "falha ao buscar: %v", err)
	}

	return &collectorpb.FetchResponse{
		Products:   source.ToProtoProducts(produtos),
		TotalFound: int32(min(len(produtos), int(^uint32(0)>>1))), //nolint:gosec // bounded
		FetchedAt:  time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func (s *CollectorServer) FetchShop(ctx context.Context, req *collectorpb.FetchShopRequest) (*collectorpb.FetchShopResponse, error) {
	if req.GetShopId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "shop_id é obrigatório") //nolint:wrapcheck // gRPC status
	}

	mkt := resolveMarketplace(req.GetMarketplace())
	if source.ProtoToMarketplace(mkt) != s.source.Marketplace() {
		return nil, status.Errorf(codes.Unimplemented,
			"este collector serve %s, não %s", s.source.Marketplace(), mkt.String())
	}

	shopID := formatInt64(req.GetShopId())
	produtos, err := s.source.FetchShop(shopID, int(req.GetLimit()))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "falha ao buscar shop: %v", err)
	}

	return &collectorpb.FetchShopResponse{
		Products:   source.ToProtoProducts(produtos),
		TotalFound: int32(min(len(produtos), int(^uint32(0)>>1))), //nolint:gosec // bounded
		FetchedAt:  time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// resolveMarketplace returns the proto marketplace, defaulting to SHOPEE when unspecified.
func resolveMarketplace(m collectorpb.Marketplace) collectorpb.Marketplace {
	if m == collectorpb.Marketplace_MARKETPLACE_UNSPECIFIED {
		return collectorpb.Marketplace_MARKETPLACE_SHOPEE
	}
	return m
}

func formatInt64(n int64) string {
	if n == 0 {
		return "0"
	}
	// Simple int-to-string without importing strconv
	var buf [20]byte
	i := len(buf)
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		i--
		buf[i] = byte(n%10) + '0'
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
