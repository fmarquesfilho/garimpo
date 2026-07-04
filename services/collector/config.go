package main

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Validation errors.
var (
	ErrMissingVersion     = errors.New("version é obrigatório")
	ErrNoReceivers        = errors.New("pelo menos um receiver é obrigatório")
	ErrMissingReceiverID  = errors.New("id é obrigatório")
	ErrDuplicateID        = errors.New("id duplicado")
	ErrInvalidType        = errors.New("type deve ser 'product' ou 'coupon'")
	ErrMissingMarketplace = errors.New("marketplace é obrigatório")
	ErrMissingSchedule    = errors.New("schedule é obrigatório")
)

// CollectorConfig é a raiz da configuração YAML do collector unificado.
type CollectorConfig struct {
	Version   string           `yaml:"version"`
	Receivers []ReceiverConfig `yaml:"receivers"`
	Exporters ExportersConfig  `yaml:"exporters"`
	Settings  SettingsConfig   `yaml:"settings"`
}

// ReceiverConfig define um pipeline de coleta (produto ou cupom).
type ReceiverConfig struct {
	ID          string            `yaml:"id"`
	Type        string            `yaml:"type"`        // "product" ou "coupon"
	Marketplace string            `yaml:"marketplace"` // "shopee", "amazon", "mercadolivre"
	Schedule    string            `yaml:"schedule"`    // cron expression
	Credentials CredentialsConfig `yaml:"credentials"`
}

// CredentialsConfig mapeia nomes de env vars para cada credencial.
type CredentialsConfig struct {
	// Shopee affiliate API environment variables
	AppIDEnv  string `yaml:"app_id_env"`
	SecretEnv string `yaml:"secret_env"`

	// Amazon Product Advertising API environment variables
	AccessKeyEnv  string `yaml:"access_key_env"`
	SecretKeyEnv  string `yaml:"secret_key_env"`
	PartnerTagEnv string `yaml:"partner_tag_env"`

	// Mercado Livre
	ClientIDEnv     string `yaml:"client_id_env"`
	ClientSecretEnv string `yaml:"client_secret_env"`
	AccessTokenEnv  string `yaml:"access_token_env"`
	RefreshTokenEnv string `yaml:"refresh_token_env"`
}

// ExportersConfig agrupa configurações de exporters.
type ExportersConfig struct {
	BigQuery BigQueryExporterConfig `yaml:"bigquery"`
}

// BigQueryExporterConfig define a config do exporter BigQuery.
type BigQueryExporterConfig struct {
	ProjectEnv    string `yaml:"project_env"`
	Dataset       string `yaml:"dataset"`
	ProductsTable string `yaml:"products_table"`
	CouponsTable  string `yaml:"coupons_table"`
}

// SettingsConfig define configurações globais do collector.
type SettingsConfig struct {
	GRPCPort               int    `yaml:"grpc_port"`
	HealthPort             int    `yaml:"health_port"`
	LogLevel               string `yaml:"log_level"`
	MaxConcurrentReceivers int    `yaml:"max_concurrent_receivers"`
}

// LoadConfig lê e faz parse do arquivo YAML de configuração.
// Prioridade: env COLLECTOR_CONFIG > path padrão /etc/garimpei/collector.yaml > ./collector.yaml
func LoadConfig(paths ...string) (*CollectorConfig, error) {
	configPath := resolveConfigPath(paths...)

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("leitura config %s: %w", configPath, err)
	}

	var cfg CollectorConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", configPath, err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validação config: %w", err)
	}

	return &cfg, nil
}

// resolveConfigPath determina qual arquivo de config usar.
func resolveConfigPath(paths ...string) string {
	// Argumento explícito (para testes)
	if len(paths) > 0 && paths[0] != "" {
		return paths[0]
	}

	// Env var
	if p := os.Getenv("COLLECTOR_CONFIG"); p != "" {
		return p
	}

	// Padrão em produção
	if _, err := os.Stat("/etc/garimpei/collector.yaml"); err == nil {
		return "/etc/garimpei/collector.yaml"
	}

	// Local (dev)
	return "collector.yaml"
}

// Validate verifica integridade mínima da configuração.
func (c *CollectorConfig) Validate() error {
	if c.Version == "" {
		return ErrMissingVersion
	}

	if len(c.Receivers) == 0 {
		return ErrNoReceivers
	}

	seen := make(map[string]bool)
	for i, r := range c.Receivers {
		if r.ID == "" {
			return fmt.Errorf("receiver[%d]: %w", i, ErrMissingReceiverID)
		}
		if seen[r.ID] {
			return fmt.Errorf("receiver[%d] %q: %w", i, r.ID, ErrDuplicateID)
		}
		seen[r.ID] = true

		if r.Type != "product" && r.Type != "coupon" {
			return fmt.Errorf("receiver %q: %w, got %q", r.ID, ErrInvalidType, r.Type)
		}
		if r.Marketplace == "" {
			return fmt.Errorf("receiver %q: %w", r.ID, ErrMissingMarketplace)
		}
		if r.Schedule == "" {
			return fmt.Errorf("receiver %q: %w", r.ID, ErrMissingSchedule)
		}
	}

	if c.Settings.GRPCPort == 0 {
		c.Settings.GRPCPort = 50051
	}
	if c.Settings.HealthPort == 0 {
		c.Settings.HealthPort = 8081
	}
	if c.Settings.MaxConcurrentReceivers == 0 {
		c.Settings.MaxConcurrentReceivers = 3
	}

	return nil
}

// ResolveCredentialEnv retorna o valor da env var referenciada, ou "" se não definida.
func ResolveCredentialEnv(envName string) string {
	if envName == "" {
		return ""
	}
	return os.Getenv(envName)
}
