package main

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	collectorpb "github.com/fmarquesfilho/garimpo/gen/go/collector/v1"
	"github.com/fmarquesfilho/garimpo/internal/source"
)

// AmazonCollectorServer implementa collector.v1.CollectorService para Amazon.
// Reutiliza exatamente a mesma lógica do collector genérico — a única diferença
// é a ProductSource injetada (AmazonAdapter em vez de ShopeeAdapter).
type AmazonCollectorServer struct {
	collectorpb.UnimplementedCollectorServiceServer
	source source.ProductSource
}

func NewAmazonCollectorServer(src source.ProductSource) *AmazonCollectorServer {
	return &AmazonCollectorServer{source: src}
}

func (s *AmazonCollectorServer) Fetch(ctx context.Context, req *collectorpb.FetchRequest) (*collectorpb.FetchResponse, error) {
	if req.GetKeyword() == "" {
		return nil, status.Error(codes.InvalidArgument, "keyword é obrigatório")
	}

	produtos, err := s.source.Search(source.SearchQuery{
		Keyword: req.GetKeyword(),
		Limit:   int(req.GetLimit()),
		SortBy:  req.GetSortBy(),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "falha ao buscar amazon: %v", err)
	}

	return &collectorpb.FetchResponse{
		Products:   source.ToProtoProducts(produtos),
		TotalFound: source.SafeInt32(len(produtos)),
		FetchedAt:  time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func (s *AmazonCollectorServer) FetchShop(_ context.Context, _ *collectorpb.FetchShopRequest) (*collectorpb.FetchShopResponse, error) {
	return nil, status.Error(codes.Unimplemented, "FetchShop não é suportado para Amazon")
}
