---
title: "Arquitetura"
---


## Visão geral

```
navegador ──https──► Cloud Run (garimpo-api)
                       ├─ /           → frontend estático (SPA, web/build)
                       ├─ /api/**     → API JSON (curadoria, publicação, coleta)
                       ├─ Secret Manager: SHOPEE_*, TELEGRAM_*, WHATSAPP_*
                       └─ grava eventos → BigQuery
                                             └─ Looker Studio / Python

Cloud Scheduler ──cron──► POST /api/coletar ──► BigQuery (snapshots)
```

## Stack

| Camada | Tecnologia |
|---|---|
| Backend | Go 1.26, Cloud Run (southamerica-east1) |
| Frontend | SvelteKit 2, Svelte 5, Vite 8 |
| Persistência | BigQuery (analítico), localStorage (favoritos local) |
| Autenticação | Firebase Auth (Bearer token) |
| Canais | Telegram Bot API, WhatsApp via Maytapi |
| CI/CD | GitHub Actions (`deploy-gcp.yml`) |
| Infra | Artifact Registry, Secret Manager, Cloud Scheduler |

## Deploy (Cloud Run)

Deploy automático via GitHub Actions em push para `main`:

1. Testes Go (`go test ./...`) — gate
2. Build Docker multi-stage (Node → frontend, Go → backend)
3. Push para Artifact Registry
4. Deploy no Cloud Run (região `southamerica-east1`)
5. Atualiza Cloud Scheduler

Custo: Cloud Run escala a zero, BigQuery free tier (10 GB storage + 1 TB consulta/mês).

Runbook completo para setup do zero: ver `docs/DEPLOY_GCP.md` (legado, ainda válido para GCP).

Ver [ADR 0003](/docs/decisoes/0003-deploy-gcp/).

## Coleta e scheduler

O Cloud Scheduler dispara `POST /api/coletar` para cada busca ativa no horário do cron.

Cada coleta:
1. Consulta a API de afiliados Shopee (keyword/categoria ou loja)
2. Aplica ranking (estratégia nicho)
3. Grava snapshot no BigQuery (`snapshots`)
4. Verifica alertas de preço (se configurados)
5. Detecta novidades (produtos que não existiam na coleta anterior)

### Amostragem rotativa (lojas)

Para lojas monitoradas, a coleta usa paginação rotativa:
- `rotation_cursor` armazena a próxima página por loja
- Cada ciclo avança 2 páginas (100 produtos)
- `full_scan_at` registra quando completou varredura do catálogo inteiro
- Loja com 500 produtos é coberta em ~5 coletas

### Throttling

- 200ms entre requisições de páginas da mesma loja
- 60s entre lojas diferentes numa mesma execução
- HTTP 429 → espera 30s, retenta até 3×

## Análise estática

- **golangci-lint** — estilo e bugs no Go (`.golangci.yml`)
- **arch-go** — restrições arquiteturais (`arch-go.yml`)
  - `internal/domain` não importa infra
  - `internal/httpapi` não acessa BigQuery diretamente
  - `cmd/` só importa `internal/`
- **eslint + stylelint** — frontend
- **check-file-size** — máx 400 linhas por arquivo de produção

## CI/CD Pipeline

```
push main → test-go → test-web → lint → build → deploy
              │          │         │
              │          │         └─ golangci-lint, arch-go, eslint
              │          └─ vitest (109 testes, <2s)
              └─ go test (~200 testes)
```

Pushes que só tocam `docs/**` são ignorados via `paths-ignore`.

## Logs

Logging estruturado via `slog`:
- Requisições: INFO (health em DEBUG)
- Erros 5xx: ERROR
- Eventos coleta/publicação: INFO com campos (`rota`, `categoria`, `coletados`)
- Cloud Run → JSON → Cloud Logging (filtro por `severity` e campos)
- Ajuste via `LOG_LEVEL=debug|info|warn|error`
