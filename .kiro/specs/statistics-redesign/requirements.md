# Spec: Redesign da Página de Estatísticas

## Problema

A página de Estatísticas atual mostra "Mercado por categoria" — uma tabela com comissão média, preço médio, vendas média por categoria. Isso era útil na fase de prova de conceito mas não reflete os fluxos reais de uso:

1. **Categorias são rótulos manuais** — a Mileny digita "cosméticos" ao buscar, não vem da API
2. **Dados desconectados da ação** — saber que "cosméticos tem 12% de comissão média" não ajuda a decidir o que publicar
3. **Evolução de lojas existe** — mas depende de dados acumulados (que agora estão chegando)
4. **Métricas de publicação ausentes** — quantas publicações, em quais canais, qual retorno (precisa da spec de conversões)

## O que deveria mostrar (baseado nos fluxos reais)

### Para a Mileny (operadora)
- **Resumo do dia/semana:** quantos produtos publicados, quantas coletas rodaram, lojas monitoradas ativas
- **Evolução de preço das lojas** (já existe, funciona quando há 2+ coletas)
- **Performance por destino** (quando conversões estiverem integradas)
- **Produtos mais publicados** — top 5 que ela mais divulga

### Para o Fernando (admin/estrategista)
- **Volume de coletas** — tabela simples (o que tem na /coletas, mas compacto)
- **Saúde do sistema** — último erro, uptime, consumo BigQuery
- **Cobertura de catálogo** — % do catálogo de cada loja já escaneado (rotation_cursor)

## Proposta de redesign

### Seção 1: Resumo rápido (cards)
| Métrica | Fonte |
|---------|-------|
| Lojas monitoradas | COUNT buscas com shop_ids |
| Produtos rastreados | SUM dos últimos snapshots |
| Publicações (7 dias) | COUNT publicacoes WHERE criada_em > 7d |
| Coletas (7 dias) | COUNT snapshots WHERE coletado_em > 7d |

### Seção 2: Evolução de preço (já existe)
- Mini gráficos por loja
- Top quedas e altas
- Filtro de período

### Seção 3: Performance de publicações (depende de spec de conversões)
- Publicações por canal (Telegram vs WhatsApp)
- Taxa de envio com sucesso vs erro
- (Futuro) Conversões reais por canal

### Seção 4: Coletas (compacto)
- Tabela simples: última coleta por loja/keyword, quantos produtos, status
- Substitui a página /coletas inteira (que pode ser removida ou virar subseção)

## O que remover
- "Mercado por categoria" (tabela com comissão média/mediana por rótulo manual)
- A seção só faz sentido se categorias vierem da API (não é o caso)

## Dependências
- Spec de conversões (para seção 3)
- Spec de entidades (para clareza nos dados)
- 2+ semanas de coleta acumulada (para evolução ter significado estatístico)

## Decisões a tomar
- [ ] Manter /coletas como página separada ou absorver em /estatisticas?
- [ ] Mostrar dados de admin (consumo, erros) para a Mileny ou só em /admin?
- [ ] Remover "Mercado por categoria" agora ou aguardar redesign completo?
- [ ] Onde mostrar performance de publicações? (Estatísticas vs Publicações)
