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
// Cada chamada a Enviar recebe o group_id via `config` — a mesma sessão
// serve N destinos (grupos).
//
// Referência: https://maytapi.com/whatsapp-api-documentation
type WhatsAppSender struct {
	productID string
	phoneID   string
	token     string
	http      *http.Client
}

// NovoWhatsAppSender cria um sender para o Maytapi.
// Requer WHATSAPP_PRODUCT_ID, WHATSAPP_PHONE_ID e WHATSAPP_API_KEY no ambiente.
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

func (w *WhatsAppSender) Enviar(ctx context.Context, o Oferta, groupID string) (Resultado, error) {
	msg := o.LegendaHTML
	if msg == "" {
		msg = o.MensagemWhatsApp()
	}

	// Adiciona link ao final
	if o.Link != "" {
		msg = msg + "\n\n🛒 " + o.Link
	}

	var payload map[string]any

	if o.Imagem != "" {
		// Envia imagem com caption
		payload = map[string]any{
			"to_number": groupID,
			"type":      "media",
			"message":   o.Imagem,
			"text":      msg,
		}
	} else {
		// Envia texto simples
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
		return Resultado{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-maytapi-key", w.token)

	resp, err := w.http.Do(req)
	if err != nil {
		return Resultado{Canal: "whatsapp", Enviado: false, Mensagem: msg, Detalhe: err.Error()}, err
	}
	defer resp.Body.Close()

	var r struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
		Data    struct {
			MsgID string `json:"msgId"`
		} `json:"data"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&r)

	if !r.Success {
		detalhe := r.Message
		if detalhe == "" {
			detalhe = fmt.Sprintf("HTTP %d", resp.StatusCode)
		}
		return Resultado{Canal: "whatsapp", Enviado: false, Mensagem: msg, Detalhe: detalhe},
			fmt.Errorf("whatsapp: %s", detalhe)
	}
	return Resultado{Canal: "whatsapp", Enviado: true, Mensagem: msg, Detalhe: "enviado"}, nil
}

// ─── Formatação WhatsApp ─────────────────────────────────────────────────────

// MensagemWhatsApp monta o texto com formatação WhatsApp (bold = *, italic = _).
// WhatsApp não suporta HTML, então usamos a marcação nativa.
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
