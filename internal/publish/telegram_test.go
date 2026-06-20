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

func TestTelegramPublicaOK(t *testing.T) {
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

	tg := NovoTelegram("TOKEN", "@canal")
	tg.apiBase = srv.URL

	res, err := tg.Publicar(context.Background(), Oferta{Nome: "Sérum", Preco: 89.9, Link: "http://l"})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Enviado || res.Canal != "telegram" {
		t.Errorf("resultado inesperado: %+v", res)
	}
	// o corpo enviado precisa ter chat_id e o texto com o nome do produto
	if recebido["chat_id"] != "@canal" {
		t.Errorf("chat_id errado: %v", recebido["chat_id"])
	}
	if txt, _ := recebido["text"].(string); !strings.Contains(txt, "Sérum") {
		t.Errorf("text não contém o produto: %q", txt)
	}
}

func TestTelegramPropagaErro(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"ok":false,"description":"chat not found"}`))
	}))
	defer srv.Close()

	tg := NovoTelegram("TOKEN", "@canal")
	tg.apiBase = srv.URL

	res, err := tg.Publicar(context.Background(), Oferta{Nome: "X"})
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

func TestNovoEscolheTelegramComEnv(t *testing.T) {
	t.Setenv("TELEGRAM_BOT_TOKEN", "tok")
	t.Setenv("TELEGRAM_CHAT_ID", "chat")
	if _, ok := Novo().(*TelegramPublicador); !ok {
		t.Error("com env deveria escolher o Telegram")
	}
}
