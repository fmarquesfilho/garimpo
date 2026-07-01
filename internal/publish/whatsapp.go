package publish

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/fmarquesfilho/garimpo/internal/apperr"
)

// WhatsAppSender implementa Sender para a Meta WhatsApp Business Cloud API.
// Usa o Graph API diretamente (sem intermediários como Maytapi).
//
// Endpoint: POST https://graph.facebook.com/v25.0/{PHONE_NUMBER_ID}/messages
// Auth: Bearer token (System User permanent token)
//
// Config pode conter múltiplos destinatários (números/group IDs) separados por vírgula.
//
// Referência: https://developers.facebook.com/docs/whatsapp/cloud-api/
type WhatsAppSender struct {
	phoneNumberID string
	accessToken   string
	http          *http.Client

	// apiBase permite apontar para outro host (testes). Vazio = Graph API oficial.
	apiBase string
}

// NovoWhatsAppSender cria um sender para a Meta Cloud API.
func NovoWhatsAppSender(phoneNumberID, accessToken string) *WhatsAppSender {
	return &WhatsAppSender{
		phoneNumberID: phoneNumberID,
		accessToken:   accessToken,
		http:          &http.Client{Timeout: 30 * time.Second},
	}
}

func (w *WhatsAppSender) Tipo() string { return "whatsapp" }

func (w *WhatsAppSender) base() string {
	if w.apiBase != "" {
		return w.apiBase
	}
	return "https://graph.facebook.com/v25.0"
}

func (w *WhatsAppSender) messagesURL() string {
	return fmt.Sprintf("%s/%s/messages", w.base(), w.phoneNumberID)
}

// Enviar publica a oferta para um ou mais destinatários (config com números separados por vírgula).
func (w *WhatsAppSender) Enviar(ctx context.Context, o Oferta, config string) (Resultado, error) {
	destinos := parseGrupos(config)
	if len(destinos) == 0 {
		return Resultado{Canal: "whatsapp", Enviado: false, Detalhe: "nenhum destinatário configurado"},
			fmt.Errorf("whatsapp config vazio: %w", apperr.ErrNoConfig)
	}

	msg := prepararMensagemWA(o)
	if o.Link != "" {
		msg = msg + "\n\n🛒 " + o.Link
	}

	var ultimoErro error
	enviados := 0
	for _, dest := range destinos {
		var err error
		if o.Imagem != "" {
			err = w.enviarImagem(ctx, dest, o.Imagem, msg)
		} else {
			err = w.enviarTexto(ctx, dest, msg)
		}
		if err != nil {
			ultimoErro = err
		} else {
			enviados++
		}
	}

	if enviados == 0 {
		return Resultado{Canal: "whatsapp", Enviado: false, Mensagem: msg, Detalhe: ultimoErro.Error()}, ultimoErro
	}

	detalhe := fmt.Sprintf("enviado para %d/%d destinatários", enviados, len(destinos))
	if ultimoErro != nil {
		detalhe += " (alguns falharam)"
	}
	return Resultado{Canal: "whatsapp", Enviado: true, Mensagem: msg, Detalhe: detalhe}, nil
}

// enviarTexto envia uma mensagem de texto via Meta Cloud API.
func (w *WhatsAppSender) enviarTexto(ctx context.Context, to, text string) error {
	payload := map[string]any{
		"messaging_product": "whatsapp",
		"to":                to,
		"type":              "text",
		"text": map[string]any{
			"preview_url": true,
			"body":        text,
		},
	}
	return w.post(ctx, payload, to)
}

// enviarImagem envia uma imagem com caption via Meta Cloud API.
func (w *WhatsAppSender) enviarImagem(ctx context.Context, to, imageURL, caption string) error {
	payload := map[string]any{
		"messaging_product": "whatsapp",
		"to":                to,
		"type":              "image",
		"image": map[string]any{
			"link":    imageURL,
			"caption": caption,
		},
	}
	return w.post(ctx, payload, to)
}

// post executa a requisição HTTP ao Graph API.
func (w *WhatsAppSender) post(ctx context.Context, payload map[string]any, to string) error {
	corpo, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.messagesURL(), bytes.NewReader(corpo))
	if err != nil {
		return fmt.Errorf("whatsapp criar request para %s: %w", to, apperr.ErrWhatsApp)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+w.accessToken)

	resp, err := w.http.Do(req)
	if err != nil {
		return fmt.Errorf("whatsapp enviar para %s: %w", to, apperr.ErrWhatsApp)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		var errResp struct {
			Error struct {
				Message string `json:"message"`
				Code    int    `json:"code"`
			} `json:"error"`
		}
		_ = json.Unmarshal(body, &errResp)
		detail := errResp.Error.Message
		if detail == "" {
			detail = fmt.Sprintf("HTTP %d", resp.StatusCode)
		}
		return fmt.Errorf("whatsapp %s: %s: %w", to, detail, apperr.ErrWhatsApp)
	}

	return nil
}

// ─── Formatação WhatsApp ─────────────────────────────────────────────────────

// prepararMensagemWA converte a legenda HTML ou gera texto com formatação WhatsApp.
func prepararMensagemWA(o Oferta) string {
	if o.LegendaHTML != "" {
		return htmlParaWhatsApp(o.LegendaHTML)
	}
	return o.MensagemWhatsApp()
}

// MensagemWhatsApp monta o texto com formatação WhatsApp (bold = *, italic = _).
func (o Oferta) MensagemWhatsApp() string {
	var b strings.Builder
	fmt.Fprintf(&b, "✨ *%s*\n", strings.TrimSpace(o.Nome))
	if o.Categoria != "" {
		fmt.Fprintf(&b, "📂 _%s_\n", o.Categoria)
	}
	if o.Preco > 0 {
		fmt.Fprintf(&b, "💸 *R$ %.2f*", o.Preco)
	}
	return strings.TrimRight(b.String(), "\n")
}

// htmlParaWhatsApp converte HTML (do editor Tiptap/Telegram) para formatação WhatsApp.
func htmlParaWhatsApp(html string) string {
	r := strings.NewReplacer(
		"<b>", "*", "</b>", "*",
		"<strong>", "*", "</strong>", "*",
		"<i>", "_", "</i>", "_",
		"<em>", "_", "</em>", "_",
		"<s>", "~", "</s>", "~",
		"<del>", "~", "</del>", "~",
		"<u>", "", "</u>", "",
		"<code>", "```", "</code>", "```",
		"<pre>", "```\n", "</pre>", "\n```",
		"<p>", "", "</p>", "\n",
		"<br>", "\n", "<br/>", "\n", "<br />", "\n",
	)
	result := r.Replace(html)
	result = stripHTMLTags(result)
	for strings.Contains(result, "\n\n\n") {
		result = strings.ReplaceAll(result, "\n\n\n", "\n\n")
	}
	return strings.TrimSpace(result)
}

// stripHTMLTags remove todas as tags HTML restantes.
func stripHTMLTags(s string) string {
	var b strings.Builder
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

// parseGrupos extrai os destinatários de uma config com vírgulas.
func parseGrupos(config string) []string {
	var grupos []string
	for _, g := range strings.Split(config, ",") {
		g = strings.TrimSpace(g)
		if g != "" {
			grupos = append(grupos, g)
		}
	}
	return grupos
}

// ─── Factory helper ──────────────────────────────────────────────────────────

// NovoWhatsAppSenderFromEnv cria o sender a partir das variáveis de ambiente (Meta Cloud API).
// Retorna nil se as variáveis não estiverem definidas.
func NovoWhatsAppSenderFromEnv() *WhatsAppSender {
	phoneNumberID := os.Getenv("WHATSAPP_PHONE_NUMBER_ID")
	accessToken := os.Getenv("WHATSAPP_ACCESS_TOKEN")
	if phoneNumberID == "" || accessToken == "" {
		return nil
	}
	return NovoWhatsAppSender(phoneNumberID, accessToken)
}
