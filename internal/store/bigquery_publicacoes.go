//go:build gcp

package store

import (
	"context"
	"time"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
)

// ─── Publicações ──────────────────────────────────────────────────────────

type linhaPublicacaoBQ struct {
	ID         string    `bigquery:"id"`
	ProdutoID  string    `bigquery:"produto_id"`
	Nome       string    `bigquery:"nome"`
	Categoria  string    `bigquery:"categoria"`
	Preco      float64   `bigquery:"preco"`
	Comissao   float64   `bigquery:"comissao"`
	Link       string    `bigquery:"link"`
	Imagem     string    `bigquery:"imagem"`
	Estrategia string    `bigquery:"estrategia"`
	DestinoID  string    `bigquery:"destino_id"`
	TemplateID string    `bigquery:"template_id"`
	AgendadaEm string    `bigquery:"agendada_em"`
	Status     string    `bigquery:"status"`
	Detalhe    string    `bigquery:"detalhe"`
	CriadaEm   time.Time `bigquery:"criada_em"`
	EnviadaEm  string    `bigquery:"enviada_em"`
	OwnerUID   string    `bigquery:"owner_uid"`
}

func (s *BigQueryStore) SalvarPublicacao(ctx context.Context, p Publicacao) error {
	criadaEm := time.Now().UTC()
	row := linhaPublicacaoBQ{
		ID: p.ID, ProdutoID: p.ProdutoID, Nome: p.Nome, Categoria: p.Categoria,
		Preco: p.Preco, Comissao: p.Comissao, Link: p.Link, Imagem: p.Imagem,
		Estrategia: p.Estrategia, DestinoID: p.DestinoID, TemplateID: p.TemplateID,
		AgendadaEm: p.AgendadaEm, Status: p.Status, Detalhe: p.Detalhe,
		CriadaEm: criadaEm, EnviadaEm: p.EnviadaEm, OwnerUID: p.OwnerUID,
	}
	return s.client.Dataset(s.dataset).Table("publicacoes").Inserter().Put(ctx, row)
}

func (s *BigQueryStore) ListarPublicacoes(ctx context.Context, status string) ([]Publicacao, error) {
	filtro := ""
	if status != "" {
		filtro = " AND status = @status"
	}
	q := s.client.Query(`
		WITH ranked AS (
		  SELECT *, ROW_NUMBER() OVER (PARTITION BY id ORDER BY criada_em DESC) AS rn
		  FROM ` + "`" + s.dataset + ".publicacoes`" + `
		  WHERE criada_em >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 30 DAY)
		)
		SELECT id, produto_id, nome, categoria, preco, comissao, link, imagem,
		       estrategia, destino_id, template_id, agendada_em, status, detalhe,
		       criada_em, enviada_em, owner_uid
		FROM ranked WHERE rn = 1` + filtro + `
		ORDER BY criada_em DESC
		LIMIT 200
	`)
	if status != "" {
		q.Parameters = []bigquery.QueryParameter{{Name: "status", Value: status}}
	}
	it, err := q.Read(ctx)
	if err != nil {
		return nil, err
	}
	var out []Publicacao
	for {
		var r linhaPublicacaoBQ
		err := it.Next(&r)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		out = append(out, Publicacao{
			ID: r.ID, ProdutoID: r.ProdutoID, Nome: r.Nome, Categoria: r.Categoria,
			Preco: r.Preco, Comissao: r.Comissao, Link: r.Link, Imagem: r.Imagem,
			Estrategia: r.Estrategia, DestinoID: r.DestinoID, TemplateID: r.TemplateID,
			AgendadaEm: r.AgendadaEm, Status: r.Status, Detalhe: r.Detalhe,
			CriadaEm: r.CriadaEm.Format(time.RFC3339), EnviadaEm: r.EnviadaEm, OwnerUID: r.OwnerUID,
		})
	}
	return out, nil
}

func (s *BigQueryStore) AtualizarPublicacao(ctx context.Context, id, status, detalhe string) error {
	row := linhaPublicacaoBQ{
		ID: id, Status: status, Detalhe: detalhe, CriadaEm: time.Now().UTC(),
	}
	return s.client.Dataset(s.dataset).Table("publicacoes").Inserter().Put(ctx, row)
}
