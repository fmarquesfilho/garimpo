package store

import (
	"fmt"
	"strings"
)

// Template é um modelo de mensagem configurável pela usuária.
// Placeholders suportados: {{nome}}, {{preco}}, {{categoria}}, {{estrategia}}, {{link}}
// Templates são independentes de destino — podem ser usados em qualquer provedor.
type Template struct {
	ID       string `json:"id"`
	Nome     string `json:"nome"`     // nome amigável (ex.: "Oferta com foto")
	Corpo    string `json:"corpo"`    // corpo com placeholders (HTML permitido)
	ComFoto  bool   `json:"com_foto"` // se true, envia foto + caption em vez de texto
	Ativo    bool   `json:"ativo"`
	CriadoEm string `json:"criado_em"` // ISO 8601
}

// Renderizar aplica os dados nos placeholders do template.
func (t Template) Renderizar(nome string, preco float64, categoria, estrategia, link string) string {
	r := strings.NewReplacer(
		"{{nome}}", strings.TrimSpace(nome),
		"{{preco}}", fmt.Sprintf("R$ %.2f", preco),
		"{{categoria}}", categoria,
		"{{estrategia}}", estrategia,
		"{{link}}", link,
	)
	return r.Replace(t.Corpo)
}
