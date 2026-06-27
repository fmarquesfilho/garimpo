//go:build gcp

package store

import (
	"context"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
)

// Novidades compara snapshots da janela para detectar produtos novos e variações de preço.
// Se buscaID não estiver vazio, filtra por keyword que contenha o id da busca.
func (s *BigQueryStore) Novidades(ctx context.Context, buscaID string, dias int) (NovidadesLojas, error) {
	if dias <= 0 {
		dias = 7
	}
	result := NovidadesLojas{BuscaID: buscaID, DiasJanela: dias}

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
			ProdutoID     string  `bigquery:"produto_id"`
			Nome          string  `bigquery:"nome"`
			Preco         float64 `bigquery:"preco"`
			Comissao      float64 `bigquery:"comissao"`
			Vendas        int     `bigquery:"vendas"`
			Nota          float64 `bigquery:"nota"`
			DetectadoEm   string  `bigquery:"detectado_em"`
			Aparicoes     int     `bigquery:"aparicoes"`
			RnDesc        int     `bigquery:"rn_desc"`
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

		if row.Aparicoes == 1 {
			result.ProdutosNovos = append(result.ProdutosNovos, ProdutoNovo{
				ProdutoID: row.ProdutoID, Nome: row.Nome, Preco: row.Preco,
				Comissao: row.Comissao, Vendas: row.Vendas, Nota: row.Nota,
				DetectadoEm: row.DetectadoEm,
			})
		}

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
func (s *BigQueryStore) EvolucaoLojas(ctx context.Context, dias int) (EvolucaoLojasResult, error) {
	if dias <= 0 {
		dias = 30
	}
	result := EvolucaoLojasResult{DiasJanela: dias}

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

		evo := EvolucaoLoja{BuscaID: b.ID, Coletas: len(pontos), Pontos: pontos}
		primeiro := pontos[0]
		ultimo := pontos[len(pontos)-1]
		evo.PrecoMedioInicio = primeiro.PrecoMedio
		evo.PrecoMedioAtual = ultimo.PrecoMedio
		evo.TotalProdutos = ultimo.Produtos

		if primeiro.PrecoMedio > 0 {
			evo.VariacaoMedia = (ultimo.PrecoMedio - primeiro.PrecoMedio) / primeiro.PrecoMedio
		}

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
