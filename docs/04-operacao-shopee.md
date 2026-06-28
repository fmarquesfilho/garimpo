# Integração Shopee

## API de Afiliados (GraphQL)

Endpoint Brasil:
```
POST https://open-api.affiliate.shopee.com.br/graphql
Content-Type: application/json
```

### Autenticação

```
Authorization: SHA256 Credential={AppId}, Timestamp={ts}, Signature={sig}
Signature = SHA256(AppId + Timestamp + Payload + Secret)
```

- Timestamp: Unix em segundos (janela ~5 min)
- Payload: corpo JSON exato (mesmos bytes que vão no body)
- Divergência de whitespace → erro 10020 (Invalid Signature)

Cada tenant configura `app_id` + `secret` no onboarding. Credenciais criptografadas
com `ENCRYPTION_KEY` antes de armazenar.

### Endpoints utilizados

**`productOfferV2`** — busca produtos com comissão.
Parâmetros: `keyword`, `productCatId`, `shopId`, `itemId`, `listType`, `sortType`,
`page`, `limit` (1–500).

Campos retornados:

| Campo API | Domínio Go | Descrição |
|---|---|---|
| `commissionRate` | `Commission` | Fração (0.0850 = 8,5%) |
| `priceMin` | `Price` | Preço mínimo |
| `sales` | `Sales30d` | Proxy de demanda |
| `ratingStar` | `Rating` | Avaliação 0–5 |
| `offerLink` | `Link` | Link de afiliado com tracking |
| `productName` | `Name` | Nome do produto |
| `shopName` | `ShopName` | Loja vendedora |
| `productCatIds` | `CatIDs` | IDs de categoria hierárquicos |

**`shopOfferV2`** — busca produtos de uma loja específica pelo `shopId`.
Mesmos campos de retorno. Usado para monitoramento de lojas.

**`generateShortLink`** — gera link curto com até 5 `subIds` de tracking.
SubIds voltam no `conversionReport` campo `utmContent`.

**`conversionReport`** — relatório de conversões (pedidos via links):
`purchaseTime`, `clickTime`, `conversionId`, `totalCommission`, `utmContent`.
Status: UNPAID → PENDING → COMPLETED → CANCELLED.
Paginação por `scrollId` (vale 30s).

### Códigos de erro comuns

| Código | Significado |
|---|---|
| 10020 | Invalid Signature |
| 10030 | Rate Limit |
| 10035 | No API Access |
| 11001 | Params Error |
| 10010 | Parse Error |

## Origem do produto

### Problema

O "País de Origem" não é exposto pela API de afiliados (GraphQL). Campos como
`brandName`, `sellerLocation`, `originCountry` causam erro 10010.

### Tentativas realizadas

1. API de Afiliados (introspecção) → campos não existem
2. API Pública v4 via Cloud Run → 403 (IP de datacenter bloqueado)
3. Cloudflare Worker como proxy → 403
4. Worker com User-Agent Googlebot → 403
5. Endpoint alternativo `/api/v4/item/get` → 403
6. Proxy residencial (Bright Data) → requer KYC

### Solução implementada: `origem_padrao` por loja

A Mileny marca "🇰🇷 Coreia" ao adicionar a loja → campo `origem_padrao` salvo na
busca → motor de coleta aplica a todos os produtos → CandidateCard exibe badge.

Limitação: não diferencia por produto individual. Aceitável porque lojas coreanas
vendem produtos coreanos (1:1).

### Alternativa futura

Se a Shopee expor campo de origem na API, basta:
1. Adicionar à query GraphQL em `shopee.go`
2. Mapear para `domain.Product.Origin`
3. Badge funciona automaticamente

Monitorar via Admin → Introspecção periodicamente.

## Schema GraphQL — Campos disponíveis

Extraídos via introspecção real (`/api/admin/shopee-introspect`):

`itemId`, `commissionRate`, `commission`, `price`, `priceMin`, `priceMax`,
`sales`, `imageUrl`, `productName`, `shopName`, `shopId`, `productLink`,
`offerLink`, `productCatIds`, `ratingStar`, `priceDiscountRate`, `shopType`,
`sellerCommissionRate`, `shopeeCommissionRate`, `periodStartTime`, `periodEndTime`,
`appExistRate`, `appNewRate`, `webExistRate`, `webNewRate`.
