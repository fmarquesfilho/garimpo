package main

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds cache sidecar configuration from environment variables.
type Config struct {
	// GRPCPort is the port the cache sidecar listens on.
	GRPCPort int
	// MaxBytes is the maximum cache size in bytes (default: 256 MB).
	MaxBytes int64
	// TTLSeconds is the entry TTL in seconds (default: 1800 = 30 minutes).
	TTLSeconds int
	// CollectorGRPCAddress is the address of the Collector gRPC service.
	CollectorGRPCAddress string
	// SchemaPath is the path to busca-contract.json for validation.
	SchemaPath string
}

// LoadConfigFromEnv reads configuration from environment variables.
func LoadConfigFromEnv() (*Config, error) {
	cfg := &Config{
		GRPCPort:             50055,
		MaxBytes:             268435456, // 256 MB
		TTLSeconds:           1800,      // 30 minutes
		CollectorGRPCAddress: "localhost:50051",
		SchemaPath:           "/schemas/busca-contract.json",
	}

	if v := os.Getenv("CACHE_GRPC_PORT"); v != "" {
		port, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("CACHE_GRPC_PORT inválido: %w", err)
		}
		cfg.GRPCPort = port
	}

	if v := os.Getenv("CACHE_MAX_BYTES"); v != "" {
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("CACHE_MAX_BYTES inválido: %w", err)
		}
		cfg.MaxBytes = n
	}

	if v := os.Getenv("CACHE_TTL_SECONDS"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("CACHE_TTL_SECONDS inválido: %w", err)
		}
		cfg.TTLSeconds = n
	}

	if v := os.Getenv("COLLECTOR_GRPC_ADDRESS"); v != "" {
		cfg.CollectorGRPCAddress = v
	}

	if v := os.Getenv("BUSCA_CONTRACT_SCHEMA_PATH"); v != "" {
		cfg.SchemaPath = v
	}

	return cfg, nil
}
