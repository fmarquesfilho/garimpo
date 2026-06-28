package store

import (
	"time"

	"github.com/fmarquesfilho/garimpo/internal/crypto"
)

// TenantConfig armazena as credenciais e configurações de um tenant (usuário).
type TenantConfig struct {
	UID              string    `json:"uid"`
	Email            string    `json:"email,omitempty"`
	ShopeeAppID      string    `json:"shopee_app_id,omitempty"`
	ShopeeSecretEnc  string    `json:"-"` // criptografado — nunca exposto
	TelegramTokenEnc string    `json:"-"` // criptografado
	TelegramChatID   string    `json:"telegram_chat_id,omitempty"`
	OnboardingStep   int       `json:"onboarding_step"` // 0=início, 4=completo
	AceitouTermos    bool      `json:"aceitou_termos"`
	AceitouTermosEm  time.Time `json:"aceitou_termos_em,omitempty"`
	CriadoEm         time.Time `json:"criado_em"`
	AtualizadoEm     time.Time `json:"atualizado_em"`
}

// Configurado retorna true se o tenant tem credenciais Shopee válidas.
func (c *TenantConfig) Configurado() bool {
	return c != nil && c.OnboardingStep >= 4 && c.ShopeeAppID != ""
}

// ShopeeSecret descriptografa e retorna o secret da Shopee.
func (c *TenantConfig) ShopeeSecret() (string, error) {
	if c == nil || c.ShopeeSecretEnc == "" {
		return "", nil
	}
	return crypto.Decrypt(c.ShopeeSecretEnc)
}

// TelegramToken descriptografa e retorna o token do bot Telegram.
func (c *TenantConfig) TelegramToken() (string, error) {
	if c == nil || c.TelegramTokenEnc == "" {
		return "", nil
	}
	return crypto.Decrypt(c.TelegramTokenEnc)
}

// SetShopeeSecret criptografa e armazena o secret.
func (c *TenantConfig) SetShopeeSecret(plain string) error {
	enc, err := crypto.Encrypt(plain)
	if err != nil {
		return err
	}
	c.ShopeeSecretEnc = enc
	return nil
}

// SetTelegramToken criptografa e armazena o token do bot.
func (c *TenantConfig) SetTelegramToken(plain string) error {
	enc, err := crypto.Encrypt(plain)
	if err != nil {
		return err
	}
	c.TelegramTokenEnc = enc
	return nil
}
