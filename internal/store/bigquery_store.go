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

// Client expõe o client BigQuery para ser reutilizado pelos stores auxiliares
// (BQDestinoStore, BQTemplateStore) sem criar conexões adicionais.
func (s *BigQueryStore) Client() *bigquery.Client { return s.client }

// Dataset retorna o nome do dataset.
func (s *BigQueryStore) Dataset() string { return s.dataset }

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

// linhaBuscaBQ mapeia a Busca para a tabela `buscas`.
// Keywords é serializado como JSON array para caber em uma coluna STRING.
type linhaBuscaBQ struct {
	ID             string    `bigquery:"id"`
	Nome           string    `bigquery:"nome"`
	Keywords       string    `bigquery:"keywords"`       // JSON array
	ShopIDs        string    `bigquery:"shop_ids"`       // JSON array de int64
	Categorias     string    `bigquery:"categorias"`     // JSON array
	Fontes         string    `bigquery:"fontes"`         // JSON array
	Categoria      string    `bigquery:"categoria"`      // legado
	Estrategia     string    `bigquery:"estrategia"`
	ComissaoMin    float64   `bigquery:"comissao_min"`
	VendasMin      int       `bigquery:"vendas_min"`
	NotaMin        float64   `bigquery:"nota_min"`
	Top            int       `bigquery:"top"`
	DiasJanela     int       `bigquery:"dias_janela"`
	Cron           string    `bigquery:"cron"`
	Ativo          bool      `bigquery:"ativo"`
	OwnerUID       string    `bigquery:"owner_uid"`
	OrigemPadrao   string    `bigquery:"origem_padrao"`
	RotationCursor string    `bigquery:"rotation_cursor"` // JSON map shopID→page
	FullScanAt     string    `bigquery:"full_scan_at"`    // JSON map shopID→timestamp
	SalvoEm        time.Time `bigquery:"salvo_em"`
}

func (s *BigQueryStore) SalvarBusca(ctx context.Context, b Busca) error {
	b = NormalizarBusca(b)
	if b.SalvoEm.IsZero() {
		b.SalvoEm = time.Now().UTC()
	}
	kw, _ := json.Marshal(b.Keywords)
	shopIDs, _ := json.Marshal(b.ShopIDs)
	categorias, _ := json.Marshal(b.Categorias)
	fontes, _ := json.Marshal(b.Fontes)
	rotCursor, _ := json.Marshal(b.RotationCursor)
	fullScan, _ := json.Marshal(b.FullScanAt)
	row := linhaBuscaBQ{
		ID: b.ID, Nome: b.Nome, Keywords: string(kw), ShopIDs: string(shopIDs),
		Categorias: string(categorias), Fontes: string(fontes),
		Categoria: b.Categoria, Estrategia: b.Estrategia,
		ComissaoMin: b.ComissaoMin, VendasMin: b.VendasMin, NotaMin: b.NotaMin, Top: b.Top,
		DiasJanela: b.DiasJanela, Cron: b.Cron, Ativo: b.Ativo, OwnerUID: b.OwnerUID,
		OrigemPadrao:   b.OrigemPadrao,
		RotationCursor: string(rotCursor), FullScanAt: string(fullScan),
		SalvoEm: b.SalvoEm,
	}
	return s.client.Dataset(s.dataset).Table("buscas").Inserter().Put(ctx, row)
}

// ListarBuscas devolve o estado atual: o último registro por ID (append-only),
// filtrando os removidos (ativo = false).
func (s *BigQueryStore) ListarBuscas(ctx context.Context) ([]Busca, error) {
	// Query compatível com tabelas que não têm as colunas novas (shop_ids, rotation_cursor, full_scan_at).
	// Usa SELECT * no CTE e seleciona apenas as colunas base. Os campos novos são lidos
	// via query separada se necessário, ou adicionados via ALTER TABLE na migração.
	q := s.client.Query(`
		WITH ranked AS (
		  SELECT *, ROW_NUMBER() OVER (PARTITION BY id ORDER BY salvo_em DESC) AS rn
		  FROM ` + "`" + s.dataset + ".buscas`" + `
		)
		SELECT id, keywords, categoria, estrategia,
		       comissao_min, vendas_min, nota_min, top, cron, ativo,
		       IFNULL(owner_uid, '') as owner_uid, salvo_em
		FROM ranked WHERE rn = 1 AND ativo = TRUE
		ORDER BY id
	`)
	it, err := q.Read(ctx)
	if err != nil {
		return nil, err
	}
	var out []Busca
	for {
		var r struct {
			ID          string    `bigquery:"id"`
			Keywords    string    `bigquery:"keywords"`
			Categoria   string    `bigquery:"categoria"`
			Estrategia  string    `bigquery:"estrategia"`
			ComissaoMin float64   `bigquery:"comissao_min"`
			VendasMin   int       `bigquery:"vendas_min"`
			NotaMin     float64   `bigquery:"nota_min"`
			Top         int       `bigquery:"top"`
			Cron        string    `bigquery:"cron"`
			Ativo       bool      `bigquery:"ativo"`
			OwnerUID    string    `bigquery:"owner_uid"`
			SalvoEm     time.Time `bigquery:"salvo_em"`
		}
		err := it.Next(&r)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var kws []string
		if e2 := json.Unmarshal([]byte(r.Keywords), &kws); e2 != nil {
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
			Cron: r.Cron, Ativo: r.Ativo, OwnerUID: r.OwnerUID, SalvoEm: r.SalvoEm,
		})
	}

	// Tenta ler campos novos (shop_ids, rotation_cursor, full_scan_at) se existirem.
	// Se a query falhar (colunas não existem), retorna o resultado sem esses campos.
	s.enriquecerBuscasComCamposNovos(ctx, out)

	return out, nil
}

// enriquecerBuscasComCamposNovos tenta ler campos adicionais (shop_ids, fontes, categorias, etc.).
// Se as colunas não existirem na tabela, simplesmente não faz nada (graceful degradation).
func (s *BigQueryStore) enriquecerBuscasComCamposNovos(ctx context.Context, buscas []Busca) {
	q := s.client.Query(`
		WITH ranked AS (
		  SELECT id, shop_ids, rotation_cursor, full_scan_at,
		         IFNULL(nome, '') as nome, IFNULL(origem_padrao, '') as origem_padrao,
		         IFNULL(categorias, '') as categorias, IFNULL(fontes, '') as fontes,
		         IFNULL(dias_janela, 0) as dias_janela,
		         ROW_NUMBER() OVER (PARTITION BY id ORDER BY salvo_em DESC) AS rn
		  FROM ` + "`" + s.dataset + ".buscas`" + `
		)
		SELECT id, IFNULL(shop_ids, '') as shop_ids,
		       IFNULL(rotation_cursor, '') as rotation_cursor,
		       IFNULL(full_scan_at, '') as full_scan_at,
		       nome, origem_padrao, categorias, fontes, dias_janela
		FROM ranked WHERE rn = 1
	`)
	it, err := q.Read(ctx)
	if err != nil {
		return
	}

	byID := make(map[string]int, len(buscas))
	for i := range buscas {
		byID[buscas[i].ID] = i
	}

	for {
		var r struct {
			ID             string `bigquery:"id"`
			ShopIDs        string `bigquery:"shop_ids"`
			RotationCursor string `bigquery:"rotation_cursor"`
			FullScanAt     string `bigquery:"full_scan_at"`
			Nome           string `bigquery:"nome"`
			OrigemPadrao   string `bigquery:"origem_padrao"`
			Categorias     string `bigquery:"categorias"`
			Fontes         string `bigquery:"fontes"`
			DiasJanela     int    `bigquery:"dias_janela"`
		}
		err := it.Next(&r)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return
		}
		idx, ok := byID[r.ID]
		if !ok {
			continue
		}
		if r.ShopIDs != "" {
			_ = json.Unmarshal([]byte(r.ShopIDs), &buscas[idx].ShopIDs)
		}
		if r.RotationCursor != "" {
			_ = json.Unmarshal([]byte(r.RotationCursor), &buscas[idx].RotationCursor)
		}
		if r.FullScanAt != "" {
			_ = json.Unmarshal([]byte(r.FullScanAt), &buscas[idx].FullScanAt)
		}
		if r.Nome != "" {
			buscas[idx].Nome = r.Nome
		}
		if r.OrigemPadrao != "" {
			buscas[idx].OrigemPadrao = r.OrigemPadrao
		}
		if r.Categorias != "" {
			_ = json.Unmarshal([]byte(r.Categorias), &buscas[idx].Categorias)
		}
		if r.Fontes != "" {
			_ = json.Unmarshal([]byte(r.Fontes), &buscas[idx].Fontes)
		}
		if r.DiasJanela > 0 {
			buscas[idx].DiasJanela = r.DiasJanela
		}
	}
}

