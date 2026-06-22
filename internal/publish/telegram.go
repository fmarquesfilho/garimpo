package publish

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// TelegramSender implementa Sender para o Telegram Bot API.
// Cada chamada a Enviar recebe o chat_id via `config` — o mesmo bot
// serve N destinos.
type TelegramSender struct {
	token string
	http  *http.Client

	// apiBase permite apontar para outro host (testes). Vazio = oficial.
	apiBase string
}

func NovoTelegramSender(token string) *TelegramSender {
	return &TelegramSender{
		token: token,
		http:  &http.Client{Timeout: 15 * time.Second},
	}
}

func (t *TelegramSender) Tipo() string { return "telegram" }

func (t *TelegramSender) base() string {
	if t.apiBase != "" {
		return t.apiBase
	}
	return "https://api.telegram.org"
}

// inlineKeyboard monta o reply_markup com um botão "🛒 Comprar" apontando para o link.
func inlineKeyboard(link string) map[string]any {
	if link == "" {
		return nil
	}
	return map[string]any{
		"inline_keyboard": []any{
			[]any{
				map[string]any{"text": "🛒 Comprar", "url": link},
			},
		},
	}
}

func (t *TelegramSender) Enviar(ctx context.Context, o Oferta, chatID string) (Resultado, error) {
	msg := o.MensagemHTML()

	payload := map[string]any{
		"chat_id":                  chatID,
		"text":                     msg,
		"parse_mode":               "HTML",
		"disable_web_page_preview": false,
	}
	if kb := inlineKeyboard(o.Link); kb != nil {
		payload["reply_markup"] = kb
	}

	corpo, _ := json.Marshal(payload)
	url := fmt.Sprintf("%s/bot%s/sendMessage", t.base(), t.token)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(corpo))
	if err != nil {
		return Resultado{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.http.Do(req)
	if err != nil {
		return Resultado{Canal: "telegram", Enviado: false, Mensagem: msg, Detalhe: err.Error()}, err
	}
	defer resp.Body.Close()

	var r struct {
		Ok          bool   `json:"ok"`
		Description string `json:"description"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&r)
	if !r.Ok {
		return Resultado{Canal: "telegram", Enviado: false, Mensagem: msg, Detalhe: r.Description},
			fmt.Errorf("telegram: %s", r.Description)
	}
	return Resultado{Canal: "telegram", Enviado: true, Mensagem: msg, Detalhe: "enviado"}, nil
}

// ─── Compat: TelegramPublicador (wrapper legado para testes que usam Publicador) ─

// TelegramPublicador é o publicador direto (sem dispatcher), mantido por
// compatibilidade. Usa um chat_id fixo. Prefira o Dispatcher em código novo.
type TelegramPublicador struct {
	sender *TelegramSender
	chatID string
}

func NovoTelegram(token, chatID string) *TelegramPublicador {
	return &TelegramPublicador{
		sender: NovoTelegramSender(token),
		chatID: chatID,
	}
}

func (t *TelegramPublicador) Nome() string { return "telegram" }

func (t *TelegramPublicador) Publicar(ctx context.Context, o Oferta) (Resultado, error) {
	return t.sender.Enviar(ctx, o, t.chatID)
}
