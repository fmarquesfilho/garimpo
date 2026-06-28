//go:build gcp

package store

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/fmarquesfilho/garimpo/internal/apperr"
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

// BQTemplateStore implementa TemplateRepo com BigQuery.
type BQTemplateStore struct {
	client  *bigquery.Client
	dataset string
}

func NovoBQTemplateStore(client *bigquery.Client, dataset string) *BQTemplateStore {
	return &BQTemplateStore{client: client, dataset: dataset}
}

func (s *BQTemplateStore) ListarTemplates(ctx context.Context) ([]Template, error) {
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
		return nil, fmt.Errorf("bq listar templates: %w", err)
	}
	var out []Template
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
			return nil, fmt.Errorf("bq templates next: %w", err)
		}
		out = append(out, Template{ID: r.ID, Nome: r.Nome, Corpo: r.Corpo, ComFoto: r.ComFoto, Ativo: r.Ativo})
	}
	return out, nil
}

func (s *BQTemplateStore) BuscarTemplate(ctx context.Context, id string) (Template, error) {
	lista, err := s.ListarTemplates(ctx)
	if err != nil {
		return Template{}, err
	}
	for _, t := range lista {
		if t.ID == id {
			return t, nil
		}
	}
	return Template{}, fmt.Errorf("template %q: %w", id, apperr.ErrNotFound)
}

func (s *BQTemplateStore) SalvarTemplate(ctx context.Context, t Template) error {
	row := linhaTemplateBQ{
		ID: t.ID, Nome: t.Nome, Corpo: t.Corpo, ComFoto: t.ComFoto,
		Ativo: t.Ativo, SalvoEm: time.Now().UTC(),
	}
	return s.client.Dataset(s.dataset).Table("templates").Inserter().Put(ctx, row)
}

func (s *BQTemplateStore) DeletarTemplate(ctx context.Context, id string) error {
	row := linhaTemplateBQ{ID: id, Ativo: false, SalvoEm: time.Now().UTC()}
	return s.client.Dataset(s.dataset).Table("templates").Inserter().Put(ctx, row)
}
