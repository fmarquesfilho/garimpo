// Package main é o entrypoint do microserviço scheduler (gRPC + cron).
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

	schedulerpb "github.com/fmarquesfilho/garimpo/gen/go/scheduler/v1"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	port := os.Getenv("PORT")
	if port == "" {
		port = "50054"
	}

	collectorAddr := os.Getenv("COLLECTOR_ADDR")
	if collectorAddr == "" {
		collectorAddr = "localhost:50051"
	}
	publisherAddr := os.Getenv("PUBLISHER_ADDR")
	if publisherAddr == "" {
		publisherAddr = "localhost:50052"
	}

	server, err := NewSchedulerServer(collectorAddr, publisherAddr, logger)
	if err != nil {
		logger.Error("falha ao criar scheduler server", slog.String("erro", err.Error()))
		os.Exit(1)
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		logger.Error("falha ao abrir porta", slog.String("port", port), slog.String("erro", err.Error()))
		os.Exit(1)
	}

	srv := grpc.NewServer()

	schedulerpb.RegisterSchedulerServiceServer(srv, server)

	healthSrv := health.NewServer()
	healthpb.RegisterHealthServer(srv, healthSrv)
	healthSrv.SetServingStatus("scheduler.v1.SchedulerService", healthpb.HealthCheckResponse_SERVING)

	reflection.Register(srv)

	// Start cron jobs
	server.Start()

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		logger.Info("shutting down scheduler")
		server.Stop()
		srv.GracefulStop()
	}()

	logger.Info("scheduler gRPC listening", slog.String("port", port))
	if err := srv.Serve(lis); err != nil {
		logger.Error("serve falhou", slog.String("erro", err.Error()))
		os.Exit(1)
	}
}
