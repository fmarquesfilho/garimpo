package httpapi

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/fmarquesfilho/garimpo/internal/tenant"
)

// ── Endpoints de onboarding ──────────────────────────────────────────────────

// onboardingStatus retorna o progresso do onboarding do tenant.
func (srv *Server) onboardingStatus(w http.ResponseWriter, r *http.Request) {
	user := usuarioDoCtx(r)

	cfg, err := srv.Tenants.Buscar(r.Context(), user.UID)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}

	if cfg == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"step":        0,
			"configurado": false,
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"step":             cfg.OnboardingStep,
		"configurado":      cfg.Configurado(),
		"aceitou_termos":   cfg.AceitouTermos,
		"shopee_app_id":    cfg.ShopeeAppID,
		"shopee_secret":    mascarar(cfg.ShopeeSecretEnc),
		"telegram_token":   mascarar(cfg.TelegramTokenEnc),
		"telegram_chat_id": cfg.TelegramChatID,
	})
}

// onboardingTermos aceita os termos de uso (LGPD).
func (srv *Server) onboardingTermos(w http.ResponseWriter, r *http.Request) {
	user := usuarioDoCtx(r)

	var req struct {
		Aceito bool `json:"aceito"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || !req.Aceito {
		writeErr(w, http.StatusBadRequest, "você deve aceitar os termos para continuar")
		return
	}

	cfg, _ := srv.Tenants.Buscar(r.Context(), user.UID)
	if cfg == nil {
		cfg = &tenant.Config{
			UID:      user.UID,
			Email:    user.Email,
			CriadoEm: time.Now().UTC(),
		}
	}

	cfg.AceitouTermos = true
	cfg.AceitouTermosEm = time.Now().UTC()
	if cfg.OnboardingStep < 1 {
		cfg.OnboardingStep = 1
	}
	cfg.AtualizadoEm = time.Now().UTC()

	if err := srv.Tenants.Salvar(r.Context(), *cfg); err != nil {
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"step": cfg.OnboardingStep, "status": "termos aceitos"})
}

// onboardingShopee salva as credenciais da Shopee Affiliate API.
func (srv *Server) onboardingShopee(w http.ResponseWriter, r *http.Request) {
	user := usuarioDoCtx(r)

	var req struct {
		AppID  string `json:"app_id"`
		Secret string `json:"secret"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "json inválido")
		return
	}

	appID := strings.TrimSpace(req.AppID)
	secret := strings.TrimSpace(req.Secret)
	if appID == "" || secret == "" {
		writeErr(w, http.StatusBadRequest, "app_id e secret são obrigatórios")
		return
	}

	cfg, _ := srv.Tenants.Buscar(r.Context(), user.UID)
	if cfg == nil {
		cfg = &tenant.Config{
			UID:      user.UID,
			Email:    user.Email,
			CriadoEm: time.Now().UTC(),
		}
	}

	if !cfg.AceitouTermos {
		writeErr(w, http.StatusForbidden, "aceite os termos antes de configurar credenciais")
		return
	}

	cfg.ShopeeAppID = appID
	if err := cfg.SetShopeeSecret(secret); err != nil {
		writeErr(w, http.StatusInternalServerError, "erro ao criptografar secret")
		return
	}
	if cfg.OnboardingStep < 2 {
		cfg.OnboardingStep = 2
	}
	cfg.AtualizadoEm = time.Now().UTC()

	if err := srv.Tenants.Salvar(r.Context(), *cfg); err != nil {
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"step": cfg.OnboardingStep, "status": "credenciais shopee salvas"})
}

// onboardingTelegram salva a configuração do bot Telegram (opcional).
func (srv *Server) onboardingTelegram(w http.ResponseWriter, r *http.Request) {
	user := usuarioDoCtx(r)

	var req struct {
		Token  string `json:"token"`
		ChatID string `json:"chat_id"`
		Pular  bool   `json:"pular"` // true = pula este step
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "json inválido")
		return
	}

	cfg, _ := srv.Tenants.Buscar(r.Context(), user.UID)
	if cfg == nil {
		writeErr(w, http.StatusBadRequest, "complete os steps anteriores primeiro")
		return
	}

	if req.Pular {
		if cfg.OnboardingStep < 3 {
			cfg.OnboardingStep = 3
		}
	} else {
		token := strings.TrimSpace(req.Token)
		chatID := strings.TrimSpace(req.ChatID)
		if token == "" || chatID == "" {
			writeErr(w, http.StatusBadRequest, "token e chat_id são obrigatórios (ou envie pular: true)")
			return
		}
		if err := cfg.SetTelegramToken(token); err != nil {
			writeErr(w, http.StatusInternalServerError, "erro ao criptografar token")
			return
		}
		cfg.TelegramChatID = chatID
		if cfg.OnboardingStep < 3 {
			cfg.OnboardingStep = 3
		}
	}

	cfg.AtualizadoEm = time.Now().UTC()
	if err := srv.Tenants.Salvar(r.Context(), *cfg); err != nil {
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"step": cfg.OnboardingStep, "status": "telegram configurado"})
}

// onboardingValidar testa as credenciais Shopee fazendo uma chamada real à API.
func (srv *Server) onboardingValidar(w http.ResponseWriter, r *http.Request) {
	user := usuarioDoCtx(r)

	cfg, _ := srv.Tenants.Buscar(r.Context(), user.UID)
	if cfg == nil || cfg.ShopeeAppID == "" {
		writeErr(w, http.StatusBadRequest, "configure as credenciais Shopee primeiro")
		return
	}

	secret, err := cfg.ShopeeSecret()
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "erro ao descriptografar secret")
		return
	}

	// Chamada de teste à API da Shopee (busca 1 produto)
	if err := validarCredenciaisShopee(cfg.ShopeeAppID, secret); err != nil {
		writeErr(w, http.StatusBadRequest, fmt.Sprintf("credenciais inválidas: %v", err))
		return
	}

	cfg.OnboardingStep = 4
	cfg.AtualizadoEm = time.Now().UTC()
	if err := srv.Tenants.Salvar(r.Context(), *cfg); err != nil {
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"step":   4,
		"status": "credenciais validadas — onboarding completo!",
	})
}

// onboardingExcluirConta exclui a conta e todos os dados do tenant (LGPD).
func (srv *Server) onboardingExcluirConta(w http.ResponseWriter, r *http.Request) {
	user := usuarioDoCtx(r)

	var req struct {
		Confirmar bool `json:"confirmar"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || !req.Confirmar {
		writeErr(w, http.StatusBadRequest, "envie {\"confirmar\": true} para confirmar a exclusão")
		return
	}

	if err := srv.Tenants.Excluir(r.Context(), user.UID); err != nil {
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "conta e dados excluídos"})
}

// ── Helpers ──────────────────────────────────────────────────────────────────

// validarCredenciaisShopee faz uma requisição de teste à API de afiliados.
func validarCredenciaisShopee(appID, secret string) error {
	query := `{ productOfferV2(listType: 1, sortType: 5, limit: 1) { nodes { itemId } pageInfo { page } } }`
	body, _ := json.Marshal(map[string]string{"query": query})

	ts := strconv.FormatInt(time.Now().Unix(), 10)
	sum := sha256.Sum256([]byte(appID + ts + string(body) + secret))
	sig := hex.EncodeToString(sum[:])

	req, err := http.NewRequest(http.MethodPost, "https://open-api.affiliate.shopee.com.br/graphql", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization",
		fmt.Sprintf("SHA256 Credential=%s, Timestamp=%s, Signature=%s", appID, ts, sig))

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("conexão falhou: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)

	var gql struct {
		Data   any `json:"data"`
		Errors []struct {
			Message    string `json:"message"`
			Extensions struct {
				Code int `json:"code"`
			} `json:"extensions"`
		} `json:"errors"`
	}
	if err := json.Unmarshal(raw, &gql); err != nil {
		return fmt.Errorf("resposta inválida da Shopee")
	}
	if len(gql.Errors) > 0 {
		e := gql.Errors[0]
		return fmt.Errorf("código %d: %s", e.Extensions.Code, e.Message)
	}
	return nil
}

// mascarar retorna "configurado" ou "" dependendo se o valor criptografado existe.
func mascarar(enc string) string {
	if enc != "" {
		return "***configurado***"
	}
	return ""
}
