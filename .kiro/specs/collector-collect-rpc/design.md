# Technical Design — Collect RPC

## Overview

Adiciona um único `Collect` RPC ao CollectorService que combina busca de produtos (keyword ou shop) com persistência assíncrona de snapshots no BigQuery. Reutiliza `store.SnapshotRepo` existente como interface do exporter. O Scheduler migra de `Fetch` + persistência local para `Collect` e perde a dependência de `internal/store`.

## Architecture

```
┌─────────────┐                          ┌──────────────────────────────────────────────┐
│  Scheduler  │  Collect(keyword="serum") │               Collector                      │
│  (cron)     │─────────────────────────→ │                                              │
└─────────────┘  Collect(shop_id=920...)  │  ┌──────────┐         ┌──────────────────┐  │
                                          │  │ Pipeline  │         │  export goroutine │  │
                                          │  │ (sources) │         │  (channel-based)  │  │
                                          │  └─────┬────┘         └────────┬─────────┘  │
                                          │        │ Search()              │             │
                                          │        ▼                       ▼             │
                                          │  ┌──────────┐         ┌──────────────────┐  │
                                          │  │  Shopee   │         │store.SnapshotRepo│  │
                                          │  │  Amazon   │         │  (BigQuery)      │  │
                                          │  │  ML       │         └──────────────────┘  │
                                          │  └──────────┘                                │
                                          └──────────────────────────────────────────────┘
```

Fluxo:
1. `Collect` recebe request (keyword ou shop_id)
2. Busca produtos no marketplace (síncrono, ~500ms)
3. Enfileira snapshot no export channel (non-blocking, ~0ms)
4. Retorna response imediatamente com `persisted=true` (aceito para export)
5. Background goroutine drena o channel e faz streaming insert no BigQuery (~100ms)

## Components and Interfaces

### 1. Proto — Um único RPC com `oneof target`

```protobuf
service CollectorService {
  // Existing (unchanged)
  rpc ResolveShop(ResolveShopRequest) returns (ResolveShopResponse);
  rpc GenerateAffiliateLink(GenerateAffiliateLinkRequest) returns (GenerateAffiliateLinkResponse);
  rpc Fetch(FetchRequest) returns (FetchResponse);
  rpc FetchShop(FetchShopRequest) returns (FetchShopResponse);

  // New: search + persist (single RPC for both keyword and shop)
  rpc Collect(CollectRequest) returns (CollectResponse);
}

message CollectRequest {
  oneof target {
    string keyword = 1;   // Busca por keyword → persiste com keyword literal
    int64 shop_id = 6;    // Busca por loja → persiste com shop_id como string
  }
  int32 limit = 2;
  string sort_by = 3;
  string owner_uid = 4;
  Marketplace marketplace = 5;
}

message CollectResponse {
  repeated Product products = 1;
  int32 total_found = 2;
  string fetched_at = 3;   // RFC3339
  bool persisted = 4;      // true = aceito para export; false = buffer cheio ou exporter desabilitado
}
```

**Decisão**: `oneof target` em vez de 2 RPCs separados. Reduz surface area (1 handler, 1 call site no Scheduler) sem perder expressividade. O handler faz switch no `oneof` para decidir entre `Search()` e `FetchShop()`.

### 2. Reusar `store.SnapshotRepo` — sem tipos novos

O Collector importa `internal/store` e usa a interface existente:

```go
type SnapshotRepo interface {
    RegistrarSnapshot(ctx context.Context, s Snapshot) error
    // ... outros métodos (não usados pelo Collector)
}
```

E os tipos existentes:

```go
type Snapshot struct {
    Keyword    string
    Estrategia string
    Em         time.Time
    Itens      []ItemSnapshot
}
```

**Decisão**: Reusar em vez de criar tipos paralelos (`ExportSnapshot`, `ExportItem`). Evita conversão intermediária, drift de campos, e código morto. Se amanhã o exporter mudar (Kafka, S3), a interface `SnapshotRepo` evolui — não precisa de indireção antecipada.

O `store.NopSnapshots()` já existe para o caso de BigQuery não estar configurado.

### 3. Export assíncrono via channel

```go
type UnifiedCollectorServer struct {
    collectorpb.UnimplementedCollectorServiceServer
    pipeline  *Pipeline
    snapshots store.SnapshotRepo
    exportCh  chan store.Snapshot  // buffer: 64 snapshots
    logger    *slog.Logger
}
```

O `Collect` handler enfileira e retorna imediatamente:

```go
func (s *UnifiedCollectorServer) enqueueExport(snap store.Snapshot) bool {
    select {
    case s.exportCh <- snap:
        return true
    default:
        s.logger.Warn("export buffer full, snapshot dropped",
            slog.String("keyword", snap.Keyword))
        return false
    }
}
```

Background goroutine (iniciada no startup, para no shutdown):

```go
func (s *UnifiedCollectorServer) runExporter(ctx context.Context) {
    for {
        select {
        case snap := <-s.exportCh:
            if err := s.snapshots.RegistrarSnapshot(ctx, snap); err != nil {
                s.logger.Error("snapshot export failed",
                    slog.String("keyword", snap.Keyword),
                    slog.Int("items", len(snap.Itens)),
                    slog.String("error", err.Error()))
            }
        case <-ctx.Done():
            // Drain remaining on shutdown
            for len(s.exportCh) > 0 {
                snap := <-s.exportCh
                _ = s.snapshots.RegistrarSnapshot(context.Background(), snap)
            }
            return
        }
    }
}
```

**Buffer size 64**: Em produção, o Scheduler roda ~10-20 jobs por ciclo (8h). Mesmo com bursts, 64 é mais que suficiente. Se lotar, o snapshot é descartado (próximo ciclo recria).

### 4. Inicialização do store — build tags

`services/collector/store_gcp.go` (build tag `gcp`):

```go
func initSnapshots(ctx context.Context, cfg BigQueryExporterConfig, logger *slog.Logger) store.SnapshotRepo {
    project := ResolveCredentialEnv(cfg.ProjectEnv)
    if project == "" || cfg.Dataset == "" || cfg.ProductsTable == "" {
        logger.Warn("BigQuery export disabled (missing config)")
        return store.NopSnapshots()
    }
    bq, err := store.NovoBigQueryStore(ctx, project, cfg.Dataset, "eventos", cfg.ProductsTable)
    if err != nil {
        logger.Warn("BigQuery init failed, export disabled", slog.String("error", err.Error()))
        return store.NopSnapshots()
    }
    logger.Info("BigQuery snapshot export enabled",
        slog.String("dataset", cfg.Dataset), slog.String("table", cfg.ProductsTable))
    return bq
}
```

`services/collector/store_nop.go` (build tag `!gcp`):

```go
func initSnapshots(_ context.Context, _ BigQueryExporterConfig, logger *slog.Logger) store.SnapshotRepo {
    logger.Info("snapshot export disabled (build without -tags gcp)")
    return store.NopSnapshots()
}
```

### 5. Scheduler simplificado

Remove:
- Campo `snapshots store.SnapshotRepo`
- Import de `internal/store`
- Função `persistSnapshot()`
- Arquivos `store_gcp.go`, `store_nogcp.go`
- Flag `-tags gcp` do Dockerfile do scheduler

Muda:
- `s.collector.Fetch()` → `s.collector.Collect()` nos jobs agendados
- Monta `CollectRequest` com `oneof target` (keyword ou shop_id)

## Data Models

### Proto Messages (novos)

```protobuf
message CollectRequest {
  oneof target {
    string keyword = 1;    // Keyword literal para busca e persistência
    int64 shop_id = 6;     // Shop ID (persiste como string no BQ)
  }
  int32 limit = 2;         // 1-500, default 50
  string sort_by = 3;      // relevance | sales | price
  string owner_uid = 4;    // UID do tenant (scoping)
  Marketplace marketplace = 5;  // Default: SHOPEE
}

message CollectResponse {
  repeated Product products = 1;
  int32 total_found = 2;
  string fetched_at = 3;    // RFC3339 timestamp
  bool persisted = 4;       // true = enqueued for export
}
```

### BigQuery Row (tabela `snapshots` — sem mudanças no schema)

| Campo | Tipo | Origem |
|-------|------|--------|
| coletado_em | TIMESTAMP | `time.Now().UTC()` |
| keyword | STRING | `CollectRequest.keyword` ou `strconv.FormatInt(shop_id, 10)` |
| estrategia | STRING | `"coleta-agendada"` |
| posicao | INTEGER | index + 1 |
| produto_id | STRING | `fmt.Sprintf("shopee-%d-%d", p.ShopId, p.ItemId)` |
| nome | STRING | `p.Name` |
| preco | FLOAT | `p.Price` |
| comissao | FLOAT | `p.Commission` |
| vendas | INTEGER | `p.Sold` |
| nota | FLOAT | `p.Rating` |
| imagem | STRING | `p.ImageUrl` |
| link | STRING | `p.ProductUrl` |
| loja | STRING | `p.ShopName` |

## Error Handling

### Busca falha (marketplace API error)

- Retorna gRPC error ao caller (`codes.Internal`)
- Nada é enfileirado para export
- Scheduler loga e continua para o próximo keyword (multi-keyword job)

### Export falha (BigQuery insert error)

- Background goroutine loga erro com keyword, item count, error message
- Snapshot é descartado (sem retry — próximo ciclo do Scheduler resolve)
- O response ao Scheduler já retornou com `persisted=true` (foi enfileirado com sucesso)
- Se o erro é persistente (permissão, quota), todos os exports falham e os logs acumulam — operador investiga

### Buffer cheio (export channel full)

- `enqueueExport` retorna `false` via `select default`
- Response retorna `persisted=false`
- Log warning com keyword
- Causa provável: BigQuery insert muito lento ou goroutine travada

### Exporter não configurado (NopSnapshots)

- `NopSnapshots().RegistrarSnapshot()` retorna nil
- Snapshot é consumido do channel e descartado silenciosamente
- `persisted` no response será `true` (aceito para export) mas nada persiste de fato
- Log warning no startup: "BigQuery export disabled"

## Correctness Properties

### Property 1: Fetch permanece puro

Nenhum call path de `Fetch`/`FetchShop` toca o channel ou o store. O campo `exportCh` existe apenas no server, e só `Collect` chama `enqueueExport`. Garantido pela ausência de referência nos métodos existentes.

**Validates: Requirements 8.1, 8.2, 8.3**

### Property 2: Keyword literal no snapshot

O campo `keyword` em BigQuery sempre contém o valor direto do request: `CollectRequest.keyword` (string literal) ou `strconv.FormatInt(CollectRequest.shop_id, 10)`. Nunca um UUID, busca ID, ou valor derivado.

**Validates: Requirements 3.3, 4.2**

### Property 3: Non-blocking export

O `Collect` handler nunca bloqueia esperando o BigQuery. O `select` com `default` garante retorno imediato mesmo com buffer cheio. A latência do `Collect` = latência do marketplace API (Shopee ~500ms), sem adição do BigQuery (~100ms).

**Validates: Requirements 3.8, 4.6, 6.2**

### Property 4: Graceful shutdown

No `Stop()`, o context é cancelado. A goroutine drena os snapshots restantes no channel antes de retornar. Nenhum snapshot em voo é perdido em shutdown limpo (SIGTERM com grace period).

**Validates: Requirements 6.3**

### Property 5: arch-go compliance

Scheduler não importa `internal/store`. Collector pode importar — não existe regra proibindo, e é arquiteturalmente coerente (Collector é produtor de dados, ADR-0018 prevê exporters BigQuery).

**Validates: Requirements 7.3, 7.6**

## Data Flow

```
1. Scheduler cron fires
2. Scheduler calls Collector.Collect(keyword="serum", owner_uid="abc", marketplace=SHOPEE)
3. Collector.Collect:
   a. Resolves marketplace → gets ProductSource
   b. Calls src.Search("serum", limit=50) → gets []Product (~500ms)
   c. Builds store.Snapshot{Keyword: "serum", Itens: [...]}
   d. Enqueues on exportCh (non-blocking, ~0ms)
   e. Returns CollectResponse{products=[50], persisted=true}
4. Scheduler receives response → enqueues price alert via Cloud Tasks
5. Background goroutine:
   a. Reads snapshot from exportCh
   b. Calls snapshots.RegistrarSnapshot() → BigQuery streaming insert (~100ms)
6. Later: Analyzer queries WHERE keyword LIKE '%serum%' → finds data → UI shows novidades/quedas
```

## File Changes Summary

| File | Action | Description |
|------|--------|-------------|
| `protos/collector/v1/collector.proto` | Modify | Add `Collect` RPC + `CollectRequest`/`CollectResponse` with oneof |
| `services/collector/server.go` | Modify | Add `snapshots` + `exportCh` fields, `Collect` method, `enqueueExport`, `runExporter` |
| `services/collector/store_gcp.go` | Create | `initSnapshots` with BigQuery (build tag gcp) |
| `services/collector/store_nop.go` | Create | `initSnapshots` nop (build tag !gcp) |
| `services/collector/main.go` | Modify | Wire snapshots + start exporter goroutine |
| `services/collector/Dockerfile` | Modify | Add `-tags gcp` |
| `services/scheduler/server.go` | Modify | Remove `snapshots` field, remove `store` import |
| `services/scheduler/jobs.go` | Modify | `Collect` instead of `Fetch`, remove `persistSnapshot` |
| `services/scheduler/store_gcp.go` | Delete | No longer needed |
| `services/scheduler/store_nogcp.go` | Delete | No longer needed |
| `services/scheduler/Dockerfile` | Modify | Remove `-tags gcp` |
| `arch-go.yml` | Restore | Scheduler rule back to original (shouldNotDependsOn store) |
| `gen/go/collector/v1/` | Regenerate | `buf generate` |
| `src/Garimpei.Protos/` | Regenerate | `buf generate` (C# stubs) |

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Channel buffer full = lost snapshots | Buffer 64 handles burst of 64 concurrent jobs. In practice Scheduler runs 10-20. If chronically full → monitor via log warn, increase buffer. |
| `persisted=true` but BQ insert fails later | Acceptable. Scheduler doesn't use `persisted` for control flow — only for observability. Next cycle produces fresh snapshots. |
| Collector imports `internal/store` | No arch-go rule against it. Aligned with ADR-0018 (exporter is part of Collector). |
| Proto oneof complexity | Single switch statement in handler. Simpler than 2 separate RPCs with duplicated code. |
| Shutdown loses in-flight snapshots | Drain loop in `runExporter` processes remaining channel items before exit. Cloud Run gives 10s grace. |

## Testing Strategy

- **Unit (Go)**: Inject `store.NopSnapshots()` → test Collect returns products correctly
- **Unit (Go)**: Inject mock SnapshotRepo → verify `RegistrarSnapshot` called with correct keyword
- **Unit (Go)**: Test oneof routing (keyword → Search, shop_id → FetchShop)
- **Unit (Go)**: Test buffer full behavior (persisted=false)
- **Unit (Go)**: Test validation (empty keyword + zero shop_id → InvalidArgument)
- **Build**: `go build -tags gcp ./services/collector/` compiles
- **Build**: `go build ./services/collector/` compiles (nop path)
- **Build**: `go build ./services/scheduler/` compiles (no store import)
- **arch-go**: `arch-go` passes (scheduler clean of internal/store)
- **E2E**: `mise run test:e2e-coleta` validates end-to-end after deploy
