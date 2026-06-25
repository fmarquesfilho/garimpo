//go:build gcp

package store

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/fmarquesfilho/garimpo/internal/publish"
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
		{Name: "shop_ids", Type: bigquery.StringFieldType}, // JSON array de int64
		{Name: "categoria", Type: bigquery.StringFieldType},
		{Name: "estrategia", Type: bigquery.StringFieldType},
		{Name: "comissao_min", Type: bigquery.FloatFieldType},
		{Name: "vendas_min", Type: bigquery.IntegerFieldType},
		{Name: "nota_min", Type: bigquery.FloatFieldType},
		{Name: "top", Type: bigquery.IntegerFieldType},
		{Name: "cron", Type: bigquery.StringFieldType},
		{Name: "ativo", Type: bigquery.BooleanFieldType},
		{Name: "owner_uid", Type: bigquery.StringFieldType},
		{Name: "rotation_cursor", Type: bigquery.StringFieldType}, // JSON map
		{Name: "full_scan_at", Type: bigquery.StringFieldType},    // JSON map
		{Name: "salvo_em", Type: bigquery.TimestampFieldType},
	}
	if err := criarSeNaoExistir(ctx, ds, "buscas", bSchema, "salvo_em"); err != nil {
		return err
	}

	// --- tabela destinos (append-only, como buscas) ---
	dSchema := bigquery.Schema{
		{Name: "id", Type: bigquery.StringFieldType},
		{Name: "nome", Type: bigquery.StringFieldType},
		{Name: "tipo", Type: bigquery.StringFieldType},
		{Name: "config", Type: bigquery.StringFieldType},
		{Name: "ativo", Type: bigquery.BooleanFieldType},
		{Name: "salvo_em", Type: bigquery.TimestampFieldType},
	}
	if err := criarSeNaoExistir(ctx, ds, "destinos", dSchema, "salvo_em"); err != nil {
		return err
	}

	// --- tabela templates (append-only) ---
	tSchema := bigquery.Schema{
		{Name: "id", Type: bigquery.StringFieldType},
		{Name: "nome", Type: bigquery.StringFieldType},
		{Name: "corpo", Type: bigquery.StringFieldType},
		{Name: "com_foto", Type: bigquery.BooleanFieldType},
		{Name: "ativo", Type: bigquery.BooleanFieldType},
		{Name: "salvo_em", Type: bigquery.TimestampFieldType},
	}
	if err := criarSeNaoExistir(ctx, ds, "templates", tSchema, "salvo_em"); err != nil {
		return err
	}

	// --- tabela publicacoes ---
	pSchema := bigquery.Schema{
		{Name: "id", Type: bigquery.StringFieldType},
		{Name: "produto_id", Type: bigquery.StringFieldType},
		{Name: "nome", Type: bigquery.StringFieldType},
		{Name: "categoria", Type: bigquery.StringFieldType},
		{Name: "preco", Type: bigquery.FloatFieldType},
		{Name: "comissao", Type: bigquery.FloatFieldType},
		{Name: "link", Type: bigquery.StringFieldType},
		{Name: "imagem", Type: bigquery.StringFieldType},
		{Name: "estrategia", Type: bigquery.StringFieldType},
		{Name: "destino_id", Type: bigquery.StringFieldType},
		{Name: "template_id", Type: bigquery.StringFieldType},
		{Name: "agendada_em", Type: bigquery.StringFieldType},
		{Name: "status", Type: bigquery.StringFieldType},
		{Name: "detalhe", Type: bigquery.StringFieldType},
		{Name: "criada_em", Type: bigquery.TimestampFieldType},
		{Name: "enviada_em", Type: bigquery.StringFieldType},
		{Name: "owner_uid", Type: bigquery.StringFieldType},
	}
	return criarSeNaoExistir(ctx, ds, "publicacoes", pSchema, "criada_em")
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
	ID             string    `bigquery:"id"`
	Keywords       string    `bigquery:"keywords"` // JSON array
	ShopIDs        string    `bigquery:"shop_ids"` // JSON array de int64
	Categoria      string    `bigquery:"categoria"`
	Estrategia     string    `bigquery:"estrategia"`
	ComissaoMin    float64   `bigquery:"comissao_min"`
	VendasMin      int       `bigquery:"vendas_min"`
	NotaMin        float64   `bigquery:"nota_min"`
	Top            int       `bigquery:"top"`
	Cron           string    `bigquery:"cron"`
	Ativo          bool      `bigquery:"ativo"`
	OwnerUID       string    `bigquery:"owner_uid"`
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
	rotCursor, _ := json.Marshal(b.RotationCursor)
	fullScan, _ := json.Marshal(b.FullScanAt)
	row := linhaBuscaBQ{
		ID: b.ID, Keywords: string(kw), ShopIDs: string(shopIDs),
		Categoria: b.Categoria, Estrategia: b.Estrategia,
		ComissaoMin: b.ComissaoMin, VendasMin: b.VendasMin, NotaMin: b.NotaMin, Top: b.Top,
		Cron: b.Cron, Ativo: b.Ativo, OwnerUID: b.OwnerUID,
		RotationCursor: string(rotCursor), FullScanAt: string(fullScan),
		SalvoEm: b.SalvoEm,
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
		SELECT id, keywords, IFNULL(shop_ids, '') as shop_ids, categoria, estrategia,
		       comissao_min, vendas_min, nota_min, top, cron, ativo, owner_uid,
		       IFNULL(rotation_cursor, '') as rotation_cursor,
		       IFNULL(full_scan_at, '') as full_scan_at, salvo_em
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
		var shopIDs []int64
		if r.ShopIDs != "" {
			_ = json.Unmarshal([]byte(r.ShopIDs), &shopIDs)
		}
		var rotCursor map[int64]int
		if r.RotationCursor != "" {
			_ = json.Unmarshal([]byte(r.RotationCursor), &rotCursor)
		}
		var fullScan map[int64]string
		if r.FullScanAt != "" {
			_ = json.Unmarshal([]byte(r.FullScanAt), &fullScan)
		}
		out = append(out, Busca{
			ID: r.ID, Keywords: kws, ShopIDs: shopIDs, Categoria: r.Categoria, Estrategia: r.Estrategia,
			ComissaoMin: r.ComissaoMin, VendasMin: r.VendasMin, NotaMin: r.NotaMin, Top: r.Top,
			Cron: r.Cron, Ativo: r.Ativo, OwnerUID: r.OwnerUID,
			RotationCursor: rotCursor, FullScanAt: fullScan, SalvoEm: r.SalvoEm,
		})
	}
	return out, nil
}

// HistoricoColetas retorna os snapshots agrupados por execução (keyword + timestamp)
// nos últimos `dias`, ordenados do mais recente ao mais antigo.
func (s *BigQueryStore) HistoricoColetas(ctx context.Context, dias int) ([]ColetaResumo, error) {
	if dias <= 0 {
		dias = 30
	}
	q := s.client.Query(`
		SELECT
		  coletado_em,
		  keyword,
		  categoria,
		  estrategia,
		  COUNT(*) AS produtos
		FROM ` + "`" + s.dataset + ".snapshots`" + `
		WHERE coletado_em >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL @dias DAY)
		GROUP BY coletado_em, keyword, categoria, estrategia
		ORDER BY coletado_em DESC
		LIMIT 200
	`)
	q.Parameters = []bigquery.QueryParameter{{Name: "dias", Value: dias}}
	it, err := q.Read(ctx)
	if err != nil {
		return nil, err
	}
	var out []ColetaResumo
	for {
		var row struct {
			ColetadoEm time.Time `bigquery:"coletado_em"`
			Keyword    string    `bigquery:"keyword"`
			Categoria  string    `bigquery:"categoria"`
			Estrategia string    `bigquery:"estrategia"`
			Produtos   int       `bigquery:"produtos"`
		}
		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		out = append(out, ColetaResumo{
			ColetadoEm: row.ColetadoEm,
			Keyword:    row.Keyword,
			Categoria:  row.Categoria,
			Estrategia: row.Estrategia,
			Produtos:   row.Produtos,
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

// Conversoes retorna o relatório de publicações agrupado por canal e sub_id,
// permitindo à usuária ver quais destinos geraram mais volume de publicação
// (e potencial de conversão).
func (s *BigQueryStore) Conversoes(ctx context.Context, dias int) ([]ConversaoResumo, error) {
	if dias <= 0 {
		dias = 30
	}
	q := s.client.Query(`
		SELECT
		  canal,
		  sub_id,
		  COUNT(*) AS publicacoes,
		  ANY_VALUE(produto_id) AS produto_id,
		  ANY_VALUE(nome) AS nome,
		  ANY_VALUE(estrategia) AS estrategia,
		  AVG(preco) AS preco,
		  SUM(comissao * preco) AS comissao_estimada,
		  FORMAT_TIMESTAMP('%Y-%m-%d', MAX(em)) AS publicado_em
		FROM ` + "`" + s.dataset + "." + s.tabela + "`" + `
		WHERE tipo = 'publicacao'
		  AND em >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL @dias DAY)
		GROUP BY canal, sub_id
		ORDER BY publicacoes DESC
		LIMIT 100
	`)
	q.Parameters = []bigquery.QueryParameter{{Name: "dias", Value: dias}}
	it, err := q.Read(ctx)
	if err != nil {
		return nil, err
	}
	var out []ConversaoResumo
	for {
		var row struct {
			Canal       string  `bigquery:"canal"`
			SubID       string  `bigquery:"sub_id"`
			Publicacoes int     `bigquery:"publicacoes"`
			ProdutoID   string  `bigquery:"produto_id"`
			Nome        string  `bigquery:"nome"`
			Estrategia  string  `bigquery:"estrategia"`
			Preco       float64 `bigquery:"preco"`
			ComissaoEst float64 `bigquery:"comissao_estimada"`
			PublicadoEm string  `bigquery:"publicado_em"`
		}
		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		out = append(out, ConversaoResumo{
			Canal:       row.Canal,
			SubID:       row.SubID,
			Publicacoes: row.Publicacoes,
			ProdutoID:   row.ProdutoID,
			Nome:        row.Nome,
			Estrategia:  row.Estrategia,
			Preco:       row.Preco,
			ComissaoEst: row.ComissaoEst,
			PublicadoEm: row.PublicadoEm,
		})
	}
	return out, nil
}

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
	criadaEm, _ := time.Parse(time.RFC3339, p.CriadaEm)
	if criadaEm.IsZero() {
		criadaEm = time.Now().UTC()
	}
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
	// Append-only: grava novo registro com o status atualizado.
	// O ROW_NUMBER em ListarPublicacoes garante que o mais recente prevalece.
	row := linhaPublicacaoBQ{
		ID: id, Status: status, Detalhe: detalhe, CriadaEm: time.Now().UTC(),
	}
	return s.client.Dataset(s.dataset).Table("publicacoes").Inserter().Put(ctx, row)
}

// ─── Destinos (BigQuery) ──────────────────────────────────────────────────

type linhaDestinoBQ struct {
	ID      string    `bigquery:"id"`
	Nome    string    `bigquery:"nome"`
	Tipo    string    `bigquery:"tipo"`
	Config  string    `bigquery:"config"`
	Ativo   bool      `bigquery:"ativo"`
	SalvoEm time.Time `bigquery:"salvo_em"`
}

// BQDestinoStore implementa publish.DestinoStore com BigQuery.
type BQDestinoStore struct {
	client  *bigquery.Client
	dataset string
}

func NovoBQDestinoStore(client *bigquery.Client, dataset string) *BQDestinoStore {
	return &BQDestinoStore{client: client, dataset: dataset}
}

func (s *BQDestinoStore) Listar(ctx context.Context) ([]publish.Destino, error) {
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
		return nil, err
	}
	var out []publish.Destino
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
			return nil, err
		}
		out = append(out, publish.Destino{ID: r.ID, Nome: r.Nome, Tipo: r.Tipo, Config: r.Config, Ativo: r.Ativo})
	}
	return out, nil
}

func (s *BQDestinoStore) Buscar(ctx context.Context, id string) (publish.Destino, error) {
	lista, err := s.Listar(ctx)
	if err != nil {
		return publish.Destino{}, err
	}
	for _, d := range lista {
		if d.ID == id {
			return d, nil
		}
	}
	return publish.Destino{}, fmt.Errorf("destino %q não encontrado", id)
}

func (s *BQDestinoStore) Salvar(ctx context.Context, d publish.Destino) error {
	row := linhaDestinoBQ{
		ID: d.ID, Nome: d.Nome, Tipo: d.Tipo, Config: d.Config,
		Ativo: d.Ativo, SalvoEm: time.Now().UTC(),
	}
	return s.client.Dataset(s.dataset).Table("destinos").Inserter().Put(ctx, row)
}

func (s *BQDestinoStore) Deletar(ctx context.Context, id string) error {
	// Append-only tombstone
	row := linhaDestinoBQ{ID: id, Ativo: false, SalvoEm: time.Now().UTC()}
	return s.client.Dataset(s.dataset).Table("destinos").Inserter().Put(ctx, row)
}

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

// Novidades compara snapshots da janela para detectar produtos novos e variações de preço.
// Se buscaID não estiver vazio, filtra por keyword que contenha o id da busca.
func (s *BigQueryStore) Novidades(ctx context.Context, buscaID string, dias int) (NovidadesLojas, error) {
	if dias <= 0 {
		dias = 7
	}
	result := NovidadesLojas{BuscaID: buscaID, DiasJanela: dias}

	// Busca os snapshots da janela, agrupados por produto_id com primeiro e último registro
	filtroKW := ""
	params := []bigquery.QueryParameter{{Name: "dias", Value: dias}}
	if buscaID != "" {
		filtroKW = " AND LOWER(keyword) LIKE @busca_id"
		params = append(params, bigquery.QueryParameter{Name: "busca_id", Value: "%" + buscaID + "%"})
	}

	q := s.client.Query(`
		WITH janela AS (
		  SELECT produto_id, nome, preco, comissao, vendas, nota, coletado_em,
		         ROW_NUMBER() OVER (PARTITION BY produto_id ORDER BY coletado_em DESC) AS rn_desc,
		         ROW_NUMBER() OVER (PARTITION BY produto_id ORDER BY coletado_em ASC) AS rn_asc,
		         COUNT(*) OVER (PARTITION BY produto_id) AS aparicoes
		  FROM ` + "`" + s.dataset + ".snapshots`" + `
		  WHERE coletado_em >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL @dias DAY)
		    ` + filtroKW + `
		)
		SELECT produto_id, nome, preco, comissao, vendas, nota,
		       FORMAT_TIMESTAMP('%Y-%m-%dT%H:%M:%SZ', coletado_em) AS detectado_em,
		       aparicoes, rn_desc,
		       -- primeiro preço registrado na janela
		       FIRST_VALUE(preco) OVER (PARTITION BY produto_id ORDER BY coletado_em ASC) AS preco_primeiro
		FROM janela
		WHERE rn_desc = 1
		ORDER BY coletado_em DESC
		LIMIT 500
	`)
	q.Parameters = params

	it, err := q.Read(ctx)
	if err != nil {
		return result, err
	}

	for {
		var row struct {
			ProdutoID    string  `bigquery:"produto_id"`
			Nome         string  `bigquery:"nome"`
			Preco        float64 `bigquery:"preco"`
			Comissao     float64 `bigquery:"comissao"`
			Vendas       int     `bigquery:"vendas"`
			Nota         float64 `bigquery:"nota"`
			DetectadoEm  string  `bigquery:"detectado_em"`
			Aparicoes    int     `bigquery:"aparicoes"`
			RnDesc       int     `bigquery:"rn_desc"`
			PrecoPrimeiro float64 `bigquery:"preco_primeiro"`
		}
		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return result, err
		}

		result.TotalAtual++

		// Produto novo: apareceu só uma vez na janela (primeira coleta)
		if row.Aparicoes == 1 {
			result.ProdutosNovos = append(result.ProdutosNovos, ProdutoNovo{
				ProdutoID: row.ProdutoID, Nome: row.Nome, Preco: row.Preco,
				Comissao: row.Comissao, Vendas: row.Vendas, Nota: row.Nota,
				DetectadoEm: row.DetectadoEm,
			})
		}

		// Variação de preço: preço atual != primeiro preço da janela
		if row.Aparicoes > 1 && row.PrecoPrimeiro > 0 && row.Preco != row.PrecoPrimeiro {
			variacao := (row.Preco - row.PrecoPrimeiro) / row.PrecoPrimeiro
			result.Variacoes = append(result.Variacoes, VariacaoPreco{
				ProdutoID: row.ProdutoID, Nome: row.Nome,
				PrecoAnterior: row.PrecoPrimeiro, PrecoAtual: row.Preco,
				Variacao: variacao, DetectadoEm: row.DetectadoEm,
			})
		}
	}

	return result, nil
}

// EvolucaoLojas retorna dados de evolução de preço das lojas monitoradas ao longo do tempo.
// Agrega: preço médio diário por keyword (=loja), total de produtos, variações.
func (s *BigQueryStore) EvolucaoLojas(ctx context.Context, dias int) (EvolucaoLojasResult, error) {
	if dias <= 0 {
		dias = 30
	}
	result := EvolucaoLojasResult{DiasJanela: dias}

	// 1. Busca as buscas ativas que têm shop_ids (são lojas monitoradas)
	buscas, err := s.ListarBuscas(ctx)
	if err != nil {
		return result, err
	}
	var buscasLoja []Busca
	for _, b := range buscas {
		if len(b.ShopIDs) > 0 && b.Ativo {
			buscasLoja = append(buscasLoja, b)
		}
	}
	if len(buscasLoja) == 0 {
		return result, nil
	}

	// 2. Query: preço médio diário dos snapshots que correspondem a lojas
	q := s.client.Query(`
		SELECT
		  keyword,
		  FORMAT_TIMESTAMP('%Y-%m-%d', coletado_em) AS dia,
		  AVG(preco) AS preco_medio,
		  COUNT(DISTINCT produto_id) AS produtos
		FROM ` + "`" + s.dataset + ".snapshots`" + `
		WHERE coletado_em >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL @dias DAY)
		  AND keyword LIKE 'loja-%'
		GROUP BY keyword, dia
		ORDER BY keyword, dia
	`)
	q.Parameters = []bigquery.QueryParameter{{Name: "dias", Value: dias}}

	it, err := q.Read(ctx)
	if err != nil {
		return result, err
	}

	// Agrupa pontos por keyword (busca_id da loja)
	pontosPorLoja := map[string][]PontoEvolucao{}
	for {
		var row struct {
			Keyword    string  `bigquery:"keyword"`
			Dia        string  `bigquery:"dia"`
			PrecoMedio float64 `bigquery:"preco_medio"`
			Produtos   int     `bigquery:"produtos"`
		}
		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return result, err
		}
		pontosPorLoja[row.Keyword] = append(pontosPorLoja[row.Keyword], PontoEvolucao{
			Data:       row.Dia,
			PrecoMedio: row.PrecoMedio,
			Produtos:   row.Produtos,
		})
	}

	// 3. Monta resultado por loja
	var totalProdutos int
	var somaVariacao float64
	var totalQuedas, totalAltas int
	var somaPreco float64
	var contaPreco int

	for _, b := range buscasLoja {
		pontos := pontosPorLoja[b.ID]
		if len(pontos) == 0 {
			continue
		}

		evo := EvolucaoLoja{
			BuscaID: b.ID,
			Coletas: len(pontos),
			Pontos:  pontos,
		}

		// Primeiro e último ponto para calcular variação
		primeiro := pontos[0]
		ultimo := pontos[len(pontos)-1]
		evo.PrecoMedioInicio = primeiro.PrecoMedio
		evo.PrecoMedioAtual = ultimo.PrecoMedio
		evo.TotalProdutos = ultimo.Produtos

		if primeiro.PrecoMedio > 0 {
			evo.VariacaoMedia = (ultimo.PrecoMedio - primeiro.PrecoMedio) / primeiro.PrecoMedio
		}

		// Busca novidades dessa loja para popular top_quedas e top_altas
		novidades, _ := s.Novidades(ctx, b.ID, dias)
		for _, v := range novidades.Variacoes {
			if v.Variacao < -0.10 {
				evo.TopQuedas = append(evo.TopQuedas, v)
				totalQuedas++
			}
			if v.Variacao > 0.10 {
				evo.TopAltas = append(evo.TopAltas, v)
				totalAltas++
			}
		}
		// Limita top 5
		if len(evo.TopQuedas) > 5 {
			evo.TopQuedas = evo.TopQuedas[:5]
		}
		if len(evo.TopAltas) > 5 {
			evo.TopAltas = evo.TopAltas[:5]
		}

		result.Lojas = append(result.Lojas, evo)
		totalProdutos += evo.TotalProdutos
		somaVariacao += evo.VariacaoMedia
		somaPreco += evo.PrecoMedioAtual
		contaPreco++
	}

	// Resumo global
	result.Resumo = EvolucaoResumo{
		TotalLojas:    len(result.Lojas),
		TotalProdutos: totalProdutos,
		TotalQuedas:   totalQuedas,
		TotalAltas:    totalAltas,
	}
	if contaPreco > 0 {
		result.Resumo.PrecoMedioGlobal = somaPreco / float64(contaPreco)
		result.Resumo.VariacaoMediaGlobal = somaVariacao / float64(contaPreco)
	}

	return result, nil
}
