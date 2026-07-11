package main

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	collectorpb "github.com/fmarquesfilho/garimpo/gen/go/collector/v1"
	couponpb "github.com/fmarquesfilho/garimpo/gen/go/coupon/v1"
	"github.com/fmarquesfilho/garimpo/internal/couponsource"
	"github.com/fmarquesfilho/garimpo/internal/domain"
	"github.com/fmarquesfilho/garimpo/internal/source"
	"github.com/fmarquesfilho/garimpo/internal/store"
)

// ─── Collector (produtos) ────────────────────────────────────────────────────

// UnifiedCollectorServer implementa collector.v1.CollectorService.
// Despacha para o ProductSource correto via Pipeline, baseado no marketplace do request.
type UnifiedCollectorServer struct {
	collectorpb.UnimplementedCollectorServiceServer
	pipeline  *Pipeline
	snapshots store.SnapshotRepo
	exportCh  chan store.Snapshot
	logger    *slog.Logger
}

func NewUnifiedCollectorServer(pipeline *Pipeline, snapshots store.SnapshotRepo, logger *slog.Logger) *UnifiedCollectorServer {
	return &UnifiedCollectorServer{
		pipeline:  pipeline,
		snapshots: snapshots,
		exportCh:  make(chan store.Snapshot, 64),
		logger:    logger,
	}
}

func (s *UnifiedCollectorServer) Fetch(ctx context.Context, req *collectorpb.FetchRequest) (*collectorpb.FetchResponse, error) {
	if req.GetKeyword() == "" {
		return nil, status.Error(codes.InvalidArgument, "keyword é obrigatório")
	}

	mkt := resolveMarketplace(req.GetMarketplace())
	marketplace := source.ProtoToMarketplace(mkt)

	src, ok := s.pipeline.GetProductSourceByMarketplace(marketplace)
	if !ok {
		return nil, status.Errorf(codes.Unimplemented,
			"nenhum receiver de produto configurado para marketplace %q", marketplace)
	}

	produtos, err := src.Search(source.SearchQuery{
		Keyword: req.GetKeyword(),
		Limit:   int(req.GetLimit()),
		SortBy:  req.GetSortBy(),
	})
	if err != nil {
		s.logger.Error("fetch falhou",
			slog.String("marketplace", marketplace),
			slog.String("keyword", req.GetKeyword()),
			slog.String("error", err.Error()))
		return nil, status.Errorf(codes.Internal, "falha ao buscar: %v", err)
	}

	return &collectorpb.FetchResponse{
		Products:   source.ToProtoProducts(produtos),
		TotalFound: source.SafeInt32(len(produtos)),
		FetchedAt:  time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func (s *UnifiedCollectorServer) FetchShop(ctx context.Context, req *collectorpb.FetchShopRequest) (*collectorpb.FetchShopResponse, error) {
	if req.GetShopId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "shop_id é obrigatório")
	}

	mkt := resolveMarketplace(req.GetMarketplace())
	marketplace := source.ProtoToMarketplace(mkt)

	src, ok := s.pipeline.GetProductSourceByMarketplace(marketplace)
	if !ok {
		return nil, status.Errorf(codes.Unimplemented,
			"nenhum receiver de produto configurado para marketplace %q", marketplace)
	}

	shopID := formatInt64(req.GetShopId())
	produtos, err := src.FetchShop(shopID, int(req.GetLimit()))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "falha ao buscar shop: %v", err)
	}

	return &collectorpb.FetchShopResponse{
		Products:   source.ToProtoProducts(produtos),
		TotalFound: source.SafeInt32(len(produtos)),
		FetchedAt:  time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// ─── Collect: search + persist ───────────────────────────────────────────────

func (s *UnifiedCollectorServer) Collect(ctx context.Context, req *collectorpb.CollectRequest) (*collectorpb.CollectResponse, error) {
	if req.GetBuscaId() == "" {
		return nil, status.Error(codes.InvalidArgument, "busca_id é obrigatório")
	}

	mkt := resolveMarketplace(req.GetMarketplace())
	marketplace := source.ProtoToMarketplace(mkt)

	var produtos []domain.Product
	var keyword string
	var err error

	switch t := req.GetTarget().(type) {
	case *collectorpb.CollectRequest_Keyword:
		if t.Keyword == "" {
			return nil, status.Error(codes.InvalidArgument, "keyword é obrigatório")
		}
		keyword = t.Keyword

		src, ok := s.pipeline.GetProductSourceByMarketplace(marketplace)
		if !ok {
			return nil, status.Errorf(codes.Unimplemented, "marketplace %q sem receiver de produto", marketplace)
		}
		produtos, err = src.Search(source.SearchQuery{
			Keyword: keyword,
			Limit:   int(req.GetLimit()),
			SortBy:  req.GetSortBy(),
		})

	case *collectorpb.CollectRequest_ShopId:
		if t.ShopId == 0 {
			return nil, status.Error(codes.InvalidArgument, "shop_id é obrigatório")
		}
		keyword = strconv.FormatInt(t.ShopId, 10)

		src, ok := s.pipeline.GetProductSourceByMarketplace(marketplace)
		if !ok {
			return nil, status.Errorf(codes.Unimplemented, "marketplace %q sem receiver de produto", marketplace)
		}
		produtos, err = src.FetchShop(formatInt64(t.ShopId), int(req.GetLimit()))

	default:
		return nil, status.Error(codes.InvalidArgument, "target (keyword ou shop_id) é obrigatório")
	}

	if err != nil {
		s.logger.Error("collect falhou",
			slog.String("marketplace", marketplace),
			slog.String("keyword", keyword),
			slog.String("error", err.Error()))
		return nil, status.Errorf(codes.Internal, "falha ao coletar: %v", err)
	}

	// Persist snapshot (non-blocking via channel)
	persisted := false
	if len(produtos) > 0 {
		snap := store.Snapshot{
			BuscaID:    req.GetBuscaId(),
			Keyword:    keyword,
			Estrategia: "coleta-agendada",
			Em:         time.Now().UTC(),
			Itens:      toSnapshotItems(produtos),
		}
		persisted = s.enqueueExport(snap)
	}

	return &collectorpb.CollectResponse{
		Products:   source.ToProtoProducts(produtos),
		TotalFound: source.SafeInt32(len(produtos)),
		FetchedAt:  time.Now().UTC().Format(time.RFC3339),
		Persisted:  persisted,
	}, nil
}

// enqueueExport sends a snapshot to the background export goroutine.
// Returns true if accepted, false if buffer is full.
func (s *UnifiedCollectorServer) enqueueExport(snap store.Snapshot) bool {
	select {
	case s.exportCh <- snap:
		return true
	default:
		s.logger.Warn("export buffer full, snapshot dropped",
			slog.String("keyword", snap.Keyword),
			slog.Int("items", len(snap.Itens)))
		return false
	}
}

// RunExporter drains the export channel and persists snapshots to BigQuery.
// Call as a goroutine; stops when ctx is cancelled (drains remaining first).
func (s *UnifiedCollectorServer) RunExporter(ctx context.Context) {
	for {
		select {
		case snap := <-s.exportCh:
			if err := s.snapshots.RegistrarSnapshot(ctx, snap); err != nil {
				s.logger.Error("snapshot export failed",
					slog.String("keyword", snap.Keyword),
					slog.Int("items", len(snap.Itens)),
					slog.String("error", err.Error()))
			}
		case <-ctx.Done():
			// Drain remaining snapshots on shutdown
			for {
				select {
				case snap := <-s.exportCh:
					_ = s.snapshots.RegistrarSnapshot(context.Background(), snap)
				default:
					return
				}
			}
		}
	}
}

// toSnapshotItems converts domain.Product slice to store.ItemSnapshot slice.
func toSnapshotItems(produtos []domain.Product) []store.ItemSnapshot {
	itens := make([]store.ItemSnapshot, 0, len(produtos))
	for i, p := range produtos {
		itens = append(itens, store.ItemSnapshot{
			Posicao:   i + 1,
			ProdutoID: fmt.Sprintf("shopee-%s-%s", p.ShopID, p.ID),
			Nome:      p.Name,
			Preco:     p.Price,
			Comissao:  p.Commission,
			Vendas:    p.Sales30d,
			Nota:      p.Rating,
			Imagem:    p.Image,
			Link:      p.Link,
			Loja:      p.ShopName,
		})
	}
	return itens
}

// ─── Coupon Collector ────────────────────────────────────────────────────────

// UnifiedCouponServer implementa coupon.v1.CouponCollectorService.
// Despacha para o CouponSource correto via Pipeline, baseado no marketplace do request.
type UnifiedCouponServer struct {
	couponpb.UnimplementedCouponCollectorServiceServer
	pipeline *Pipeline
	logger   *slog.Logger
}

func NewUnifiedCouponServer(pipeline *Pipeline, logger *slog.Logger) *UnifiedCouponServer {
	return &UnifiedCouponServer{pipeline: pipeline, logger: logger}
}

func (s *UnifiedCouponServer) FetchCoupons(ctx context.Context, req *couponpb.FetchCouponsRequest) (*couponpb.FetchCouponsResponse, error) {
	marketplace := source.ProtoToMarketplace(req.GetMarketplace())

	src, ok := s.pipeline.GetCouponSourceByMarketplace(marketplace)
	if !ok {
		return nil, status.Errorf(codes.Unimplemented,
			"nenhum receiver de cupom configurado para marketplace %q", marketplace)
	}

	pageSize := int(req.GetPageSize())
	if pageSize <= 0 {
		pageSize = 500
	}

	coupons, err := src.FetchCoupons(couponsource.FetchConfig{
		OwnerUID: req.GetOwnerUid(),
		PageSize: pageSize,
	})
	if err != nil {
		s.logger.Error("coupon fetch failed",
			slog.String("marketplace", marketplace),
			slog.String("owner_uid", req.GetOwnerUid()),
			slog.String("error", err.Error()))
		return nil, status.Errorf(codes.Internal, "falha ao buscar cupons: %v", err)
	}

	s.logger.Info("coupons collected",
		slog.String("marketplace", marketplace),
		slog.Int("count", len(coupons)))

	return &couponpb.FetchCouponsResponse{
		Coupons:    toProtoCoupons(coupons),
		TotalFound: source.SafeInt32(len(coupons)),
		FetchedAt:  time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

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
	}
	return result
}

func formatUnix(ts int64) string {
	if ts == 0 {
		return ""
	}
	return time.Unix(ts, 0).UTC().Format(time.RFC3339)
}
