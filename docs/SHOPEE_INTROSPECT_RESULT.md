# Resultado da Introspecção — API Afiliados Shopee

Data: 2026-06-27
AppID: 18325460168
Endpoint: https://open-api.affiliate.shopee.com.br/graphql

## Conclusão

A API de afiliados da Shopee **NÃO expõe** campos de origem do produto (país de fabricação).
Campos testados e **rejeitados** (erro 10010 "Cannot query field"):
- `brandName` — não existe em ProductOfferV2
- `sellerLocation` — não existe
- `originCountry` — não existe
- `productOrigin` — não existe

O campo **`shopType`** (tipo LIST de enum ShopType) **EXISTE** em ProductOfferV2 e pode ser pedido.
Valores possíveis do enum `ShopType`: indica tipo de loja (mall, preferred, overseas, etc).

## Campos disponíveis no tipo ProductOfferV2

Extraídos via `__schema` (introspecção completa):

| Campo | Tipo | Nullable |
|-------|------|----------|
| itemId | NON_NULL | não |
| commissionRate | NON_NULL | não |
| appExistRate | NON_NULL | não |
| appNewRate | NON_NULL | não |
| webExistRate | NON_NULL | não |
| webNewRate | NON_NULL | não |
| commission | NON_NULL | não |
| price | NON_NULL | não |
| sales | NON_NULL | não |
| imageUrl | NON_NULL | não |
| productName | NON_NULL | não |
| shopName | NON_NULL | não |
| productLink | NON_NULL | não |
| offerLink | NON_NULL | não |
| periodEndTime | NON_NULL | não |
| periodStartTime | NON_NULL | não |
| priceMin | NON_NULL | não |
| priceMax | NON_NULL | não |
| productCatIds | LIST | sim |
| ratingStar | NON_NULL | não |
| priceDiscountRate | NON_NULL | não |
| shopId | NON_NULL | não |
| **shopType** | **LIST** | sim |
| sellerCommissionRate | NON_NULL | não |
| shopeeCommissionRate | NON_NULL | não |

## Campos NÃO disponíveis (confirmado)

- `brandName` — não existe
- `sellerLocation` — não existe
- `originCountry` — não existe
- `productOrigin` — não existe
- `origin` — não existe
- `country` — não existe

## Decisão de design

Como a API não expõe país de origem do produto, a abordagem é:

1. **`shopType`** — pedir na query e exibir como badge (ex: "Mall", "Preferred").
   Lojas "overseas" ou cross-border podem indicar origem estrangeira, mas não especificam o país.

2. **Fallback `origem_padrao`** — campo na Busca (configurado pelo usuário ao adicionar loja).
   A Mileny marca manualmente quais lojas são coreanas/japonesas.
   Todos os produtos da loja herdam o badge automaticamente.

Esta é a solução definitiva até que a Shopee adicione campos de origem à API de afiliados.
