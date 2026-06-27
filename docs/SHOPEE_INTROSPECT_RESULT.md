# Origem do Produto — Investigação Completa

Data: 2026-06-27
Objetivo: Obter automaticamente o "País de Origem" dos produtos da Shopee para diferenciar originais de falsificações.

---

## Resultado Final

**Não é possível obter a origem do produto de forma programática sem serviço externo pago.**

A solução implementada é `origem_padrao` por loja monitorada: a Mileny marca "Coreia" uma vez ao adicionar a loja e todos os produtos herdam o badge automaticamente.

---

## Tentativas realizadas (em ordem cronológica)

### 1. API de Afiliados (GraphQL) — campo de origem na query

**Hipótese:** A API GraphQL pode ter campos de origem não documentados.

**Ação:** Introspecção do schema via endpoint `/api/admin/shopee-introspect`.

**Resultado:** A API **NÃO expõe** campos de origem. Campos testados e rejeitados (erro 10010):
- `brandName` — "Cannot query field on type ProductOfferV2"
- `sellerLocation` — idem
- `originCountry` — idem
- `productOrigin` — idem

O campo `shopType` (LIST enum) existe mas indica apenas tipo de loja (mall/preferred), não país de origem.

**Nota:** Adicionar `shopType` à query causou erro 502 em produção — removido.

---

### 2. API Pública v4 via Cloud Run

**Hipótese:** O endpoint `shopee.com.br/api/v4/pdp/get_pc` retorna atributos do produto incluindo origem.

**Ação:** Chamada direta do Cloud Run à API pública.

**Resultado:** `status 403` — a Shopee bloqueia IPs de datacenter (GCP/Cloud Run).

---

### 3. Cloudflare Worker como proxy

**Hipótese:** IPs da edge da Cloudflare não são bloqueados.

**Ação:** Rota `/shopee-proxy/pdp` no Worker faz fetch à API pública.

**Resultado:** `status 403` — Cloudflare Workers também são bloqueados.

---

### 4. Cloudflare Worker com User-Agent Googlebot

**Hipótese:** A Shopee serve dados para crawlers (SEO).

**Ação:** Worker faz fetch ao HTML da página com `User-Agent: Googlebot`.

**Resultado:** `status 403` — o bloqueio é por IP, não por User-Agent.

---

### 5. Cloudflare Worker com endpoint alternativo (`/api/v4/item/get`)

**Hipótese:** Endpoint mais antigo pode ter regras diferentes.

**Ação:** Worker tenta `item/get` como fallback do `pdp/get_pc`.

**Resultado:** `status 403` — ambos os endpoints bloqueados.

---

### 6. Proxy residencial (Bright Data)

**Hipótese:** IP residencial brasileiro não é bloqueado.

**Ação:** Configuração de proxy `brd.superproxy.io:33335` no Cloud Run.

**Resultado:** `Auth failed` / `407` — o Bright Data exige verificação de identidade (KYC) para ativar proxies residenciais com geo-targeting. O modo "Immediate access" não funcionou com a Shopee.

**Custo estimado:** ~$3-9/mês (se KYC fosse concluído).

---

## Motivo do bloqueio

A Shopee implementa proteção anti-bot em múltiplas camadas:
- Bloqueio por IP (datacenter, CDN, e proxies não-verificados)
- Assinatura `af-ac-enc-dat` nos requests à API v4
- Páginas carregadas 100% via JavaScript (HTML vazio sem JS)
- Rate limiting agressivo

A única forma de acessar os dados é:
1. Browser real com IP residencial (o browser da Mileny funciona)
2. Proxy residencial verificado (custo mensal + KYC)
3. Serviço de scraping dedicado (Bright Data Scraper API, Apify, etc.)

---

## Solução definitiva implementada

**`origem_padrao` por loja monitorada** — zero dependências externas.

Fluxo:
1. Mileny adiciona loja em `/lojas`
2. Seleciona "🇰🇷 Coreia" no campo "Origem dos produtos"
3. Campo `origem_padrao` salvo na Busca (BigQuery)
4. Motor de coleta aplica `origem_padrao` a todos os produtos da loja
5. CandidateCard exibe badge 🇰🇷/🇯🇵/🇨🇳

Limitação: não diferencia origem por produto individual dentro da mesma loja. Aceitável porque lojas coreanas vendem produtos coreanos (1:1).

---

## Schema GraphQL — Campos disponíveis no ProductOfferV2

Extraídos via introspecção real (referência para desenvolvimento futuro):

| Campo | Tipo |
|-------|------|
| itemId | NON_NULL |
| commissionRate | NON_NULL |
| commission | NON_NULL |
| price | NON_NULL |
| priceMin | NON_NULL |
| priceMax | NON_NULL |
| sales | NON_NULL |
| imageUrl | NON_NULL |
| productName | NON_NULL |
| shopName | NON_NULL |
| shopId | NON_NULL |
| productLink | NON_NULL |
| offerLink | NON_NULL |
| productCatIds | LIST |
| ratingStar | NON_NULL |
| priceDiscountRate | NON_NULL |
| shopType | LIST (enum ShopType) |
| sellerCommissionRate | NON_NULL |
| shopeeCommissionRate | NON_NULL |
| periodStartTime | NON_NULL |
| periodEndTime | NON_NULL |
| appExistRate | NON_NULL |
| appNewRate | NON_NULL |
| webExistRate | NON_NULL |
| webNewRate | NON_NULL |

---

## Alternativa futura

Se a Shopee eventualmente expor o campo de origem na API de afiliados (possível — a integração com ChatGPT sugere abertura de dados), basta:
1. Adicionar o campo à query GraphQL em `shopee.go`
2. Mapear para `domain.Product.Origin`
3. O badge já funciona automaticamente

Monitorar via endpoint Admin > Introspecção periodicamente.
