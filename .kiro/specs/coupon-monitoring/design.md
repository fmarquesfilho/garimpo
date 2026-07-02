# Technical Design: Coupon Monitoring

## Overview

Coupon Monitoring extends Garimpei's existing product pipeline with a parallel coupon-specific flow. It reuses the same architectural patterns already proven in the codebase: Strategy + Registry in Go for marketplace adapters, Keyed Services in C# for DI-driven polymorphism, the Scheduler for cron-based dispatching, BigQuery for append-only analytics storage, and the Alerter gRPC service for notification delivery.

The feature adds a new **coupon-collector** Go gRPC microservice (mirroring the product collector), a **coupon detection** step in the Python analyzer service, and **alert matching + deduplication** logic in the C# API (which already owns PostgreSQL access and tenant context).

## Architecture

### System Flow

```
┌───────────┐    gRPC dispatch     ┌──────────────────────┐
│ Scheduler │ ──────────────────► │  Coupon Collector    │
│  (Go)     │  per marketplace     │  (Go gRPC service)   │
└───────────┘                      └──────────┬───────────┘
                                              │ writes
                                              ▼
                                   ┌──────────────────────┐
                                   │  BigQuery            │
                                   │  coupon_snapshots    │
                                   └──────────┬───────────┘
                                              │ triggers
                                              ▼
                                   ┌──────────────────────┐
                                   │  Coupon Detector     │
                                   │  (Python analyzer)   │
                                   └──────────┬───────────┘
                                              │ detection events
                                              ▼
                                   ┌──────────────────────┐
                                   │  Alert Matcher       │
                                   │  (C# API)           │
                                   └──────────┬───────────┘
                                              │ gRPC
                                              ▼
                                   ┌──────────────────────┐
                                   │  Alerter             │
                                   │  (Go gRPC service)   │
                                   └──────────────────────┘
                                              │
                                    Telegram / WhatsApp
```

### Component Diagram

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           SCHEDULER (Go)                                │
│  job_type: "coupon_collection"                                          │
│  sequential dispatch: Shopee → Amazon → Mercado Livre                   │
└────────┬──────────────────────┬────────────────────────┬────────────────┘
         │                      │                        │
    gRPC │                 gRPC │                   gRPC │
         ▼                      ▼                        ▼
┌────────────────┐  ┌────────────────────┐  ┌─────────────────────┐
│ Shopee Coupon  │  │  Amazon Coupon     │  │  ML Coupon          │
│ Adapter (Go)   │  │  Adapter (Go)      │  │  Adapter (Go)       │
│ :50061         │  │  :50062            │  │  :50063             │
└───────┬────────┘  └────────┬───────────┘  └──────────┬──────────┘
        │                    │                         │
        └────────────────────┼─────────────────────────┘
                             │ INSERT (append-only)
                             ▼
              ┌──────────────────────────────┐
              │  BigQuery: coupon_snapshots   │
              │  partitioned by collected_at  │
              └──────────────┬───────────────┘
                             │ diff query
                             ▼
              ┌──────────────────────────────┐
              │  Python Analyzer             │
              │  POST /detect-coupons        │
              │  (BigQuery diff logic)       │
              └──────────────┬───────────────┘
                             │ HTTP callback
                             ▼
              ┌──────────────────────────────┐
              │  C# API (Alert Matcher)      │
              │  POST /internal/coupon-alerts │
              │  PostgreSQL: dedup + rules    │
              └──────────────┬───────────────┘
                             │ gRPC
                             ▼
              ┌──────────────────────────────┐
              │  Alerter (Go gRPC)           │
              │  Telegram / WhatsApp         │
              └──────────────────────────────┘
```

## Data Models

### Go Domain Model (Coupon)

File: `internal/domain/coupon.go`

```go
package domain

// Coupon represents a marketplace coupon/voucher collected from affiliate APIs.
type Coupon struct {
    ID                   string   // unique coupon identifier from marketplace
    Marketplace          string   // "shopee", "amazon", "mercadolivre"
    Code                 string   // voucher code or claiming URL
    DiscountType         string   // "percentage" or "fixed"
    DiscountValue        float64  // e.g. 20.0 for 20% or 15.00 for R$15
    MinSpend             float64  // minimum purchase amount (0 = no minimum)
    StartTime            int64    // Unix timestamp
    EndTime              int64    // Unix timestamp
    ApplicableCategories []string // category IDs/names this coupon applies to
    Status               string   // "active", "expired", "claimed"
    OwnerUID             string   // tenant that collected this coupon
    CollectedAt          int64    // Unix timestamp of collection
}

// DetectionStatus classifies a coupon after snapshot comparison.
type DetectionStatus string

const (
    DetectionNewlyDiscovered  DetectionStatus = "newly_discovered"
    DetectionModified         DetectionStatus = "modified"
    DetectionExpiredOrRemoved DetectionStatus = "expired_or_removed"
    DetectionUnchanged        DetectionStatus = "unchanged"
)

// CouponDetection is the result of comparing two snapshots.
type CouponDetection struct {
    CouponID    string
    Marketplace string
    OwnerUID    string
    Status      DetectionStatus
    DetectedAt  int64 // Unix timestamp
}
```

### Proto Definition (coupon_collector.proto)

File: `protos/coupon/v1/coupon.proto`

```protobuf
syntax = "proto3";

package coupon.v1;

option go_package = "github.com/fmarquesfilho/garimpo/gen/go/coupon/v1;couponv1";
option csharp_namespace = "Garimpei.Protos.Coupon.V1";

import "collector/v1/collector.proto"; // reuse Marketplace enum

service CouponCollectorService {
  // Fetch available coupons from a marketplace for a given tenant.
  rpc FetchCoupons(FetchCouponsRequest) returns (FetchCouponsResponse);
}

message FetchCouponsRequest {
  string owner_uid = 1;
  collector.v1.Marketplace marketplace = 2;
  int32 page_size = 3; // max items per page (default 500 for Shopee)
}

message FetchCouponsResponse {
  repeated CouponProto coupons = 1;
  int32 total_found = 2;
  string fetched_at = 3; // RFC3339
}

message CouponProto {
  string coupon_id = 1;
  collector.v1.Marketplace marketplace = 2;
  string code = 3;
  DiscountType discount_type = 4;
  double discount_value = 5;
  double min_spend = 6;
  string start_time = 7;  // RFC3339
  string end_time = 8;    // RFC3339
  repeated string applicable_categories = 9;
  string status = 10;     // active, expired
}

enum DiscountType {
  DISCOUNT_TYPE_UNSPECIFIED = 0;
  DISCOUNT_TYPE_PERCENTAGE = 1;
  DISCOUNT_TYPE_FIXED = 2;
}
```

### C# Domain Entities

File: `src/Garimpei.Domain/Entities/CouponAlertRule.cs`

```csharp
public sealed class CouponAlertRule : IOwnedEntity
{
    public Guid Id { get; init; } = Guid.NewGuid();
    public string OwnerUid { get; set; } = string.Empty;

    // Filter criteria
    public string DiscountType { get; set; } = "percentage"; // "percentage" or "fixed"
    public double MinDiscount { get; set; }                   // e.g. 10.0 for 10% or R$10
    public List<string> Marketplaces { get; set; } = [];      // ["shopee","amazon","mercadolivre"]
    public List<string> Categories { get; set; } = [];        // up to 10 categories (optional)

    // Notification
    public string Channel { get; set; } = "telegram";         // "telegram" or "whatsapp"

    // State
    public bool IsActive { get; set; } = true;
    public DateTime CreatedAt { get; init; } = DateTime.UtcNow;
    public DateTime UpdatedAt { get; set; } = DateTime.UtcNow;
}
```

File: `src/Garimpei.Domain/Entities/CouponAlertHistory.cs`

```csharp
public sealed class CouponAlertHistory : IOwnedEntity
{
    public Guid Id { get; init; } = Guid.NewGuid();
    public string OwnerUid { get; set; } = string.Empty;

    public string CouponId { get; set; } = string.Empty;
    public Guid AlertRuleId { get; set; }
    public double AlertedDiscountValue { get; set; }
    public DateTime AlertedAt { get; set; } = DateTime.UtcNow;

    // Dedup: records expire 72h after AlertedAt (48h beyond 24h window)
    public DateTime ExpiresAt { get; set; }
}
```

### BigQuery Schema (coupon_snapshots table)

File: `deploy/bigquery_coupon_schema.sql`

```sql
-- Coupon snapshots: append-only log of collected coupons.
-- Partitioned by collected_at for cost-efficient time-range queries.
CREATE TABLE IF NOT EXISTS `SEU_PROJECT.garimpo.coupon_snapshots` (
  coupon_id             STRING    NOT NULL,
  marketplace           STRING    NOT NULL,  -- "shopee", "amazon", "mercadolivre"
  code                  STRING,              -- voucher code or claiming URL
  discount_type         STRING    NOT NULL,  -- "percentage" or "fixed"
  discount_value        FLOAT64   NOT NULL,
  min_spend             FLOAT64,
  start_time            TIMESTAMP,
  end_time              TIMESTAMP,
  applicable_categories STRING,              -- JSON array
  status                STRING    NOT NULL,  -- "active", "expired", "claimed"
  detection_status      STRING,              -- "newly_discovered", "modified", "expired_or_removed", "unchanged"
  owner_uid             STRING    NOT NULL,
  collected_at          TIMESTAMP NOT NULL
)
PARTITION BY DATE(collected_at)
OPTIONS (
  partition_expiration_days = 90,
  description = "Append-only coupon snapshots for detection and analytics"
);
```

## Interfaces & Abstractions

### Go: CouponSource Interface

File: `internal/couponsource/source.go`

Follows the exact same pattern as `internal/source/source.go` — interface + factory + config.

```go
package couponsource

import "github.com/fmarquesfilho/garimpo/internal/domain"

// CouponSource is the port for coupon collection. Each marketplace implements this.
type CouponSource interface {
    // FetchCoupons retrieves available coupons for the given tenant credentials.
    FetchCoupons(cfg FetchConfig) ([]domain.Coupon, error)

    // Marketplace returns the marketplace identifier.
    Marketplace() string

    // Name returns a descriptive name for logging.
    Name() string
}

// FetchConfig holds per-request parameters for coupon fetching.
type FetchConfig struct {
    OwnerUID string
    PageSize int // max per page (500 for Shopee, 100 for Amazon)
}

// CouponSourceFactory creates a CouponSource with the given credentials.
type CouponSourceFactory func(cfg SourceConfig) CouponSource

// SourceConfig reuses the same credential fields as source.SourceConfig.
type SourceConfig struct {
    // Shopee
    AppID  string
    Secret string

    // Amazon (OAuth 2.0)
    AccessKey    string
    SecretKey    string
    PartnerTag   string
    RefreshToken string
    AccessToken  string

    // Mercado Livre (OAuth 2.0)
    ClientID     string
    ClientSecret string
    AccessToken  string
    RefreshToken string
}
```

### Go: CouponSource Registry

File: `internal/couponsource/registry.go`

Same thread-safe registry pattern as `internal/source/registry.go`.

```go
package couponsource

var DefaultRegistry = NewRegistry()

// Registry maps marketplace → CouponSourceFactory.
// Adapters register in init() or service setup.
type Registry struct { /* same as source.Registry */ }

func init() {
    DefaultRegistry.Register(domain.MarketplaceShopee, func(cfg SourceConfig) CouponSource {
        return NewShopeeCouponAdapter(cfg.AppID, cfg.Secret)
    })
    DefaultRegistry.Register(domain.MarketplaceAmazon, func(cfg SourceConfig) CouponSource {
        return NewAmazonCouponAdapter(cfg.AccessKey, cfg.SecretKey, cfg.PartnerTag, cfg.RefreshToken)
    })
    DefaultRegistry.Register(domain.MarketplaceMercadoLivre, func(cfg SourceConfig) CouponSource {
        return NewMLCouponAdapter(cfg.ClientID, cfg.ClientSecret, cfg.AccessToken, cfg.RefreshToken)
    })
}
```

### C#: ICouponSource Interface

File: `src/Garimpei.Domain/Interfaces/ICouponSource.cs`

Follows the same Keyed Services pattern as `IProductSource`.

```csharp
namespace Garimpei.Domain.Interfaces;

/// <summary>
/// Port for coupon collection per marketplace.
/// Registered via AddKeyedScoped — same Strategy pattern as IProductSource.
/// </summary>
public interface ICouponSource
{
    string MarketplaceId { get; }
    Task<CouponSourceResult> FetchCouponsAsync(string ownerUid, CancellationToken ct = default);
}

public sealed record CouponSourceResult
{
    public required IReadOnlyList<CouponCandidate> Coupons { get; init; }
    public int TotalFound { get; init; }
    public DateTime FetchedAt { get; init; } = DateTime.UtcNow;
}

public sealed record CouponCandidate
{
    public required string CouponId { get; init; }
    public required string Marketplace { get; init; }
    public string? Code { get; init; }
    public required string DiscountType { get; init; }
    public required double DiscountValue { get; init; }
    public double MinSpend { get; init; }
    public DateTime? StartTime { get; init; }
    public DateTime? EndTime { get; init; }
    public List<string> ApplicableCategories { get; init; } = [];
}
```

Registration in `DependencyInjection.cs`:

```csharp
// ─── Keyed ICouponSource (same pattern as IProductSource) ────────────
services.AddKeyedScoped<ICouponSource, ShopeeCouponSource>(Marketplaces.Shopee);
services.AddKeyedScoped<ICouponSource, AmazonCouponSource>(Marketplaces.Amazon);
services.AddKeyedScoped<ICouponSource, MLCouponSource>(Marketplaces.MercadoLivre);
```

## Component Details

### 1. coupon-collector (Go gRPC service)

**Location:** `services/coupon-collector/`

Mirrors the existing product collector (`services/collector/server.go`). One binary serving one marketplace at a time, determined at startup by env var `MARKETPLACE`.

**Architecture:**
- Proto: `protos/coupon/v1/coupon.proto` → generated to `gen/go/coupon/v1/`
- Server struct holds a `CouponSource` (from registry), same as `CollectorServer` holds `ProductSource`
- BigQuery writer for append-only inserts into `coupon_snapshots`

**Key behaviors:**
- Shopee adapter: paginates `productOfferV2` endpoint (500/page), 200ms delay between requests, HMAC-SHA256 auth
- Amazon adapter: OAuth 2.0 with token refresh, 1 req/s rate limit, HTTP 429 → 60s backoff
- Mercado Livre adapter: OAuth 2.0, token refresh, normalize promotions to coupon schema
- Retry: up to 2 retries with 5s exponential backoff on errors/timeouts (30s timeout)
- On complete: marks coupons with `end_time < now` as "expired" in BigQuery

**Server implementation sketch:**

```go
// services/coupon-collector/server.go
type CouponCollectorServer struct {
    couponpb.UnimplementedCouponCollectorServiceServer
    source  couponsource.CouponSource
    bq      *bigquery.Client
    logger  *slog.Logger
}

func (s *CouponCollectorServer) FetchCoupons(ctx context.Context, req *couponpb.FetchCouponsRequest) (*couponpb.FetchCouponsResponse, error) {
    mkt := source.ProtoToMarketplace(req.GetMarketplace())
    if mkt != s.source.Marketplace() {
        return nil, status.Errorf(codes.Unimplemented, "este collector serve %s", s.source.Marketplace())
    }

    coupons, err := s.source.FetchCoupons(couponsource.FetchConfig{
        OwnerUID: req.GetOwnerUid(),
        PageSize: int(req.GetPageSize()),
    })
    if err != nil {
        return nil, status.Errorf(codes.Internal, "falha ao buscar cupons: %v", err)
    }

    // Write to BigQuery (append-only)
    s.writeToBigQuery(ctx, coupons, req.GetOwnerUid())

    return &couponpb.FetchCouponsResponse{
        Coupons:    toProtoCoupons(coupons),
        TotalFound: int32(len(coupons)),
        FetchedAt:  time.Now().UTC().Format(time.RFC3339),
    }, nil
}
```

### 2. Coupon Detector (Python Analyzer)

**Location:** `services/analyzer/` (existing service, extended with new endpoint)

The Python analyzer already has BigQuery access and runs as a FastAPI service. Detection is a BigQuery diff query — no need for a new microservice.

**New endpoint:** `POST /detect-coupons`

**Input:**
```json
{
  "owner_uid": "abc123",
  "marketplace": "shopee",
  "snapshot_timestamp": "2025-01-15T10:00:00Z"
}
```

**Detection logic (BigQuery SQL):**

```sql
-- Newly discovered: in current snapshot but not in previous
WITH current AS (
  SELECT coupon_id, discount_value, end_time
  FROM `garimpo.coupon_snapshots`
  WHERE owner_uid = @owner_uid
    AND marketplace = @marketplace
    AND collected_at = @current_ts
),
previous AS (
  SELECT coupon_id, discount_value, end_time
  FROM `garimpo.coupon_snapshots`
  WHERE owner_uid = @owner_uid
    AND marketplace = @marketplace
    AND collected_at = (
      SELECT MAX(collected_at)
      FROM `garimpo.coupon_snapshots`
      WHERE owner_uid = @owner_uid
        AND marketplace = @marketplace
        AND collected_at < @current_ts
    )
)
SELECT
  c.coupon_id,
  CASE
    WHEN p.coupon_id IS NULL THEN 'newly_discovered'
    WHEN c.discount_value != p.discount_value OR c.end_time != p.end_time THEN 'modified'
    ELSE 'unchanged'
  END AS detection_status,
  c.discount_value,
  c.end_time
FROM current c
LEFT JOIN previous p ON c.coupon_id = p.coupon_id

UNION ALL

-- Expired/removed: in previous but not in current
SELECT
  p.coupon_id,
  'expired_or_removed' AS detection_status,
  p.discount_value,
  p.end_time
FROM previous p
LEFT JOIN current c ON p.coupon_id = c.coupon_id
WHERE c.coupon_id IS NULL
```

**Safety check:** If `current` has 0 rows, skip classification and log warning (R5-AC3).

**Output:** HTTP POST to C# API internal endpoint with detection results:
```json
{
  "owner_uid": "abc123",
  "marketplace": "shopee",
  "detections": [
    {
      "coupon_id": "SHOPEE_VOUCHER_123",
      "detection_status": "newly_discovered",
      "discount_type": "percentage",
      "discount_value": 20.0,
      "end_time": "2025-01-20T23:59:59Z",
      "applicable_categories": ["electronics", "fashion"]
    }
  ]
}
```

### 3. Alert Matcher (C# API)

**Location:** `src/Garimpei.Api/Endpoints/CouponAlerts/`

Receives detection results from the analyzer and evaluates them against tenant's `CouponAlertRule` records.

**Internal endpoint:** `POST /internal/coupon-alerts/evaluate`

**Logic flow:**
1. Receive detection events from Python analyzer
2. Filter to `newly_discovered` and `modified` only
3. For each detection, load all active `CouponAlertRule` for that `owner_uid`
4. Match: discount_type matches, discount_value >= min_discount, marketplace in rule's list, category overlap (if specified)
5. Deduplication check: query `CouponAlertHistory` for same coupon_id + rule_id within 24h window
   - If `newly_discovered` and already alerted in window → skip
   - If `modified` and already alerted → only re-alert if discount_value increased
6. Consolidate: group by (owner_uid, channel) to avoid duplicate messages
7. Dispatch via Alerter gRPC with formatted message

**Deduplication service:**

```csharp
public sealed class CouponDeduplicationService
{
    private readonly AppDbContext _db;

    public async Task<bool> ShouldAlertAsync(
        string couponId, Guid ruleId, string detectionStatus,
        double currentDiscountValue, CancellationToken ct)
    {
        var recent = await _db.CouponAlertHistories
            .Where(h => h.CouponId == couponId
                     && h.AlertRuleId == ruleId
                     && h.AlertedAt > DateTime.UtcNow.AddHours(-24))
            .OrderByDescending(h => h.AlertedAt)
            .FirstOrDefaultAsync(ct);

        if (recent is null) return true;

        // Modified coupon: only re-alert if discount increased
        if (detectionStatus == "modified")
            return currentDiscountValue > recent.AlertedDiscountValue;

        return false; // still within dedup window
    }
}
```

### 4. Scheduler Extension

**Location:** `services/scheduler/server.go` (extend existing)

New job type `coupon_collection` follows the same `executeJob` pattern.

**Changes:**
- Add gRPC client for `CouponCollectorService` (one per marketplace port)
- New method `executeCouponCollectionJob`:
  1. For each marketplace with valid tenant credentials (sequential: Shopee → Amazon → ML)
  2. Call `coupon-collector` gRPC `FetchCoupons`
  3. On success, call Python analyzer `POST /detect-coupons` to trigger detection
- Auto-registration: when C# API creates a tenant's first `CouponAlertRule`, it calls `SetSchedule` on the scheduler with default cron `0 */2 * * *` (BRT)

```go
func (s *SchedulerServer) executeCouponCollectionJob(job *registeredJob, params map[string]string) {
    ownerUID := params["owner_uid"]
    marketplaces := []couponpb.Marketplace{
        collectorpb.Marketplace_MARKETPLACE_SHOPEE,
        collectorpb.Marketplace_MARKETPLACE_AMAZON,
        collectorpb.Marketplace_MARKETPLACE_MERCADOLIVRE,
    }

    for _, mkt := range marketplaces {
        ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
        resp, err := s.couponCollector.FetchCoupons(ctx, &couponpb.FetchCouponsRequest{
            OwnerUid:    ownerUID,
            Marketplace: mkt,
            PageSize:    500,
        })
        cancel()

        if err != nil {
            s.logger.Warn("coupon collection failed",
                slog.String("marketplace", mkt.String()),
                slog.String("error", err.Error()))
            continue // next marketplace
        }

        // Trigger detection in analyzer
        s.triggerCouponDetection(ownerUID, mkt.String(), resp.GetFetchedAt())
    }
}
```

### 5. API Endpoints (C# Minimal API)

**Location:** `src/Garimpei.Api/Endpoints/Coupons/`

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v2/cupons/regras` | Create CouponAlertRule |
| GET | `/api/v2/cupons/regras` | List tenant's alert rules |
| PUT | `/api/v2/cupons/regras/{id}` | Update alert rule |
| DELETE | `/api/v2/cupons/regras/{id}` | Delete alert rule |
| PATCH | `/api/v2/cupons/regras/{id}/toggle` | Activate/deactivate rule |
| GET | `/api/v2/cupons` | List active coupons (from BigQuery via analyzer proxy) |
| GET | `/api/v2/cupons/historico` | Coupon analytics (time-range, marketplace filter) |
| POST | `/internal/coupon-alerts/evaluate` | Internal: receive detection events (no auth, internal network) |

**Validation rules:**
- Max 20 active rules per tenant (R6-AC5)
- `min_discount`: 1-99 for percentage, 0.01-99999.99 for fixed (R6-AC1)
- `categories`: max 10 items (R6-AC1)
- `marketplaces`: at least one, valid values only

## Database Migrations

### PostgreSQL (EF Core)

Migration: `AddCouponAlertRulesAndHistory`

```sql
-- CouponAlertRules
CREATE TABLE "CouponAlertRules" (
    "Id"            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    "OwnerUid"      text NOT NULL,
    "DiscountType"  text NOT NULL DEFAULT 'percentage',
    "MinDiscount"   double precision NOT NULL,
    "Marketplaces"  jsonb NOT NULL DEFAULT '[]',
    "Categories"    jsonb NOT NULL DEFAULT '[]',
    "Channel"       text NOT NULL DEFAULT 'telegram',
    "IsActive"      boolean NOT NULL DEFAULT true,
    "CreatedAt"     timestamp with time zone NOT NULL DEFAULT now(),
    "UpdatedAt"     timestamp with time zone NOT NULL DEFAULT now()
);

CREATE INDEX "IX_CouponAlertRules_OwnerUid" ON "CouponAlertRules" ("OwnerUid");
CREATE INDEX "IX_CouponAlertRules_Active" ON "CouponAlertRules" ("OwnerUid", "IsActive")
    WHERE "IsActive" = true;

-- CouponAlertHistory (deduplication tracking)
CREATE TABLE "CouponAlertHistory" (
    "Id"                    uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    "OwnerUid"              text NOT NULL,
    "CouponId"              text NOT NULL,
    "AlertRuleId"           uuid NOT NULL REFERENCES "CouponAlertRules"("Id") ON DELETE CASCADE,
    "AlertedDiscountValue"  double precision NOT NULL,
    "AlertedAt"             timestamp with time zone NOT NULL DEFAULT now(),
    "ExpiresAt"             timestamp with time zone NOT NULL
);

CREATE INDEX "IX_CouponAlertHistory_Dedup"
    ON "CouponAlertHistory" ("CouponId", "AlertRuleId", "AlertedAt" DESC);
CREATE INDEX "IX_CouponAlertHistory_Cleanup"
    ON "CouponAlertHistory" ("ExpiresAt");
```

**AppDbContext additions:**

```csharp
public DbSet<CouponAlertRule> CouponAlertRules => Set<CouponAlertRule>();
public DbSet<CouponAlertHistory> CouponAlertHistories => Set<CouponAlertHistory>();
```

Both entities get tenant query filter: `entity.HasQueryFilter(e => e.OwnerUid == _tenantContext.OwnerUid);`

### BigQuery

File: `deploy/bigquery_coupon_schema.sql` (shown in Data Models section above)

Partition expiration: 90 days (R8-AC2). No manual cleanup needed — BigQuery handles it.

## Configuration

### Environment Variables

**coupon-collector service:**

| Variable | Description | Example |
|----------|-------------|---------|
| `MARKETPLACE` | Which marketplace this instance serves | `shopee` |
| `GRPC_PORT` | gRPC listen port | `50061` |
| `BQ_PROJECT` | BigQuery project ID | `garimpei-prod` |
| `BQ_DATASET` | BigQuery dataset | `garimpo` |
| `BIGQUERY_EMULATOR_HOST` | (dev only) emulator address | `bigquery-emulator:9050` |

Marketplace credentials are fetched per-tenant from the C# API's `TenantConfig` table (existing pattern). The scheduler passes `owner_uid` in each request, and the collector loads credentials dynamically.

**Scheduler (new env vars):**

| Variable | Description | Example |
|----------|-------------|---------|
| `COUPON_COLLECTOR_SHOPEE_ADDR` | Shopee coupon collector address | `coupon-collector-shopee:50061` |
| `COUPON_COLLECTOR_AMAZON_ADDR` | Amazon coupon collector address | `coupon-collector-amazon:50062` |
| `COUPON_COLLECTOR_ML_ADDR` | ML coupon collector address | `coupon-collector-ml:50063` |
| `ANALYZER_URL` | Python analyzer HTTP base URL | `http://analyzer:8060` |

### Docker Compose

New services added to `docker-compose.yml`:

```yaml
  # ── Coupon Collectors (gRPC) ─────────────────────────────────────────────
  coupon-collector-shopee:
    build:
      context: .
      dockerfile: services/coupon-collector/Dockerfile
    ports:
      - "50061:50061"
    environment:
      MARKETPLACE: shopee
      GRPC_PORT: "50061"
      BQ_PROJECT: garimpei-dev
      BQ_DATASET: garimpo
      BIGQUERY_EMULATOR_HOST: bigquery-emulator:9050
    depends_on:
      - bigquery-emulator

  coupon-collector-amazon:
    build:
      context: .
      dockerfile: services/coupon-collector/Dockerfile
    ports:
      - "50062:50062"
    environment:
      MARKETPLACE: amazon
      GRPC_PORT: "50062"
      BQ_PROJECT: garimpei-dev
      BQ_DATASET: garimpo
      BIGQUERY_EMULATOR_HOST: bigquery-emulator:9050
    depends_on:
      - bigquery-emulator

  coupon-collector-ml:
    build:
      context: .
      dockerfile: services/coupon-collector/Dockerfile
    ports:
      - "50063:50063"
    environment:
      MARKETPLACE: mercadolivre
      GRPC_PORT: "50063"
      BQ_PROJECT: garimpei-dev
      BQ_DATASET: garimpo
      BIGQUERY_EMULATOR_HOST: bigquery-emulator:9050
    depends_on:
      - bigquery-emulator
```

## Tracing Requirements to Design

| Requirement | Component(s) | Key Design Decisions |
|---|---|---|
| **R1: Coupon Collection from Shopee** | coupon-collector (Shopee adapter), BigQuery | Shopee `productOfferV2` paginated at 500/page, 200ms throttle, HMAC-SHA256 auth, 30s timeout + 2 retries |
| **R2: Coupon Collection from Amazon** | coupon-collector (Amazon adapter), BigQuery | OAuth 2.0 token refresh, 1 req/s rate limit, HTTP 429 → 60s wait, skip if no credentials |
| **R3: Coupon Collection from Mercado Livre** | coupon-collector (ML adapter), BigQuery | OAuth 2.0, normalize promotions to unified schema, skip if no credentials |
| **R4: Coupon Scheduling** | Scheduler (extended) | New job type `coupon_collection`, default cron `0 */2 * * *`, sequential per marketplace, auto-create on first alert rule |
| **R5: New Coupon Detection** | Python Analyzer (BigQuery diff query) | SQL FULL OUTER JOIN on coupon_id between snapshots, empty-snapshot safety check, 60s SLA |
| **R6: Coupon Alert Rules** | C# API (CRUD endpoints), PostgreSQL | `CouponAlertRules` table, max 20/tenant, Keyed Services pattern, discount_type matching |
| **R7: Coupon Alert Notifications** | C# Alert Matcher → Alerter gRPC | Markdown for Telegram, emoji plain-text for WhatsApp, urgency tag for <24h expiry, affiliate link injection |
| **R8: Coupon Data Persistence** | BigQuery `coupon_snapshots` | Append-only, partitioned by `collected_at`, 90-day expiration, detection_status field |
| **R9: Coupon Deduplication** | C# Alert Matcher, PostgreSQL `CouponAlertHistory` | 24h window per coupon_id × rule_id, re-alert only on discount increase, 72h record retention, reset on rule edit |

## Alerter Proto Extension

The existing `AlerterService` needs a new RPC for coupon notifications (or the alert matcher can format and call `CheckAndNotify` with a coupon-specific payload). Recommended approach: add a new RPC to keep concerns separate.

File: `protos/alerter/v1/alerter.proto` (addition)

```protobuf
// New RPC for coupon-specific alerts
rpc SendCouponAlert(SendCouponAlertRequest) returns (SendCouponAlertResponse);

message SendCouponAlertRequest {
  string owner_uid = 1;
  string channel = 2;           // "telegram" or "whatsapp"
  string group_id = 3;          // chat_id or phone_number_id
  repeated CouponAlertPayload coupons = 4;
}

message CouponAlertPayload {
  string coupon_id = 1;
  string marketplace = 2;
  string discount_description = 3; // "20% OFF" or "R$15 de desconto"
  double min_spend = 4;
  string start_time = 5;
  string end_time = 6;
  repeated string categories = 7;
  string link_or_code = 8;
  bool expires_soon = 9;        // end_time within 24h
  string affiliate_link = 10;   // tenant's affiliate-tagged URL
}

message SendCouponAlertResponse {
  bool delivered = 1;
  string notified_at = 2; // RFC3339
}
```

## Open Decisions

1. **Single binary vs per-marketplace binary for coupon-collector:** Recommended single binary with `MARKETPLACE` env var (same as product collector). Simpler to maintain, one Dockerfile, different deploy targets.
2. **Detection trigger mechanism:** Scheduler calls analyzer HTTP after successful collection (push model). Alternative: analyzer polls BigQuery on its own cron — rejected for latency reasons.
3. **Credential routing:** Coupon collector receives `owner_uid`, calls C# API to fetch decrypted credentials. Alternative: pass credentials in gRPC metadata — rejected for security (credentials shouldn't travel over internal gRPC without mTLS).

## Components and Interfaces

| Component | Language | Interface | Location |
|-----------|----------|-----------|----------|
| Coupon Collector Server | Go | `CouponCollectorService` (gRPC proto) | `services/coupon-collector/` |
| Shopee Coupon Adapter | Go | `CouponSource` interface | `internal/couponsource/shopee_adapter.go` |
| Amazon Coupon Adapter | Go | `CouponSource` interface | `internal/couponsource/amazon_adapter.go` |
| ML Coupon Adapter | Go | `CouponSource` interface | `internal/couponsource/ml_adapter.go` |
| Coupon Source Registry | Go | `Registry` (map-based factory) | `internal/couponsource/registry.go` |
| Coupon Detector | Python | FastAPI endpoint `POST /detect-coupons` | `services/analyzer/` |
| Alert Matcher | C# | Internal endpoint `POST /internal/coupon-alerts/evaluate` | `src/Garimpei.Api/Endpoints/Coupons/` |
| Deduplication Service | C# | `CouponDeduplicationService` class | `src/Garimpei.Domain/Services/` |
| Alert Rules CRUD | C# | REST endpoints `/api/v2/cupons/regras` | `src/Garimpei.Api/Endpoints/Coupons/` |
| Shopee Coupon Source (C#) | C# | `ICouponSource` (Keyed "shopee") | `src/Garimpei.Infrastructure/Sources/` |
| Amazon Coupon Source (C#) | C# | `ICouponSource` (Keyed "amazon") | `src/Garimpei.Infrastructure/Sources/` |
| ML Coupon Source (C#) | C# | `ICouponSource` (Keyed "mercadolivre") | `src/Garimpei.Infrastructure/Sources/` |
| Scheduler Extension | Go | `executeCouponCollectionJob` method | `services/scheduler/server.go` |
| Alerter Extension | Go | `SendCouponAlert` RPC | `services/alerter/` |

## Correctness Properties

### Property 1: Append-Only Snapshot Integrity
Coupon snapshots in BigQuery are never updated or deleted — only appended. This ensures detection diff is always computed against pristine historical data.
**Validates: Requirements 8.3**

### Property 2: Idempotent Detection
Running detection twice on the same snapshot produces the same results, since it's a pure comparison between two immutable snapshots.
**Validates: Requirements 5.1**

### Property 3: Deduplication Invariant
A given (coupon_id, alert_rule_id) pair generates at most 1 alert per 24h window unless discount_value strictly increases. This is enforced by the PostgreSQL `CouponAlertHistory` table with a unique check on (coupon_id, rule_id, window).
**Validates: Requirements 9.1, 9.2**

### Property 4: Sequential Marketplace Execution
The scheduler processes marketplaces sequentially per tenant to prevent concurrent load spikes and ensure predictable resource usage.
**Validates: Requirements 4.3**

### Property 5: Graceful Degradation
If one marketplace fails (API down, auth expired), the system continues collecting from remaining marketplaces. Individual failures don't cascade.
**Validates: Requirements 1.6, 2.8, 3.5**

### Property 6: Tenant Isolation
All queries include `owner_uid` filter. EF Core global query filters and BigQuery WHERE clauses ensure no cross-tenant data leakage.
**Validates: Requirements 6.2, 8.1**

## Error Handling

| Scenario | Component | Behavior |
|----------|-----------|----------|
| Shopee API timeout (>30s) | Coupon Collector | Log + retry up to 2x with 5s backoff, then skip cycle |
| Amazon HTTP 429 | Coupon Collector | Wait 60s, retry up to 2x, then skip |
| Amazon/ML OAuth token expired | Coupon Collector | Attempt refresh; if refresh fails, log auth failure and skip marketplace |
| BigQuery write failure | Coupon Collector | Retry once; on failure, log and skip (snapshot is lost for this cycle) |
| Empty snapshot (0 coupons) | Coupon Detector | Skip detection for this cycle, log warning, do not mark previous coupons as expired |
| Detection processing error | Coupon Detector | Discard partial results, log failure with snapshot ID, retry on next cycle |
| Alert delivery failure (Telegram/WhatsApp) | Alerter | Retry once after 30s; if retry fails, log with rule_id + coupon_id |
| Max 20 rules exceeded | C# API | Return HTTP 409 with descriptive error message |
| Invalid rule parameters | C# API | Return HTTP 400 with validation details |
| Credential not configured | Coupon Collector | Skip marketplace silently (no error log) |

## Testing Strategy

### Unit Tests

| Component | What to test | Framework |
|-----------|--------------|-----------|
| CouponSource adapters (Go) | Mock HTTP responses, verify mapping to domain.Coupon | `go test` + `httptest` |
| Detection SQL logic | Verify newly_discovered/modified/expired classification | BigQuery emulator + `go test` |
| Alert matching logic (C#) | Rule evaluation against coupons, discount_type matching | xUnit + in-memory DB |
| Deduplication service (C#) | 24h window, discount increase bypass, rule edit reset | xUnit + in-memory DB |
| Proto mapping | Verify Go domain ↔ proto conversion | `go test` |

### Integration Tests

| Scenario | Components involved | Method |
|----------|---------------------|--------|
| End-to-end coupon collection | Scheduler → Collector → BigQuery | Docker Compose + BigQuery emulator |
| Detection + alert flow | BigQuery → Analyzer → C# API → (mock) Alerter | Integration test with test containers |
| Multi-tenant isolation | Two tenants, same coupons, independent rules | xUnit + in-memory PostgreSQL |
| Dedup across cycles | Run 3 collection cycles, verify single alert | Integration test with clock mocking |

### Architecture Tests (existing pattern)

Add to `ArchitectureTests.cs`:
- `CouponAlertRule` must implement `IOwnedEntity`
- `CouponAlertHistory` must implement `IOwnedEntity`
- `ICouponSource` must reside in `Garimpei.Domain.Interfaces`

