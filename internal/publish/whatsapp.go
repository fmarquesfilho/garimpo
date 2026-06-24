package publish

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// WhatsAppSender implementa Sender para a Maytapi (cloud-hosted).
// Config pode conter múltiplos group IDs separados por vírgula.
// A mesma mensagem é enviada para todos os grupos listados.
//
// Referência: https://maytapi.com/whatsapp-api-documentation
type WhatsAppSender struct {
	productID string
	phoneID   string
	token     string
	http      *http.Client
}

// NovoWhatsAppSender cria um sender para o Maytapi.
func NovoWhatsAppSender(productID, phoneID, token string) *WhatsAppSender {
	return &WhatsAppSender{
		productID: productID,
		phoneID:   phoneID,
		token:     token,
		http:      &http.Client{Timeout: 30 * time.Second},
	}
}

func (w *WhatsAppSender) Tipo() string { return "whatsapp" }

func (w *WhatsAppSender) base() string {
	return fmt.Sprintf("https://api.maytapi.com/api/%s", w.productID)
}

// Enviar publica a oferta para um ou mais grupos (config com IDs separados por vírgula).
func (w *WhatsAppSender) Enviar(ctx context.Context, o Oferta, config string) (Resultado, error) {
	grupos := parseGrupos(config)
	if len(grupos) == 0 {
		return Resultado{Canal: "whatsapp", Enviado: false, Detalhe: "nenhum grupo configurado"},
			fmt.Errorf("whatsapp: config vazio")
	}

	// Monta a mensagem (converte HTML se necessário)
	msg := prepararMensagemWA(o)

	// Adiciona link ao final
	if o.Link != "" {
		msg = msg + "\n\n🛒 " + o.Link
	}

	// Envia para cada grupo
	var ultimoErro error
	enviados := 0
	for _, groupID := range grupos {
		err := w.enviarParaGrupo(ctx, o, groupID, msg)
		if err != nil {
			ultimoErro = err
		} else {
			enviados++
		}
	}

	if enviados == 0 {
		return Resultado{Canal: "whatsapp", Enviado: false, Mensagem: msg, Detalhe: ultimoErro.Error()}, ultimoErro
	}

	detalhe := fmt.Sprintf("enviado para %d/%d grupos", enviados, len(grupos))
	if ultimoErro != nil {
		detalhe += " (alguns falharam)"
	}
	return Resultado{Canal: "whatsapp", Enviado: true, Mensagem: msg, Detalhe: detalhe}, nil
}

func (w *WhatsAppSender) enviarParaGrupo(ctx context.Context, o Oferta, groupID, msg string) error {
	var payload map[string]any

	if o.Imagem != "" {
		payload = map[string]any{
			"to_number": groupID,
			"type":      "media",
			"message":   o.Imagem,
			"text":      msg,
		}
	} else {
		payload = map[string]any{
			"to_number": groupID,
			"type":      "text",
			"message":   msg,
		}
	}

	corpo, _ := json.Marshal(payload)
	url := fmt.Sprintf("%s/%s/sendMessage", w.base(), w.phoneID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(corpo))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-maytapi-key", w.token)

	resp, err := w.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var r struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&r)

	if !r.Success {
		msg := r.Message
		if msg == "" {
			msg = fmt.Sprintf("HTTP %d", resp.StatusCode)
		}
		return fmt.Errorf("grupo %s: %s", groupID, msg)
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
// WhatsApp aceita: *bold*, _italic_, ~strikethrough~, ```monospace```
func htmlParaWhatsApp(html string) string {
	r := strings.NewReplacer(
		"<b>", "*", "</b>", "*",
		"<strong>", "*", "</strong>", "*",
		"<i>", "_", "</i>", "_",
		"<em>", "_", "</em>", "_",
		"<s>", "~", "</s>", "~",
		"<del>", "~", "</del>", "~",
		"<u>", "", "</u>", "", // WhatsApp não suporta underline
		"<code>", "```", "</code>", "```",
		"<pre>", "```\n", "</pre>", "\n```",
		"<p>", "", "</p>", "\n",
		"<br>", "\n", "<br/>", "\n", "<br />", "\n",
	)
	result := r.Replace(html)

	// Remove quaisquer tags HTML restantes (ex.: <a>, <span>, etc.)
	result = stripHTMLTags(result)

	// Limpa linhas vazias duplicadas
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

// parseGrupos extrai os group IDs de uma config com vírgulas.
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

// NovoWhatsAppSenderFromEnv cria o sender a partir das variáveis de ambiente.
// Retorna nil se as variáveis não estiverem definidas.
func NovoWhatsAppSenderFromEnv() *WhatsAppSender {
	productID := os.Getenv("WHATSAPP_PRODUCT_ID")
	phoneID := os.Getenv("WHATSAPP_PHONE_ID")
	token := os.Getenv("WHATSAPP_API_KEY")
	if productID == "" || phoneID == "" || token == "" {
		return nil
	}
	return NovoWhatsAppSender(productID, phoneID, token)
}
