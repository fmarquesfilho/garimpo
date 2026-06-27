package tenant

import "time"

// Config armazena as credenciais e configurações de um tenant (usuário).
type Config struct {
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
func (c *Config) Configurado() bool {
	return c != nil && c.OnboardingStep >= 4 && c.ShopeeAppID != ""
}

// ShopeeSecret descriptografa e retorna o secret da Shopee.
func (c *Config) ShopeeSecret() (string, error) {
	if c == nil || c.ShopeeSecretEnc == "" {
		return "", nil
	}
	return Decrypt(c.ShopeeSecretEnc)
}

// TelegramToken descriptografa e retorna o token do bot Telegram.
func (c *Config) TelegramToken() (string, error) {
	if c == nil || c.TelegramTokenEnc == "" {
		return "", nil
	}
	return Decrypt(c.TelegramTokenEnc)
}

// SetShopeeSecret criptografa e armazena o secret.
func (c *Config) SetShopeeSecret(plain string) error {
	enc, err := Encrypt(plain)
	if err != nil {
		return err
	}
	c.ShopeeSecretEnc = enc
	return nil
}

// SetTelegramToken criptografa e armazena o token do bot.
func (c *Config) SetTelegramToken(plain string) error {
	enc, err := Encrypt(plain)
	if err != nil {
		return err
	}
	c.TelegramTokenEnc = enc
	return nil
}
