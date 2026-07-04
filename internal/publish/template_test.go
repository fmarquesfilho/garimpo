package publish

import (
	"context"
	"testing"
)

func TestTemplateRenderizar(t *testing.T) {
	tmpl := Template{
		Corpo: "✨ <b>{{nome}}</b>\n💸 {{preco}}\n📂 {{categoria}}\n🎯 {{estrategia}}\n🛒 {{link}}",
	}
	oferta := Oferta{
		Nome: "Sérum", Preco: 49.90, Categoria: "Beleza",
		Estrategia: "nicho", Link: "https://example.com",
	}
	got := tmpl.Renderizar(oferta)
	expects := []string{
		"<b>Sérum</b>",
		"R$ 49.90",
		"Beleza",
		"nicho",
		"https://example.com",
	}
	for _, e := range expects {
		if !contains(got, e) {
			t.Errorf("renderizado não contém %q:\n%s", e, got)
		}
	}
}

func TestTemplateRenderizarSemCampos(t *testing.T) {
	tmpl := Template{Corpo: "Oferta: {{nome}} por {{preco}}"}
	oferta := Oferta{Nome: "Produto", Preco: 0}
	got := tmpl.Renderizar(oferta)
	if !contains(got, "Produto") {
		t.Errorf("deveria conter o nome: %q", got)
	}
	if !contains(got, "R$ 0.00") {
		t.Errorf("preço zero deveria renderizar como R$ 0.00: %q", got)
	}
}

func TestMemTemplateStoreCRUD(t *testing.T) {
	ctx := context.Background()
	s := NovoMemTemplateStore()

	// Listar (templates padrão embutidos)
	lista, err := s.Listar(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(lista) < 2 {
		t.Errorf("deveria ter ao menos 2 templates padrão, veio %d", len(lista))
	}

	// Buscar existente
	tmpl, err := s.Buscar(ctx, "padrao")
	if err != nil {
		t.Fatal(err)
	}
	if tmpl.Nome != "Padrão" {
		t.Errorf("esperava nome=Padrão, veio %q", tmpl.Nome)
	}

	// Buscar inexistente
	_, err = s.Buscar(ctx, "inexistente")
	if err == nil {
		t.Error("deveria retornar erro para template inexistente")
	}

	// Salvar novo
	novo := Template{ID: "promo", Nome: "Promoção", Corpo: "🔥 {{nome}}", Ativo: true}
	if err := s.Salvar(ctx, novo); err != nil {
		t.Fatal(err)
	}
	lista, _ = s.Listar(ctx)
	if len(lista) < 3 {
		t.Errorf("após salvar deveria ter 3+, veio %d", len(lista))
	}

	// Deletar — remove template and verify it's no longer retrievable
	if err := s.Deletar(ctx, "promo"); err != nil {
		t.Fatal(err)
	}
	_, err = s.Buscar(ctx, "promo")
	if err == nil {
		t.Error("após deletar deveria dar erro")
	}
}

func TestTemplateComFotoFlag(t *testing.T) {
	foto := Template{Corpo: "{{nome}}", ComFoto: true}
	texto := Template{Corpo: "{{nome}}", ComFoto: false}

	if !foto.ComFoto {
		t.Error("foto deveria ter ComFoto=true")
	}
	if texto.ComFoto {
		t.Error("texto deveria ter ComFoto=false")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
