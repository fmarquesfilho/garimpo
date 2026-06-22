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
