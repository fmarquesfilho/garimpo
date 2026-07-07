# Bugfix Requirements Document

## Introduction

Após a migração arquitetural do monolito Go para a arquitetura poliglota (C# API + Go Collector + Python Analyzer), a página "Buscas Agendadas" (`/lojas`) não foi validada end-to-end para os dois fluxos principais de monitoramento de lojas.

O sistema segue um modelo de **data ownership** estrito:
- **C# API**: dono do PostgreSQL (CRUD de Buscas, Destinos, etc.) — único ponto de entrada para o frontend
- **Go Collector**: I/O externo (Shopee API) — escreve snapshots no BigQuery, nunca acessa PostgreSQL
- **Go Scheduler**: orquestra coletas periódicas via cron — toda busca agendada DEVE ser registrada no Scheduler via gRPC `SetSchedule`, que então dispara Cloud Tasks para executar coletas
- **Python Analyzer**: leitura analítica do BigQuery — detecta novidades e variações de preço

**Regra arquitetural:** Nenhum agendamento pode ocorrer fora do Scheduler. O C# API DEVE chamar `Scheduler.SetSchedule` via gRPC ao criar/atualizar uma busca agendada. O Scheduler é o único serviço que detém cron jobs e Cloud Tasks.

Atualmente, o `POST /api/lojas` persiste a Busca no PostgreSQL (via C# API) e resolve o shop_id (via Collector gRPC), mas **NÃO** registra o agendamento no Scheduler. Isso significa que buscas criadas pós-migração não são coletadas periodicamente.

Os dois fluxos que precisam de validação E2E:
1. **Monitorar todos os produtos de uma loja** — Busca com `ShopIds` preenchido e sem keywords → Scheduler agenda coleta de TODOS os produtos
2. **Monitorar produtos filtrados por keywords** — Busca com `keywords[]` → Scheduler agenda coleta filtrada por cada keyword

## Bug Analysis

### Current Behavior (Defect)

1.1 WHEN a shop is added via POST /api/lojas THEN the system persists the Busca in PostgreSQL but does NOT call Scheduler gRPC SetSchedule to register periodic collection, so the busca is never actually collected

1.2 WHEN a shop-based Busca exists without keywords (monitoring all products) THEN the Scheduler has no job registered for that busca, so no periodic Collector gRPC Fetch is triggered for that shop_id

1.3 WHEN a scheduled search with keywords AND shop_ids is created THEN the Scheduler has no job registered to periodically fetch filtered products matching those keywords from the monitored shop

1.4 WHEN the frontend calls GET /api/lojas/novidades with a busca_id (C# proxies to Analyzer) THEN the response is always empty because no snapshots were ever collected by the Scheduler → Collector → BigQuery pipeline

1.5 WHEN a busca is removed (soft-deleted) via DELETE /api/lojas THEN the system does NOT call Scheduler to remove/pause the corresponding job, leaving orphan jobs in the Scheduler

### Expected Behavior (Correct)

2.1 WHEN a shop is added via POST /api/lojas THEN the C# API SHALL persist the Busca in PostgreSQL AND call Scheduler gRPC SetSchedule to register a periodic collection job with the resolved shop_id and default cron (e.g., every 8h)

2.2 WHEN a shop-based Busca exists without keywords THEN the Scheduler SHALL have a registered job that periodically calls Collector gRPC Fetch with the shop_id, collecting ALL products from that shop into BigQuery snapshots

2.3 WHEN a scheduled search exists with keywords AND shop_ids THEN the Scheduler SHALL have a registered job that periodically calls Collector gRPC Fetch with both shop_id and keyword parameters, collecting only matching products

2.4 WHEN the frontend calls GET /api/lojas/novidades with a valid busca_id THEN the C# API SHALL proxy to Analyzer (GET /novidades?busca_id=X&dias=7), and the Analyzer SHALL return produtos_novos[] and variacoes[] computed from BigQuery snapshots produced by the Scheduler's periodic collections

2.5 WHEN a busca is removed (soft-deleted) via DELETE /api/lojas THEN the C# API SHALL call Scheduler gRPC SetSchedule with enabled=false (or equivalent) to pause/remove the corresponding collection job

### Unchanged Behavior (Regression Prevention)

3.1 WHEN a shop URL or username is submitted via POST /api/lojas THEN the system SHALL CONTINUE TO resolve the shop_id via C# API → Collector gRPC ResolveShop and persist a Busca entity with ShopIds[] in PostgreSQL (C# owns the write)

3.2 WHEN an invalid shop URL is submitted via POST /api/lojas THEN the system SHALL CONTINUE TO return HTTP 400 with a user-friendly error message (gRPC NotFound/InvalidArgument mapped to BadRequest)

3.3 WHEN GET /api/lojas is called THEN the system SHALL CONTINUE TO return all active Buscas from PostgreSQL with id, keyword, shop_ids, source_url, ativo, and criado_em fields

3.4 WHEN the Analyzer service is unavailable THEN GET /api/lojas/novidades SHALL CONTINUE TO return a graceful empty response (produtos_novos=[], variacoes=[], total_atual=0) — respecting isolamento de falhas

3.5 WHEN the Collector gRPC is unavailable THEN POST /api/lojas SHALL CONTINUE TO return HTTP 400 with error "Falha ao resolver o ID da loja via Collector" — the C# API never accesses marketplace APIs directly

3.6 WHEN the Scheduler gRPC is unavailable at the time of busca creation THEN the C# API SHALL still persist the Busca in PostgreSQL (eventual consistency) but SHALL log a warning — the busca can be scheduled later via reconciliation
