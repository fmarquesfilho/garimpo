//go:build gcp

package store

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/fmarquesfilho/garimpo/internal/publish"
	"google.golang.org/api/iterator"
)

// ─── Templates (BigQuery) ─────────────────────────────────────────────────

type linhaTemplateBQ struct {
	ID      string    `bigquery:"id"`
	Nome    string    `bigquery:"nome"`
	Corpo   string    `bigquery:"corpo"`
	ComFoto bool      `bigquery:"com_foto"`
	Ativo   bool      `bigquery:"ativo"`
	SalvoEm time.Time `bigquery:"salvo_em"`
}

// BQTemplateStore implementa publish.TemplateStore com BigQuery.
type BQTemplateStore struct {
	client  *bigquery.Client
	dataset string
}

func NovoBQTemplateStore(client *bigquery.Client, dataset string) *BQTemplateStore {
	return &BQTemplateStore{client: client, dataset: dataset}
}

func (s *BQTemplateStore) Listar(ctx context.Context) ([]publish.Template, error) {
	q := s.client.Query(`
		WITH ranked AS (
		  SELECT *, ROW_NUMBER() OVER (PARTITION BY id ORDER BY salvo_em DESC) AS rn
		  FROM ` + "`" + s.dataset + ".templates`" + `
		)
		SELECT id, nome, corpo, com_foto, ativo
		FROM ranked WHERE rn = 1 AND ativo = TRUE
		ORDER BY nome
	`)
	it, err := q.Read(ctx)
	if err != nil {
		return nil, err
	}
	var out []publish.Template
	for {
		var r struct {
			ID      string `bigquery:"id"`
			Nome    string `bigquery:"nome"`
			Corpo   string `bigquery:"corpo"`
			ComFoto bool   `bigquery:"com_foto"`
			Ativo   bool   `bigquery:"ativo"`
		}
		err := it.Next(&r)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		out = append(out, publish.Template{ID: r.ID, Nome: r.Nome, Corpo: r.Corpo, ComFoto: r.ComFoto, Ativo: r.Ativo})
	}
	return out, nil
}

func (s *BQTemplateStore) Buscar(ctx context.Context, id string) (publish.Template, error) {
	lista, err := s.Listar(ctx)
	if err != nil {
		return publish.Template{}, err
	}
	for _, t := range lista {
		if t.ID == id {
			return t, nil
		}
	}
	return publish.Template{}, fmt.Errorf("template %q não encontrado", id)
}

func (s *BQTemplateStore) Salvar(ctx context.Context, t publish.Template) error {
	row := linhaTemplateBQ{
		ID: t.ID, Nome: t.Nome, Corpo: t.Corpo, ComFoto: t.ComFoto,
		Ativo: t.Ativo, SalvoEm: time.Now().UTC(),
	}
	return s.client.Dataset(s.dataset).Table("templates").Inserter().Put(ctx, row)
}

func (s *BQTemplateStore) Deletar(ctx context.Context, id string) error {
	row := linhaTemplateBQ{ID: id, Ativo: false, SalvoEm: time.Now().UTC()}
	return s.client.Dataset(s.dataset).Table("templates").Inserter().Put(ctx, row)
}
