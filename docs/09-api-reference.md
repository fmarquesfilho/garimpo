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

---

## Detalhes de endpoints-chave

### GET /api/buscas

Retorna buscas ativas do tenant autenticado.

**Response:**
```json
{
  "buscas": [
    {
      "id": "uuid",
      "keywords": ["serum"],
      "shop_ids": [920292999, 282170857],
      "shop_names": { "920292999": "Glory of Seoul", "282170857": "Le Botanic" },
      "cron": "0 */8 * * *",
      "comissao_min": 0.07,
      "vendas_min": 10,
      "categorias": ["Skincare"],
      "fontes": ["curadoria", "lojas"],
      "marketplaces": "shopee",
      "ativo": true,
      "sort_by": "relevance",
      "limit": 50
    }
  ],
  "total": 1
}
```

**Notas:**
- `shop_names` substituiu o campo `nome` (deprecated). É um dict `id→nome` para todas as lojas.
- Para buscas keyword-only (sem ShopIds), `shop_names` é `null`.
- Fallback legacy: buscas antigas sem `ShopNames` persistido usam `Keyword` como nome.

### GET /api/candidatos

Busca produtos com scoring e ranking.

**Query params:**
| Param | Tipo | Obrigatório | Descrição |
|-------|------|-------------|-----------|
| `keyword` | string | Sim (exceto shop) | Termo de busca |
| `top` | int | Não (default 50) | Limite de resultados |
| `comissao_min` | double | Não | Filtro comissão mínima |
| `vendas_min` | int | Não | Filtro vendas mínimas |
| `nota_min` | double | Não | Filtro nota mínima |
| `fonte` | string | Não | `"shopee-shop"` para busca por loja |
| `shop_ids` | string | Não | IDs separados por vírgula (requer `fonte=shopee-shop`) |

**Roteamento:**
- `fonte=shopee-shop` + `shop_ids`: usa `FetchShop` no Collector (sem keyword obrigatório)
- Caso contrário: usa `Fetch` com keyword obrigatório

### POST /api/buscas

Cria ou atualiza busca (upsert por keyword).

**Request body:**
```json
{
  "keywords": ["serum"],
  "shop_ids": [920292999],
  "shop_names": { "920292999": "Glory of Seoul" },
  "cron": "0 */8 * * *",
  "comissao_min": 0.07,
  "categorias": ["Skincare"],
  "fontes": ["curadoria"],
  "marketplaces": "shopee"
}
```

**Notas:**
- `shop_names` é persistido junto com a busca para o GET retornar nomes corretos.
- `?remover=true` no query string desativa a busca.

