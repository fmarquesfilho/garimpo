// Package alerts envia notificações automáticas quando variações de preço
// significativas são detectadas nas lojas monitoradas. Integra com Telegram
// para enviar mensagens em tempo real para um grupo configurado.
package alerts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fmarquesfilho/garimpo/internal/store"
)

// Config define as configurações do sistema de alertas.
type Config struct {
	// TelegramToken é o token do bot Telegram (mesmo do publish).
	TelegramToken string
	// TelegramChatID é o chat_id do grupo/canal onde os alertas serão enviados.
	TelegramChatID string
	// Threshold é a variação mínima (valor absoluto) para disparar alerta.
	// Ex: 0.15 = 15%. Default: 0.15.
	Threshold float64
	// ApenasQuedas: se true, só envia alertas de queda (oportunidades).
	ApenasQuedas bool
	// Logger para registrar alertas enviados.
	Logger *slog.Logger
}

// ConfigFromEnv cria Config a partir de variáveis de ambiente.
func ConfigFromEnv() Config {
	threshold := 0.15
	if s := os.Getenv("ALERTAS_THRESHOLD"); s != "" {
		if v, err := strconv.ParseFloat(s, 64); err == nil && v > 0 && v < 1 {
			threshold = v
		}
	}
	apenasQuedas := os.Getenv("ALERTAS_APENAS_QUEDAS") == "true"

	// Token de alertas: usa ALERTAS_TELEGRAM_TOKEN se disponível,
	// senão cai pro TELEGRAM_BOT_TOKEN (bot de publicações como fallback).
	token := os.Getenv("ALERTAS_TELEGRAM_TOKEN")
	if token == "" {
		token = os.Getenv("TELEGRAM_BOT_TOKEN")
	}

	return Config{
		TelegramToken:  token,
		TelegramChatID: os.Getenv("ALERTAS_TELEGRAM_CHAT_ID"),
		Threshold:      threshold,
		ApenasQuedas:   apenasQuedas,
		Logger:         slog.Default(),
	}
}

// Ativo retorna true se o sistema de alertas está configurado.
func (c Config) Ativo() bool {
	return c.TelegramToken != "" && c.TelegramChatID != ""
}

// Alerter verifica variações de preço e envia alertas.
type Alerter struct {
	cfg    Config
	client *http.Client
}

// Novo cria um Alerter com a configuração fornecida.
func Novo(cfg Config) *Alerter {
	return &Alerter{
		cfg:    cfg,
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

// VerificarENotificar compara as novidades de uma busca e envia alertas
// para variações que excedam o threshold.
func (a *Alerter) VerificarENotificar(ctx context.Context, eventos store.EventoStore, buscaID string) {
	if !a.cfg.Ativo() {
		return
	}

	novidades, err := eventos.Novidades(ctx, buscaID, 7)
	if err != nil {
		a.cfg.Logger.Error("alertas: falha ao buscar novidades",
			slog.String("busca", buscaID), slog.String("erro", err.Error()))
		return
	}

	var alertas []store.VariacaoPreco
	for _, v := range novidades.Variacoes {
		absVar := v.Variacao
		if absVar < 0 {
			absVar = -absVar
		}
		if absVar < a.cfg.Threshold {
			continue
		}
		if a.cfg.ApenasQuedas && v.Variacao > 0 {
			continue
		}
		alertas = append(alertas, v)
	}

	if len(alertas) == 0 {
		return
	}

	// Monta mensagem de alertas
	msg := a.formatarMensagem(buscaID, alertas)

	if err := a.enviarTelegram(ctx, msg); err != nil {
		a.cfg.Logger.Error("alertas: falha ao enviar telegram",
			slog.String("busca", buscaID), slog.String("erro", err.Error()))
		return
	}

	a.cfg.Logger.Info("alertas enviados",
		slog.String("busca", buscaID),
		slog.Int("alertas", len(alertas)),
		slog.String("chat_id", a.cfg.TelegramChatID),
	)
}

// VerificarNovos envia alerta sobre produtos novos detectados.
func (a *Alerter) VerificarNovos(ctx context.Context, eventos store.EventoStore, buscaID string) {
	if !a.cfg.Ativo() {
		return
	}

	novidades, err := eventos.Novidades(ctx, buscaID, 2) // últimos 2 dias para pegar novos recentes
	if err != nil || len(novidades.ProdutosNovos) == 0 {
		return
	}

	msg := a.formatarNovos(buscaID, novidades.ProdutosNovos)
	if err := a.enviarTelegram(ctx, msg); err != nil {
		a.cfg.Logger.Error("alertas: falha ao enviar novidades telegram",
			slog.String("busca", buscaID), slog.String("erro", err.Error()))
	}
}

func (a *Alerter) formatarMensagem(buscaID string, alertas []store.VariacaoPreco) string {
	var b strings.Builder
	b.WriteString("🔔 <b>Alerta de Preço</b>\n")
	b.WriteString(fmt.Sprintf("🏪 Loja: <code>%s</code>\n\n", buscaID))

	for i, v := range alertas {
		if i >= 10 { // Limita a 10 alertas por mensagem
			b.WriteString(fmt.Sprintf("\n<i>...e mais %d alertas</i>", len(alertas)-10))
			break
		}
		emoji := "📉"
		direcao := "↓"
		if v.Variacao > 0 {
			emoji = "📈"
			direcao = "↑"
		}
		pct := v.Variacao * 100
		b.WriteString(fmt.Sprintf("%s <b>%s</b>\n", emoji, truncar(v.Nome, 40)))
		b.WriteString(fmt.Sprintf("   R$ %.2f → R$ %.2f (%s%.1f%%)\n\n",
			v.PrecoAnterior, v.PrecoAtual, direcao, abs(pct)))
	}

	b.WriteString(fmt.Sprintf("⏰ %s", time.Now().Format("02/01 15:04")))
	return b.String()
}

func (a *Alerter) formatarNovos(buscaID string, novos []store.ProdutoNovo) string {
	var b strings.Builder
	b.WriteString("🆕 <b>Produtos Novos Detectados</b>\n")
	b.WriteString(fmt.Sprintf("🏪 Loja: <code>%s</code>\n\n", buscaID))

	for i, p := range novos {
		if i >= 5 { // Limita a 5 por mensagem
			b.WriteString(fmt.Sprintf("\n<i>...e mais %d novos</i>", len(novos)-5))
			break
		}
		b.WriteString(fmt.Sprintf("• <b>%s</b>\n", truncar(p.Nome, 40)))
		b.WriteString(fmt.Sprintf("  R$ %.2f", p.Preco))
		if p.Comissao > 0 {
			b.WriteString(fmt.Sprintf(" · %.0f%% comissão", p.Comissao*100))
		}
		b.WriteString("\n\n")
	}

	b.WriteString(fmt.Sprintf("⏰ %s", time.Now().Format("02/01 15:04")))
	return b.String()
}

func (a *Alerter) enviarTelegram(ctx context.Context, msg string) error {
	payload := map[string]any{
		"chat_id":                  a.cfg.TelegramChatID,
		"text":                     msg,
		"parse_mode":               "HTML",
		"disable_web_page_preview": true,
	}

	corpo, _ := json.Marshal(payload)
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", a.cfg.TelegramToken)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(corpo))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var r struct {
		Ok          bool   `json:"ok"`
		Description string `json:"description"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&r)
	if !r.Ok {
		return fmt.Errorf("telegram: %s", r.Description)
	}
	return nil
}

// EnviarTeste envia uma mensagem de teste para verificar que o bot e chat_id estão corretos.
func (a *Alerter) EnviarTeste(ctx context.Context) error {
	msg := fmt.Sprintf("✅ <b>Garimpo — Alertas Ativos!</b>\n\n"+
		"Os alertas de preço estão configurados corretamente.\n"+
		"Threshold: <b>%.0f%%</b>\n"+
		"Apenas quedas: <b>%v</b>\n\n"+
		"Você receberá notificações quando produtos monitorados tiverem "+
		"variações de preço significativas.\n\n"+
		"⏰ %s", a.cfg.Threshold*100, a.cfg.ApenasQuedas, time.Now().Format("02/01/2006 15:04"))
	return a.enviarTelegram(ctx, msg)
}

func truncar(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
