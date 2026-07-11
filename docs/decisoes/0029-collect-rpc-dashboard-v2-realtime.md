# ADR-0029: Collect RPC + Dashboard v2 + Realtime Polling

**Data:** 2026-07-11
**Status:** Aceita
**Decisores:** Fernando
**PR:** [#44](https://github.com/fmarquesfilho/garimpo/pull/44)

## Contexto

O sistema de monitoramento Garimpei tinha três lacunas arquiteturais:

1. **Persistência de snapshots acoplada ao Scheduler** — o Scheduler importava `internal/store` para gravar no BigQuery após buscar produtos via `Fetch`/`FetchShop`. Isso violava a separação de responsabilidades (ADR-0018 prevê exporters no Collector) e impedia o Analyzer de ter dados para detectar novidades e quedas.

2. **Dashboard baseado em métricas de vaidade** — a página `/estatisticas` mostrava total de produtos, preço médio e comissão média — dados sem ação. O usuário não sabia se as coletas estavam rodando, o que publicar agora, ou se estava ganhando dinheiro.

3. **Dashboard estático** — exigia refresh manual para ver dados novos. Sem indicação de freshness, sem detecção de mudanças, sem percepção de "sistema vivo".

## Decisão

Implementar três specs complementares em um único PR:

### 1. Collect RPC (`collector-collect-rpc`)

- Novo `Collect` RPC no CollectorService com `oneof target` (keyword ou shop_id)
- Combina busca de produtos com persistência assíncrona no BigQuery via buffered channel (cap: 64)
- Scheduler migra de `Fetch` + persistência local para `Collect` — perde dependência de `internal/store`
- Export é config-driven: BigQuery exporter controlado pelo YAML do Collector
- Graceful degradation: falha no BigQuery nunca impacta o response ao Scheduler

### 2. Analyzer Dashboard v2 (`analyzer-dashboard-v2`)

- 4 novos endpoints no Analyzer Python: `/coletas/saude`, `/oportunidades/agora`, `/conversoes/resumo`, `/alertas/eficacia`
- Frontend reestruturado em 3 seções orientadas por perguntas: "Coletas estão rodando?", "O que publicar agora?", "Estou ganhando dinheiro?"
- Proxy via C# API (padrão existente) — 4 novos routes em `AnalyzerProxyEndpoints.cs`
- Endpoints existentes (7) mantidos sem alteração (backward compatible)
- Graceful degradation: tabelas inexistentes retornam empty-state estruturado, nunca HTTP 500

### 3. Dashboard Realtime (`dashboard-realtime`)

- Smart polling a cada 30s via endpoint leve `/api/dashboard/changes` (3 timestamps MAX)
- Refresh seletivo: apenas seções cujo dado mudou são refetched
- Tab visibility: polling pausa quando tab está em background
- Freshness indicator: dot pulsante + countdown + "há Xs"
- Animated transitions: valores numéricos interpolados via Svelte tweened (600ms ease-out)
- Backoff exponencial em falhas (3 erros consecutivos → intervalo dobra)

## Consequências

### Positivas

- Snapshots fluem para BigQuery a cada coleta → Analyzer detecta quedas e novidades
- Scheduler 100% limpo de `internal/store` → arch-go passa sem exceções
- Dashboard responde às 3 perguntas do usuário em < 3s (parallel fetch)
- Percepção de real-time sem infra nova (sem WebSockets, sem Durable Objects, sem Firestore)
- Multi-tenant isolado em todos os endpoints (owner_uid via JWT + EF Core query filter + BigQuery WHERE)
- Custo mínimo de polling: MAX() em colunas timestamp é near-free no BigQuery (< $0.01/mês)

### Negativas / Riscos

- Latência de até 30s para refletir uma coleta no dashboard (aceitável para o caso de uso)
- Buffer full no Collector descarta snapshots (mitigado: 64 slots >> 10-20 jobs por ciclo)
- `persisted=true` no response não garante escrita — mas próximo ciclo do Scheduler regenera

### Neutras

- Fetch/FetchShop RPCs permanecem inalterados (clientes existentes compatíveis)
- Proto `oneof target` é mais simples que 2 RPCs separados (1 handler, 1 call site)
- Polling configurável via `VITE_POLL_INTERVAL_MS` (10s-120s, default 30s)

## Alternativas Consideradas

### Persistência

| Alternativa | Descartada porque |
|-------------|-------------------|
| Scheduler persiste (status quo) | Viola boundaries, Scheduler não deveria importar `internal/store` |
| 2 RPCs separados (`CollectKeyword` + `CollectShop`) | Duplicação de código, surface area maior |
| Persistência síncrona no Collect | Adicionaria ~100ms à latência do response |
| Retry de snapshots falhados | Complexidade desnecessária — ciclo seguinte regenera |

### Dashboard

| Alternativa | Descartada porque |
|-------------|-------------------|
| Server-Sent Events (SSE) | Cloud Run fecha conexões longas (timeout 60s), requer infra de reconexão |
| Cloudflare Durable Objects (WebSocket) | Infra nova, custo adicional, overengineering para ≤30s de latência |
| Firebase Realtime Database | Vendor coupling, exige listener no backend para população |
| Long polling | Mais complexo que smart polling sem benefício real (30s é aceitável) |

## Validação

Testes E2E cross-service (61 checks, 0 falhas):

| Suite | Checks |
|-------|--------|
| Full Pipeline (trace único) | 17 ✅ |
| Coleta (Scheduler → Collector → BigQuery) | 13 ✅ |
| Alertas (Analyzer → Cloud Tasks → Publisher) | 12 ✅ |
| Publicações (API → Publisher gRPC → Telegram) | 9 ✅ |
| Dashboard (change detection + polling) | 13 ✅ |
| Services (health + endpoints) | 8 ✅ |
| Scheduler (agendamento + trigger) | 5 ✅ |
| Traces (OTel propagation) | 3 ✅ |
| Analyzer (BigQuery queries) | 3 ✅ |

Drift checks (8/8):
- fixtures-contract, api-contract, schema-sync, service-contracts
- stale-refs, data-ownership, docs-drift, config-consistency

Unit tests: 298 (frontend) + 72 (C# integration) = 370 passando.

## Evolução Futura

- Se sub-segundo se tornar necessário: Cloudflare Durable Objects com WebSocket (frontend contract não muda — swap timer por message handler)
- Se volume crescer: SharedWorker para coordenar polling entre tabs
- Se alertas precisarem de push imediato: Firebase Cloud Messaging para notificações mobile

## Arquivos-Chave

| Arquivo | O quê |
|---------|-------|
| `protos/collector/v1/collector.proto` | Definição do Collect RPC + messages |
| `services/collector/server.go` | Implementação: Collect + export goroutine |
| `services/scheduler/jobs.go` | Migração para Collect (sem `internal/store`) |
| `services/analyzer/routes/saude.py` | `/coletas/saude` |
| `services/analyzer/routes/oportunidades.py` | `/oportunidades/agora` |
| `services/analyzer/routes/resumo_conversoes.py` | `/conversoes/resumo` |
| `services/analyzer/routes/eficacia.py` | `/alertas/eficacia` |
| `services/analyzer/routes/changes.py` | `/dashboard/changes` (change detection) |
| `src/Garimpei.Api/Endpoints/AnalyzerProxyEndpoints.cs` | Proxy C# → Analyzer |
| `web/src/lib/polling.svelte.js` | PollingTimer (30s, visibility, backoff) |
| `web/src/lib/components/ui/FreshnessBar.svelte` | Indicador de freshness |
| `web/src/lib/components/ui/AnimatedMetric.svelte` | Transições numéricas |
| `web/src/routes/estatisticas/+page.svelte` | Dashboard reestruturado |
