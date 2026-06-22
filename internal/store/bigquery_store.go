//go:build gcp

package store

import (
	"context"
	"encoding/json"
	"strings"
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

// EnsureSchema cria as tabelas do dataset se ainda não existirem. Idempotente —
// chamado no startup toda vez, de modo que o banco evolui automaticamente ao
// deploy sem passo manual de "criar tabelas".
func (s *BigQueryStore) EnsureSchema(ctx context.Context) error {
	ds := s.client.Dataset(s.dataset)

	// --- tabela eventos ---
	eSchema := bigquery.Schema{
		{Name: "tipo", Type: bigquery.StringFieldType},
		{Name: "produto_id", Type: bigquery.StringFieldType},
		{Name: "nome", Type: bigquery.StringFieldType},
		{Name: "categoria", Type: bigquery.StringFieldType},
		{Name: "estrategia", Type: bigquery.StringFieldType},
		{Name: "canal", Type: bigquery.StringFieldType},
		{Name: "sub_id", Type: bigquery.StringFieldType},
		{Name: "comissao", Type: bigquery.FloatFieldType},
		{Name: "preco", Type: bigquery.FloatFieldType},
		{Name: "vendas", Type: bigquery.IntegerFieldType},
		{Name: "score", Type: bigquery.FloatFieldType},
		{Name: "em", Type: bigquery.TimestampFieldType},
	}
	if err := criarSeNaoExistir(ctx, ds, s.tabela, eSchema, "em"); err != nil {
		return err
	}

	// --- tabela snapshots ---
	sSchema := bigquery.Schema{
		{Name: "coletado_em", Type: bigquery.TimestampFieldType},
		{Name: "categoria", Type: bigquery.StringFieldType},
		{Name: "keyword", Type: bigquery.StringFieldType},
		{Name: "estrategia", Type: bigquery.StringFieldType},
		{Name: "posicao", Type: bigquery.IntegerFieldType},
		{Name: "produto_id", Type: bigquery.StringFieldType},
		{Name: "nome", Type: bigquery.StringFieldType},
		{Name: "preco", Type: bigquery.FloatFieldType},
		{Name: "comissao", Type: bigquery.FloatFieldType},
		{Name: "vendas", Type: bigquery.IntegerFieldType},
		{Name: "nota", Type: bigquery.FloatFieldType},
		{Name: "score", Type: bigquery.FloatFieldType},
	}
	if err := criarSeNaoExistir(ctx, ds, s.tabelaSnap, sSchema, "coletado_em"); err != nil {
		return err
	}

	// --- tabela buscas ---
	bSchema := bigquery.Schema{
		{Name: "id", Type: bigquery.StringFieldType},
		{Name: "keywords", Type: bigquery.StringFieldType}, // JSON array
		{Name: "categoria", Type: bigquery.StringFieldType},
		{Name: "estrategia", Type: bigquery.StringFieldType},
		{Name: "comissao_min", Type: bigquery.FloatFieldType},
		{Name: "vendas_min", Type: bigquery.IntegerFieldType},
		{Name: "nota_min", Type: bigquery.FloatFieldType},
		{Name: "top", Type: bigquery.IntegerFieldType},
		{Name: "cron", Type: bigquery.StringFieldType},
		{Name: "ativo", Type: bigquery.BooleanFieldType},
		{Name: "salvo_em", Type: bigquery.TimestampFieldType},
	}
	return criarSeNaoExistir(ctx, ds, "buscas", bSchema, "salvo_em")
}

// criarSeNaoExistir cria a tabela particionada por dia se ainda não existir.
func criarSeNaoExistir(ctx context.Context, ds *bigquery.Dataset, nome string, schema bigquery.Schema, campoPartition string) error {
	t := ds.Table(nome)
	_, err := t.Metadata(ctx)
	if err == nil {
		return nil // já existe
	}
	// se o erro não for "não encontrado", propaga
	meta := &bigquery.TableMetadata{
		Schema: schema,
		TimePartitioning: &bigquery.TimePartitioning{
			Type:  bigquery.DayPartitioningType,
			Field: campoPartition,
		},
	}
	return t.Create(ctx, meta)
}

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

// linhaBuscaBQ mapeia a Busca para a tabela `buscas`.
// Keywords é serializado como JSON array para caber em uma coluna STRING.
type linhaBuscaBQ struct {
	ID          string    `bigquery:"id"`
	Keywords    string    `bigquery:"keywords"` // JSON array
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
	b = NormalizarBusca(b)
	if b.SalvoEm.IsZero() {
		b.SalvoEm = time.Now().UTC()
	}
	kw, _ := json.Marshal(b.Keywords)
	row := linhaBuscaBQ{
		ID: b.ID, Keywords: string(kw), Categoria: b.Categoria, Estrategia: b.Estrategia,
		ComissaoMin: b.ComissaoMin, VendasMin: b.VendasMin, NotaMin: b.NotaMin, Top: b.Top,
		Cron: b.Cron, Ativo: b.Ativo, SalvoEm: b.SalvoEm,
	}
	return s.client.Dataset(s.dataset).Table("buscas").Inserter().Put(ctx, row)
}

// ListarBuscas devolve o estado atual: o último registro por ID (append-only),
// filtrando os removidos (ativo = false).
func (s *BigQueryStore) ListarBuscas(ctx context.Context) ([]Busca, error) {
	q := s.client.Query(`
		WITH ranked AS (
		  SELECT *, ROW_NUMBER() OVER (PARTITION BY id ORDER BY salvo_em DESC) AS rn
		  FROM ` + "`" + s.dataset + ".buscas`" + `
		)
		SELECT id, keywords, categoria, estrategia, comissao_min, vendas_min,
		       nota_min, top, cron, ativo, salvo_em
		FROM ranked WHERE rn = 1 AND ativo = TRUE
		ORDER BY id
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
		var kws []string
		// tenta deserializar o JSON array; cai num slice com o valor bruto se falhar
		if e2 := json.Unmarshal([]byte(r.Keywords), &kws); e2 != nil {
			// compatibilidade: campo keywords pode ser string simples em dados antigos
			if r.Keywords != "" {
				kws = strings.Split(r.Keywords, ",")
				for i := range kws {
					kws[i] = strings.TrimSpace(kws[i])
				}
			}
		}
		out = append(out, Busca{
			ID: r.ID, Keywords: kws, Categoria: r.Categoria, Estrategia: r.Estrategia,
			ComissaoMin: r.ComissaoMin, VendasMin: r.VendasMin, NotaMin: r.NotaMin, Top: r.Top,
			Cron: r.Cron, Ativo: r.Ativo, SalvoEm: r.SalvoEm,
		})
	}
	return out, nil
}

// Estatisticas agrega os snapshots dos últimos `dias` por categoria.
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
