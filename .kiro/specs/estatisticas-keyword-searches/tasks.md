# Tasks

## Task 1: Segmentar GET /estatisticas por fonte no Analyzer

- [x] Em `services/analyzer/routes/estatisticas.py`, adicionar uma segunda query que agrupa por `CASE WHEN keyword LIKE 'loja-%' THEN 'loja' ELSE 'keyword' END AS fonte`
- [x] Retornar novo campo `por_fonte` com sub-objetos `lojas` e `keywords`, cada um com `total_produtos`, `preco_medio`, `comissao_media`, `total_coletas` (COUNT DISTINCT keyword)
- [x] Manter os campos existentes (`total_amostras`, `resumo`) para backward compatibility
- [x] Se mock_data ativo, adicionar `por_fonte` ao mock response em `mock_data.py`
- [x] Verificar que `ruff check services/analyzer/` passa

## Task 2: Separar arrays lojas/keywords no GET /evolucao do Analyzer

- [x] Em `services/analyzer/routes/evolucao.py`, ao construir `lojas_lista`, separar em dois arrays: `lojas` (keyword começa com `loja-`) e `keywords` (demais)
- [x] Adicionar campo `resumo_keywords` com `total_quedas` e `total_altas` computados apenas de snapshots onde keyword NOT LIKE 'loja-%'
- [x] Manter o array `lojas` no response (para backward compat)
- [x] Adicionar novo array `keywords` no response (mesma shape: busca_id, pontos, variacao_media_pct, produtos)
- [x] Se mock_data ativo, adicionar `keywords` e `resumo_keywords` ao mock
- [x] Verificar com `ruff check services/analyzer/`

## Task 3: Refatorar página /estatisticas — novos MetricCards de keywords

- [x] Em `web/src/routes/estatisticas/+page.svelte`, extrair `por_fonte` do response de `buscarEstatisticas`
- [x] Adicionar MetricCard "Buscas keyword" com valor `por_fonte?.keywords?.total_coletas ?? 0`
- [x] Adicionar MetricCard "Produtos (kw)" com valor `por_fonte?.keywords?.total_produtos ?? 0`
- [x] Quando `total_coletas == 0`, os valores já aparecem como "0" naturalmente
- [x] Reorganizar grid para acomodar os novos cards (2 rows: row 1 = Lojas + Produtos + Publicações + Taxa; row 2 = Buscas kw + Produtos kw + Quedas + Altas)

## Task 4: Refatorar página /estatisticas — painel evolução keywords

- [x] Extrair `keywords` do response de `buscarEvolucaoLojas` (que agora retorna `keywords[]` além de `lojas[]`)
- [x] Adicionar DashPanel "📈 Preço médio (keywords)" abaixo do painel de lojas existente
- [x] Renderizar até 3 keyword entries (sorted por `|variacao_media_pct|` DESC) usando MiniChart
- [x] Se keyword tem apenas 1 ponto, renderizar Badge com preço em vez de MiniChart
- [x] Se `keywords` vazio, mostrar mensagem "Sem dados de buscas por keyword no período."
- [x] Extrair `resumo_keywords` e exibir breakdown abaixo dos MetricCards de Quedas/Altas: "{N} lojas · {M} kw"
- [x] Omitir breakdown se `resumo_keywords` ausente ou com zeros
- [x] Verificar `npm run check` e `npm run build` no diretório `web/`

## Task 5: Atualizar documentação — Fluxo 9 (Dashboard) em docs/08-fluxos-sequencia.md

- [x] Atualizar Fluxo 9 no mermaid diagram para mostrar os novos campos `por_fonte` e `keywords` na response do Analyzer
- [x] Adicionar nota na seção `<details>` explicando a lógica de classificação por prefixo `loja-`
- [x] Documentar que `resumo` continua global e `resumo_keywords` é o subset
- [x] Manter consistência com o data ownership model existente (sem novas fronteiras cruzadas)
- [x] Atualizar tabela de resumo de data ownership por caso de uso
