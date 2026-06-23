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

func TestTelegramEnviaFotoQuandoTemImagem(t *testing.T) {
	var endpoint string
	var recebido map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		endpoint = r.URL.Path
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &recebido)
		w.Write([]byte(`{"ok":true,"result":{"message_id":1}}`))
	}))
	defer srv.Close()

	sender := NovoTelegramSender("TOKEN")
	sender.apiBase = srv.URL

	oferta := Oferta{
		Nome:   "Sérum",
		Preco:  49.90,
		Link:   "https://shope.ee/abc",
		Imagem: "https://img.shopee.com/foto.jpg",
	}

	res, err := sender.Enviar(context.Background(), oferta, "@canal")
	if err != nil {
		t.Fatal(err)
	}
	if !res.Enviado {
		t.Error("deveria ter enviado")
	}
	if !strings.Contains(res.Detalhe, "foto") {
		t.Errorf("detalhe deveria mencionar foto: %q", res.Detalhe)
	}

	// Verifica que usou sendPhoto
	if !strings.HasSuffix(endpoint, "/sendPhoto") {
		t.Errorf("esperava sendPhoto, usou: %s", endpoint)
	}

	// Verifica campos do payload
	if recebido["photo"] != "https://img.shopee.com/foto.jpg" {
		t.Errorf("photo errado: %v", recebido["photo"])
	}
	if recebido["chat_id"] != "@canal" {
		t.Errorf("chat_id errado: %v", recebido["chat_id"])
	}
	if recebido["parse_mode"] != "HTML" {
		t.Errorf("parse_mode deveria ser HTML: %v", recebido["parse_mode"])
	}
	// caption deve conter o nome
	caption, _ := recebido["caption"].(string)
	if !strings.Contains(caption, "Sérum") {
		t.Errorf("caption deveria conter o nome: %q", caption)
	}
	// reply_markup com botão
	rm, ok := recebido["reply_markup"].(map[string]any)
	if !ok {
		t.Fatal("reply_markup ausente no sendPhoto")
	}
	kb, _ := rm["inline_keyboard"].([]any)
	if len(kb) == 0 {
		t.Error("inline_keyboard vazio no sendPhoto")
	}
}

func TestTelegramUsaSendMessageSemImagem(t *testing.T) {
	var endpoint string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		endpoint = r.URL.Path
		w.Write([]byte(`{"ok":true,"result":{"message_id":1}}`))
	}))
	defer srv.Close()

	sender := NovoTelegramSender("TOKEN")
	sender.apiBase = srv.URL

	oferta := Oferta{Nome: "Produto", Preco: 10.0, Imagem: ""} // sem imagem
	_, err := sender.Enviar(context.Background(), oferta, "@canal")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(endpoint, "/sendMessage") {
		t.Errorf("sem imagem deveria usar sendMessage, usou: %s", endpoint)
	}
}

func TestTelegramUsaLegendaCustomQuandoFornecida(t *testing.T) {
	var recebido map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &recebido)
		w.Write([]byte(`{"ok":true,"result":{"message_id":1}}`))
	}))
	defer srv.Close()

	sender := NovoTelegramSender("TOKEN")
	sender.apiBase = srv.URL

	oferta := Oferta{
		Nome:        "Sérum",
		Preco:       49.90,
		Link:        "https://shope.ee/abc",
		LegendaHTML: "<b>Oferta imperdível!</b>\nSó hoje por R$ 49,90",
	}

	_, err := sender.Enviar(context.Background(), oferta, "@canal")
	if err != nil {
		t.Fatal(err)
	}

	// Deve usar a legenda custom, não a gerada por MensagemHTML
	txt, _ := recebido["text"].(string)
	if !strings.Contains(txt, "Oferta imperdível!") {
		t.Errorf("deveria usar legenda_custom, veio: %q", txt)
	}
	if strings.Contains(txt, "✨") {
		t.Errorf("não deveria usar MensagemHTML quando legenda_custom está preenchida: %q", txt)
	}
}

func TestTelegramUsaMensagemHTMLSemLegendaCustom(t *testing.T) {
	var recebido map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &recebido)
		w.Write([]byte(`{"ok":true,"result":{"message_id":1}}`))
	}))
	defer srv.Close()

	sender := NovoTelegramSender("TOKEN")
	sender.apiBase = srv.URL

	oferta := Oferta{
		Nome:  "Sérum",
		Preco: 49.90,
		Link:  "https://shope.ee/abc",
		// LegendaHTML vazia → usa MensagemHTML()
	}

	_, err := sender.Enviar(context.Background(), oferta, "@canal")
	if err != nil {
		t.Fatal(err)
	}

	txt, _ := recebido["text"].(string)
	if !strings.Contains(txt, "✨") {
		t.Errorf("sem legenda_custom deveria usar MensagemHTML (com ✨): %q", txt)
	}
}

func TestTelegramEnviaFotoComLegendaCustom(t *testing.T) {
	var recebido map[string]any
	var endpoint string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		endpoint = r.URL.Path
		b, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(b, &recebido)
		w.Write([]byte(`{"ok":true,"result":{"message_id":1}}`))
	}))
	defer srv.Close()

	sender := NovoTelegramSender("TOKEN")
	sender.apiBase = srv.URL

	oferta := Oferta{
		Nome:        "Sérum",
		Preco:       49.90,
		Link:        "https://shope.ee/abc",
		Imagem:      "https://img.shopee.com/foto.jpg",
		LegendaHTML: "<b>Com foto e legenda!</b>",
	}

	_, err := sender.Enviar(context.Background(), oferta, "@canal")
	if err != nil {
		t.Fatal(err)
	}

	// Deve usar sendPhoto (tem imagem)
	if !strings.HasSuffix(endpoint, "/sendPhoto") {
		t.Errorf("deveria usar sendPhoto, usou: %s", endpoint)
	}
	// Caption deve ser a legenda custom
	caption, _ := recebido["caption"].(string)
	if !strings.Contains(caption, "Com foto e legenda!") {
		t.Errorf("caption deveria ser a legenda_custom: %q", caption)
	}
}
