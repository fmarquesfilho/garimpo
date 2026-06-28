package publish

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/fmarquesfilho/garimpo/internal/apperr"
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

// Renderizar aplica os dados da oferta nos placeholders do template.
func (t Template) Renderizar(o Oferta) string {
	r := strings.NewReplacer(
		"{{nome}}", strings.TrimSpace(o.Nome),
		"{{preco}}", fmt.Sprintf("R$ %.2f", o.Preco),
		"{{categoria}}", o.Categoria,
		"{{estrategia}}", o.Estrategia,
		"{{link}}", o.Link,
	)
	return r.Replace(t.Corpo)
}

// TemplateStore persiste templates. Implementações: MemTemplateStore (dev),
// BigQuery (produção).
type TemplateStore interface {
	Listar(ctx context.Context) ([]Template, error)
	Buscar(ctx context.Context, id string) (Template, error)
	Salvar(ctx context.Context, t Template) error
	Deletar(ctx context.Context, id string) error
}

// MemTemplateStore é uma implementação em memória (dev/testes).
type MemTemplateStore struct {
	mu        sync.RWMutex
	templates map[string]Template
}

func NovoMemTemplateStore() *MemTemplateStore {
	m := &MemTemplateStore{templates: make(map[string]Template)}
	// Template padrão embutido
	m.templates["padrao"] = Template{
		ID:       "padrao",
		Nome:     "Padrão",
		Corpo:    "✨ <b>{{nome}}</b>\n📂 <i>{{categoria}}</i>\n💸 <b>{{preco}}</b>\n🎯 {{estrategia}}",
		ComFoto:  false,
		Ativo:    true,
		CriadoEm: time.Now().UTC().Format(time.RFC3339),
	}
	m.templates["foto"] = Template{
		ID:       "foto",
		Nome:     "Com foto",
		Corpo:    "✨ <b>{{nome}}</b>\n💸 <b>{{preco}}</b>",
		ComFoto:  true,
		Ativo:    true,
		CriadoEm: time.Now().UTC().Format(time.RFC3339),
	}
	return m
}

func (m *MemTemplateStore) Listar(_ context.Context) ([]Template, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	lista := make([]Template, 0, len(m.templates))
	for _, t := range m.templates {
		if t.Ativo {
			lista = append(lista, t)
		}
	}
	return lista, nil
}

func (m *MemTemplateStore) Buscar(_ context.Context, id string) (Template, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	t, ok := m.templates[id]
	if !ok {
		return Template{}, fmt.Errorf("template %q: %w", id, apperr.ErrNotFound)
	}
	return t, nil
}

func (m *MemTemplateStore) Salvar(_ context.Context, t Template) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.templates[t.ID] = t
	return nil
}

func (m *MemTemplateStore) Deletar(_ context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.templates, id)
	return nil
}
