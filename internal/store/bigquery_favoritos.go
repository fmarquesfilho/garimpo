//go:build gcp

package store

import (
	"context"
	"time"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
)

type linhaFavoritoBQ struct {
	ProdutoID string    `bigquery:"produto_id"`
	Nome      string    `bigquery:"nome"`
	Preco     float64   `bigquery:"preco"`
	Comissao  float64   `bigquery:"comissao"`
	Link      string    `bigquery:"link"`
	Imagem    string    `bigquery:"imagem"`
	Loja      string    `bigquery:"loja"`
	Categoria string    `bigquery:"categoria"`
	Origem    string    `bigquery:"origem"`
	Ativo     bool      `bigquery:"ativo"`
	OwnerUID  string    `bigquery:"owner_uid"`
	SalvoEm   time.Time `bigquery:"salvo_em"`
}

func (s *BigQueryStore) SalvarFavorito(ctx context.Context, f Favorito) error {
	row := linhaFavoritoBQ{
		ProdutoID: f.ProdutoID, Nome: f.Nome, Preco: f.Preco,
		Comissao: f.Comissao, Link: f.Link, Imagem: f.Imagem,
		Loja: f.Loja, Categoria: f.Categoria, Origem: f.Origem,
		Ativo: true, OwnerUID: f.OwnerUID, SalvoEm: time.Now().UTC(),
	}
	return s.client.Dataset(s.dataset).Table("favoritos").Inserter().Put(ctx, row)
}

func (s *BigQueryStore) ListarFavoritos(ctx context.Context, ownerUID string) ([]Favorito, error) {
	q := s.client.Query(`
		WITH ranked AS (
		  SELECT *, ROW_NUMBER() OVER (PARTITION BY produto_id ORDER BY salvo_em DESC) AS rn
		  FROM ` + "`" + s.dataset + ".favoritos`" + `
		  WHERE owner_uid = @uid
		)
		SELECT produto_id, nome, preco, comissao, link, imagem, loja, categoria, origem, ativo, salvo_em
		FROM ranked WHERE rn = 1 AND ativo = TRUE
		ORDER BY salvo_em DESC
		LIMIT 200
	`)
	q.Parameters = []bigquery.QueryParameter{{Name: "uid", Value: ownerUID}}

	it, err := q.Read(ctx)
	if err != nil {
		return nil, err
	}
	var out []Favorito
	for {
		var r linhaFavoritoBQ
		err := it.Next(&r)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		out = append(out, Favorito{
			ProdutoID: r.ProdutoID, Nome: r.Nome, Preco: r.Preco,
			Comissao: r.Comissao, Link: r.Link, Imagem: r.Imagem,
			Loja: r.Loja, Categoria: r.Categoria, Origem: r.Origem,
			SalvoEm: r.SalvoEm, OwnerUID: r.OwnerUID, Ativo: r.Ativo,
		})
	}
	return out, nil
}

func (s *BigQueryStore) RemoverFavorito(ctx context.Context, ownerUID, produtoID string) error {
	// Append-only tombstone
	row := linhaFavoritoBQ{
		ProdutoID: produtoID, OwnerUID: ownerUID,
		Ativo: false, SalvoEm: time.Now().UTC(),
	}
	return s.client.Dataset(s.dataset).Table("favoritos").Inserter().Put(ctx, row)
}
