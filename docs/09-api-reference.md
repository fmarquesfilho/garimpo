# API Reference

Portal unificado de referência de todos os serviços do Garimpei.

## Endpoints HTTP (C# API — :8080)

Autenticação: Bearer token (Firebase JWT). Exceto `/api/health` e `/internal/*`.

### Core

| Método | Rota | Descrição |
|--------|------|-----------|
| GET | `/api/health` | Health check (público) |
| GET | `/api/admin/me` | Verifica admin + links de ferramentas |
| GET | `/api/candidatos` | Busca + scoring de produtos (4 fontes) |
| GET | `/api/categorias` | Categorias por marketplace |

### Lojas & Buscas Agendadas

| Método | Rota | Descrição |
|--------|------|-----------|
| GET | `/api/lojas` | Listar lojas monitoradas (shop_ids, keywords, cron) |
| POST | `/api/lojas` | Adicionar loja (ResolveShop + SetSchedule) |
| DELETE | `/api/lojas?id={uuid}` | Remover loja (soft-delete + pausa Scheduler) |
| GET | `/api/lojas/novidades?busca_id=X&dias=7` | Produtos novos + variações (proxy Analyzer) |
| GET | `/api/lojas/evolucao?dias=30` | Série temporal de preço (proxy Analyzer) |
| GET | `/api/buscas` | Listar perfis de busca (keywords, shop_ids) |
| POST | `/api/buscas` | Criar/atualizar busca por keyword |

### Publicação

| Método | Rota | Descrição |
|--------|------|-----------|
| POST | `/api/publicar` | Publicar imediatamente (GenerateAffiliateLink + Publisher gRPC) |
| GET | `/api/publicacoes` | Listar publicações (status: pendente/agendada/enviada/erro) |
| POST | `/api/publicacoes` | Agendar ou enviar publicação (SetSchedule se agendada) |
| POST | `/api/resolver-link` | Resolver link curto Shopee → dados do produto |

### Favoritos, Destinos, Templates

| Método | Rota | Descrição |
|--------|------|-----------|
| GET/POST/DELETE | `/api/favoritos` | CRUD de produtos favoritos |
| GET/POST/DELETE | `/api/destinos` | CRUD de canais (Telegram/WhatsApp) |
| GET/POST/DELETE | `/api/templates` | CRUD de templates de mensagem |
| POST | `/api/templates/preview` | Preview de template com dados reais |

### Alertas

| Método | Rota | Descrição |
|--------|------|-----------|
| GET | `/api/alertas` | Configuração atual (habilitado, threshold, chat_id) |
| POST | `/api/alertas/configurar` | Atualizar threshold e filtros |
| POST | `/api/alertas/testar` | Enviar alerta de teste |

### Onboarding

| Método | Rota | Descrição |
|--------|------|-----------|
| GET | `/api/onboarding/status` | Status atual do multi-step |
| POST | `/api/onboarding/termos` | Step 1: aceitar termos |
| POST | `/api/onboarding/shopee` | Step 2: credenciais Shopee |
| POST | `/api/onboarding/telegram` | Step 3a: bot Telegram |
| POST | `/api/onboarding/whatsapp` | Step 3b: WhatsApp Meta |
| POST | `/api/onboarding/validar` | Step 4: validar credenciais |
| POST | `/api/onboarding/excluir-conta` | LGPD: excluir dados |

### Analytics (proxy Analyzer Python)

| Método | Rota | Descrição |
|--------|------|-----------|
| GET | `/api/estatisticas?dias=30` | Dashboard (resumo por categoria) |
| GET | `/api/coletas?dias=30` | Histórico de coletas executadas |
| GET | `/api/conversoes/reais?dias=30` | Conversões Shopee (conversionReport) |

### Endpoints Internos (sem auth — rede interna)

| Método | Rota | Descrição |
|--------|------|-----------|
| POST | `/internal/publish-scheduled` | Callback do Scheduler para publicação agendada |
| POST | `/internal/coupon-alerts/evaluate` | Callback do Analyzer para alertas de cupons |
| POST | `/process-alert` | Proxy para Scheduler HTTP (Cloud Tasks) |

---

## Serviços gRPC (Go sidecars)

Comunicação interna via localhost (Cloud Run multi-container). Stubs pré-gerados em `gen/go/` e `src/Garimpei.Protos/`.

### Collector (`:50051`) — `collector.v1.CollectorService`

| RPC | Request | Response | Descrição |
|-----|---------|----------|-----------|
| `ResolveShop` | username_or_url, marketplace | shop_id, shop_name | Resolve URL/username → shop_id (API pública Shopee v4) |
| `GenerateAffiliateLink` | original_url, sub_ids[], marketplace | short_link, long_link | Gera link curto de afiliada (Shopee generateShortLink) |
| `Fetch` | keyword, limit, marketplace | products[], total_found | Busca produtos por keyword (GraphQL productOfferV2) |
| `FetchShop` | shop_id, limit, marketplace | products[], total_found | Busca todos os produtos de uma loja |

### Publisher (`:50052`) — `publisher.v1.PublisherService`

| RPC | Request | Response | Descrição |
|-----|---------|----------|-----------|
| `Publish` | channel, group_id, content | success, message_id | Envia mensagem para Telegram/WhatsApp |
| `ListGroups` | channel | groups[] | Lista destinos configurados |

### Scheduler (`:50054`) — `scheduler.v1.SchedulerService`

| RPC | Request | Response | Descrição |
|-----|---------|----------|-----------|
| `SetSchedule` | job_id, cron_expression, enabled, params | success, job | Criar/atualizar/pausar job de coleta |
| `ListJobs` | status_filter | jobs[] | Listar jobs registrados |
| `TriggerJob` | job_id, params | accepted, execution_id | Executar job manualmente |

### Analyzer (`:8060`) — Python FastAPI (REST, não gRPC)

| Método | Rota | Descrição |
|--------|------|-----------|
| GET | `/novidades?busca_id=X&dias=7` | Produtos novos + variações de preço |
| GET | `/quedas?dias=7&threshold=0.15` | Quedas de preço (para alertas) |
| GET | `/evolucao?dias=30` | Série temporal por loja |
| GET | `/estatisticas?dias=30` | Resumo por categoria |
| GET | `/coletas?dias=30` | Histórico de coletas |
| GET | `/conversoes?dias=30` | Conversões reais Shopee |
| POST | `/detect-coupons` | Detecção de cupons novos/modificados |
| GET | `/health` | Health check |

---

## Frontend (SvelteKit — Cloudflare Pages)

| Rota | Página | Descrição |
|------|--------|-----------|
| `/` | Home/Descobrir | Curadoria + 4 fontes (busca, quedas, novos, favoritos) |
| `/lojas` | Buscas Agendadas | Monitorar lojas + novidades + variações |
| `/publicar` | Publicar | Enviar/agendar oferta para canais |
| `/publicacoes` | Histórico | Lista de publicações enviadas/agendadas |
| `/canais` | Destinos | Configurar Telegram/WhatsApp |
| `/estatisticas` | Dashboard | Métricas e evolução de preço |
| `/coletas` | Coletas | Histórico de execuções do Scheduler |
| `/configurar` | Onboarding | Multi-step de configuração |
| `/admin` | Admin | Painel administrativo |

---

## Contratos de Serviço

Definidos em `contracts/registry.yaml` (ADR-0020). Validados automaticamente no CI.

- **27 fronteiras** documentadas (HTTP + gRPC)
- **8 JSON Schemas** para payloads críticos
- **7 métodos gRPC** nos protos
- Validação: `mise run check:service-contracts`

Proto files: `protos/collector/v1/collector.proto`, `protos/publisher/v1/publisher.proto`, `protos/scheduler/v1/scheduler.proto`
