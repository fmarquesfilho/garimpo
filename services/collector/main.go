// Package main é o entrypoint do microserviço shopee-collector (gRPC).
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
		port = "50051"
	}

	appID := os.Getenv("SHOPEE_APP_ID")
	secret := os.Getenv("SHOPEE_SECRET")
	if appID == "" || secret == "" {
		logger.Error("SHOPEE_APP_ID e SHOPEE_SECRET são obrigatórios")
		os.Exit(1)
	}

	// Cria a source via Registry — Open/Closed principle:
	// adicionar marketplace = registrar factory, zero mudança aqui.
	shopeeSource := source.NewShopeeAdapter(appID, secret)

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		logger.Error("falha ao abrir porta", slog.String("port", port), slog.String("erro", err.Error()))
		os.Exit(1)
	}

	srv := grpc.NewServer()

	// Registra o serviço collector com a source injetada
	collectorpb.RegisterCollectorServiceServer(srv, NewCollectorServer(shopeeSource))

	// Health check
	healthSrv := health.NewServer()
	healthpb.RegisterHealthServer(srv, healthSrv)
	healthSrv.SetServingStatus("collector.v1.CollectorService", healthpb.HealthCheckResponse_SERVING)

	// Reflection (dev)
	reflection.Register(srv)

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		logger.Info("shutting down collector")
		srv.GracefulStop()
	}()

	logger.Info("collector gRPC listening",
		slog.String("port", port),
		slog.String("marketplace", shopeeSource.Marketplace()))
	if err := srv.Serve(lis); err != nil {
		logger.Error("serve falhou", slog.String("erro", err.Error()))
		os.Exit(1)
	}
}
