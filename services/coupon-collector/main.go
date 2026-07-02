// Package main is the entrypoint for the coupon-collector gRPC microservice.
// Determined by MARKETPLACE env var: one binary, multiple deploy targets.
package main

import (
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	couponpb "github.com/fmarquesfilho/garimpo/gen/go/coupon/v1"
	"github.com/fmarquesfilho/garimpo/internal/couponsource"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "50061"
	}

	marketplace := os.Getenv("MARKETPLACE")
	if marketplace == "" {
		marketplace = "shopee"
	}

	// Build source config from env (credentials vary by marketplace)
	cfg := couponsource.SourceConfig{
		AppID:      os.Getenv("SHOPEE_APP_ID"),
		Secret:     os.Getenv("SHOPEE_SECRET"),
		AccessKey:  os.Getenv("AMAZON_ACCESS_KEY"),
		SecretKey:  os.Getenv("AMAZON_SECRET_KEY"),
		PartnerTag: os.Getenv("AMAZON_PARTNER_TAG"),
	}

	src, err := couponsource.DefaultRegistry.Create(marketplace, cfg)
	if err != nil {
		logger.Error("marketplace não suportado", slog.String("marketplace", marketplace), slog.String("erro", err.Error()))
		os.Exit(1)
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		logger.Error("falha ao abrir porta", slog.String("port", port), slog.String("erro", err.Error()))
		os.Exit(1)
	}

	srv := grpc.NewServer()
	couponpb.RegisterCouponCollectorServiceServer(srv, NewCouponCollectorServer(src, logger))

	healthSrv := health.NewServer()
	healthpb.RegisterHealthServer(srv, healthSrv)
	healthSrv.SetServingStatus("coupon.v1.CouponCollectorService", healthpb.HealthCheckResponse_SERVING)

	reflection.Register(srv)

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		logger.Info("shutting down coupon-collector")
		srv.GracefulStop()
	}()

	logger.Info("coupon-collector gRPC listening",
		slog.String("port", port),
		slog.String("marketplace", src.Marketplace()))
	if err := srv.Serve(lis); err != nil {
		logger.Error("serve falhou", slog.String("erro", err.Error()))
		os.Exit(1)
	}
}
