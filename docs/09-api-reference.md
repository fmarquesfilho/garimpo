# API Reference

Portal unificado de referência de todos os serviços do Garimpei.

## Endpoints HTTP (C# API — :8080)

Autenticação: Bearer token (Firebase JWT). Exceto `/api/health` e `/internal/*`.

### Endpoints extraídos do código

| Método | Rota | Arquivo |
|--------|------|---------|
| Get | `/` | AlertasEndpoints.cs |
| Delete | `/` | DestinosEndpoints.cs |
| Get | `/` | DestinosEndpoints.cs |
| Post | `/` | DestinosEndpoints.cs |
| Delete | `/` | FavoritosEndpoints.cs |
| Get | `/` | FavoritosEndpoints.cs |
| Post | `/` | FavoritosEndpoints.cs |
| Post | `/` | PublicacaoEndpoints.cs |
| Get | `/` | PublicacoesEndpoints.cs |
| Post | `/` | PublicacoesEndpoints.cs |
| Delete | `/` | TemplatesEndpoints.cs |
| Get | `/` | TemplatesEndpoints.cs |
| Post | `/` | TemplatesEndpoints.cs |
| Post | `/amazon` | OnboardingEndpoints.cs |
| Get | `/api/admin/logs` | LogsEndpoints.cs |
| Get | `/api/admin/me` | CoreEndpoints.cs |
| Get | `/api/buscas` | BuscasEndpoints.cs |
| Post | `/api/buscas` | BuscasEndpoints.cs |
| Get | `/api/candidatos` | CoreEndpoints.cs |
| Get | `/api/categorias` | CoreEndpoints.cs |
| Get | `/api/coletas` | AnalyticsEndpoints.cs |
| Get | `/api/conversoes/reais` | AnalyticsEndpoints.cs |
| Get | `/api/conversoes` | AnalyticsEndpoints.cs |
| Get | `/api/estatisticas` | AnalyticsEndpoints.cs |
| Get | `/api/health` | CoreEndpoints.cs |
| Get | `/api/lojas/evolucao` | LojasEndpoints.cs |
| Get | `/api/lojas/novidades` | LojasEndpoints.cs |
| Delete | `/api/lojas` | LojasEndpoints.cs |
| Get | `/api/lojas` | LojasEndpoints.cs |
| Post | `/api/lojas` | LojasEndpoints.cs |
| Post | `/api/publicar` | PublicacoesEndpoints.cs |
| Post | `/api/resolver-link` | ResolverLinkEndpoints.cs |
| Post | `/api/telemetry` | LogsEndpoints.cs |
| Post | `/api/templates/preview` | TemplatesEndpoints.cs |
| Post | `/configurar` | AlertasEndpoints.cs |
| Get | `/destinos` | PublicacaoEndpoints.cs |
| Post | `/excluir-conta` | OnboardingEndpoints.cs |
| Get | `/favoritos` | CuradoriaEndpoints.cs |
| Post | `/internal/publish-scheduled` | ScheduledPublishEndpoints.cs |
| Get | `/novos` | CuradoriaEndpoints.cs |
| Post | `/process-alert` | AlertProxyEndpoints.cs |
| Get | `/quedas` | CuradoriaEndpoints.cs |
| Get | `/ranking/shop` | CuradoriaEndpoints.cs |
| Get | `/ranking` | CuradoriaEndpoints.cs |
| Post | `/shopee` | OnboardingEndpoints.cs |
| Get | `/status` | OnboardingEndpoints.cs |
| Post | `/telegram` | OnboardingEndpoints.cs |
| Post | `/termos` | OnboardingEndpoints.cs |
| Post | `/testar` | AlertasEndpoints.cs |
| Post | `/validar` | OnboardingEndpoints.cs |
| Post | `/whatsapp` | OnboardingEndpoints.cs |

---

## Serviços gRPC (Go sidecars)

Comunicação interna via localhost (Cloud Run multi-container).

### alerter — `AlerterService`

| RPC | Descrição |
|-----|-----------|
| `CheckAndNotify(CheckAndNotifyRequest)` | |
| `SendCouponAlert(SendCouponAlertRequest)` | |

### collector — `CollectorService`

| RPC | Descrição |
|-----|-----------|
| `ResolveShop(ResolveShopRequest)` | |
| `GenerateAffiliateLink(GenerateAffiliateLinkRequest)` | |
| `Fetch(FetchRequest)` | |
| `FetchShop(FetchShopRequest)` | |

### coupon — `CouponCollectorService`

| RPC | Descrição |
|-----|-----------|
| `FetchCoupons(FetchCouponsRequest)` | |

### publisher — `PublisherService`

| RPC | Descrição |
|-----|-----------|
| `Publish(PublishRequest)` | |
| `ListGroups(ListGroupsRequest)` | |

### scheduler — `SchedulerService`

| RPC | Descrição |
|-----|-----------|
| `TriggerJob(TriggerJobRequest)` | |
| `ListJobs(ListJobsRequest)` | |
| `SetSchedule(SetScheduleRequest)` | |

### Analyzer (`:8060`) — Python FastAPI

| Método | Rota | Descrição |
|--------|------|-----------|
| GET | `/coletas` | |
| GET | `/conversoes` | |
| POST | `/detect-coupons` | |
| GET | `/estatisticas` | |
| GET | `/evolucao` | |
| GET | `/novidades` | |
| GET | `/quedas` | |

---

## Frontend (SvelteKit — Cloudflare Pages)

| Rota | Página |
|------|--------|
| `/` | |
| `/admin` | |
| `/admin/logs` | |
| `/canais` | |
| `/coletas` | |
| `/configurar` | |
| `/estatisticas` | |
| `/publicacoes` | |
| `/publicar` | |

---

## Contratos de Serviço

Definidos em `contracts/registry.yaml` (ADR-0020). Validação: `mise run check:service-contracts`

Proto files: `protos/collector/v1/collector.proto`, `protos/publisher/v1/publisher.proto`, `protos/scheduler/v1/scheduler.proto`
