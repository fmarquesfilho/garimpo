package publish

// TelegramSender implementa envio de ofertas via Telegram Bot API.
// Suporta sendPhoto (com fallback para sendMessage se CDN inacessível)
// e sendMessage (texto com botão inline de compra).

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/fmarquesfilho/garimpo/internal/apperr"
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
	// Usa legenda customizada se fornecida, senão gera do template padrão
	msg := o.LegendaHTML
	if msg == "" {
		msg = o.MensagemHTML()
	}
	// Sanitiza HTML: Telegram só aceita b, i, u, s, a, code, pre
	msg = sanitizarHTMLTelegram(msg)

	// Se tem imagem, usa sendPhoto com caption
	if o.Imagem != "" {
		return t.enviarFoto(ctx, o, chatID, msg)
	}

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
		return Resultado{}, fmt.Errorf("telegram criar request: %w", apperr.ErrTelegram)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.http.Do(req)
	if err != nil {
		return Resultado{Canal: "telegram", Enviado: false, Mensagem: msg, Detalhe: err.Error()}, fmt.Errorf("telegram enviar: %w", apperr.ErrTelegram)
	}
	defer resp.Body.Close()

	var r struct {
		Ok          bool   `json:"ok"`
		Description string `json:"description"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&r)
	if !r.Ok {
		return Resultado{Canal: "telegram", Enviado: false, Mensagem: msg, Detalhe: r.Description},
			fmt.Errorf("telegram %s: %w", r.Description, apperr.ErrTelegram)
	}
	return Resultado{Canal: "telegram", Enviado: true, Mensagem: msg, Detalhe: "enviado"}, nil
}

// enviarFoto usa sendPhoto do Telegram (foto + caption + botão inline).
func (t *TelegramSender) enviarFoto(ctx context.Context, o Oferta, chatID, caption string) (Resultado, error) {
	// Timeout curto para sendPhoto — se falhar, faz fallback para sendMessage.
	fotoCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	payload := map[string]any{
		"chat_id":    chatID,
		"photo":      o.Imagem,
		"caption":    caption,
		"parse_mode": "HTML",
	}
	if kb := inlineKeyboard(o.Link); kb != nil {
		payload["reply_markup"] = kb
	}

	corpo, _ := json.Marshal(payload)
	fotoURL := fmt.Sprintf("%s/bot%s/sendPhoto", t.base(), t.token)

	req, err := http.NewRequestWithContext(fotoCtx, http.MethodPost, fotoURL, bytes.NewReader(corpo))
	if err != nil {
		semFoto := o
		semFoto.Imagem = ""
		return t.Enviar(ctx, semFoto, chatID)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.http.Do(req)
	if err != nil {
		// Timeout ou erro de rede — fallback para texto
		semFoto := o
		semFoto.Imagem = ""
		return t.Enviar(ctx, semFoto, chatID)
	}
	defer resp.Body.Close()

	var r struct {
		Ok          bool   `json:"ok"`
		Description string `json:"description"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&r)
	if !r.Ok {
		// Foto inacessível (CDN Shopee bloqueada) — fallback para texto
		semFoto := o
		semFoto.Imagem = ""
		return t.Enviar(ctx, semFoto, chatID)
	}
	return Resultado{Canal: "telegram", Enviado: true, Mensagem: caption, Detalhe: "enviado com foto"}, nil
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

// sanitizarHTMLTelegram converte/remove tags não suportadas pelo Telegram.
// Telegram aceita apenas: b, i, u, s, a, code, pre, tg-spoiler.
// Tiptap gera <p>, <br>, <em>, <strong> — precisam ser convertidos.
func sanitizarHTMLTelegram(html string) string {
	r := strings.NewReplacer(
		"<p>", "",
		"</p>", "\n",
		"<br>", "\n",
		"<br/>", "\n",
		"<br />", "\n",
		"<strong>", "<b>",
		"</strong>", "</b>",
		"<em>", "<i>",
		"</em>", "</i>",
		"<del>", "<s>",
		"</del>", "</s>",
	)
	result := r.Replace(html)
	// Remove linhas vazias duplicadas
	for strings.Contains(result, "\n\n\n") {
		result = strings.ReplaceAll(result, "\n\n\n", "\n\n")
	}
	return strings.TrimSpace(result)
}
