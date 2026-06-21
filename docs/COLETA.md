# Coleta de Dados — o que é guardado, onde, e como começar já

Responde direto: **sim, a aplicação coleta dados**, e este documento descreve o
quê, onde fica, e como ligar a coleta periódica agora.

## O que é coletado e onde fica

Todo o armazenamento é no **BigQuery** (dataset `garimpo`), no build de produção
(imagem com a tag `gcp`). Localmente, sem o BigQuery, nada persiste (NopStore) —
o uso é o mesmo, só não grava.

| Tabela | Quem grava | Quando | Para quê |
|---|---|---|---|
| `snapshots` | `POST /api/coletar` | coleta agendada (Cloud Scheduler) | série temporal do mercado por categoria (preço, comissão, vendas, nota, teor, posição) com `coletado_em` |
| `eventos` | `POST /api/eventos` (seleção) e `POST /api/publicar` | quando ela garimpa/publica | comportamento de curadoria + atribuição (`sub_id`, canal) |
| `buscas` | `POST /api/buscas` | ao salvar um perfil de busca | perfis de coleta (filtros + cron), append-only/versionado |
| `conversoes` | (futuro) `conversionReport` da Shopee | quando ligado | fecha o laço receita ↔ curadoria |

O schema está em `deploy/bigquery_schema.sql` (rode-o uma vez; ver
`docs/DEPLOY_GCP.md`). Cada tabela é particionada por data para consulta barata.

## Como ler o que foi coletado

`GET /api/estatisticas?dias=30` devolve o resumo descritivo dos snapshots por
categoria (média e mediana de comissão, preço médio, vendas média, teor médio).
É o primeiro passo do pipeline de análise e já aparece na tela **Estatísticas**
do app. Internamente é uma consulta agregada na tabela `snapshots`.

## Buscas salvas → coleta agendada (o ciclo)

```
[buscas salvas]  ──save──►  localStorage (uso manual imediato)
   (perfis de         └sync─►  /api/buscas ──► BigQuery `buscas`
    filtros+cron)                                   │
                                                    ▼
                          scheduler-exemplo.sh lê os perfis com cron
                                                    │
                                                    ▼
                         Cloud Scheduler: 1 job por perfil (no cron dele)
                                                    │  POST /api/coletar
                                                    ▼
                                 BigQuery `snapshots`  ──►  /api/estatisticas
                                  (com coletado_em)            (tela Estatísticas)
```

Uma **busca salva** é um conjunto nomeado de filtros (keyword, categoria,
estratégia, pisos, top) com um **cron** opcional. Ela serve para dois fins:
reusar filtros num clique (manual) e definir a coleta periódica (automático).
Vivem no navegador (localStorage) **e** sincronizam no servidor (BigQuery), para
a coleta não depender do navegador estar aberto.

## Começar a coletar AGORA (passo a passo)

A coleta a longo prazo só rende se começar cedo — então vale ligar já, mesmo com
poucos perfis.

1. **Crie o token de coleta** (uma vez), que protege o endpoint:
   ```bash
   printf '%s' "$(openssl rand -hex 24)" | gcloud secrets create COLETA_TOKEN --data-file=-
   gcloud run services update garimpo-api --region southamerica-east1 \
     --update-secrets COLETA_TOKEN=COLETA_TOKEN:latest
   ```
2. **Salve 2–4 buscas** na tela de Curadoria (campo "nome do perfil" + cron). Ex.:
   perfumaria `0 8 * * *`, skincare `10 8 * * *`, maquiagem `20 8 * * *`
   (horários espaçados para respeitar o rate limit da Shopee). Ao salvar, elas
   sincronizam para o BigQuery.
3. **Crie os jobs do Scheduler** a partir dos perfis:
   ```bash
   # precisa de gcloud autenticado e jq instalado
   ./deploy/scheduler-exemplo.sh           # use DRY_RUN=1 antes para conferir
   ```
4. **Confira**:
   ```bash
   gcloud scheduler jobs list --location southamerica-east1
   # dispara uma coleta na hora para testar:
   gcloud scheduler jobs run coleta-perfumaria-diaria --location southamerica-east1
   ```
5. **Veja os dados chegando** na tela **Estatísticas** (ou em `/api/estatisticas`).
   Os primeiros snapshots aparecem após a primeira execução.

## Logs

A aplicação loga de forma estruturada por criticidade (`slog`): requisições em
INFO (health em DEBUG), 5xx em ERROR, e eventos de `coleta`/`publicacao` em INFO
com seus campos. No Cloud Run sai em JSON e o Cloud Logging deixa filtrar por
`severity` e por campos (`rota`, `categoria`, `coletados`…). Ajuste o volume com
`LOG_LEVEL=debug|info|warn|error`.

## Próximo incremento (não agora)

Com `snapshots` acumulando, o pipeline de análise cresce sobre `/api/estatisticas`:
tendência por categoria (variação período-a-período), depois STL/change-point
quando houver semanas de histórico (ver `docs/CIENCIA_DE_DADOS.md`). O endpoint
já devolve o agregado por categoria; o próximo passo é a evolução temporal.
