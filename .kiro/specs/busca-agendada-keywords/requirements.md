# Requirements Document

## Introduction

Permitir que o usuário crie buscas agendadas por palavras-chave que são executadas diretamente na API da Shopee via Scheduler, sem necessidade de ter uma loja (shop_id) cadastrada. O backend já suporta a mecânica (`SchedulerJobs` + Scheduler Go com branch `keyword_search`), mas o frontend envia o cron ao backend de forma inconsistente e o fluxo de criação/visualização precisa funcionar de ponta a ponta sem dependência de loja.

## Glossary

- **Busca_Agendada**: Entidade `Busca` no PostgreSQL com `CronExpression` definida e sem `ShopIds`, representando uma coleta periódica por palavras-chave na Shopee inteira.
- **Scheduler**: Microserviço Go (gRPC :50054) que gerencia jobs periódicos via robfig/cron e executa coletas no Collector.
- **SchedulerJobs**: Helper C# compartilhado que monta e registra/pausa jobs no Scheduler a partir de uma `Busca`.
- **Collector**: Microserviço Go (gRPC :50051) que busca produtos na API de afiliados da Shopee.
- **GerenciarBuscas**: Componente Svelte (`GerenciarBuscas.svelte`) que exibe o formulário de criação e a lista de buscas por palavra-chave.
- **AgendadorBusca**: Componente Svelte que permite ao usuário selecionar uma expressão cron (atalhos ou avançado).
- **Analyzer**: Microserviço Python (REST :8060) que consulta BigQuery para retornar novidades e variações de preço.
- **Snapshot**: Registro no BigQuery com produtos coletados em uma execução do Scheduler.
- **API_Principal**: API C# ASP.NET Core (POST /api/buscas, GET /api/buscas, GET /api/lojas/novidades).

## Requirements

### Requirement 1: Criação de busca agendada por keywords via API

**User Story:** As a usuário, I want to create a scheduled keyword search via the API so that the Scheduler periodically collects products matching my keywords from the Shopee API.

#### Acceptance Criteria

1. WHEN a POST /api/buscas request is received with keywords and a cron expression but without shop_ids, THE API_Principal SHALL persist the Busca_Agendada in PostgreSQL with the provided keywords and CronExpression.
2. WHEN a Busca_Agendada is persisted with a non-null CronExpression, THE API_Principal SHALL call SchedulerJobs.RegisterAsync to register a job of type "keyword_search" in the Scheduler.
3. WHEN a POST /api/buscas request is received with keywords but without a cron expression, THE API_Principal SHALL persist the Busca_Agendada without registering a job in the Scheduler.
4. WHEN a POST /api/buscas request is received without keywords and without shop_ids, THE API_Principal SHALL return HTTP 400 with an error message indicating that keywords are required.
5. WHEN a Busca_Agendada already exists for the same keywords and a POST is received with an updated cron, THE API_Principal SHALL update the CronExpression and re-register the job in the Scheduler.
6. IF the Scheduler is unavailable during registration, THEN THE API_Principal SHALL persist the Busca_Agendada in PostgreSQL and log a warning without returning an error to the client.

### Requirement 2: Envio de cron do frontend ao backend

**User Story:** As a usuário, I want the cron expression I select in the scheduling UI to be sent to the backend so that my keyword search runs on the schedule I chose.

#### Acceptance Criteria

1. WHEN the user submits the "nova busca" form in GerenciarBuscas with keywords and a cron selection, THE GerenciarBuscas SHALL include the cron field in the payload sent to POST /api/buscas.
2. WHEN the user selects "Nunca" (empty cron) in AgendadorBusca, THE GerenciarBuscas SHALL send the request without a cron field, resulting in a manual-only search with no scheduled job.
3. THE GerenciarBuscas SHALL send the keywords as an array field in the POST /api/buscas payload.
4. WHEN the backend responds with success, THE GerenciarBuscas SHALL update the local store and display the saved search with its active cron schedule.

### Requirement 3: Listagem de buscas agendadas por keyword

**User Story:** As a usuário, I want to see my keyword-only scheduled searches with their schedule status so that I can manage them.

#### Acceptance Criteria

1. THE API_Principal SHALL return all active keyword-only Busca_Agendada records (where ShopIds is null or empty) via GET /api/buscas, including the cron expression and keywords.
2. WHEN the user opens the /lojas page, THE GerenciarBuscas SHALL fetch and display keyword-only searches from the server, showing keywords, cron schedule, and active status.
3. WHEN a Busca_Agendada has a non-null CronExpression, THE GerenciarBuscas SHALL display a visual indicator showing the search is scheduled (e.g., frequency label).

### Requirement 4: Desativação de busca agendada por keyword

**User Story:** As a usuário, I want to deactivate a scheduled keyword search so that the Scheduler stops collecting products for it.

#### Acceptance Criteria

1. WHEN the user clicks "remover" on a keyword search in GerenciarBuscas, THE GerenciarBuscas SHALL send a POST /api/buscas?remover request with the keyword identifier.
2. WHEN a POST /api/buscas?remover request is received, THE API_Principal SHALL set the Busca_Agendada as inactive (Active=false) and call SchedulerJobs.PauseAsync to stop the Scheduler job.
3. WHEN the Busca_Agendada is deactivated, THE GerenciarBuscas SHALL remove the search from the displayed list.

### Requirement 5: Visualização de resultados (novidades) de buscas por keyword

**User Story:** As a usuário, I want to view the collected results (new products and price changes) from my keyword searches so that I can find relevant products to publish.

#### Acceptance Criteria

1. WHEN the user selects a keyword search from GerenciarBuscas, THE GerenciarBuscas SHALL fetch results from GET /api/lojas/novidades?busca_id={id} passing the Busca_Agendada ID.
2. THE API_Principal SHALL proxy the novidades request to the Analyzer regardless of whether the Busca_Agendada has ShopIds or not.
3. WHEN the Analyzer returns products, THE GerenciarBuscas SHALL display new products and price variations grouped by collection date.
4. IF the Analyzer is unavailable, THEN THE API_Principal SHALL return an empty result set with zero totals instead of an error.

### Requirement 6: Execução do job keyword_search pelo Scheduler

**User Story:** As a system operator, I want the Scheduler to correctly execute keyword_search jobs so that products are collected from Shopee on schedule.

#### Acceptance Criteria

1. WHEN a registered keyword_search job fires on its cron schedule, THE Scheduler SHALL call Collector.Fetch with the keyword parameter for each keyword in the job params.
2. WHEN the Collector returns products, THE Scheduler SHALL log the total found and enqueue a price alert via Cloud Tasks for the keyword.
3. IF the Collector returns an error, THEN THE Scheduler SHALL log the error and continue to the next keyword without crashing the job.
4. WHEN a keyword_search job has multiple keywords (comma-separated), THE Scheduler SHALL iterate over each keyword and call Collector.Fetch for each one individually.
