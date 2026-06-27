package tenant

import (
	"context"
	"testing"
	"time"
)

func TestConfig_Configurado(t *testing.T) {
	cases := []struct {
		name   string
		cfg    *Config
		expect bool
	}{
		{"nil config", nil, false},
		{"step 0", &Config{OnboardingStep: 0}, false},
		{"step 3 sem appid", &Config{OnboardingStep: 3}, false},
		{"step 4 sem appid", &Config{OnboardingStep: 4}, false},
		{"step 4 com appid", &Config{OnboardingStep: 4, ShopeeAppID: "123"}, true},
		{"step 5 com appid", &Config{OnboardingStep: 5, ShopeeAppID: "456"}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.cfg.Configurado()
			if got != tc.expect {
				t.Errorf("Configurado() = %v, want %v", got, tc.expect)
			}
		})
	}
}

func TestConfig_SetAndGetShopeeSecret(t *testing.T) {
	cfg := &Config{UID: "user1"}
	err := cfg.SetShopeeSecret("meu-secret-123")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ShopeeSecretEnc == "" {
		t.Error("secret enc não deveria estar vazio")
	}
	if cfg.ShopeeSecretEnc == "meu-secret-123" {
		t.Error("secret enc não deveria ser plaintext")
	}

	got, err := cfg.ShopeeSecret()
	if err != nil {
		t.Fatal(err)
	}
	if got != "meu-secret-123" {
		t.Errorf("ShopeeSecret() = %q, want 'meu-secret-123'", got)
	}
}

func TestConfig_SetAndGetTelegramToken(t *testing.T) {
	cfg := &Config{UID: "user1"}
	err := cfg.SetTelegramToken("123456:ABC-DEF")
	if err != nil {
		t.Fatal(err)
	}

	got, err := cfg.TelegramToken()
	if err != nil {
		t.Fatal(err)
	}
	if got != "123456:ABC-DEF" {
		t.Errorf("TelegramToken() = %q, want '123456:ABC-DEF'", got)
	}
}

func TestConfig_EmptySecrets(t *testing.T) {
	cfg := &Config{UID: "user1"}

	secret, err := cfg.ShopeeSecret()
	if err != nil || secret != "" {
		t.Errorf("empty ShopeeSecret should return ('', nil), got (%q, %v)", secret, err)
	}

	token, err := cfg.TelegramToken()
	if err != nil || token != "" {
		t.Errorf("empty TelegramToken should return ('', nil), got (%q, %v)", token, err)
	}
}

func TestMemoryStore_CRUD(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()

	// Buscar inexistente
	cfg, err := store.Buscar(ctx, "user1")
	if err != nil || cfg != nil {
		t.Errorf("buscar inexistente: cfg=%v, err=%v", cfg, err)
	}

	// Salvar
	err = store.Salvar(ctx, Config{
		UID:            "user1",
		Email:          "test@test.com",
		ShopeeAppID:    "app123",
		OnboardingStep: 4,
		CriadoEm:       time.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}

	// Buscar existente
	cfg, err = store.Buscar(ctx, "user1")
	if err != nil {
		t.Fatal(err)
	}
	if cfg == nil {
		t.Fatal("cfg não deveria ser nil")
	}
	if cfg.ShopeeAppID != "app123" {
		t.Errorf("ShopeeAppID = %q, want 'app123'", cfg.ShopeeAppID)
	}
	if cfg.OnboardingStep != 4 {
		t.Errorf("OnboardingStep = %d, want 4", cfg.OnboardingStep)
	}

	// Atualizar
	cfg.OnboardingStep = 5
	err = store.Salvar(ctx, *cfg)
	if err != nil {
		t.Fatal(err)
	}
	cfg2, _ := store.Buscar(ctx, "user1")
	if cfg2.OnboardingStep != 5 {
		t.Errorf("após update, step = %d, want 5", cfg2.OnboardingStep)
	}

	// Excluir
	err = store.Excluir(ctx, "user1")
	if err != nil {
		t.Fatal(err)
	}
	cfg3, _ := store.Buscar(ctx, "user1")
	if cfg3 != nil {
		t.Error("após excluir, deveria retornar nil")
	}
}

func TestMemoryStore_IsolamentoEntreUsers(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()

	_ = store.Salvar(ctx, Config{UID: "alice", ShopeeAppID: "alice-app"})
	_ = store.Salvar(ctx, Config{UID: "bob", ShopeeAppID: "bob-app"})

	alice, _ := store.Buscar(ctx, "alice")
	bob, _ := store.Buscar(ctx, "bob")

	if alice.ShopeeAppID != "alice-app" {
		t.Errorf("alice app = %q", alice.ShopeeAppID)
	}
	if bob.ShopeeAppID != "bob-app" {
		t.Errorf("bob app = %q", bob.ShopeeAppID)
	}

	// Excluir alice não afeta bob
	_ = store.Excluir(ctx, "alice")
	bob2, _ := store.Buscar(ctx, "bob")
	if bob2 == nil {
		t.Error("bob não deveria ser afetado pela exclusão de alice")
	}
}
