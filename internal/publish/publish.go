// Package publish é a PORTA DE SAÍDA do Garimpo: espelho do ProductSource (porta
// de entrada). Cada provedor (Telegram, WhatsApp, …) implementa Sender, e o
// Dispatcher roteia a oferta para o destino correto baseado no tipo/id.
//
// Padrão de projeto: Strategy + Registry. Cada provedor é uma strategy de envio;
// o Dispatcher é o registry que mapeia tipo → sender.
package publish

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// ─── Modelo de Destino (genérico, agnóstico de provedor) ─────────────────────

// Destino é um canal de publicação: um grupo de Telegram, um número de
// WhatsApp, etc. É o que a usuária cadastra no dashboard.
type Destino struct {
	ID     string `json:"id"`     // slug único (ex.: "ofertas-beleza")
	Nome   string `json:"nome"`   // nome amigável
	Tipo   string `json:"tipo"`   // "telegram" | "whatsapp" | ...
	Config string `json:"config"` // dado específico do provedor (chat_id, telefone, ...)
	Ativo  bool   `json:"ativo"`
}

// DestinoStore persiste destinos. Implementações: MemDestinoStore (dev),
// BigQuery (produção).
type DestinoStore interface {
	Listar(ctx context.Context) ([]Destino, error)
	Buscar(ctx context.Context, id string) (Destino, error)
	Salvar(ctx context.Context, d Destino) error
	Deletar(ctx context.Context, id string) error
}

// ─── Oferta e Resultado ──────────────────────────────────────────────────────

// Oferta é o que sai para um canal: o produto curado + o link rastreável.
// Comissão entra aqui só para registro interno — NUNCA aparece na mensagem ao público.
type Oferta struct {
	ProdutoID  string
	Nome       string
	Categoria  string
	Preco      float64
	Comissao   float64
	Link       string
	Imagem     string // URL da foto do produto
	Estrategia string
	DestinoID  string // qual destino usar (vazio = padrão do provedor)
	TemplateID string // qual template usar (vazio = MensagemHTML padrão)
}

// Resultado descreve o que aconteceu (para o front mostrar o que "saiu").
type Resultado struct {
	Canal    string `json:"canal"`
	Enviado  bool   `json:"enviado"`
	Mensagem string `json:"mensagem"`
	Detalhe  string `json:"detalhe"`
	SubID    string `json:"sub_id"` // identificador de atribuição (canal_estrategia_data)
}

// ─── Formatação de mensagem ──────────────────────────────────────────────────

// Mensagem monta o texto voltado ao público (texto plano, fallback).
func (o Oferta) Mensagem() string {
	var b strings.Builder
	fmt.Fprintf(&b, "✨ %s\n", strings.TrimSpace(o.Nome))
	if o.Preco > 0 {
		fmt.Fprintf(&b, "💸 R$ %.2f\n", o.Preco)
	}
	if o.Link != "" {
		fmt.Fprintf(&b, "🛒 %s", o.Link)
	}
	return strings.TrimRight(b.String(), "\n")
}

// MensagemHTML monta o texto com formatação rica (Telegram/WhatsApp HTML).
// O link NÃO vai no corpo — vai no botão inline, mantendo a mensagem limpa.
func (o Oferta) MensagemHTML() string {
	var b strings.Builder
	fmt.Fprintf(&b, "✨ <b>%s</b>\n", strings.TrimSpace(o.Nome))
	if o.Categoria != "" {
		fmt.Fprintf(&b, "📂 <i>%s</i>\n", o.Categoria)
	}
	if o.Preco > 0 {
		fmt.Fprintf(&b, "💸 <b>R$ %.2f</b>\n", o.Preco)
	}
	if o.Estrategia != "" {
		fmt.Fprintf(&b, "🎯 %s", o.Estrategia)
	}
	return strings.TrimRight(b.String(), "\n")
}

// ─── SubID (atribuição para rastrear conversão) ──────────────────────────────

// SubID compõe o identificador de atribuição embutido no link (utm_content da
// Shopee): canal_estrategia_AAAAMMDD. É a peça que permite saber QUAL destino
// converteu no conversionReport.
func SubID(canal, estrategia string, t time.Time) string {
	limpa := func(s string) string {
		s = strings.ToLower(s)
		var b strings.Builder
		for _, r := range s {
			if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
				b.WriteRune(r)
			}
		}
		return b.String()
	}
	return fmt.Sprintf("%s_%s_%s", limpa(canal), limpa(estrategia), t.Format("20060102"))
}

// ─── Interfaces de envio ─────────────────────────────────────────────────────

// Sender é a interface que cada provedor implementa (Telegram, WhatsApp, ...).
// Diferente de Publicador (que era monolítico), Sender foca no envio puro;
// o roteamento fica no Dispatcher.
type Sender interface {
	// Tipo retorna o identificador do provedor (ex.: "telegram", "whatsapp").
	Tipo() string
	// Enviar publica a oferta para o destino (config contém chat_id/telefone/etc).
	Enviar(ctx context.Context, o Oferta, config string) (Resultado, error)
}

// Publicador é a interface de alto nível consumida pelo httpapi. O Dispatcher
// a implementa, roteando para o Sender correto via destino.
type Publicador interface {
	Nome() string
	Publicar(ctx context.Context, o Oferta) (Resultado, error)
}
