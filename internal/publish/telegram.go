package publish

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// TelegramPublicador posta de verdade via Bot API (só net/http, sem dependência).
// Requer um bot (criado no @BotFather) que seja admin do canal/grupo, e o chat_id
// do destino. Não é usado até você definir TELEGRAM_BOT_TOKEN e TELEGRAM_CHAT_ID.
type TelegramPublicador struct {
	token  string
	chatID string
	http   *http.Client
}

func NovoTelegram(token, chatID string) *TelegramPublicador {
	return &TelegramPublicador{
		token:  token,
		chatID: chatID,
		http:   &http.Client{Timeout: 15 * time.Second},
	}
}

func (t *TelegramPublicador) Nome() string { return "telegram" }

func (t *TelegramPublicador) Publicar(ctx context.Context, o Oferta) (Resultado, error) {
	msg := o.Mensagem()
	corpo, _ := json.Marshal(map[string]any{
		"chat_id":                  t.chatID,
		"text":                     msg,
		"disable_web_page_preview": false,
	})
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.token)

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
