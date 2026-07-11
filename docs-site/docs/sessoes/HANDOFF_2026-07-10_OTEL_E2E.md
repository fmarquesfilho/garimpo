# Handoff — OTel + E2E + Bugs v3 (2026-07-10)

> Próxima sessão: testes E2E cross-service (scheduler → collector → analyzer → publisher)
> usando OTel para debug. Foco em fluxos de monitoramento (quedas de preço + produtos novos).
> Branch: main.

## Estado atual

- **298 unit tests passando** (frontend vitest)
- **72 C# tests passando** (arch + integration)
- **24 E2E local passando** (Playwright, mock)
- **8 E2E produção passando** (Playwright, APIs reais)
- **16 E2E services passando** (scheduler + analyzer + services)
- **3 E2E traces passando** (OTel propagation validada)
- OTel em produção: Cloud Trace habilitado, role `cloudtrace.agent` adicionada
- Banco de produção (Neon) limpo em 2026-07-10
- Frontend em Cloudflare Pages (adapter-cloudflare)
- Bun como package manager (substituiu npm)

## Decisões tomadas nesta sessão

### ADR-0028: adapter-static → adapter-cloudflare + OTel
- SvelteKit com `instrumentation.server.js` (traces nativos)
- OTel SDK direto para `telemetry.googleapis.com:443` (sem sidecar collector)
- Frontend propaga `traceparent` em todo fetch

### ADR-0025 (revisada): Deploys independentes
- deploy-backend não depende mais de frontend/contracts
- Cada stack deploya somente se seus próprios jobs passam
- `.github/workflows/**` agora trigga deploy-backend

### Arquitetura OTel (zero sidecar)
```
Go/C#/Python → OTLP gRPC → telemetry.googleapis.com:443 → Cloud Trace
Browser → traceparent header → C# API → propaga para Go/Python
```

### Migração npm → bun
- `bun.lock` substitui `package-lock.json`
- CI usa `oven-sh/setup-bun@v2`
- mise.toml: todos os comandos `npm run` → `bun run`

### Engine: init sem busca automática
- `#inicializar` não chama `#executarBusca()` se ctx não tem keyword/loja/categoria
- Evita timeout de 25s com buscas salvas que têm lojas
- Empty state instantâneo na primeira visita

### Filtros client-side corrigidos
- `!r.vendas` → `r.vendas == null` (0 é valor válido, não ausência)
- Mesmo para comissão

### shopNomes persistido corretamente
- `#salvar()` agora inclui `shopNomes: this.ctx.shopNomes` no payload
- Backend persiste `ShopNames` (jsonb) no PostgreSQL

## Objetivo da próxima sessão

### Testes E2E cross-service com OTel debug

Exercitar os fluxos que envolvem **todos os services** juntos:

1. **Coleta agendada** (Scheduler → Collector → BigQuery)
   - Criar busca com cron → verificar que o Scheduler registra job
   - Trigger manual do job → verificar que Collector.Fetch executa
   - Verificar que snapshots são salvos no BigQuery

2. **Detecção de quedas** (Analyzer → BigQuery)
   - Inserir snapshots com preço variado → chamar /quedas
   - Verificar que variação é calculada corretamente
   - Verificar threshold de alerta

3. **Detecção de produtos novos** (Analyzer → BigQuery)
   - Inserir snapshot novo → chamar /novidades
   - Verificar que produto aparece como "novo"

4. **Alerta de preço** (Cloud Tasks → Scheduler → Analyzer → Publisher)
   - Enqueue alert task → verificar processamento
   - Verificar que Publisher.Publish é chamado com mensagem formatada
   - Verificar entrega no Telegram (chat de teste)

5. **Publicação agendada** (Scheduler → C# API → Publisher)
   - Criar publicação agendada → verificar job no Scheduler
   - Trigger → verificar envio via Publisher

### Usando OTel para debug

- Cada teste gera um `traceparent` com trace_id conhecido
- Após execução, buscar trace no Cloud Trace
- Verificar spans de cada service na timeline
- Se falha: usar `mise run debug:trace <id>` para investigar

### Multi-tenant (LGPD)

Os testes devem considerar isolamento:
- Dados do tenant de teste (`e2e@garimpei.app.br`) isolados por `owner_uid`
- Buscas/coletas de um tenant não vazam para outro
- O `EF Core query filter` garante isolamento no PostgreSQL
- BigQuery: filtro por `owner_uid` nas queries do Analyzer

### Dados fictícios para testes

Usar os `fixtures/` compartilhados como base:
- `fixtures/lojas.json` — 3 lojas de teste
- `fixtures/produtos.json` — 10 produtos determinísticos
- Para BigQuery: inserir snapshots via API ou script direto

## Arquivos-chave

| Arquivo | O quê |
|---------|-------|
| `internal/otel/otel.go` | SDK OTel Go (Init, interceptors, slog handler) |
| `services/scheduler/jobs.go` | Lógica de coleta (Fetch/FetchShop) + Cloud Tasks |
| `services/scheduler/alerts.go` | Processamento de alertas (/process-alert) |
| `services/analyzer/routes/quedas.py` | Detecção de quedas de preço |
| `services/analyzer/routes/novidades.py` | Detecção de produtos novos |
| `src/Garimpei.Api/Endpoints/SchedulerJobs.cs` | Montagem do SetScheduleRequest |
| `src/Garimpei.Api/Endpoints/ScheduledPublishEndpoints.cs` | Publicação agendada |
| `internal/taskqueue/taskqueue.go` | Cloud Tasks enqueue |
| `deploy/cloud-run-deploy-now.yaml` | Manifest multi-container |
| `.mise/tasks/test/e2e-scheduler` | Teste E2E do fluxo scheduler |
| `.mise/tasks/debug/trace` | Investigação de trace por ID |

## Documentos importantes

| Documento | Relevância |
|-----------|-----------|
| `docs/observability.md` | Guia completo de OTel (arquitetura, config, Jaeger local) |
| `docs/decisoes/0028-adapter-cloudflare-otel.md` | ADR do OTel + adapter |
| `docs/decisoes/0025-ci-deploy-conservador.md` | ADR revisada (deploys independentes) |
| `.kiro/specs/unified-observability/` | Spec completo (requirements + design + tasks) |
| `fixtures/README.md` | Como usar os dados de teste cross-stack |
| `docs/09-api-reference.md` | Referência da API (com shop_names, FetchShop) |
| `docs/componentes.md` | Arquitetura frontend (BuscaEngine, modos, raias) |
| `docs/06-qualidade-e-testes.md` | Infraestrutura de testes e drift checks |

## Como verificar

```bash
# Testes unitários
cd web && bunx vitest run

# E2E local (24 testes, mock)
mise run test:e2e-local

# E2E produção (8 testes, APIs reais)
mise run test:e2e-prod

# Services (16 checks contra produção)
mise run test:e2e-services
mise run test:e2e-scheduler
mise run test:e2e-analyzer

# OTel traces
mise run test:e2e-traces

# Debug
mise run debug:health
mise run debug:logs -- --severity ERROR --minutes 60
mise run debug:trace <trace_id>

# Limpar banco produção
mise run db:reset -- --prod
mise run db:reset -- --prod --bq
```

## Steering rules ativas

- `git.md` — nunca `--no-verify`, nunca push automático
- `ci.md` — nunca E2E real no CI, deploy conservador
- `dependencies.md` — sem deps com >3 meses sem release
