// Package publish é a PORTA DE SAÍDA do Garimpo: espelho do ProductSource (porta
// de entrada). Cada canal — Telegram hoje, outros depois — implementa Publicador,
// e o resto do sistema não precisa saber qual é.
package publish

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Oferta é o que sai para um canal: o produto curado + o link rastreável.
// Comissão entra aqui só para registro interno — NUNCA aparece na mensagem ao público.
type Oferta struct {
	ProdutoID  string
	Nome       string
	Categoria  string
	Preco      float64
	Comissao   float64
	Link       string
	Estrategia string
}

// Mensagem monta o texto voltado ao público (sem comissão, que é margem dela).
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

// Resultado descreve o que aconteceu (para o front mostrar o que "saiu").
type Resultado struct {
	Canal    string `json:"canal"`
	Enviado  bool   `json:"enviado"`
	Mensagem string `json:"mensagem"`
	Detalhe  string `json:"detalhe"`
	SubID    string `json:"sub_id"` // identificador de atribuição (canal_estrategia_data)
}

// SubID compõe o identificador de atribuição embutido no link (utm_content da
// Shopee): canal_estrategia_AAAAMMDD. É a peça que, lá na frente, permite saber
// QUAL canal converteu (Telegram vs. Instagram) no conversionReport.
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

// Publicador é a porta: cada canal a implementa.
type Publicador interface {
	Nome() string
	Publicar(ctx context.Context, o Oferta) (Resultado, error)
}
