package publish

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// WhatsAppSender implementa Sender para a WaSenderAPI (cloud-hosted).
// Cada chamada a Enviar recebe o group_id via `config` — a mesma sessão
// serve N destinos (grupos).
//
// Referência: https://wasenderapi.com/api-docs/groups/send-group-message
type WhatsAppSender struct {
	apiKey  string
	http    *http.Client
	apiBase string // vazio = "https://www.wasenderapi.com"
}

func NovoWhatsAppSender(apiKey string) *WhatsAppSender {
	return &WhatsAppSender{
		apiKey: apiKey,
		http:   &http.Client{Timeout: 30 * time.Second},
	}
}

func (w *WhatsAppSender) Tipo() string { return "whatsapp" }

func (w *WhatsAppSender) base() string {
	if w.apiBase != "" {
		return w.apiBase
	}
	return "https://www.wasenderapi.com"
}

func (w *WhatsAppSender) Enviar(ctx context.Context, o Oferta, groupID string) (Resultado, error) {
	msg := o.LegendaHTML
	if msg == "" {
		msg = o.MensagemWhatsApp()
	}

	payload := map[string]any{
		"to":   groupID,
		"text": msg,
	}

	// Se tem imagem, envia junto como imageUrl
	if o.Imagem != "" {
		payload["imageUrl"] = o.Imagem
	}

	// Se tem link e não tem imagem, inclui o link no texto (WhatsApp gera preview)
	if o.Link != "" && o.Imagem == "" && !strings.Contains(msg, o.Link) {
		payload["text"] = msg + "\n\n🛒 " + o.Link
	}

	// Se tem link e tem imagem, adiciona o link ao final do texto (caption)
	if o.Link != "" && o.Imagem != "" {
		payload["text"] = msg + "\n\n🛒 " + o.Link
	}

	corpo, _ := json.Marshal(payload)
	url := fmt.Sprintf("%s/api/send-message", w.base())

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(corpo))
	if err != nil {
		return Resultado{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+w.apiKey)

	resp, err := w.http.Do(req)
	if err != nil {
		return Resultado{Canal: "whatsapp", Enviado: false, Mensagem: msg, Detalhe: err.Error()}, err
	}
	defer resp.Body.Close()

	var r struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
		Data    struct {
			MsgID  int    `json:"msgId"`
			Status string `json:"status"`
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
