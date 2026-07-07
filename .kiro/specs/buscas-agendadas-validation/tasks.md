# Buscas Agendadas Validation — Tasks

## Task 1: Register SchedulerServiceClient in DI
- [ ] Add `SchedulerServiceClient` gRPC client registration in `src/Garimpei.Infrastructure/DependencyInjection.cs`
- [ ] Read address from config: `configuration["Grpc:SchedulerAddress"] ?? "http://localhost:50054"`
- [ ] Add `Grpc:SchedulerAddress` to `appsettings.Development.json`
- [ ] Verify build passes with new client registered

## Task 2: Add Keywords and CronExpression fields to Busca entity
- [ ] Add `public string[]? Keywords { get; set; }` to `src/Garimpei.Domain/Entities/Busca.cs`
- [ ] Add `public string? CronExpression { get; set; }` to `src/Garimpei.Domain/Entities/Busca.cs`
- [ ] Create EF Core migration: `dotnet ef migrations add AddKeywordsAndCronToBusca`
- [ ] Verify migration generates correct columns (text[] for Keywords, text for CronExpression)
- [ ] Apply migration locally and in production (`mise run deploy:migrate`)

## Task 3: Update AdicionarLojaRequest to accept Keywords
- [ ] Add `public string[]? Keywords { get; init; }` to `AdicionarLojaRequest` record
- [ ] Update `POST /api/lojas` handler to populate `busca.Keywords` from request
- [ ] Update `POST /api/lojas` handler to populate `busca.CronExpression` from `req.Cron`
- [ ] Update JSON schema `contracts/schemas/lojas.request.json` to include `keywords` field
- [ ] Verify `mise run check:service-contracts` passes

## Task 4: Integrate Scheduler SetSchedule in POST /api/lojas
- [ ] Inject `Scheduler.V1.SchedulerService.SchedulerServiceClient` in POST handler
- [ ] After `db.SaveChangesAsync()`, call `SetScheduleAsync` with: job_id=`busca-{Id}`, cron (default `0 */8 * * *`), enabled=true, params={shop_id, owner_uid, type, keywords?}
- [ ] Wrap in try/catch: if Scheduler unavailable, log warning but still return 200 (eventual consistency)
- [ ] Inject `ILogger<>` for warning logging
- [ ] Verify POST still returns same response shape (id, keyword, shop_ids, source_url, status)

## Task 5: Integrate Scheduler SetSchedule in DELETE /api/lojas
- [ ] Inject `Scheduler.V1.SchedulerService.SchedulerServiceClient` in DELETE handler
- [ ] After soft-delete (`Active=false`), call `SetScheduleAsync` with: job_id=`busca-{id}`, enabled=false
- [ ] Wrap in try/catch: if Scheduler unavailable, log warning but still return 200
- [ ] Verify DELETE still returns same response shape (status, id)

## Task 6: Update contracts registry and response schema
- [ ] Add boundary `api-scheduler-set-schedule` in `contracts/registry.yaml` (source: csharp-api, target: scheduler, protocol: grpc, service: scheduler.v1.SchedulerService, method: SetSchedule)
- [ ] Update `contracts/schemas/lojas.request.json` with `keywords` field (type: array, items: string)
- [ ] Update GET /api/lojas response to include `keywords` and `cron_expression` fields
- [ ] Run `mise run check:service-contracts` — verify all pass

## Task 7: Update frontend API and components
- [ ] Update `adicionarLoja()` in `web/src/lib/api.js` to accept and send `keywords` field
- [ ] Update `FormAdicionarLoja.svelte` or `GerenciarBuscas.svelte` to allow keyword input for filtered searches
- [ ] Ensure GET /api/lojas response displays keywords in the lojas list UI

## Task 8: Write E2E test for scheduled search flows
- [ ] Create `web/tests/buscas-agendadas.spec.js` with test cases:
  - POST loja without keywords → verify Scheduler job created (all products)
  - POST loja with keywords → verify Scheduler job has keywords in params
  - DELETE loja → verify Scheduler job is paused
  - Scheduler offline → verify Busca persists (200 returned)
  - GET /api/lojas returns keywords field
- [ ] Add mise task `test:e2e:buscas-agendadas` (similar to `test:e2e:lojas` but includes Scheduler Go)

## Task 9: Write C# integration tests
- [ ] Test: POST /api/lojas with mock Scheduler → verify SetSchedule called with correct params
- [ ] Test: POST /api/lojas with keywords → verify Keywords persisted in DB and passed to Scheduler
- [ ] Test: DELETE /api/lojas → verify SetSchedule called with enabled=false
- [ ] Test: Scheduler unavailable → verify Busca persists, warning logged, HTTP 200 returned
- [ ] Test: Preservation — GET /api/lojas response shape unchanged
- [ ] Test: Preservation — POST error handling (Collector NotFound → 400) unchanged

## Task 10: Update documentation
- [ ] Update `docs/03-fluxos-e-modelo.md` — buscas agendadas section with Scheduler integration
- [ ] Update `docs/08-fluxos-sequencia.md` — add/update sequence diagram for scheduled search creation
- [ ] Update `docs/06-qualidade-e-testes.md` — add E2E buscas-agendadas test documentation
- [ ] Update `docs/02-arquitetura.md` — scheduler section to mention busca-triggered jobs
- [ ] Commit and push all changes
