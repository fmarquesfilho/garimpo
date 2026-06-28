package tenant

import (
	"context"

	"github.com/fmarquesfilho/garimpo/internal/store"
)

// RepoAdapter adapta tenant.Store para satisfazer store.TenantRepo.
// Faz a conversão entre tenant.Config e store.TenantConfig.
type RepoAdapter struct {
	inner Store
}

// NewRepoAdapter cria um adapter sobre um tenant.Store existente.
func NewRepoAdapter(s Store) *RepoAdapter {
	return &RepoAdapter{inner: s}
}

func (a *RepoAdapter) BuscarTenant(ctx context.Context, uid string) (*store.TenantConfig, error) {
	cfg, err := a.inner.Buscar(ctx, uid)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		return nil, nil
	}
	return toStoreTenantConfig(cfg), nil
}

func (a *RepoAdapter) SalvarTenant(ctx context.Context, tc store.TenantConfig) error {
	cfg := fromStoreTenantConfig(&tc)
	return a.inner.Salvar(ctx, *cfg)
}

func (a *RepoAdapter) ExcluirTenant(ctx context.Context, uid string) error {
	return a.inner.Excluir(ctx, uid)
}

func toStoreTenantConfig(c *Config) *store.TenantConfig {
	return &store.TenantConfig{
		UID:              c.UID,
		Email:            c.Email,
		ShopeeAppID:      c.ShopeeAppID,
		ShopeeSecretEnc:  c.ShopeeSecretEnc,
		TelegramTokenEnc: c.TelegramTokenEnc,
		TelegramChatID:   c.TelegramChatID,
		OnboardingStep:   c.OnboardingStep,
		AceitouTermos:    c.AceitouTermos,
		AceitouTermosEm:  c.AceitouTermosEm,
		CriadoEm:         c.CriadoEm,
		AtualizadoEm:     c.AtualizadoEm,
	}
}

func fromStoreTenantConfig(tc *store.TenantConfig) *Config {
	return &Config{
		UID:              tc.UID,
		Email:            tc.Email,
		ShopeeAppID:      tc.ShopeeAppID,
		ShopeeSecretEnc:  tc.ShopeeSecretEnc,
		TelegramTokenEnc: tc.TelegramTokenEnc,
		TelegramChatID:   tc.TelegramChatID,
		OnboardingStep:   tc.OnboardingStep,
		AceitouTermos:    tc.AceitouTermos,
		AceitouTermosEm:  tc.AceitouTermosEm,
		CriadoEm:         tc.CriadoEm,
		AtualizadoEm:     tc.AtualizadoEm,
	}
}
