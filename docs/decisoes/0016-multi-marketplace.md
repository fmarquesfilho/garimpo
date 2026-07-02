# ADR-0016: Suporte multi-marketplace (Shopee + Amazon + Mercado Livre)

## Status

Aceito (2026-07-02) — Fase 1 implementada, Fase 2 em progresso

## Contexto

O Garimpei foi construído exclusivamente para a Shopee (API de afiliados, coleta,
publicação com links de afiliado). O modelo de dados, o collector Go, o scoring
engine e o frontend assumem que a fonte é sempre a Shopee.

Para crescer o produto, queremos suportar **múltiplos marketplaces** — inicialmente
Amazon Brasil e Mercado Livre — mantendo a mesma experiência: curadoria, ranking,
publicação em canais, e rastreamento de conversão.

### APIs disponíveis

| Marketplace | API | Auth | Busca | Comissão | Link afiliado | Status |
|-------------|-----|------|-------|----------|---------------|--------|
| **Shopee** | Affiliate API (OpenAPI) | HMAC-SHA256 (AppID + Secret) | `productOfferV2` por keyword/shop | Retornado no payload (%) | Gerado pela API | ✅ Em uso |
| **Amazon BR** | Creators API (substitui PA-API 5.0, deprecada mai/2026) | OAuth 2.0 (Creators platform) | `SearchItems`, `GetItems` | Tabela fixa por categoria (1-15%) | Tag no URL (PartnerTag) | 🔮 Novo |
| **Mercado Livre** | Items/Search API (REST, OAuth 2.0) | OAuth 2.0 (Client ID + Secret) | `GET /sites/MLB/search?q=...` | Programa afiliados (7-16% por categoria) | URL com tag `?aff_source=...` | 🔮 Novo |

### Diferenças-chave entre as APIs

| Aspecto | Shopee | Amazon (Creators) | Mercado Livre |
|---------|--------|-------------------|---------------|
| Modelo de produto | `item_id` + `shop_id` | ASIN | `item_id` (MLB...) |
| Preço | `price` (float) | `Amount` (int cents ou string) | `price` (float) |
| Comissão | No payload (%) | Tabela fixa por categoria | Tabela fixa por categoria |
| Vendas/demanda | `sold` (30 dias) | Não disponível via API | `sold_quantity` |
| Avaliação | `rating_star` (0-5) | `star_rating` (0-5) | Não diretamente (reviews) |
| Imagem | `image` (URL) | `Images[0].Large.URL` | `pictures[0].url` |
| Link afiliado | Gerado pela API | Tag no URL | Tag no URL |
| Rate limiting | 200ms entre requests | 1 req/s (padrão) | 12k req/min (autenticado) |
| País | Brasil (shopee.com.br) | Brasil (amazon.com.br) | Brasil (MLB) |

## Decisão

### 1. Abstrair o conceito de "fonte" (Source/Marketplace)

Criar uma interface `IProductSource` que cada marketplace implementa:

```csharp
public interface IProductSource
{
    string MarketplaceId { get; }  // "shopee" | "amazon" | "mercadolivre"
    Task<SourceResult> SearchAsync(SearchQuery query, CancellationToken ct);
    Task<SourceResult> FetchByShopAsync(string shopId, int limit, CancellationToken ct);
    string GenerateAffiliateLink(string productUrl, string affiliateTag);
}
```

### 2. Modelo de produto unificado (ProductCandidate)

O `ProductCandidate` existente já é agnóstico de marketplace — ele tem:
- `Id`, `Name`, `Price`, `Commission`, `Sales`, `Rating`, `Link`, `ImageUrl`

Adicionar campo `Marketplace` (string) para identificar a origem:
```csharp
public string? Marketplace { get; init; }  // "shopee", "amazon", "mercadolivre"
```

### 3. Collector multi-source (gRPC → HTTP adapters)

O collector Go atual chama a Shopee API. Para multi-marketplace:

**Opção A: Um collector por marketplace (microserviços separados)**
- `collector-shopee` (Go, gRPC :50051) — já existe
- `collector-amazon` (Go ou C#, gRPC :50055) — novo
- `collector-mercadolivre` (Go ou C#, gRPC :50056) — novo

**Opção B: Collector único com adapters internos**
- Um único serviço com `switch` por marketplace
- Mais simples de deployar (1 container), mas menos isolado

**Escolha: Opção A** — cada marketplace tem rate limits, auth e throttling
diferentes. Isolá-los permite falhar independentemente (resiliência).

### 4. Credenciais por marketplace no TenantConfig

Expandir `TenantConfig` para suportar múltiplas credenciais:

```csharp
// Shopee (existente)
public string? ShopeeAppId { get; set; }
public string? ShopeeSecretEnc { get; set; }

// Amazon (Creators API)
public string? AmazonAccessKeyEnc { get; set; }
public string? AmazonPartnerTag { get; set; }

// Mercado Livre (OAuth)
public string? MeliClientId { get; set; }
public string? MeliClientSecretEnc { get; set; }
public string? MeliAccessTokenEnc { get; set; }
public string? MeliRefreshTokenEnc { get; set; }
```

### 5. Busca multi-source na Busca entity

A entidade `Busca` precisa indicar quais marketplaces consultar:

```csharp
public string[] Marketplaces { get; set; } = ["shopee"];  // ["shopee", "amazon", "mercadolivre"]
```

### 6. Scoring unificado

O scoring engine (`ScoringService`) já é agnóstico — recebe `ProductCandidate[]`
e rankeia. A única mudança é que a lista de candidatos pode vir de múltiplas
fontes mescladas. O campo `Marketplace` no resultado permite ao frontend mostrar
de onde veio.

### 7. Publicação com link correto

O publisher precisa gerar o link de afiliado correto por marketplace:
- Shopee: link já vem pronto da API
- Amazon: `productUrl + ?tag=PARTNER_TAG`
- Mercado Livre: `productUrl + ?aff_source=TAG`

### 8. Frontend — indicador de marketplace

Na UI, cada produto mostra um badge com o marketplace de origem.
O formulário de busca permite selecionar quais marketplaces consultar.

### 9. BigQuery — campo marketplace nos snapshots

```sql
ALTER TABLE snapshots ADD COLUMN marketplace STRING;  -- "shopee" | "amazon" | "mercadolivre"
```

Permite análise de variações cross-marketplace (mesmo produto em lojas diferentes).

## Modificações necessárias

### Backend (C# API)

| Mudança | Esforço | Impacto |
|---------|---------|---------|
| `ProductCandidate.Marketplace` field | P | Baixo |
| `Busca.Marketplaces` field | P | Baixo |
| `TenantConfig` + campos Amazon/ML | M | Migration |
| Onboarding steps opcionais por marketplace | M | Frontend + Backend |
| Endpoint `/api/candidatos` aceita `marketplace` param | P | Retrocompat |

### Microserviços (Go gRPC)

| Mudança | Esforço | Impacto |
|---------|---------|---------|
| `collector-amazon` (novo serviço) | G | Novo container |
| `collector-mercadolivre` (novo serviço) | G | Novo container |
| Proto `FetchRequest.marketplace` field | P | Retrocompat |
| Scheduler: despachar para collector correto | M | Lógica de routing |

### Frontend (SvelteKit)

| Mudança | Esforço | Impacto |
|---------|---------|---------|
| Badge de marketplace nos cards | P | Visual |
| Seletor de marketplace na busca | M | UX |
| Onboarding com steps opcionais | M | Wizard |

### BigQuery

| Mudança | Esforço | Impacto |
|---------|---------|---------|
| Campo `marketplace` em snapshots | P | Schema evolution |
| Analyzer: filtro por marketplace | P | Query params |

## Fases de implementação

### Fase 1: Abstração (sem marketplace novo)
- Adicionar `Marketplace` field no modelo
- Refactor collector para aceitar `marketplace` no proto
- Tudo continua funcionando como antes (default: "shopee")

### Fase 2: Amazon (Creators API)
- `collector-amazon` Go service
- OAuth 2.0 flow no onboarding
- Comissão por tabela de categorias

### Fase 3: Mercado Livre
- `collector-mercadolivre` Go service
- OAuth 2.0 flow no onboarding (com refresh token)
- Busca via `/sites/MLB/search`

## Consequências

### Positivas
- Produto não depende de um único marketplace (reduz risco de negócio)
- Mais produtos disponíveis para curadoria (pool maior = scoring melhor)
- Comparação cross-marketplace (mesmo produto, preço diferente)
- Possibilidade de arbitragem (publicar o mais barato entre os 3)

### Negativas
- Complexidade de manutenção (3 APIs, 3 auth flows, 3 rate limits)
- Scoring precisa normalizar dados heterogêneos (Amazon não tem `sales`)
- Cada marketplace pode mudar sua API independentemente
- Custos de infra: +2 containers no Cloud Run

### Riscos
- Amazon Creators API é nova (substituiu PA-API em mai/2026) — pode ter instabilidades
- Mercado Livre pode restringir acesso à API de busca (requer app certificado)
- Comissões da Amazon/ML são fixas por categoria (menos precisas que Shopee)
- Rate limits mais restritivos na Amazon (1 req/s)

## Referências

- [Amazon Creators API](https://affiliate-program.amazon.com/creatorsapi/docs/en-us/introduction) — substituto do PA-API 5.0
- [Mercado Livre Developers](https://developers.mercadolivre.com.br) — Items & Search API
- [Shopee Affiliate API](https://affiliate.shopee.com.br/open_api) — API atual
