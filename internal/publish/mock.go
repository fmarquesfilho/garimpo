package publish

import (
	"context"
	"log"
)

// MockPublicador não envia nada de verdade: monta a mensagem, registra no log e
// devolve o que SERIA enviado. É o padrão enquanto não há token/canal reais.
type MockPublicador struct {
	Canal string
}

func NovoMock(canal string) *MockPublicador {
	if canal == "" {
		canal = "telegram"
	}
	return &MockPublicador{Canal: canal}
}

func (m *MockPublicador) Nome() string { return "mock:" + m.Canal }

func (m *MockPublicador) Publicar(ctx context.Context, o Oferta) (Resultado, error) {
	_ = ctx
	msg := o.Mensagem()
	log.Printf("[publicador mock/%s] enviaria:\n%s", m.Canal, msg)
	return Resultado{
		Canal:    m.Canal,
		Enviado:  true,
		Mensagem: msg,
		Detalhe:  "mock — nada foi enviado de verdade",
	}, nil
}
