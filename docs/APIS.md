# APIs — Shopee Afiliados e Instagram (referência)

Levantamento feito em junho/2026. APIs mudam; confira sempre as fontes oficiais
(links no fim). Os números de versão e limites abaixo podem ter mudado.

---

## 1. Shopee — API de Afiliados (Open API)

**É uma API GraphQL**, diferente da Open Platform de sellers. Endpoint Brasil:

```
POST https://open-api.affiliate.shopee.com.br/graphql
Content-Type: application/json
```

### 1.1 Acesso e autenticação

Credenciais (App ID + Secret) são solicitadas no painel de afiliado
(`affiliate.shopee.com.br/open_api` → Central de Ajuda → formulário). Aprovação
manual, normalmente alguns dias a ~2 semanas. Você já tem acesso, então é só
pegar App ID e Secret.

Toda requisição vai assinada:

```
Authorization: SHA256 Credential={AppId}, Timestamp={ts}, Signature={sig}
Signature = SHA256(AppId + Timestamp + Payload + Secret)
```

- `Timestamp`: Unix em **segundos** (não ms); janela de ~5 min.
- `Payload`: o corpo JSON **exato** enviado. Assine os mesmos bytes que vão no
  body — divergência de whitespace dá erro 10020 (Invalid Signature). O adaptador
  Go faz exatamente isso: marshala uma vez, assina e envia o mesmo `[]byte`.

### 1.2 Endpoints que importam para o Garimpo

**`productOfferV2`** — o mais usado. Busca produtos com comissão detalhada.
Parâmetros úteis: `keyword`, `productCatId` (categoria nível 1/2/3 — é como você
puxa o nicho), `shopId`, `itemId`, `listType` (0=Recomendados, 1=Maior comissão,
2=Top performance), `sortType` (1=Relevância, 2=Vendidos, 3=Maior preço,
4=Menor preço, 5=Comissão), `isKeySeller`, `page`, `limit` (1–500).

> ⚠️ A semântica de `listType` aparece diferente entre fontes/versões (algumas
> listam 3=landing category, 4=detail category, 5=detail shop). Confirme no seu
> painel antes de fixar.

Campos retornados (mapeamento no adaptador):

| Campo da API | Vira no domínio | Observação |
|---|---|---|
| `commissionRate` | `Commission` | já é fração ("0.0850" = 8,5%); soma Shopee + seller |
| `priceMin` | `Price` | há também `priceMax`, `priceDiscountRate` |
| `sales` | `Sales30d` | proxy de demanda — é o volume reportado, **não** janela fixa de 30d |
| `ratingStar` | `Rating` | avaliação 0–5 |
| `offerLink` | `Link` | já vem com seu tracking |
| `productName`, `shopName` | `Name` | categoria não vem no nó: carimbamos via `productCatId` consultado |

Isso é o ponto central: **demanda (`sales`) e avaliação (`ratingStar`) vêm de
graça**, então o scoring não depende de proxy inventado.

**`generateShortLink`** (mutation) — transforma uma URL Shopee em link curto com
até **5 `subIds`** de tracking. Os subIds voltam no `conversionReport` (campo
`utmContent`). Ex.: `subIds: ["instagram", "stories", "{data}", "{itemId}"]`.

**`conversionReport`** — relatório de conversões (pedidos via seus links):
`purchaseTime`, `clickTime`, `conversionId`, `totalCommission`, `utmContent`,
e `orders { items { itemId itemName itemTotalCommission ... } }`. Status:
UNPAID → PENDING → COMPLETED → CANCELLED. Paginação por `scrollId` (vale 30s).

**`validatedReport`** — conversões já validadas, com valor **final** de comissão
(inclui `refundAmount`). É o que conferir antes do pagamento.

**`shopOfferV2`** — lojas com comissão diferenciada (Key Sellers, Mall/Star/Star+).

### 1.3 O que isso destrava

A atribuição fecha **dentro da própria Shopee**: `generateShortLink(subIds)` na
saída + `conversionReport(utmContent)` na volta. Ou seja, para os links Shopee
você **não precisa** de um encurtador próprio só para saber o que converteu —
basta padronizar os subIds (ex.: incluir a estratégia e a data) e cruzar com o
relatório. O encurtador caseiro continua útil para (a) redes que não são Shopee,
e (b) volume bruto de cliques (o `conversionReport` só mostra pedidos, não todo
clique).

### Códigos de erro comuns

`10020` Invalid Signature · `10030` Rate Limit · `10035` No API Access ·
`11001` Params Error · `10010` Parse Error.

---

## 2. Instagram — Graph API

Em 2026, **só existe a Instagram Graph API** (a Basic Display foi desligada em
dez/2024). Chamadas vão para `graph.facebook.com` (versão ~v21.0 em meados de
2026, versionada a cada trimestre). Exige conta **Profissional (Business ou
Creator)** vinculada, autenticada via Instagram Business Login ou Facebook Login
for Business. Limite de ~200 chamadas/hora por conta.

### 2.1 O que dá para fazer

- **Publicar** imagem, carrossel (até 10), Reels e Stories — via URL pública
  (sem upload direto de arquivo); fluxo de 2 passos: cria o container → publica.
- **Insights de conteúdo**: alcance, impressões, e por mídia likes/comentários/
  salvamentos/compartilhamentos/visualizações (métricas ampliadas em 2026).
- **Insights de conta**: alcance, impressões, visitas ao perfil, e demografia
  da audiência (com `instagram_manage_insights`, só da própria conta).
- **Comentários/DMs**: moderar, responder. **Rótulo de parceria** já pode ser
  aplicado via API (novidade 2026) — útil se houver publi.

### 2.2 Os limites que moldam a arquitetura

Estes três importam direto para a operação dela:

1. **Sem link clicável em post de feed.** É limitação da plataforma, não da API.
   Link clicável só na bio, no link-in-bio, ou no sticker de link em Stories.
2. **A API não adiciona sticker de link em Stories** nem elementos interativos
   (enquetes, stickers). Então a colocação do link de afiliado **continua
   manual** (bio / página link-in-bio / sticker feito à mão).
3. **A API não expõe clique no link de afiliado.** Os Insights medem
   alcance/engajamento do conteúdo, não o clique de saída para a Shopee.

**Conclusão de design:** o Instagram entra como **sinal de performance de
conteúdo** (alcance/engajamento por post), que você correlaciona com qual produto
e qual estratégia foram publicados. Mas o **clique e a conversão vêm da Shopee**
(subIds → conversionReport), e a **colocação do link permanece manual**. É por
isso que o encurtador próprio + a página link-in-bio seguem valendo a pena: são
onde você captura o clique que o Instagram não te dá.

---

## Fontes oficiais

- Shopee Afiliados — `https://affiliate.shopee.com.br/open_api`
- Instagram Graph API — `https://developers.facebook.com/docs/instagram-platform`
- Content Publishing — `https://developers.facebook.com/docs/instagram-platform/content-publishing`
