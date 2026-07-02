# ADR-0018: Unificar collectors em binário único config-driven

## Status

Aceito (2026-07-02)

## Contexto

Atualmente o Garimpei tem **5 binários separados** para coleta de dados:

| Binário | Porta | O que faz |
|---------|-------|-----------|
| `services/collector` | :50051 | Produtos Shopee |
| `services/collector-amazon` | :50055 | Produtos Amazon |
| `services/coupon-collector` (Shopee) | :50061 | Cupons Shopee |
| `services/coupon-collector` (Amazon) | :50062 | Cupons Amazon |
| `services/coupon-collector` (ML) | :50063 | Cupons Mercado Livre |

Cada um é um binário Go separado com seu Dockerfile, que lê env vars para credenciais
e serve um único marketplace. Isso gera:

- **5 Dockerfiles** praticamente idênticos
- **5 containers** no Cloud Run (custo de cold start × 5)
- **5 health checks** para monitorar
- **15+ env vars** para configurar
- Código duplicado entre `services/collector/server.go` e `services/collector-amazon/server.go`

### Estado da arte: OpenTelemetry Collector

O padrão moderno em Go para collectors é o modelo do [OpenTelemetry Collector](https://opentelemetry.io/docs/collector/architecture/):

```yaml
# Um binário, múltiplos pipelines definidos em YAML
receivers:
  shopee:
    app_id: "123"
    secret_enc: "vault://shopee-secret"
    type: products
  amazon:
    access_key_enc: "vault://amazon-key"
    type: coupons

processors:
  batch:
    timeout: 5s

exporters:
  bigquery:
    project: garimpei-prod
    dataset: garimpo
    table: snapshots

pipelines:
  products:
    receivers: [shopee, amazon]
    processors: [batch]
    exporters: [bigquery]
```

Princípio: **configuração, não código** para decidir o que coletar.

### Cannectors (referência prática)

[Cannectors](https://cannectors.com/) é um Go binary que faz exatamente isso para
HTTP APIs: single binary + YAML config + pull from APIs + transform + export.
Retries, OAuth2, HMAC, state management — tudo por configuração.

## Decisão

### Unificar em um único binário `garimpei-collector`

```
services/collector/          ← binário único
├── main.go                  ← startup: lê config YAML, monta pipelines
├── config.go                ← parser da config YAML
├── server.go                ← gRPC server (genérico, delega para pipeline)
├── pipeline.go              ← orquestrador: receiver → processor → exporter
└── Dockerfile               ← UM Dockerfile

internal/
├── source/                  ← receivers (produto): Shopee, Amazon, ML adapters
├── couponsource/            ← receivers (cupom): Shopee, Amazon, ML adapters
└── exporter/                ← exporters: BigQuery, (futuro: Kafka, S3)
```

### Config YAML (`collector.yaml`)

```yaml
# /etc/garimpei/collector.yaml ou env COLLECTOR_CONFIG
version: "1"

receivers:
  - id: shopee-products
    type: product
    marketplace: shopee
    schedule: "*/30 * * * *"  # a cada 30 min
    credentials:
      app_id_env: SHOPEE_APP_ID
      secret_env: SHOPEE_SECRET

  - id: amazon-products
    type: product
    marketplace: amazon
    schedule: "0 * * * *"     # a cada hora
    credentials:
      access_key_env: AMAZON_ACCESS_KEY
      secret_key_env: AMAZON_SECRET_KEY
      partner_tag_env: AMAZON_PARTNER_TAG

  - id: shopee-coupons
    type: coupon
    marketplace: shopee
    schedule: "0 */2 * * *"   # a cada 2h
    credentials:
      app_id_env: SHOPEE_APP_ID
      secret_env: SHOPEE_SECRET

  - id: amazon-coupons
    type: coupon
    marketplace: amazon
    schedule: "0 */2 * * *"
    credentials:
      access_key_env: AMAZON_ACCESS_KEY
      secret_key_env: AMAZON_SECRET_KEY
      partner_tag_env: AMAZON_PARTNER_TAG

exporters:
  bigquery:
    project_env: BQ_PROJECT
    dataset: garimpo
    products_table: snapshots
    coupons_table: coupon_snapshots

settings:
  grpc_port: 50051
  health_port: 8081
  log_level: info
  max_concurrent_receivers: 3
```

### Benefícios

1. **Um container** em vez de 5 → custo Cloud Run reduz 80%
2. **Um Dockerfile** → manutenção simplificada
3. **Adicionar marketplace** = adicionar bloco YAML (zero código novo se adapter existe)
4. **Scheduling interno** (cron embutido) → não precisa do scheduler para coleta básica
5. **Rate limiting centralizado** → evita que múltiplos receivers estourem limites
6. **Testável localmente** com um único `docker run`

### O que mantém do design atual

- **Interfaces** `ProductSource` e `CouponSource` continuam iguais
- **Registry** de adapters continua (lookup by marketplace string)
- **Adapters** (Shopee, Amazon, ML) continuam como estão
- **gRPC API** continua disponível (scheduler pode chamar sob demanda)
- **BigQuery exporter** continua append-only

### Migração (incremental)

1. Criar `services/collector/config.go` + `pipeline.go`
2. O `main.go` lê YAML e instancia receivers do registry
3. Manter gRPC API para chamadas sob demanda (scheduler)
4. Remover `services/collector-amazon/` e `services/coupon-collector/`
5. Um único Dockerfile + um único serviço no docker-compose

## Consequências

### Positivas

- 80% menos containers (1 em vez de 5)
- Config-as-code (YAML versionado no repo)
- Adicionar marketplace = config change, não code change
- Scheduling embutido elimina dependência do scheduler para coleta
- Rate limit centralizado (1 binary controla todos os requests)
- Cold start único (1 container wakes up, não 5)

### Negativas

- Blast radius maior (se o collector cai, todas as coletas param)
  - Mitigação: health check granular por receiver, restart on failure
- Mais memória no container (múltiplos adapters carregados)
  - Mitigação: Go é leve (~50MB total para todos os adapters)
- Config YAML precisa de validação robusta na startup

### Alternativa rejeitada

**Manter N binários com env var `MARKETPLACE`**: já é o que temos, e funciona.
Mas não escala — cada marketplace novo = novo Dockerfile + novo deploy +
novo health check + novas env vars. O overhead operacional cresce linearmente.

## Referências

- [OpenTelemetry Collector Architecture](https://opentelemetry.io/docs/collector/architecture/)
- [Cannectors](https://cannectors.com/) — Go binary, YAML config, HTTP API collector
- ADR-0016: Multi-marketplace (define os adapters por marketplace)
