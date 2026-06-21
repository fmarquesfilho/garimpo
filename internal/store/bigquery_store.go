//go:build gcp

package store

import (
	"context"
	"time"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
)

// BigQueryStore grava eventos via streaming insert. Volume é baixo (decisões de
// curadoria), então inserts diretos bastam. Requer:
//
//	go get cloud.google.com/go/bigquery
//
// e credenciais (ADC) — no Cloud Run, a service account da revisão.
type BigQueryStore struct {
	client     *bigquery.Client
	dataset    string
	tabela     string // eventos
	tabelaSnap string // snapshots
}

func NovoBigQueryStore(ctx context.Context, projeto, dataset, tabela, tabelaSnap string) (*BigQueryStore, error) {
	c, err := bigquery.NewClient(ctx, projeto)
	if err != nil {
		return nil, err
	}
	return &BigQueryStore{client: c, dataset: dataset, tabela: tabela, tabelaSnap: tabelaSnap}, nil
}

func (s *BigQueryStore) Nome() string { return "bigquery" }

// linhaBQ mapeia o Evento para as colunas da tabela (ver deploy/bigquery_schema.sql).
type linhaBQ struct {
	Tipo       string    `bigquery:"tipo"`
	ProdutoID  string    `bigquery:"produto_id"`
	Nome       string    `bigquery:"nome"`
	Categoria  string    `bigquery:"categoria"`
	Estrategia string    `bigquery:"estrategia"`
	Canal      string    `bigquery:"canal"`
	SubID      string    `bigquery:"sub_id"`
	Comissao   float64   `bigquery:"comissao"`
	Preco      float64   `bigquery:"preco"`
	Vendas     int       `bigquery:"vendas"`
	Score      float64   `bigquery:"score"`
	Em         time.Time `bigquery:"em"`
}

func (s *BigQueryStore) Registrar(ctx context.Context, e Evento) error {
	if e.Em.IsZero() {
		e.Em = time.Now().UTC()
	}
	row := linhaBQ{
		Tipo:       e.Tipo,
		ProdutoID:  e.ProdutoID,
		Nome:       e.Nome,
		Categoria:  e.Categoria,
		Estrategia: e.Estrategia,
		Canal:      e.Canal,
		SubID:      e.SubID,
		Comissao:   e.Comissao,
		Preco:      e.Preco,
		Vendas:     e.Vendas,
		Score:      e.Score,
		Em:         e.Em,
	}
	return s.client.Dataset(s.dataset).Table(s.tabela).Inserter().Put(ctx, row)
}

// linhaSnapBQ mapeia cada item do snapshot para a tabela `snapshots`.
type linhaSnapBQ struct {
	ColetadoEm time.Time `bigquery:"coletado_em"`
	Categoria  string    `bigquery:"categoria"`
	Keyword    string    `bigquery:"keyword"`
	Estrategia string    `bigquery:"estrategia"`
	Posicao    int       `bigquery:"posicao"`
	ProdutoID  string    `bigquery:"produto_id"`
	Nome       string    `bigquery:"nome"`
	Preco      float64   `bigquery:"preco"`
	Comissao   float64   `bigquery:"comissao"`
	Vendas     int       `bigquery:"vendas"`
	Nota       float64   `bigquery:"nota"`
	Score      float64   `bigquery:"score"`
}

func (s *BigQueryStore) RegistrarSnapshot(ctx context.Context, snap Snapshot) error {
	if len(snap.Itens) == 0 {
		return nil
	}
	em := snap.Em
	if em.IsZero() {
		em = time.Now().UTC()
	}
	linhas := make([]linhaSnapBQ, 0, len(snap.Itens))
	for _, it := range snap.Itens {
		linhas = append(linhas, linhaSnapBQ{
			ColetadoEm: em,
			Categoria:  snap.Categoria,
			Keyword:    snap.Keyword,
			Estrategia: snap.Estrategia,
			Posicao:    it.Posicao,
			ProdutoID:  it.ProdutoID,
			Nome:       it.Nome,
			Preco:      it.Preco,
			Comissao:   it.Comissao,
			Vendas:     it.Vendas,
			Nota:       it.Nota,
			Score:      it.Score,
		})
	}
	return s.client.Dataset(s.dataset).Table(s.tabelaSnap).Inserter().Put(ctx, linhas)
}

// linhaBuscaBQ mapeia a Busca para a tabela `buscas` (append-only/versionada).
type linhaBuscaBQ struct {
	Nome        string    `bigquery:"nome"`
	Keyword     string    `bigquery:"keyword"`
	Categoria   string    `bigquery:"categoria"`
	Estrategia  string    `bigquery:"estrategia"`
	ComissaoMin float64   `bigquery:"comissao_min"`
	VendasMin   int       `bigquery:"vendas_min"`
	NotaMin     float64   `bigquery:"nota_min"`
	Top         int       `bigquery:"top"`
	Cron        string    `bigquery:"cron"`
	Ativo       bool      `bigquery:"ativo"`
	SalvoEm     time.Time `bigquery:"salvo_em"`
}

func (s *BigQueryStore) SalvarBusca(ctx context.Context, b Busca) error {
	if b.SalvoEm.IsZero() {
		b.SalvoEm = time.Now().UTC()
	}
	row := linhaBuscaBQ{
		Nome: b.Nome, Keyword: b.Keyword, Categoria: b.Categoria, Estrategia: b.Estrategia,
		ComissaoMin: b.ComissaoMin, VendasMin: b.VendasMin, NotaMin: b.NotaMin, Top: b.Top,
		Cron: b.Cron, Ativo: b.Ativo, SalvoEm: b.SalvoEm,
	}
	return s.client.Dataset(s.dataset).Table("buscas").Inserter().Put(ctx, row)
}

// ListarBuscas devolve o estado atual: o último registro por nome (append-only),
// filtrando os removidos (ativo = false).
func (s *BigQueryStore) ListarBuscas(ctx context.Context) ([]Busca, error) {
	q := s.client.Query(`
		WITH ranked AS (
		  SELECT *, ROW_NUMBER() OVER (PARTITION BY nome ORDER BY salvo_em DESC) AS rn
		  FROM ` + "`" + s.dataset + ".buscas`" + `
		)
		SELECT nome, keyword, categoria, estrategia, comissao_min, vendas_min,
		       nota_min, top, cron, ativo, salvo_em
		FROM ranked WHERE rn = 1 AND ativo = TRUE
		ORDER BY nome
	`)
	it, err := q.Read(ctx)
	if err != nil {
		return nil, err
	}
	var out []Busca
	for {
		var r linhaBuscaBQ
		err := it.Next(&r)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		out = append(out, Busca{
			Nome: r.Nome, Keyword: r.Keyword, Categoria: r.Categoria, Estrategia: r.Estrategia,
			ComissaoMin: r.ComissaoMin, VendasMin: r.VendasMin, NotaMin: r.NotaMin, Top: r.Top,
			Cron: r.Cron, Ativo: r.Ativo, SalvoEm: r.SalvoEm,
		})
	}
	return out, nil
}

// Estatisticas agrega os snapshots dos últimos `dias` por categoria: média de
// comissão/preço/vendas/teor e mediana de comissão (APPROX_QUANTILES). É a
// primeira consulta do pipeline de análise — descritiva, barata, e base para o
// painel no frontend.
func (s *BigQueryStore) Estatisticas(ctx context.Context, dias int) (Estatisticas, error) {
	if dias <= 0 {
		dias = 30
	}
	est := Estatisticas{Fonte: "bigquery", DiasJanela: dias, GeradoEm: time.Now().UTC()}

	q := s.client.Query(`
		SELECT
		  categoria,
		  COUNT(*)                                   AS amostras,
		  AVG(comissao)                              AS comissao_media,
		  APPROX_QUANTILES(comissao, 2)[OFFSET(1)]   AS comissao_mediana,
		  AVG(preco)                                 AS preco_medio,
		  AVG(vendas)                                AS vendas_media,
		  AVG(score)                                 AS teor_medio
		FROM ` + "`" + s.dataset + ".snapshots`" + `
		WHERE coletado_em >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL @dias DAY)
		GROUP BY categoria
		ORDER BY amostras DESC
	`)
	q.Parameters = []bigquery.QueryParameter{{Name: "dias", Value: dias}}

	it, err := q.Read(ctx)
	if err != nil {
		return est, err
	}
	for {
		var row struct {
			Categoria       string  `bigquery:"categoria"`
			Amostras        int     `bigquery:"amostras"`
			ComissaoMedia   float64 `bigquery:"comissao_media"`
			ComissaoMediana float64 `bigquery:"comissao_mediana"`
			PrecoMedio      float64 `bigquery:"preco_medio"`
			VendasMedia     float64 `bigquery:"vendas_media"`
			TeorMedio       float64 `bigquery:"teor_medio"`
		}
		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return est, err
		}
		est.TotalAmostras += row.Amostras
		est.PorCategoria = append(est.PorCategoria, EstatCategoria{
			Categoria:       row.Categoria,
			Amostras:        row.Amostras,
			ComissaoMedia:   row.ComissaoMedia,
			ComissaoMediana: row.ComissaoMediana,
			PrecoMedio:      row.PrecoMedio,
			VendasMedia:     row.VendasMedia,
			TeorMedio:       row.TeorMedio,
		})
	}
	return est, nil
}
