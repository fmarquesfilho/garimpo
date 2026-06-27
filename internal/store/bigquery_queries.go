//go:build gcp

package store

import (
	"context"
	"time"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
)

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

// Conversoes retorna o relatório de publicações agrupado por canal e sub_id.
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
