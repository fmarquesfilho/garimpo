// Package main é o entrypoint do garimpei-collector unificado.
// Um único binário config-driven que substitui collector, collector-amazon e coupon-collector.
// Modelo: OpenTelemetry Collector — receivers definidos em YAML, gRPC para chamadas on-demand.
// Ref: docs/decisoes/0018-collector-unificado.md
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
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	collectorpb "github.com/fmarquesfilho/garimpo/gen/go/collector/v1"
	couponpb "github.com/fmarquesfilho/garimpo/gen/go/coupon/v1"
	garimpotel "github.com/fmarquesfilho/garimpo/internal/otel"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Initialize OpenTelemetry (graceful: no-op if no collector configured)
	ctx := context.Background()
	otelShutdown, otelErr := garimpotel.Init(ctx, "collector")
	if otelErr == nil {
		logger = slog.New(garimpotel.SlogHandler(""))
		slog.SetDefault(logger)
		defer otelShutdown(ctx) //nolint:errcheck
	}

	// Carrega configuração
	cfg, err := LoadConfig()
	if err != nil {
		logger.Error("falha ao carregar config", slog.String("erro", err.Error()))
		os.Exit(1)
	}

	// Ajusta log level
	if cfg.Settings.LogLevel == "debug" {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
		slog.SetDefault(logger)
	}

	logger.Info("garimpei-collector unificado iniciando",
		slog.String("version", cfg.Version),
		slog.Int("receivers", len(cfg.Receivers)),
		slog.Int("grpc_port", cfg.Settings.GRPCPort))

	// Monta pipeline (receivers + cron scheduler)
	pipeline, err := NewPipeline(cfg, logger)
	if err != nil {
		logger.Error("falha ao montar pipeline", slog.String("erro", err.Error()))
		os.Exit(1)
	}

	// Inicia scheduler
	pipeline.Start()

	// gRPC server
	port := fmt.Sprintf(":%d", cfg.Settings.GRPCPort)
	lis, err := net.Listen("tcp", port)
	if err != nil {
		logger.Error("falha ao abrir porta gRPC", slog.String("port", port), slog.String("erro", err.Error()))
		os.Exit(1)
	}

	srv := grpc.NewServer(garimpotel.GRPCServerInterceptors()...)

	// Registra CollectorService (produtos) — genérico, delega por marketplace
	collectorpb.RegisterCollectorServiceServer(srv, NewUnifiedCollectorServer(pipeline, logger))

	// Registra CouponCollectorService (cupons) — genérico, delega por marketplace
	couponpb.RegisterCouponCollectorServiceServer(srv, NewUnifiedCouponServer(pipeline, logger))

	// Health check
	healthSrv := health.NewServer()
	healthpb.RegisterHealthServer(srv, healthSrv)
	healthSrv.SetServingStatus("collector.v1.CollectorService", healthpb.HealthCheckResponse_SERVING)
	healthSrv.SetServingStatus("coupon.v1.CouponCollectorService", healthpb.HealthCheckResponse_SERVING)

	// Reflection (dev/debug)
	reflection.Register(srv)

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		logger.Info("shutdown signal recebido")

		// Para pipeline (aguarda jobs ativos)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		pipeline.Stop(ctx)

		// Para gRPC
		srv.GracefulStop()
	}()

	logger.Info("garimpei-collector gRPC listening", slog.String("port", port))
	if err := srv.Serve(lis); err != nil {
		logger.Error("serve falhou", slog.String("erro", err.Error()))
		os.Exit(1)
	}
}
