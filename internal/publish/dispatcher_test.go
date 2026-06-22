package publish

import (
	"context"
	"testing"
)

func TestDispatcherDestinoInativo(t *testing.T) {
	spy := &spySender{tipo: "telegram"}
	store := NovoMemDestinoStore()
	_ = store.Salvar(context.Background(), Destino{
		ID: "inativo", Nome: "Canal Inativo", Tipo: "telegram", Config: "@x", Ativo: false,
	})

	d := NovoDispatcher(DispatcherConfig{
		Destinos: store, TipoPadrao: "telegram", ConfigPadrao: "@padrao",
	}, spy)

	_, err := d.Publicar(context.Background(), Oferta{Nome: "Test", DestinoID: "inativo"})
	if err == nil {
		t.Error("deveria retornar erro para destino inativo")
	}
}

func TestDispatcherDestinoNaoEncontrado(t *testing.T) {
	spy := &spySender{tipo: "telegram"}
	store := NovoMemDestinoStore()

	d := NovoDispatcher(DispatcherConfig{
		Destinos: store, TipoPadrao: "telegram", ConfigPadrao: "@padrao",
	}, spy)

	_, err := d.Publicar(context.Background(), Oferta{Nome: "Test", DestinoID: "inexistente"})
	if err == nil {
		t.Error("deveria retornar erro para destino inexistente")
	}
}

func TestDispatcherSemConfigPadrao(t *testing.T) {
	spy := &spySender{tipo: "telegram"}
	store := NovoMemDestinoStore()

	d := NovoDispatcher(DispatcherConfig{
		Destinos: store, TipoPadrao: "telegram", ConfigPadrao: "", // vazio
	}, spy)

	_, err := d.Publicar(context.Background(), Oferta{Nome: "Test"}) // sem DestinoID
	if err == nil {
		t.Error("deveria retornar erro quando não há config padrão e DestinoID está vazio")
	}
}

func TestDispatcherProvedorNaoRegistrado(t *testing.T) {
	spy := &spySender{tipo: "telegram"} // só telegram registrado
	store := NovoMemDestinoStore()
	_ = store.Salvar(context.Background(), Destino{
		ID: "whats", Nome: "WhatsApp", Tipo: "whatsapp", Config: "+5511", Ativo: true,
	})

	d := NovoDispatcher(DispatcherConfig{
		Destinos: store, TipoPadrao: "telegram", ConfigPadrao: "@padrao",
	}, spy)

	_, err := d.Publicar(context.Background(), Oferta{Nome: "Test", DestinoID: "whats"})
	if err == nil {
		t.Error("deveria retornar erro para provedor whatsapp não registrado")
	}
}

func TestMemDestinoStoreCRUD(t *testing.T) {
	ctx := context.Background()
	s := NovoMemDestinoStore()

	// Listar vazio
	lista, _ := s.Listar(ctx)
	if len(lista) != 0 {
		t.Errorf("deveria estar vazio, veio %d", len(lista))
	}

	// Salvar
	_ = s.Salvar(ctx, Destino{ID: "a", Nome: "A", Tipo: "telegram", Config: "@a", Ativo: true})
	_ = s.Salvar(ctx, Destino{ID: "b", Nome: "B", Tipo: "telegram", Config: "@b", Ativo: false})

	lista, _ = s.Listar(ctx)
	if len(lista) != 1 { // b é inativo
		t.Errorf("esperava 1 ativo, veio %d", len(lista))
	}

	// Buscar
	d, err := s.Buscar(ctx, "a")
	if err != nil || d.Nome != "A" {
		t.Errorf("buscar 'a' falhou: %v %+v", err, d)
	}

	// Buscar inexistente
	_, err = s.Buscar(ctx, "z")
	if err == nil {
		t.Error("deveria dar erro para destino inexistente")
	}

	// Deletar
	_ = s.Deletar(ctx, "a")
	lista, _ = s.Listar(ctx)
	if len(lista) != 0 {
		t.Errorf("após deletar deveria estar vazio, veio %d", len(lista))
	}
}
