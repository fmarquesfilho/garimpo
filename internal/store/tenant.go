package store

import "time"

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
