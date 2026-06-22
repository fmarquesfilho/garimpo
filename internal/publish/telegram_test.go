package publish

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTelegramPublicaComHTMLeBotao(t *testing.T) {
	var recebido map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/sendMessage") {
			t.Errorf("path inesperado: %s", r.URL.Path)
		}
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &recebido)
		w.Write([]byte(`{"ok":true,"result":{"message_id":1}}`))
	}))
	defer srv.Close()

	sender := NovoTelegramSender("TOKEN")
	sender.apiBase = srv.URL

	oferta := Oferta{
		Nome:       "Sérum Vitamina C",
		Categoria:  "Beleza",
		Preco:      89.90,
		Link:       "https://shope.ee/abc123",
		Estrategia: "nicho",
	}

	res, err := sender.Enviar(context.Background(), oferta, "@canal")
	if err != nil {
		t.Fatal(err)
	}
	if !res.Enviado || res.Canal != "telegram" {
		t.Errorf("resultado inesperado: %+v", res)
	}

	// Verifica parse_mode HTML
	if pm, _ := recebido["parse_mode"].(string); pm != "HTML" {
		t.Errorf("esperava parse_mode=HTML, veio: %q", pm)
	}

	// Verifica que o texto contém tags HTML
	txt, _ := recebido["text"].(string)
	if !strings.Contains(txt, "<b>Sérum Vitamina C</b>") {
		t.Errorf("texto deveria conter nome em negrito, veio: %q", txt)
	}
	if !strings.Contains(txt, "<i>Beleza</i>") {
		t.Errorf("texto deveria conter categoria em itálico, veio: %q", txt)
	}
	if !strings.Contains(txt, "<b>R$ 89.90</b>") {
		t.Errorf("texto deveria conter preço em negrito, veio: %q", txt)
	}

	// Verifica chat_id
	if recebido["chat_id"] != "@canal" {
		t.Errorf("chat_id errado: %v", recebido["chat_id"])
	}

	// Verifica reply_markup com botão inline
	rm, ok := recebido["reply_markup"].(map[string]any)
	if !ok {
		t.Fatal("reply_markup ausente ou tipo errado")
	}
	kb, ok := rm["inline_keyboard"].([]any)
	if !ok || len(kb) == 0 {
		t.Fatal("inline_keyboard ausente ou vazio")
	}
	row, ok := kb[0].([]any)
	if !ok || len(row) == 0 {
		t.Fatal("primeira row do inline_keyboard vazia")
	}
	btn, ok := row[0].(map[string]any)
	if !ok {
		t.Fatal("botão não é um mapa")
	}
	if btn["text"] != "🛒 Comprar" {
		t.Errorf("texto do botão errado: %v", btn["text"])
	}
	if btn["url"] != "https://shope.ee/abc123" {
		t.Errorf("url do botão errado: %v", btn["url"])
	}
}

func TestTelegramSemLinkNaoEnviaBotao(t *testing.T) {
	var recebido map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &recebido)
		w.Write([]byte(`{"ok":true,"result":{"message_id":2}}`))
	}))
	defer srv.Close()

	sender := NovoTelegramSender("TOKEN")
	sender.apiBase = srv.URL

	oferta := Oferta{Nome: "Produto Sem Link", Preco: 10.0}
	_, err := sender.Enviar(context.Background(), oferta, "@canal")
	if err != nil {
		t.Fatal(err)
	}

	// Sem link → sem reply_markup
	if _, ok := recebido["reply_markup"]; ok {
		t.Error("não deveria enviar reply_markup sem link")
	}
}

func TestTelegramPropagaErro(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"ok":false,"description":"chat not found"}`))
	}))
	defer srv.Close()

	sender := NovoTelegramSender("TOKEN")
	sender.apiBase = srv.URL

	res, err := sender.Enviar(context.Background(), Oferta{Nome: "X"}, "@canal")
	if err == nil || !strings.Contains(err.Error(), "chat not found") {
		t.Errorf("esperava erro com a descrição da API, veio: %v", err)
	}
	if res.Enviado {
		t.Error("não deveria marcar enviado em erro")
	}
}

func TestNovoEscolheMockSemEnv(t *testing.T) {
	t.Setenv("TELEGRAM_BOT_TOKEN", "")
	t.Setenv("TELEGRAM_CHAT_ID", "")
	if _, ok := Novo().(*MockPublicador); !ok {
		t.Error("sem env deveria escolher o Mock")
	}
}

func TestNovoEscolheDispatcherComEnv(t *testing.T) {
	t.Setenv("TELEGRAM_BOT_TOKEN", "tok")
	t.Setenv("TELEGRAM_CHAT_ID", "") // chat vazio — Dispatcher funciona mesmo assim
	if _, ok := Novo().(*Dispatcher); !ok {
		t.Error("com token deveria escolher o Dispatcher (mesmo sem CHAT_ID)")
	}
}

func TestDispatcherRoteiaParaDestinoCorreto(t *testing.T) {
	var configRecebida string
	mockSender := &spySender{tipo: "telegram"}
	store := NovoMemDestinoStore()
	_ = store.Salvar(context.Background(), Destino{
		ID: "beleza", Nome: "Ofertas Beleza", Tipo: "telegram", Config: "@beleza_ofertas", Ativo: true,
	})

	d := NovoDispatcher(DispatcherConfig{
		Destinos: store, TipoPadrao: "telegram", ConfigPadrao: "@padrao",
	}, mockSender)

	// Publica para destino específico
	_, _ = d.Publicar(context.Background(), Oferta{Nome: "Test", DestinoID: "beleza"})
	configRecebida = mockSender.ultimaConfig
	if configRecebida != "@beleza_ofertas" {
		t.Errorf("esperava config=@beleza_ofertas, veio: %q", configRecebida)
	}

	// Publica para o padrão
	_, _ = d.Publicar(context.Background(), Oferta{Nome: "Test"})
	configRecebida = mockSender.ultimaConfig
	if configRecebida != "@padrao" {
		t.Errorf("esperava config=@padrao, veio: %q", configRecebida)
	}
}

// spySender registra a última config usada para verificação.
type spySender struct {
	tipo         string
	ultimaConfig string
}

func (s *spySender) Tipo() string { return s.tipo }
func (s *spySender) Enviar(_ context.Context, o Oferta, config string) (Resultado, error) {
	s.ultimaConfig = config
	return Resultado{Canal: s.tipo, Enviado: true, Mensagem: o.Mensagem(), Detalhe: "spy"}, nil
}
