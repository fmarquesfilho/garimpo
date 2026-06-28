---
title: "Dados e IA"
---


## Dados disponíveis

O BigQuery acumula snapshots de mercado (coletas periódicas) que sustentam análises:

- **Evolução de preço** por produto/loja ao longo do tempo
- **Detecção de tendências** — produtos em alta/queda
- **Sazonalidade** — variações por dia da semana / mês
- **Performance de afiliado** — conversões × publicações × canais
- **Comportamento de curadoria** — seleções por estratégia, hora, categoria

## Análises implementadas

| Análise | Endpoint | Descrição |
|---|---|---|
| Estatísticas de mercado | `GET /api/estatisticas` | Comissão média/mediana, preço, vendas por categoria |
| Evolução de lojas | `GET /api/lojas/evolucao` | Série temporal de preço médio por loja |
| Novidades de loja | `GET /api/lojas/novidades` | Produtos novos e variações de preço |
| Conversões reais | `GET /api/conversoes/reais` | Vendas do conversionReport Shopee |
| Histórico de coletas | `GET /api/coletas` | Coletas executadas nos últimos N dias |

## Pipeline de análise

```
BigQuery (snapshots, eventos, conversoes)
    │
    ├── /api/estatisticas → tela Estatísticas (resumo por categoria)
    ├── /api/lojas/evolucao → série temporal + top variações
    ├── /api/conversoes/reais → receita por canal e estratégia
    │
    └── (futuro) Python/Looker → análises offline
```

### Consultas úteis

```sql
-- Evolução de comissão por categoria
SELECT DATE(coletado_em) dia, categoria,
       AVG(comissao) comissao_media, AVG(preco) preco_medio
FROM `PROJECT.garimpo.snapshots`
WHERE coletado_em >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 30 DAY)
GROUP BY dia, categoria
ORDER BY dia;

-- Seleções por estratégia na última semana
SELECT estrategia, COUNT(*) selecoes, AVG(comissao) comissao_media, AVG(score) teor_medio
FROM `PROJECT.garimpo.eventos`
WHERE em >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 7 DAY)
GROUP BY estrategia;
```

## Motor de ranking (teor)

O score atual é simples:
```
teor = f(comissão, vendas, avaliação)
```

Com filtros de elegibilidade (pisos opcionais):
- `comissao_min` — descarta produtos com comissão abaixo do piso
- `vendas_min` — descarta produtos sem tração
- `nota_min` — descarta produtos mal avaliados

Exploração (epsilon-greedy): uma fração configurável de produtos fora do ranking
usual é incluída para descobrir oportunidades.

## Roadmap de IA

> 🔮 Planejado

### Curto prazo (Now)

- Score de teor composto (já implementado)
- Exploração epsilon-greedy (já implementado)
- Detecção de produtos-fantasma via selos

### Médio prazo (Next)

- **Recomendação personalizada** — aprender preferências do afiliado
  (categorias, faixa de preço, lojas favoritas)
- **Predição de conversão** — usar histórico de conversões para priorizar
  produtos com maior chance
- **Comparação de estratégias com dados reais** — quando houver volume
  suficiente de conversões

### Longo prazo (Later)

- **Geração de copy** — IA gerando legendas otimizadas para engajamento
- **Detecção de anomalias** — alertar sobre mudanças bruscas no mercado
  (change-point detection sobre série temporal de snapshots)
- **Clustering de nicho** — agrupar produtos similares para sugestão de
  novos nichos inexplorados
- **Sazonalidade** — STL decomposition quando houver semanas de histórico

## Ferramentas para análise offline

- **Looker Studio** (grátis): conecte ao dataset `garimpo` e monte painéis
- **Python**: `pip install pandas-gbq` →
  `pandas_gbq.read_gbq("SELECT * FROM garimpo.eventos", project_id="...")`
