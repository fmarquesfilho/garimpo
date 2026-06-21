package publish

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestMensagemComPrecoELink(t *testing.T) {
	o := Oferta{Nome: "Perfume Floral 100ml", Preco: 99.9, Link: "https://s.shopee/x", Comissao: 0.55}
	msg := o.Mensagem()

	for _, querConter := range []string{"Perfume Floral 100ml", "R$ 99.90", "https://s.shopee/x"} {
		if !strings.Contains(msg, querConter) {
			t.Errorf("mensagem deveria conter %q, veio:\n%s", querConter, msg)
		}
	}
	// A comissão é margem dela — NUNCA pode aparecer na mensagem ao público.
	for _, naoQuer := range []string{"55", "0.55", "comiss", "Comiss"} {
		if strings.Contains(msg, naoQuer) {
			t.Errorf("mensagem NÃO deveria conter %q (vazou comissão):\n%s", naoQuer, msg)
		}
	}
}

func TestMensagemSemPrecoSemLink(t *testing.T) {
	o := Oferta{Nome: "Só o nome"}
	msg := o.Mensagem()
	if strings.Contains(msg, "R$") {
		t.Errorf("sem preço não deveria ter 'R$': %q", msg)
	}
	if !strings.Contains(msg, "Só o nome") {
		t.Errorf("deveria conter o nome: %q", msg)
	}
}

func TestMockPublica(t *testing.T) {
	m := NovoMock("telegram")
	if m.Nome() != "mock:telegram" {
		t.Errorf("Nome()=%q", m.Nome())
	}
	o := Oferta{Nome: "X", Preco: 10, Link: "l"}
	res, err := m.Publicar(context.Background(), o)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Enviado || res.Canal != "telegram" {
		t.Errorf("resultado inesperado: %+v", res)
	}
	if res.Mensagem != o.Mensagem() {
		t.Errorf("mensagem do resultado diverge da composta")
	}
	if !strings.Contains(res.Detalhe, "mock") {
		t.Errorf("detalhe deveria deixar claro que é mock: %q", res.Detalhe)
	}
}

func TestMockCanalPadrao(t *testing.T) {
	if NovoMock("").Nome() != "mock:telegram" {
		t.Error("canal vazio deveria virar telegram")
	}
}

func TestSubIDComposicao(t *testing.T) {
	id := SubID("Telegram", "nicho", mustTime())
	if id != "telegram_nicho_20260315" {
		t.Errorf("subId inesperado: %q", id)
	}
}

func TestSubIDSanitiza(t *testing.T) {
	// acentos, espaços e símbolos saem; fica só [a-z0-9] em cada parte
	id := SubID("Insta Gram!", "bem-estar", mustTime())
	if id != "instagram_bemestar_20260315" {
		t.Errorf("subId não sanitizado: %q", id)
	}
}

func mustTime() time.Time {
	return time.Date(2026, 3, 15, 10, 0, 0, 0, time.UTC)
}
