# Fluxos de Requisição — Diagramas de Sequência

Documentação detalhada de como cada caso de uso se traduz em chamadas entre
componentes, quais dados são armazenados onde, e como o data ownership é respeitado.

**Convenções:**
- 🟢 PostgreSQL (C# API é o dono exclusivo)
- 🔵 BigQuery (Go escreve, Python lê)
- ☁️ APIs externas (Shopee, Amazon, Telegram, WhatsApp)
- 📦 Cloud Tasks (barramento durável entre serviços)

---

## 1. Descobrir Produtos (Curadoria)

O usuário busca por palavras-chave. O sistema consulta a API de afiliados em tempo real,
aplica scoring e retorna candidatos rankeados.

```mermaid
sequenceDiagram
    participant U as 👤 Usuário
    participant F as 🌐 Frontend (SvelteKit)
    participant A as 🔷 API C# (:8080)
    participant C as ⚙️ Collector Go (:50051)
    participant S as ☁️ Shopee API

    U->>F: digita "sérum vitamina c"
    F->>A: GET /api/candidatos?keyword=sérum+vitamina+c&top=20
    A->>C: gRPC Fetch(keyword, limit=50, marketplace=SHOPEE)
    C->>S: POST GraphQL productOfferV2 (HMAC-SHA256 auth)
    S-->>C: JSON {products: [{commissionRate, priceMin, sales, ratingStar, offerLink}]}
    C-->>A: FetchResponse {products[], totalFound, fetchedAt}
    A->>A: ScoringService.Rank(filter: comissão≥7%, vendas≥0)
    A-->>F: {estrategia: "nicho", candidatos: [{id, nome, preco, comissao, score}], total_bruto}
    F-->>U: grade de cards com produtos rankeados
```

<details>
<summary>📋 Detalhes técnicos</summary>

**Data ownership:** Nenhum dado é armazenado neste fluxo. É uma consulta real-time
pura (stateless). O C# API não grava no PostgreSQL nem no BigQuery.

**Scoring:** `Score = 0.45×norm(comissão) + 0.35×norm(EV) + 0.20×norm(rating)`
onde EV = preço × comissão × vendas.

**Autenticação Shopee:** `Signature = SHA256(AppId + timestamp + jsonBody + Secret)`.
Header: `Authorization: SHA256 Credential={AppId}, Timestamp={ts}, Signature={sig}`

**Rate limit:** 200ms throttle entre páginas, 60s entre lojas diferentes.

**Escalabilidade:** O Collector é stateless — pode escalar horizontalmente. Cada
request é independente. Não há cache (dados sempre frescos da API).

</details>

---

## 2. Publicar Oferta (Manual)

O usuário seleciona um produto, escolhe destino e template, e publica para Telegram/WhatsApp.
O link enviado é um link curto de afiliada gerado via Collector (com sub_ids para rastreamento).

```mermaid
sequenceDiagram
    participant U as 👤 Usuário
    participant F as 🌐 Frontend
    participant A as 🔷 API C# (:8080)
    participant PG as 🟢 PostgreSQL
    participant C as ⚙️ Collector (:50051)
    participant S as ☁️ Shopee API
    participant P as 📤 Publisher Go (:50052)
    participant T as ☁️ Telegram Bot API

    U->>F: clica "📤 Publicar" com produto selecionado
    F->>A: POST /api/publicar {nome, preco, link, imagem, destino_id, estrategia}
    A->>PG: SELECT Config FROM Destinos WHERE Id = destino_id
    PG-->>A: Config = "@mileseleciona" (chat_id resolvido)
    A->>C: gRPC GenerateAffiliateLink(url, sub_ids=["mileseleciona","nicho","20260707"])
    C->>S: POST GraphQL generateShortLink(originUrl, subIds)
    S-->>C: {shortLink: "https://s.shopee.com.br/xyz123"}
    C-->>A: GenerateAffiliateLinkResponse {short_link}
    A->>P: gRPC Publish(groupId="@mileseleciona", content={productUrl=short_link})
    P->>T: POST /bot{token}/sendPhoto {chat_id, photo, caption com link curto}
    T-->>P: {ok: true, message_id: 12345}
    P-->>A: PublishResponse {success: true, messageId: "12345"}
    A->>PG: INSERT INTO Publicacoes {status="enviada", link=short_link}
    A-->>F: {success: true, publicacao_id: "uuid", message_id: "12345"}
    F-->>U: ✅ "Publicação enviada!"
```

<details>
<summary>📋 Detalhes técnicos</summary>

**Data ownership:**
- 🟢 PostgreSQL: `Destinos` (leitura do chat_id), `Publicacoes` (escrita do registro)
- O Publisher Go **nunca** acessa o PostgreSQL — recebe o chat_id já resolvido via gRPC
- O Collector Go gera o link de afiliada via Shopee GraphQL (I/O externo)

**Link de afiliada (GenerateAffiliateLink):**
- Recebe a URL original do produto + sub_ids para rastreamento
- Sub IDs: `[canal, estrategia, data]` → voltam no `conversionReport.utmContent`
- Se falhar (Collector indisponível), usa o link original como fallback

**Resolução de destino:** O `destino_id` do frontend é um UUID do PostgreSQL. O C#
resolve para o `Config` real (chat_id ou telefone) antes de chamar o Publisher.

**Fallback:** Se `sendPhoto` falha (CDN Shopee bloqueada), o Publisher tenta
`sendMessage` com texto puro (graceful degradation).

**Escalabilidade:** Publisher é stateless. Telegram limita 30 msg/s global e
1 msg/s por chat. O Cloud Tasks cuida do throttle em cenários automáticos.

</details>

---

## 2.1. Publicar Oferta Agendada

O usuário agenda uma publicação para o futuro. O Scheduler dispara no horário correto.

```mermaid
sequenceDiagram
    participant U as 👤 Usuário
    participant F as 🌐 Frontend
    participant A as 🔷 API C# (:8080)
    participant PG as 🟢 PostgreSQL
    participant SC as ⏱️ Scheduler (:50054)
    participant C as ⚙️ Collector (:50051)
    participant P as 📤 Publisher (:50052)
    participant T as ☁️ Telegram

    U->>F: seleciona data/hora futura e clica "Agendar"
    F->>A: POST /api/publicacoes {nome, link, destino_id, agendada_em: "2026-07-08T14:30:00Z"}
    A->>PG: INSERT Publicacoes {status="agendada", agendada_em}
    A->>SC: gRPC SetSchedule(job_id="pub-{id}", cron="30 14 8 7 *", type="scheduled_publish")
    SC-->>A: SetScheduleResponse {success, job: {status: "active"}}
    A-->>F: {publicacao: {id, status: "agendada"}}
    F-->>U: ⏱️ "Publicação agendada para 08/07 às 14:30"

    Note over SC: Às 14:30 do dia 08/07...
    SC->>A: POST /internal/publish-scheduled {publicacao_id}
    A->>PG: SELECT Publicacao WHERE Id = pubId AND Status = "agendada"
    A->>PG: SELECT Config FROM Destinos WHERE Id = destino_id
    A->>C: gRPC GenerateAffiliateLink(url, sub_ids)
    C-->>A: {short_link}
    A->>P: gRPC Publish(groupId=chat_id, content={productUrl=short_link})
    P->>T: POST sendPhoto
    T-->>P: {ok: true}
    A->>PG: UPDATE Publicacoes SET status="enviada", link=short_link
    Note over SC: Remove job one-shot (cleanup)
```

<details>
<summary>📋 Detalhes técnicos</summary>

**Separação de responsabilidades:**
- C# API: CRUD + resolve dados + gera link + chama Publisher (O QUÊ)
- Scheduler Go: cron one-shot + disparo no horário correto (QUANDO)
- Collector Go: GenerateAffiliateLink (I/O externo com Shopee)
- Publisher Go: envio Telegram/WhatsApp (COMO entregar)

**Cron one-shot:** O SetSchedule recebe `cron="30 14 8 7 *"` (minuto 30, hora 14,
dia 8, mês 7, qualquer dia da semana). O Scheduler executa uma vez e remove o job.

**Eventual consistency:** Se o Scheduler estiver indisponível ao criar, a Publicacao
persiste no PG com status "agendada" — pode ser reconciliada depois.

**Endpoint interno:** `/internal/publish-scheduled` não requer auth (rede interna
Cloud Run). O Scheduler faz HTTP POST com `{publicacao_id}`.

</details>

---

## 3. Coleta Agendada + Alerta de Preço (Automático)

O scheduler dispara coletas periódicas. Após cada coleta, verifica se houve quedas
significativas e notifica o usuário via Telegram.

```mermaid
sequenceDiagram
    participant CR as ⏱️ Cron (Scheduler)
    participant C as ⚙️ Collector (:50051)
    participant S as ☁️ Shopee API
    participant BQ as 🔵 BigQuery
    participant CT as 📦 Cloud Tasks
    participant SH as ⏱️ Scheduler HTTP (:8054)
    participant AN as 🐍 Analyzer (:8060)
    participant P as 📤 Publisher (:50052)
    participant T as ☁️ Telegram

    CR->>C: gRPC Fetch(keyword="loja-920292999", limit=50)
    C->>S: POST GraphQL productOfferV2
    S-->>C: JSON {products[]}
    C->>BQ: INSERT INTO snapshots (coletado_em, keyword, produto_id, nome, preco, ...)
    C-->>CR: FetchResponse {totalFound: 149}
    CR->>CT: CreateTask(queue=price-alerts, payload={keyword, threshold, chat_id})
    Note over CT: Rate: 1 msg/s, retry 5x, dedup keyword+dia
    CT->>SH: POST /process-alert {keyword, threshold, chat_id}
    SH->>AN: GET /quedas?dias=2&threshold=0.15&limit=10
    AN->>BQ: SELECT com window functions (preco_primeiro vs preco_atual)
    BQ-->>AN: [{produto_id, nome, preco_anterior, preco_atual, variacao}]
    AN-->>SH: {quedas: [...], total: 9}
    SH->>SH: formatAlertMessage(keyword, quedas)
    SH->>P: gRPC Publish(channel="telegram", groupId=chatId, content={title=HTML})
    P->>T: POST /bot{token}/sendMessage {chat_id, text, parse_mode=HTML}
    T-->>P: {ok: true}
    P-->>SH: PublishResponse {success: true}
```

<details>
<summary>📋 Detalhes técnicos</summary>

**Data ownership:**
- 🔵 BigQuery: Collector **escreve** snapshots (append-only, particionado por dia)
- 🔵 BigQuery: Analyzer **lê** snapshots (queries analíticas com window functions)
- 🟢 PostgreSQL: **não envolvido** neste fluxo automático
- O Scheduler **orquestra** mas não toca dados diretamente

**Cloud Tasks (barramento durável):**
- Queue: `price-alerts` em `southamerica-east1`
- Rate limit: 1 dispatch/s (Telegram safe)
- Retry: 5 tentativas, backoff 10s→300s
- Deduplicação: task name = `alert-{keyword}-{YYYY-MM-DD}` (1 alerta por keyword por dia)

**Detecção de variações (query BigQuery):**
```sql
FIRST_VALUE(preco) OVER (PARTITION BY produto_id ORDER BY coletado_em ASC) AS preco_primeiro
LAST_VALUE(preco) OVER (PARTITION BY produto_id ORDER BY coletado_em ASC ...) AS preco_atual
SAFE_DIVIDE(preco_atual - preco_primeiro, preco_primeiro) AS variacao
WHERE variacao <= -threshold
```

**Escalabilidade:**
- Collector escala horizontalmente (Cloud Run auto-scale)
- BigQuery escala infinitamente para leitura
- Cloud Tasks controla o rate entre serviços
- Múltiplos tenants = múltiplas tasks na queue, processadas 1/s

</details>

---

## 4. Monitorar Lojas — Novidades e Variações

O usuário navega para `/lojas`, seleciona uma loja monitorada, e vê produtos novos
e variações de preço dos últimos dias.

```mermaid
sequenceDiagram
    participant U as 👤 Usuário
    participant F as 🌐 Frontend
    participant A as 🔷 API C# (:8080)
    participant PG as 🟢 PostgreSQL
    participant AN as 🐍 Analyzer (:8060)
    participant BQ as 🔵 BigQuery

    U->>F: navega para /lojas
    F->>A: GET /api/buscas
    A->>PG: SELECT * FROM Buscas WHERE Active=true AND OwnerUid=@uid
    PG-->>A: [{id, keyword, shop_ids, criado_em}]
    A-->>F: {buscas: [...]}
    F-->>U: lista de lojas monitoradas

    U->>F: seleciona "Glory of Seoul"
    F->>A: GET /api/lojas/novidades?busca_id=loja-920292999&dias=7
    A->>AN: GET /novidades?busca_id=loja-920292999&dias=7
    AN->>BQ: WITH recentes AS (SELECT ... WHERE keyword LIKE '%loja-920292999%' AND coletado_em >= ...)
    BQ-->>AN: rows com aparicoes, preco_primeiro, preco_atual, variacao
    AN-->>A: {produtos_novos: [...], variacoes: [...]}
    A-->>F: JSON (proxy direto, sem transformação)
    F-->>U: aba "📉 Preços" com tabela de variações + aba "🆕 Novidades"
```

<details>
<summary>📋 Detalhes técnicos</summary>

**Data ownership:**
- 🟢 PostgreSQL: `Buscas` — lista de lojas monitoradas (CRUD do C#)
- 🔵 BigQuery: `snapshots` — dados históricos (escrito pelo Collector, lido pelo Analyzer)
- O C# API faz **proxy transparente** para o Analyzer (não transforma dados)

**Classificação de produtos:**
- **Produto novo**: aparece apenas 1× na janela (nunca coletado antes)
- **Variação**: `|preco_atual - preco_primeiro| / preco_primeiro > 1%`

**Não armazena detecções:** As variações são calculadas em runtime via query BQ.
Não há tabela de "detecções" — é sempre recalculado. Isso garante zero dados obsoletos.

**Escalabilidade:** BigQuery aceita centenas de queries concorrentes. 1000 tenants
consultando novidades = 1000 queries BQ paralelas (cada uma ~200ms).

</details>

---

## 5. Adicionar Loja (Resolver Shop ID)

O usuário adiciona uma loja informando URL ou username. O sistema resolve o shop_id
real via Collector gRPC, persiste a busca, e registra um job no Scheduler para coleta periódica.

```mermaid
sequenceDiagram
    participant U as 👤 Usuário
    participant F as 🌐 Frontend
    participant A as 🔷 API C# (:8080)
    participant C as ⚙️ Collector Go (:50051)
    participant S as ☁️ Shopee API (v4 pública)
    participant PG as 🟢 PostgreSQL
    participant SC as ⏱️ Scheduler (:50054)

    U->>F: digita "https://shopee.com.br/belezanaweb_oficial"
    F->>A: POST /api/lojas {input: "https://shopee.com.br/belezanaweb_oficial", keywords?: ["serum"]}
    A->>A: resolve marketplace enum (shopee/amazon/ml)
    A->>C: gRPC ResolveShop(username_or_url, marketplace=SHOPEE)
    C->>C: parse URL → extrai username "belezanaweb_oficial"
    C->>S: GET /api/v4/shop/get_shop_detail?username=belezanaweb_oficial
    S-->>C: {error: 0, data: {shopid: 920292999, name: "Beleza Na Web"}}
    C-->>A: ResolveShopResponse {shop_id: 920292999, shop_name: "Beleza Na Web"}
    A->>A: cria Busca com ShopIds=[920292999], Keywords=["serum"]
    A->>PG: INSERT INTO Buscas {Keyword, ShopIds, Keywords, CronExpression, OwnerUid}
    PG-->>A: persisted
    A->>SC: gRPC SetSchedule(job_id="busca-{Id}", cron="0 */8 * * *", enabled=true, params={shop_id, owner_uid, keywords})
    SC-->>A: SetScheduleResponse {success: true, job: {status: "active", next_run_at}}
    A-->>F: {id: "uuid", keyword: "Beleza Na Web", shop_ids: [920292999], status: "adicionada"}
    F-->>U: ✅ Loja adicionada com sucesso

    Note over SC: A cada 8h, o Scheduler dispara a coleta →
    SC->>C: gRPC FetchShop(shop_id=920292999, limit=50)
    C->>S: POST GraphQL productOfferV2(shopId: 920292999)
    S-->>C: JSON {products[]}
    C-->>SC: FetchShopResponse {products[], totalFound}
    Note over SC: Collector grava snapshots no BigQuery (append-only)
```

<details>
<summary>📋 Detalhes técnicos</summary>

**Data ownership:**
- 🟢 PostgreSQL: `Buscas` — C# API é o dono exclusivo, grava o perfil de monitoramento
- ⏱️ Scheduler: dono dos jobs periódicos — registra/pausa jobs via SetSchedule
- O Collector Go **não acessa o PostgreSQL** — faz apenas I/O externo (Shopee API v4)
- O C# API **não faz scraping direto** — delega ao Collector via gRPC (bounded context)

**Integração com Scheduler:**
- O C# API chama `SetSchedule(enabled=true)` ao criar a busca
- O C# API chama `SetSchedule(enabled=false)` ao deletar a busca
- Se o Scheduler estiver indisponível, a Busca persiste no PG (eventual consistency)
- O Scheduler armazena os params do job (shop_id, owner_uid, keywords) em memória

**Fluxo de coleta periódica (disparado pelo cron do Scheduler):**
1. Cron trigger → `dispatchJob()` → `executeJob()`
2. Se `params["type"] == "shop_collection"` → usa FetchShop(shop_id)
3. Se keywords estão presentes → passa como filtro para Fetch(keyword, shop_id)
4. Collector consulta Shopee → grava snapshots no BigQuery
5. Scheduler enfileira alerta via Cloud Tasks (se configurado)

**Pipeline de detecção (pós-coleta):**
```
Scheduler coleta → BigQuery (snapshots)
                        ↓
Frontend GET /api/lojas/novidades → C# proxy → Analyzer /novidades
                                                    ↓
                                            BigQuery query (window functions)
                                                    ↓
                                            {produtos_novos[], variacoes[]}
```

**Modos de monitoramento:**
- Sem keywords: `FetchShop(shop_id)` — coleta TODOS os produtos da loja
- Com keywords: `Fetch(keyword)` para cada keyword — coleta apenas produtos que matcham

</details>

---

## 6. Coleta de Cupons + Detecção

O scheduler coleta cupons de múltiplos marketplaces e o analyzer detecta novos/modificados/expirados.

```mermaid
sequenceDiagram
    participant CR as ⏱️ Scheduler
    participant C as ⚙️ Collector (:50051)
    participant S as ☁️ Shopee/Amazon/ML
    participant BQ as 🔵 BigQuery
    participant AN as 🐍 Analyzer (:8060)
    participant A as 🔷 API C# (:8080)
    participant PG as 🟢 PostgreSQL

    CR->>C: gRPC FetchCoupons(ownerUid, marketplace=SHOPEE, pageSize=500)
    C->>S: API de cupons do marketplace
    S-->>C: [{coupon_id, discount_type, discount_value, end_time, categories}]
    C->>BQ: INSERT INTO coupon_snapshots (collected_at, coupon_id, marketplace, ...)
    C-->>CR: FetchCouponsResponse {totalFound, fetchedAt}
    CR->>AN: POST /detect-coupons {owner_uid, marketplace, snapshot_timestamp}
    AN->>BQ: diff query (current snapshot vs previous) — UNION ALL newly/modified/expired
    BQ-->>AN: [{coupon_id, detection_status, discount_value}]
    AN->>A: POST /internal/coupon-alerts/evaluate {ownerUid, detections[]}
    A->>PG: SELECT CouponAlertRules WHERE OwnerUid=X AND IsActive=true
    A->>PG: CHECK CouponAlertHistory (dedup 72h)
    A->>PG: INSERT CouponAlertHistory {couponId, ruleId, alertedAt}
    A-->>AN: {alerts_sent: N}
```

<details>
<summary>📋 Detalhes técnicos</summary>

**Data ownership:**
- 🔵 BigQuery: `coupon_snapshots` — append-only, TTL 90 dias (Collector escreve)
- 🔵 BigQuery: Analyzer **lê** para fazer o diff entre snapshots
- 🟢 PostgreSQL: `CouponAlertRules` + `CouponAlertHistory` — regras e dedup (C# API gerencia)

**Cross-boundary communication:**
O Analyzer detecta cupons no BigQuery e **não acessa o PostgreSQL**. Ele faz HTTP POST
para o C# API que é o dono do PG. O dado nunca pula a fronteira.

**Deduplicação:** Um mesmo cupom+regra só gera alerta 1× a cada 72h (CouponAlertHistory).
Se o desconto muda (detection_status="modified"), um novo alerta é permitido.

**Sequencial por marketplace:** Shopee → Amazon → Mercado Livre (3 chamadas sequenciais).
Se um falha, os outros continuam (graceful degradation).

</details>

---

## 7. Publicar a Partir de Variação de Preço

O usuário vê uma queda na aba Preços e clica "📤" para publicar aquela oferta.

```mermaid
sequenceDiagram
    participant U as 👤 Usuário
    participant F as 🌐 Frontend
    participant A as 🔷 API C# (:8080)
    participant PG as 🟢 PostgreSQL
    participant P as 📤 Publisher (:50052)
    participant T as ☁️ Telegram

    U->>F: clica 📤 na linha "CK One R$189→R$151 (-20%)"
    F->>F: prepararPublicacao({id, nome, preco: 151.90})
    F->>F: navega para /publicar (dados pré-preenchidos)
    U->>F: confirma destino e clica "Enviar"
    F->>A: POST /api/publicar {nome: "CK One", preco: 151.90, link, imagem, destino_id}
    A->>PG: SELECT Config FROM Destinos WHERE Id=destino_id
    PG-->>A: "@mileseleciona"
    A->>P: gRPC Publish(groupId="@mileseleciona", content={title, price, imageUrl, productUrl})
    P->>T: POST /bot{token}/sendPhoto
    T-->>P: {ok: true}
    P-->>A: {success: true}
    A->>PG: INSERT Publicacoes {status="enviada"}
    A-->>F: {success: true}
    F-->>U: ✅ "Oferta publicada!"
```

<details>
<summary>📋 Detalhes técnicos</summary>

**Data ownership:**
- 🟢 PostgreSQL: `Destinos` (lê chat_id), `Publicacoes` (grava registro)
- Nenhum dado do BigQuery é acessado neste fluxo

**Dados pré-preenchidos:** O frontend usa `prepararPublicacao()` que serializa
o produto no URL params para a página /publicar. O preço atual (pós-queda) é
o que vai na publicação.

**Sem duplicação de dados:** A variação de preço não é "salva" em nenhum lugar.
O produto vai direto para publicação com o preço atual da Shopee.

</details>

---

## 8. Onboarding (Configuração Multi-Tenant)

Fluxo multi-step de cadastro do tenant. Puramente CRUD no PostgreSQL.

```mermaid
sequenceDiagram
    participant U as 👤 Usuário
    participant F as 🌐 Frontend
    participant A as 🔷 API C# (:8080)
    participant PG as 🟢 PostgreSQL

    U->>F: acessa /configurar
    F->>A: GET /api/onboarding/status
    A->>PG: SELECT TenantConfigs WHERE OwnerUid=@uid
    PG-->>A: {onboardingStep: 0}
    A-->>F: {step: 0, configurado: false}

    U->>F: aceita termos
    F->>A: POST /api/onboarding/termos
    A->>PG: UPDATE TenantConfigs SET AceitouTermos=true, Step=1

    U->>F: preenche AppId + Secret da Shopee
    F->>A: POST /api/onboarding/shopee {appId, secret}
    A->>PG: UPDATE TenantConfigs SET ShopeeAppId=X, ShopeeSecretEnc=Y, Step=2

    U->>F: configura Telegram (token + chat_id)
    F->>A: POST /api/onboarding/telegram {token, chatId}
    A->>PG: UPDATE TenantConfigs SET TelegramTokenEnc=T, TelegramChatId=C, Step=3

    U->>F: valida credenciais
    F->>A: POST /api/onboarding/validar
    A->>PG: UPDATE TenantConfigs SET Step=4, Configurado=true
    A-->>F: {configurado: true}
```

<details>
<summary>📋 Detalhes técnicos</summary>

**Data ownership:** 100% PostgreSQL via C# API. Nenhum sidecar envolvido.

**Multi-tenancy:** Todos os dados são filtrados por `owner_uid` via EF Core
global query filters. Tenant A nunca vê dados de Tenant B.

**Segurança:** Credenciais (secret, tokens) são armazenadas com sufixo `Enc`
indicando que deverão ser encriptadas (T-0045 pendente).

**Escalabilidade:** CRUD simples com PostgreSQL. Escala naturalmente com o banco.

</details>

---

## 9. Dashboard (Estatísticas e Evolução)

A página de estatísticas mostra métricas agregadas, evolução de preço, e contagem de quedas/altas.

```mermaid
sequenceDiagram
    participant U as 👤 Usuário
    participant F as 🌐 Frontend
    participant A as 🔷 API C# (:8080)
    participant PG as 🟢 PostgreSQL
    participant AN as 🐍 Analyzer (:8060)
    participant BQ as 🔵 BigQuery

    U->>F: navega para /estatisticas (dias=7)
    par Chamadas paralelas
        F->>A: GET /api/estatisticas?dias=7
        A->>AN: GET /estatisticas?dias=7
        AN->>BQ: SELECT AVG(preco), AVG(comissao), COUNT(DISTINCT produto_id) ...
        BQ-->>AN: {total_produtos, preco_medio, comissao_media, ...}
        AN-->>A: JSON
        A-->>F: {resumo: {...}}
    and
        F->>A: GET /api/lojas/evolucao?dias=7
        A->>AN: GET /evolucao?dias=7
        AN->>BQ: SELECT DATE(coletado_em), keyword, AVG(preco) ... GROUP BY dia, keyword
        AN->>BQ: COUNTIF(variacao < -0.01) AS total_quedas, COUNTIF(variacao > 0.01) AS total_altas
        BQ-->>AN: {lojas: [...], resumo: {total_quedas, total_altas}}
        AN-->>A: JSON
        A-->>F: {lojas: [...], resumo: {...}}
    and
        F->>A: GET /api/publicacoes?status=
        A->>PG: SELECT * FROM Publicacoes ORDER BY CreatedAt DESC
        PG-->>A: [{status, nome, ...}]
        A-->>F: {publicacoes: [...]}
    end
    F-->>U: Dashboard com MetricCards, MiniCharts, RankList
```

<details>
<summary>📋 Detalhes técnicos</summary>

**Data ownership:**
- 🔵 BigQuery: Analyzer lê snapshots para estatísticas e evolução
- 🟢 PostgreSQL: publicações (C# API lê para contagem e ranking)

**Proxy transparente:** O C# API faz proxy para o Analyzer sem transformar dados.
Se o Analyzer estiver offline, retorna fallback vazio (graceful degradation).

**3 chamadas paralelas:** O frontend dispara as 3 requisições simultaneamente
(`Promise.all`). O dashboard carrega assim que todas respondem.

**Escalabilidade:** Queries BQ são independentes e podem rodar em paralelo.
O C# API é stateless (proxying only). Suporta N tenants simultâneos.

</details>

---

## 10. Resolver Link Shopee

Utilitário para extrair dados de um link curto da Shopee (s.shopee.com.br/xxx).

```mermaid
sequenceDiagram
    participant U as 👤 Usuário
    participant F as 🌐 Frontend
    participant A as 🔷 API C# (:8080)
    participant S as ☁️ Shopee CDN

    U->>F: cola link "https://s.shopee.com.br/4Vatt0MTDy"
    F->>A: POST /api/resolver-link {url: "https://s.shopee.com.br/4Vatt0MTDy"}
    A->>S: HEAD request (seguir redirects)
    S-->>A: Location: https://shopee.com.br/product/920292999/25000641551
    A->>A: extrai shop_id=920292999, item_id=25000641551
    A-->>F: {shop_id, item_id, url_final, imagem, nome, preco}
    F-->>U: preenche formulário com dados do produto
```

<details>
<summary>📋 Detalhes técnicos</summary>

**Data ownership:** Nenhum dado armazenado. Operação stateless.

**Uso:** Na página /publicar, quando o usuário cola um link em vez de selecionar
um produto da curadoria. Resolve o link para obter imagem, nome e preço.

</details>

---

## 11. CRUD de Canais e Templates

Gerenciamento de destinos de publicação e templates de mensagem. Puramente PostgreSQL.

```mermaid
sequenceDiagram
    participant U as 👤 Usuário
    participant F as 🌐 Frontend
    participant A as 🔷 API C# (:8080)
    participant PG as 🟢 PostgreSQL

    U->>F: navega para /canais
    F->>A: GET /api/destinos
    A->>PG: SELECT * FROM Destinos WHERE OwnerUid=@uid AND Ativo=true
    PG-->>A: [{id, nome, tipo, config}]
    A-->>F: {destinos: [{nome: "Mileseleciona", tipo: "telegram", config: "@mileseleciona"}]}
    F-->>U: lista de canais configurados

    U->>F: adiciona novo canal WhatsApp
    F->>A: POST /api/destinos {nome: "Grupo VIP", tipo: "whatsapp", config: "+5511999..."}
    A->>PG: INSERT Destinos {nome, tipo, config, OwnerUid=@uid}
    A-->>F: {id: "uuid", status: "salvo"}
```

<details>
<summary>📋 Detalhes técnicos</summary>

**Data ownership:** 100% PostgreSQL. Nenhum sidecar envolvido.

**Multi-tenancy:** `Destinos`, `Templates` têm `OwnerUid` com global query filter.
Cada tenant vê apenas seus próprios canais.

**O Publisher não conhece Destinos:** O Publisher recebe `groupId` (chat_id/telefone)
já resolvido. Ele não sabe que existem "destinos" no PostgreSQL.

</details>

---

## Resumo: Data Ownership por Caso de Uso

| Caso de Uso | PostgreSQL (C#) | BigQuery (Go→Python) | APIs Externas |
|---|---|---|---|
| 1. Descobrir | — | — | Shopee/Amazon (real-time) |
| 2. Publicar | ✍ Destinos, Publicacoes | — | Telegram/WhatsApp + Shopee (GenerateAffiliateLink) |
| 2.1. Publicar Agendada | ✍ Publicacoes | — | Scheduler (timer) + Telegram/WhatsApp + Shopee |
| 3. Coleta+Alerta | — | ✍ snapshots | Shopee → BQ → Telegram |
| 4. Monitorar Lojas | 📖 Buscas | 📖 snapshots (via Analyzer) | — |
| 5. Adicionar Loja | ✍ Buscas | — | Shopee v4 (via Collector gRPC) + Scheduler SetSchedule |
| 6. Cupons | ✍ AlertRules, AlertHistory | ✍ coupon_snapshots, 📖 diff | Shopee/Amazon/ML |
| 7. Publicar Variação | ✍ Destinos, Publicacoes | — | Telegram |
| 8. Onboarding | ✍ TenantConfigs | — | — |
| 9. Dashboard | 📖 Publicacoes | 📖 snapshots (via Analyzer) | — |
| 10. Resolver Link | — | — | Shopee CDN (redirect) |
| 11. Canais/Templates | ✍ Destinos, Templates | — | — |

**Legenda:** ✍ = escreve, 📖 = lê, — = não envolvido

---

## Fluxos Pendentes de Implementação

| Fluxo | Status | Descrição |
|---|---|---|
| Publicações agendadas | ✅ Implementado | Usuário agenda para data futura → C# persiste + Scheduler.SetSchedule(one-shot cron) → Scheduler dispara POST /internal/publish-scheduled → C# gera link de afiliada + envia via Publisher. |
| Coleta por shop_id no Scheduler | ✅ Implementado | O Scheduler executa `Collector.FetchShop(shop_id)` para jobs do tipo `shop_collection`. Jobs com keywords fazem `Fetch(keyword)` para cada keyword individualmente. |
| Alertas de cupons → Telegram | ⬜ Parcial (T-0045) | Detecção funciona mas envio para Telegram do tenant ainda não wired. |

---

## Princípios de Escalabilidade

| Princípio | Como está implementado |
|---|---|
| **Stateless services** | Todos os serviços são stateless — estado fica no PG/BQ |
| **Horizontal scaling** | Cloud Run auto-scale 0→N para cada container |
| **Rate limiting externo** | Cloud Tasks controla throughput entre serviços |
| **Event-driven alerts** | Coleta → task → processamento assíncrono |
| **Query isolation** | Cada query BQ é independente (sem locks, sem transações) |
| **Graceful degradation** | Analyzer offline → fallback vazio; Publisher fail → retry |
| **Deduplication** | Cloud Tasks (keyword+dia), CouponAlertHistory (72h window) |
| **Data locality** | PostgreSQL para CRUD, BigQuery para analytics (each at what it does best) |
