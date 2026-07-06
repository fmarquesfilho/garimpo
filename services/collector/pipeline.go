package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/robfig/cron/v3"

	"github.com/fmarquesfilho/garimpo/internal/couponsource"
	"github.com/fmarquesfilho/garimpo/internal/source"
)

// ErrUnknownReceiverType indicates a receiver type that is not product or coupon.
var ErrUnknownReceiverType = errors.New("tipo de receiver desconhecido")

// Pipeline orquestra múltiplos receivers com scheduling cron.
// Modelo inspirado no OpenTelemetry Collector: receivers → (processor) → exporter.
type Pipeline struct {
	mu        sync.RWMutex
	receivers map[string]*Receiver
	cron      *cron.Cron
	sem       chan struct{} // limita concorrência
	logger    *slog.Logger
}

// Receiver é um pipeline individual de coleta.
type Receiver struct {
	Config        ReceiverConfig
	ProductSource source.ProductSource
	CouponSource  couponsource.CouponSource
	EntryID       cron.EntryID
}

// NewPipeline cria o orquestrador a partir da configuração.
func NewPipeline(cfg *CollectorConfig, logger *slog.Logger) (*Pipeline, error) {
	p := &Pipeline{
		receivers: make(map[string]*Receiver),
		cron:      cron.New(cron.WithParser(cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow))),
		sem:       make(chan struct{}, cfg.Settings.MaxConcurrentReceivers),
		logger:    logger,
	}

	for _, rcfg := range cfg.Receivers {
		if err := p.addReceiver(rcfg); err != nil {
			return nil, fmt.Errorf("receiver %q: %w", rcfg.ID, err)
		}
	}

	return p, nil
}

// addReceiver instancia o adapter e registra o cron job.
func (p *Pipeline) addReceiver(rcfg ReceiverConfig) error {
	recv := &Receiver{Config: rcfg}

	switch rcfg.Type {
	case "product":
		src, err := p.createProductSource(rcfg)
		if err != nil {
			return err
		}
		recv.ProductSource = src

	case "coupon":
		src, err := p.createCouponSource(rcfg)
		if err != nil {
			return err
		}
		recv.CouponSource = src

	default:
		return fmt.Errorf("%w: %s", ErrUnknownReceiverType, rcfg.Type)
	}

	entryID, err := p.cron.AddFunc(rcfg.Schedule, func() {
		p.runReceiver(recv)
	})
	if err != nil {
		return fmt.Errorf("cron inválido %q: %w", rcfg.Schedule, err)
	}
	recv.EntryID = entryID

	p.mu.Lock()
	p.receivers[rcfg.ID] = recv
	p.mu.Unlock()

	p.logger.Info("receiver registrado",
		slog.String("id", rcfg.ID),
		slog.String("type", rcfg.Type),
		slog.String("marketplace", rcfg.Marketplace),
		slog.String("schedule", rcfg.Schedule))

	return nil
}

// createProductSource instancia um ProductSource via Registry.
func (p *Pipeline) createProductSource(rcfg ReceiverConfig) (source.ProductSource, error) {
	cfg := source.SourceConfig{
		AppID:      ResolveCredentialEnv(rcfg.Credentials.AppIDEnv),
		Secret:     ResolveCredentialEnv(rcfg.Credentials.SecretEnv),
		AccessKey:  ResolveCredentialEnv(rcfg.Credentials.AccessKeyEnv),
		SecretKey:  ResolveCredentialEnv(rcfg.Credentials.SecretKeyEnv),
		PartnerTag: ResolveCredentialEnv(rcfg.Credentials.PartnerTagEnv),
	}
	return source.DefaultRegistry.Create(rcfg.Marketplace, cfg)
}

// createCouponSource instancia um CouponSource via Registry.
func (p *Pipeline) createCouponSource(rcfg ReceiverConfig) (couponsource.CouponSource, error) {
	cfg := couponsource.SourceConfig{
		AppID:      ResolveCredentialEnv(rcfg.Credentials.AppIDEnv),
		Secret:     ResolveCredentialEnv(rcfg.Credentials.SecretEnv),
		AccessKey:  ResolveCredentialEnv(rcfg.Credentials.AccessKeyEnv),
		SecretKey:  ResolveCredentialEnv(rcfg.Credentials.SecretKeyEnv),
		PartnerTag: ResolveCredentialEnv(rcfg.Credentials.PartnerTagEnv),
	}
	return couponsource.DefaultRegistry.Create(rcfg.Marketplace, cfg)
}

// Start inicia o cron scheduler (não-blocking).
func (p *Pipeline) Start() {
	p.cron.Start()
	p.logger.Info("pipeline iniciado", slog.Int("receivers", len(p.receivers)))
}

// Stop para o scheduler e aguarda jobs ativos.
func (p *Pipeline) Stop(ctx context.Context) {
	stopCtx := p.cron.Stop()
	select {
	case <-stopCtx.Done():
		p.logger.Info("pipeline parou graciosamente")
	case <-ctx.Done():
		p.logger.Warn("pipeline stop atingiu timeout")
	}
}

// runReceiver executa a coleta de um receiver, respeitando o semáforo de concorrência.
func (p *Pipeline) runReceiver(recv *Receiver) {
	p.sem <- struct{}{} // acquire
	defer func() { <-p.sem }()

	logger := p.logger.With(
		slog.String("receiver", recv.Config.ID),
		slog.String("marketplace", recv.Config.Marketplace),
		slog.String("type", recv.Config.Type),
	)

	switch recv.Config.Type {
	case "product":
		logger.Info("coleta de produtos iniciada")
		// A coleta real é delegada ao ProductSource.
		// Em uso completo, aqui chamaria Search() com keywords de um store/owner.
		// Por agora, o cron garante que o receiver está vivo e pronto para gRPC on-demand.
		logger.Info("receiver product pronto para chamadas gRPC on-demand")

	case "coupon":
		logger.Info("coleta de cupons iniciada")
		logger.Info("receiver coupon pronto para chamadas gRPC on-demand")
	}
}

// GetProductSource retorna a ProductSource de um receiver pelo ID.
func (p *Pipeline) GetProductSource(id string) (source.ProductSource, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	recv, ok := p.receivers[id]
	if !ok || recv.ProductSource == nil {
		return nil, false
	}
	return recv.ProductSource, true
}

// GetCouponSource retorna a CouponSource de um receiver pelo ID.
func (p *Pipeline) GetCouponSource(id string) (couponsource.CouponSource, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	recv, ok := p.receivers[id]
	if !ok || recv.CouponSource == nil {
		return nil, false
	}
	return recv.CouponSource, true
}

// GetProductSourceByMarketplace retorna o primeiro ProductSource para o marketplace dado.
func (p *Pipeline) GetProductSourceByMarketplace(marketplace string) (source.ProductSource, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, recv := range p.receivers {
		if recv.Config.Type == "product" && recv.Config.Marketplace == marketplace {
			return recv.ProductSource, true
		}
	}
	return nil, false
}

// GetCouponSourceByMarketplace retorna o primeiro CouponSource para o marketplace dado.
func (p *Pipeline) GetCouponSourceByMarketplace(marketplace string) (couponsource.CouponSource, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, recv := range p.receivers {
		if recv.Config.Type == "coupon" && recv.Config.Marketplace == marketplace {
			return recv.CouponSource, true
		}
	}
	return nil, false
}

// ReceiverIDs retorna a lista de IDs de receivers registrados.
func (p *Pipeline) ReceiverIDs() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	ids := make([]string, 0, len(p.receivers))
	for id := range p.receivers {
		ids = append(ids, id)
	}
	return ids
}

// ShopeeCredentials retorna AppID e Secret do primeiro receiver Shopee configurado.
func (p *Pipeline) ShopeeCredentials() (appID, secret string) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, recv := range p.receivers {
		if recv.Config.Marketplace == "shopee" {
			appID = ResolveCredentialEnv(recv.Config.Credentials.AppIDEnv)
			secret = ResolveCredentialEnv(recv.Config.Credentials.SecretEnv)
			if appID != "" && secret != "" {
				return
			}
		}
	}
	return "", ""
}
