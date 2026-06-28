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

// ─── Destinos (BigQuery) ──────────────────────────────────────────────────

type linhaDestinoBQ struct {
	ID      string    `bigquery:"id"`
	Nome    string    `bigquery:"nome"`
	Tipo    string    `bigquery:"tipo"`
	Config  string    `bigquery:"config"`
	Ativo   bool      `bigquery:"ativo"`
	SalvoEm time.Time `bigquery:"salvo_em"`
}

// BQDestinoStore implementa DestinoRepo com BigQuery.
type BQDestinoStore struct {
	client  *bigquery.Client
	dataset string
}

func NovoBQDestinoStore(client *bigquery.Client, dataset string) *BQDestinoStore {
	return &BQDestinoStore{client: client, dataset: dataset}
}

func (s *BQDestinoStore) ListarDestinos(ctx context.Context) ([]Destino, error) {
	q := s.client.Query(`
		WITH ranked AS (
		  SELECT *, ROW_NUMBER() OVER (PARTITION BY id ORDER BY salvo_em DESC) AS rn
		  FROM ` + "`" + s.dataset + ".destinos`" + `
		)
		SELECT id, nome, tipo, config, ativo
		FROM ranked WHERE rn = 1 AND ativo = TRUE
		ORDER BY nome
	`)
	it, err := q.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("bq listar destinos: %w", err)
	}
	var out []Destino
	for {
		var r struct {
			ID     string `bigquery:"id"`
			Nome   string `bigquery:"nome"`
			Tipo   string `bigquery:"tipo"`
			Config string `bigquery:"config"`
			Ativo  bool   `bigquery:"ativo"`
		}
		err := it.Next(&r)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("bq destinos next: %w", err)
		}
		out = append(out, Destino{ID: r.ID, Nome: r.Nome, Tipo: r.Tipo, Config: r.Config, Ativo: r.Ativo})
	}
	return out, nil
}

func (s *BQDestinoStore) BuscarDestino(ctx context.Context, id string) (Destino, error) {
	lista, err := s.ListarDestinos(ctx)
	if err != nil {
		return Destino{}, err
	}
	for _, d := range lista {
		if d.ID == id {
			return d, nil
		}
	}
	return Destino{}, fmt.Errorf("destino %q: %w", id, apperr.ErrNotFound)
}

func (s *BQDestinoStore) SalvarDestino(ctx context.Context, d Destino) error {
	row := linhaDestinoBQ{
		ID: d.ID, Nome: d.Nome, Tipo: d.Tipo, Config: d.Config,
		Ativo: d.Ativo, SalvoEm: time.Now().UTC(),
	}
	return s.client.Dataset(s.dataset).Table("destinos").Inserter().Put(ctx, row)
}

func (s *BQDestinoStore) DeletarDestino(ctx context.Context, id string) error {
	// Append-only tombstone
	row := linhaDestinoBQ{ID: id, Ativo: false, SalvoEm: time.Now().UTC()}
	return s.client.Dataset(s.dataset).Table("destinos").Inserter().Put(ctx, row)
}
