// Package main é o entrypoint do microserviço publisher (gRPC).
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

	publisherpb "github.com/fmarquesfilho/garimpo/gen/go/publisher/v1"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	port := os.Getenv("PORT")
	if port == "" {
		port = "50052"
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		logger.Error("falha ao abrir porta", slog.String("port", port), slog.String("erro", err.Error()))
		os.Exit(1)
	}

	srv := grpc.NewServer()

	publisherpb.RegisterPublisherServiceServer(srv, NewPublisherServer())

	// Health check
	healthSrv := health.NewServer()
	healthpb.RegisterHealthServer(srv, healthSrv)
	healthSrv.SetServingStatus("publisher.v1.PublisherService", healthpb.HealthCheckResponse_SERVING)

	reflection.Register(srv)

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		logger.Info("shutting down publisher")
		srv.GracefulStop()
	}()

	logger.Info("publisher gRPC listening", slog.String("port", port))
	if err := srv.Serve(lis); err != nil {
		logger.Error("serve falhou", slog.String("erro", err.Error()))
		os.Exit(1)
	}
}
