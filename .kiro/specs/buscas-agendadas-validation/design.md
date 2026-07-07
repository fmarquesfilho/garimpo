# Buscas Agendadas Validation — Bugfix Design

## Overview

Após a migração do monolito Go para a arquitetura poliglota, o endpoint `POST /api/lojas` persiste a Busca no PostgreSQL e resolve o `shop_id` via Collector gRPC, mas **não registra o agendamento no Scheduler**. Da mesma forma, `DELETE /api/lojas` faz soft-delete mas não remove/pausa o job correspondente. O resultado é que buscas criadas pós-migração nunca são coletadas periodicamente — o pipeline `Scheduler → Collector → BigQuery` nunca dispara.

A correção integra o Scheduler gRPC client (`SetSchedule`) nos endpoints de criação e remoção de lojas, adiciona suporte a `Keywords` na entidade `Busca` para buscas filtradas, e valida o fluxo end-to-end com testes E2E que exercitam a cadeia completa.

## Glossary

- **Bug_Condition (C)**: Busca criada/removida via `/api/lojas` sem chamada correspondente ao Scheduler gRPC `SetSchedule`
- **Property (P)**: Toda Busca persistida com `ShopIds` não-nulo DEVE ter um job registrado no Scheduler; toda Busca soft-deleted DEVE ter o job pausado
- **Preservation**: Resolução de shop_id via Collector, persistência no PostgreSQL, tratamento de erros existente — tudo deve permanecer inalterado
- **SetSchedule**: RPC `scheduler.v1.SchedulerService.SetSchedule(job_id, cron_expression, enabled, params)` — cria ou atualiza um job no Scheduler Go
- **Keywords**: Campo `string[]?` na entidade `Busca` — quando preenchido, o Scheduler passa cada keyword como parâmetro para coletas filtradas (em vez de coletar todos os produtos da loja)
- **Job ID Convention**: `busca-{Busca.Id}` — identificador único e determinístico que vincula uma Busca PG a um job no Scheduler

## Bug Details

### Bug Condition

O bug manifesta quando uma Busca é criada ou removida via `POST /api/lojas` ou `DELETE /api/lojas`. O C# API persiste a mudança no PostgreSQL mas não comunica ao Scheduler, resultando em buscas que nunca são coletadas (criação) ou jobs órfãos (remoção).

**Formal Specification:**
```
FUNCTION isBugCondition(input)
  INPUT: input of type HttpRequest (POST or DELETE /api/lojas)
  OUTPUT: boolean

  IF input.method == POST AND input resolves shop_id successfully THEN
    RETURN schedulerSetScheduleCalled == false
           AND busca.ShopIds IS NOT NULL
           AND busca persisted in PostgreSQL
  END IF

  IF input.method == DELETE AND busca exists with Active=true THEN
    RETURN schedulerSetScheduleCalled == false
           AND busca.Active set to false in PostgreSQL
  END IF

  RETURN false
END FUNCTION
```

### Examples

- **POST loja sem keywords**: Usuário adiciona `shopee.com.br/belezanaweb_oficial` → Busca persistida com `ShopIds=[1674883556]`, mas Scheduler não tem job `busca-{id}` → coleta nunca ocorre → `/api/lojas/novidades` retorna vazio
- **POST loja com keywords**: Usuário cria busca com `shop_ids=[920292999]` e `keywords=["serum", "protetor solar"]` → Busca persistida, mas Scheduler não tem job filtrado → snapshots filtrados nunca são gerados
- **DELETE loja**: Usuário remove loja → `Active=false` no PG, mas se job existisse no Scheduler ele continuaria rodando como órfão
- **Scheduler indisponível no POST**: Busca criada deve persistir (eventual consistency) mas warning logado — job pode ser reconciliado depois

## Expected Behavior

### Preservation Requirements

**Unchanged Behaviors:**
- Resolução de `shop_id` via `Collector.ResolveShop` gRPC deve continuar funcionando exatamente como antes
- Persistência da `Busca` no PostgreSQL via EF Core (multi-tenant, `OwnerUid` filter) inalterada
- Resposta HTTP do `POST /api/lojas` (200 com `{id, keyword, shop_ids, source_url, status}`) inalterada
- Tratamento de erro quando Collector retorna `NotFound`/`InvalidArgument` → HTTP 400
- Tratamento genérico de falha do Collector → HTTP 400 "Falha ao resolver o ID da loja via Collector"
- `GET /api/lojas` retorna todas as Buscas ativas com campos existentes
- `GET /api/lojas/novidades` continua proxying para o Analyzer com fallback graceful
- `DELETE /api/lojas` continua fazendo soft-delete com `Active=false`

**Scope:**
Inputs que NÃO envolvem criação/remoção de lojas devem ser completamente inalterados:
- Mouse clicks e navegação no frontend
- `GET /api/lojas`, `GET /api/lojas/novidades`, `GET /api/lojas/evolucao`
- Outros endpoints (`/api/candidatos`, `/api/buscas`, `/api/publicar`, etc.)
- Coletas já agendadas manualmente no Scheduler (jobs pré-existentes)

## Hypothesized Root Cause

O endpoint `POST /api/lojas` em `LojasCompatEndpoints.cs` foi portado do monolito Go sem incluir a integração com o Scheduler. No monolito, o scheduling era interno ao mesmo processo. Na nova arquitetura, o Scheduler é um sidecar separado acessível via gRPC na porta 50054 — a chamada `SetSchedule` simplesmente nunca foi adicionada ao endpoint C#.

Causas específicas:

1. **Falta de registro do SchedulerServiceClient no DI**: `DependencyInjection.cs` registra `CollectorServiceClient` e `PublisherServiceClient`, mas não `SchedulerServiceClient` — o endpoint não tem como injetar o client
2. **Nenhuma chamada SetSchedule no fluxo de criação**: O handler do `POST /api/lojas` termina após `db.SaveChangesAsync()` sem interagir com o Scheduler
3. **Nenhuma chamada SetSchedule(enabled=false) no fluxo de remoção**: O handler do `DELETE /api/lojas` faz `Active=false` mas não notifica o Scheduler
4. **Entidade Busca sem campo Keywords**: A entity `Busca.cs` não tem `Keywords` — buscas filtradas não podem ser configuradas

## Correctness Properties

Property 1: Bug Condition - Scheduler Integration on Busca Creation

_For any_ POST `/api/lojas` request that successfully resolves a `shop_id` and persists a Busca with non-null `ShopIds`, the fixed endpoint SHALL call `Scheduler.SetSchedule` with `job_id=busca-{Busca.Id}`, `cron_expression` (default `0 */8 * * *`), `enabled=true`, and params containing `shop_id`, `owner_uid`, and optionally `keywords`. If the Scheduler is unavailable, the Busca SHALL still be persisted (eventual consistency) and a warning SHALL be logged.

**Validates: Requirements 2.1, 2.2, 2.3, 3.6**

Property 2: Bug Condition - Scheduler Integration on Busca Removal

_For any_ DELETE `/api/lojas` request that soft-deletes a Busca, the fixed endpoint SHALL call `Scheduler.SetSchedule` with `job_id=busca-{Busca.Id}` and `enabled=false` to pause/remove the corresponding collection job.

**Validates: Requirements 2.5**

Property 3: Preservation - Existing Behavior Unchanged

_For any_ request to `POST /api/lojas` or `DELETE /api/lojas`, the fixed code SHALL produce the same HTTP response (status code, body shape, error messages) as the original code. The Collector resolution, PostgreSQL persistence, error handling, and response format SHALL remain identical.

**Validates: Requirements 3.1, 3.2, 3.3, 3.4, 3.5**

Property 4: Bug Condition - Keywords Support for Filtered Searches

_For any_ Busca created with both `ShopIds` and `Keywords` populated, the Scheduler job params SHALL include the keywords array, enabling the Collector to perform filtered fetches (by keyword) instead of fetching all products from the shop.

**Validates: Requirements 2.3**

## Fix Implementation

### Changes Required

Assuming our root cause analysis is correct:

**File**: `src/Garimpei.Infrastructure/DependencyInjection.cs`

**Change**: Register `SchedulerServiceClient` gRPC client

**Specific Changes**:
1. **Add Scheduler gRPC client registration**:
   ```csharp
   var schedulerAddr = configuration["Grpc:SchedulerAddress"] ?? "http://localhost:50054";
   services.AddGrpcClient<Scheduler.V1.SchedulerService.SchedulerServiceClient>(o =>
   {
       o.Address = new Uri(schedulerAddr);
   });
   ```

---

**File**: `src/Garimpei.Domain/Entities/Busca.cs`

**Change**: Add `Keywords` property

**Specific Changes**:
2. **Add Keywords field to Busca entity**:
   ```csharp
   /// <summary>
   /// Keywords de filtragem para coletas agendadas. Se nulo/vazio, coleta todos os produtos da loja.
   /// </summary>
   public string[]? Keywords { get; set; }
   ```

3. **Add Cron field to Busca entity** (para permitir cron customizado):
   ```csharp
   /// <summary>
   /// Expressão cron para agendamento. Default: "0 */8 * * *" (a cada 8h).
   /// Se null, usa o default.
   /// </summary>
   public string? CronExpression { get; set; }
   ```

---

**File**: `src/Garimpei.Api/Endpoints/LojasCompatEndpoints.cs`

**Function**: `MapPost("/api/lojas", ...)`

**Specific Changes**:
4. **Inject SchedulerServiceClient** no handler do POST:
   ```csharp
   Scheduler.V1.SchedulerService.SchedulerServiceClient schedulerClient
   ```

5. **Call SetSchedule after successful persist** (fire-and-forget com logging de warning):
   ```csharp
   // Após SaveChangesAsync — registra job no Scheduler
   try
   {
       var jobId = $"busca-{busca.Id}";
       var cronExpr = req.Cron ?? "0 */8 * * *";
       var jobParams = new Dictionary<string, string>
       {
           ["shop_id"] = busca.ShopIds![0].ToString(),
           ["owner_uid"] = busca.OwnerUid,
           ["type"] = "shop_collection"
       };
       if (busca.Keywords is { Length: > 0 })
           jobParams["keywords"] = string.Join(",", busca.Keywords);

       await schedulerClient.SetScheduleAsync(new Scheduler.V1.SetScheduleRequest
       {
           JobId = jobId,
           CronExpression = cronExpr,
           Enabled = true,
           Params = { jobParams }
       }, cancellationToken: ct);
   }
   catch (Exception ex)
   {
       // Eventual consistency: Busca persiste, job pode ser reconciliado depois
       logger.LogWarning(ex, "Falha ao registrar job no Scheduler para busca {BuscaId}", busca.Id);
   }
   ```

6. **Populate Keywords from request** (se presente no body):
   - Aceitar campo `keywords` no `AdicionarLojaRequest`
   - Mapear para `busca.Keywords`

---

**Function**: `MapDelete("/api/lojas", ...)`

**Specific Changes**:
7. **Inject SchedulerServiceClient** no handler do DELETE

8. **Call SetSchedule(enabled=false) após soft-delete**:
   ```csharp
   try
   {
       await schedulerClient.SetScheduleAsync(new Scheduler.V1.SetScheduleRequest
       {
           JobId = $"busca-{guid}",
           CronExpression = "0 */8 * * *", // required pelo proto
           Enabled = false
       }, cancellationToken: ct);
   }
   catch (Exception ex)
   {
       logger.LogWarning(ex, "Falha ao pausar job no Scheduler para busca {BuscaId}", guid);
   }
   ```

---

**File**: `src/Garimpei.Api/Endpoints/LojasCompatEndpoints.cs`

**Record**: `AdicionarLojaRequest`

**Specific Changes**:
9. **Add Keywords to request record**:
   ```csharp
   public sealed record AdicionarLojaRequest
   {
       public string? Input { get; init; }
       public string? Cron { get; init; }
       public string? OrigemPadrao { get; init; }
       public string[]? Keywords { get; init; }
   }
   ```

---

**File**: EF Core migration

10. **Add migration for Keywords + CronExpression columns** na tabela `Buscas`

## Testing Strategy

### Validation Approach

The testing strategy follows a two-phase approach: first, surface counterexamples that demonstrate the bug on unfixed code, then verify the fix works correctly and preserves existing behavior. Os testes E2E exercitam a cadeia completa: Frontend → C# API → Scheduler gRPC → (Collector gRPC no job dispatch).

### Exploratory Bug Condition Checking

**Goal**: Confirmar que o código atual NÃO chama `SetSchedule` ao criar/remover lojas. Rodar contra o código sem fix para observar a falha.

**Test Plan**: Escrever testes de integração C# que adicionam uma loja via `POST /api/lojas` e verificam se o Scheduler recebeu uma chamada `SetSchedule`. No código unfixed, o teste DEVE falhar (nenhuma chamada ao Scheduler).

**Test Cases**:
1. **POST loja → Scheduler não chamado**: Adicionar loja via POST, verificar que `ListJobs` retorna 0 jobs para `busca-{id}` (falha no código unfixed)
2. **DELETE loja → Scheduler não chamado**: Remover loja via DELETE, verificar que nenhum job foi pausado (falha no código unfixed)
3. **POST loja com keywords → Keywords não persistidas**: Enviar `keywords: ["serum"]`, verificar que o campo não é salvo (falha no código unfixed — campo não existe)
4. **Novidades vazias**: Após POST sem fix, `GET /api/lojas/novidades` retorna `produtos_novos=[]` porque nenhum snapshot foi coletado

**Expected Counterexamples**:
- `Scheduler.ListJobs()` retorna lista vazia após `POST /api/lojas`
- Campo `Keywords` não existe na entidade `Busca` (compilation error)
- Possible causes: falta de registro do SchedulerClient no DI, falta de chamada no handler

### Fix Checking

**Goal**: Verificar que para toda Busca criada com `ShopIds` preenchido, o Scheduler tem um job ativo correspondente.

**Pseudocode:**
```
FOR ALL input WHERE isBugCondition(input) DO
  result := POST /api/lojas (input)
  ASSERT result.status == 200
  ASSERT schedulerJobs CONTAINS job WITH id == "busca-{result.id}"
  ASSERT job.status == "active"
  ASSERT job.params["shop_id"] == result.shop_ids[0]
  IF input.keywords IS NOT NULL THEN
    ASSERT job.params["keywords"] == join(input.keywords, ",")
  END IF
END FOR
```

### Preservation Checking

**Goal**: Verificar que para todos os inputs que NÃO envolvem criação/remoção de lojas, o comportamento é idêntico ao original.

**Pseudocode:**
```
FOR ALL input WHERE NOT isBugCondition(input) DO
  ASSERT originalEndpoint(input) = fixedEndpoint(input)
END FOR
```

**Testing Approach**: Testes de integração C# (xUnit + WebApplicationFactory) que verificam:
- HTTP response codes inalterados
- JSON response shapes inalteradas
- Erros do Collector continuam mapeados para HTTP 400
- `GET /api/lojas` continua retornando dados do PG

**Test Plan**: Observar comportamento no código atual para GET, error cases, e mouse-click flows. Escrever testes que capturam esse comportamento e rodar após o fix.

**Test Cases**:
1. **GET /api/lojas preservation**: Verificar que a listagem continua retornando os mesmos campos (`id`, `keyword`, `shop_ids`, `source_url`, `ativo`, `criado_em`)
2. **POST /api/lojas error handling**: Verificar que Collector NotFound → 400, Collector Unavailable → 400 com mensagem correta
3. **Scheduler unavailable on POST**: Verificar que a Busca é persistida mesmo quando Scheduler falha (eventual consistency), response continua 200
4. **DELETE /api/lojas soft-delete**: Verificar que `Active=false` e `UpdatedAt` são setados corretamente

### Unit Tests

- Teste unitário do mapeamento de `Busca` → `SetScheduleRequest` (job_id convention, params)
- Teste unitário do campo `Keywords` na entidade (serialização/deserialização do array)
- Teste unitário do fallback quando Scheduler está indisponível (Busca persiste, warning logado)
- Teste unitário do `AdicionarLojaRequest` com campos opcionais (`Keywords`, `Cron`)

### Property-Based Tests

- Gerar `shop_ids` aleatórios e verificar que o `job_id` gerado segue a convenção `busca-{Guid}`
- Gerar combinações aleatórias de `keywords` (0..N strings) e verificar que `params["keywords"]` é serializado corretamente (join por vírgula)
- Gerar requests com/sem Scheduler disponível e verificar que a Busca SEMPRE persiste no PG (independente do Scheduler)

### Integration Tests

- **E2E local** (`mise run test:e2e:buscas-agendadas`): Frontend → C# API → Collector gRPC (ResolveShop real) → Scheduler gRPC (SetSchedule real) → verificar job criado via `ListJobs`
- **E2E com keywords**: POST com `keywords=["serum"]` → verificar que job params incluem keywords
- **E2E delete**: POST + DELETE → verificar que job status é `paused` via `ListJobs`
- **E2E Scheduler offline**: Parar Scheduler → POST loja → verificar 200 (Busca persistida) + warning no log

### Mise Task para E2E

Nova task `test:e2e:buscas-agendadas` que:
1. Sobe PostgreSQL (docker compose)
2. Sobe API C# (docker compose com override gRPC)
3. Sobe Collector Go (real, para ResolveShop)
4. Sobe Scheduler Go (real, para SetSchedule/ListJobs)
5. Inicia Firebase Auth Emulator
6. Roda Playwright tests com tag `buscas-agendadas`
7. Cleanup: para processos iniciados

Padrão similar ao `test:e2e:lojas` existente, mas adicionando o Scheduler Go como dependência.
