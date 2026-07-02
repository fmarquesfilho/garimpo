// Package main é o entrypoint do microserviço collector-amazon (gRPC).
// Usa a Amazon Creators API (substituta da PA-API 5.0) com AWS Sig V4.
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

	collectorpb "github.com/fmarquesfilho/garimpo/gen/go/collector/v1"
	"github.com/fmarquesfilho/garimpo/internal/source"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	port := os.Getenv("PORT")
	if port == "" {
		port = "50055"
	}

	partnerTag := os.Getenv("AMAZON_PARTNER_TAG")
	accessKey := os.Getenv("AMAZON_ACCESS_KEY")
	secretKey := os.Getenv("AMAZON_SECRET_KEY")
	if partnerTag == "" || accessKey == "" || secretKey == "" {
		logger.Error("AMAZON_PARTNER_TAG, AMAZON_ACCESS_KEY e AMAZON_SECRET_KEY são obrigatórios")
		os.Exit(1)
	}

	amazonSource := source.NewAmazonAdapter(accessKey, secretKey, partnerTag)

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		logger.Error("falha ao abrir porta", slog.String("port", port), slog.String("erro", err.Error()))
		os.Exit(1)
	}

	srv := grpc.NewServer()

	collectorpb.RegisterCollectorServiceServer(srv, NewAmazonCollectorServer(amazonSource))

	healthSrv := health.NewServer()
	healthpb.RegisterHealthServer(srv, healthSrv)
	healthSrv.SetServingStatus("collector.v1.CollectorService", healthpb.HealthCheckResponse_SERVING)

	reflection.Register(srv)

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		logger.Info("shutting down collector-amazon")
		srv.GracefulStop()
	}()

	logger.Info("collector-amazon gRPC listening",
		slog.String("port", port),
		slog.String("marketplace", amazonSource.Marketplace()))
	if err := srv.Serve(lis); err != nil {
		logger.Error("serve falhou", slog.String("erro", err.Error()))
		os.Exit(1)
	}
}
