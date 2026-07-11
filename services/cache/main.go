// Package main is the entrypoint for the garimpei cache sidecar.
// In-memory LRU cache (L2) between C# API and Collector.
// Protects against thundering herd via singleflight and serves as
// fast read path for product data.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	cachepb "github.com/fmarquesfilho/garimpo/gen/go/cache/v1"
	collectorpb "github.com/fmarquesfilho/garimpo/gen/go/collector/v1"
	garimpotel "github.com/fmarquesfilho/garimpo/internal/otel"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Initialize OpenTelemetry (graceful: no-op if no collector configured)
	ctx := context.Background()
	otelShutdown, otelErr := garimpotel.Init(ctx, "cache-sidecar")
	if otelErr == nil {
		logger = slog.New(garimpotel.SlogHandler(""))
		slog.SetDefault(logger)
		defer otelShutdown(ctx) //nolint:errcheck
	}

	cfg, err := LoadConfigFromEnv()
	if err != nil {
		logger.Error("falha ao carregar config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("cache-sidecar iniciando",
		slog.Int("port", cfg.GRPCPort),
		slog.Int64("max_bytes", cfg.MaxBytes),
		slog.Int("ttl_seconds", cfg.TTLSeconds))

	// Connect to Collector
	collectorConn, err := grpc.NewClient(
		cfg.CollectorGRPCAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		logger.Error("falha ao conectar ao Collector",
			slog.String("address", cfg.CollectorGRPCAddress),
			slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer collectorConn.Close()

	collectorClient := collectorpb.NewCollectorServiceClient(collectorConn)

	// Create LRU cache
	ttl := time.Duration(cfg.TTLSeconds) * time.Second
	lruCache := NewLRUCache(cfg.MaxBytes, ttl)

	// Create cache server
	cacheServer := NewCacheServer(lruCache, collectorClient, ttl, logger)

	// gRPC server with OTel interceptors
	port := fmt.Sprintf(":%d", cfg.GRPCPort)
	lis, err := net.Listen("tcp", port)
	if err != nil {
		logger.Error("falha ao abrir porta gRPC",
			slog.String("port", port),
			slog.String("error", err.Error()))
		os.Exit(1)
	}

	srv := grpc.NewServer(garimpotel.GRPCServerInterceptors()...)
	cachepb.RegisterCacheServiceServer(srv, cacheServer)

	// Health check
	healthSrv := health.NewServer()
	healthpb.RegisterHealthServer(srv, healthSrv)
	healthSrv.SetServingStatus("cache.v1.CacheService", healthpb.HealthCheckResponse_SERVING)

	// Reflection (dev/debug)
	reflection.Register(srv)

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		logger.Info("shutdown signal recebido")
		srv.GracefulStop()
	}()

	logger.Info("cache-sidecar gRPC listening", slog.String("port", port))
	if err := srv.Serve(lis); err != nil {
		logger.Error("serve falhou", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
