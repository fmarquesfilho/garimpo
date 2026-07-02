# Próxima Sessão — Itens Prioritários

## Contexto

A migração arquitetural (ADR-0012) está completa. O C# serve toda a API, os
microserviços Go estão rodando como sidecars, o frontend está no Cloudflare Pages.
O monólito Go foi decomissionado.

---

## Status: Endpoints Portados ✅ (sessão anterior)

Todos os endpoints que o frontend (`web/src/lib/api.js`) consome foram portados
para o C# na camada de compatibilidade (`/api/*`):

| Endpoint | Arquivo | Status |
|----------|---------|--------|
| `/api/candidatos` | CompatEndpoints.cs | ✅ já existia |
| `/api/admin/me` | CompatEndpoints.cs | ✅ já existia |
| `/api/health` | CompatEndpoints.cs | ✅ já existia |
| `/api/buscas` (GET/POST) | BuscasCompatEndpoints.cs | ✅ portado |
| `/api/lojas` (GET/POST/DELETE) | LojasCompatEndpoints.cs | ✅ portado |
| `/api/lojas/novidades` | LojasCompatEndpoints.cs | ✅ portado (proxy → analyzer) |
| `/api/lojas/evolucao` | LojasCompatEndpoints.cs | ✅ portado (proxy → analyzer) |
| `/api/favoritos` (GET/POST/DELETE) | FavoritosEndpoints.cs | ✅ portado |
| `/api/destinos` (GET/POST/DELETE) | DestinosEndpoints.cs | ✅ portado |
| `/api/templates` (GET/POST/DELETE) | TemplatesEndpoints.cs | ✅ portado |
| `/api/templates/preview` | TemplatesEndpoints.cs | ✅ portado |
| `/api/publicar` (POST) | PublicacoesEndpoints.cs | ✅ portado |
| `/api/publicacoes` (GET/POST) | PublicacoesEndpoints.cs | ✅ portado |
| `/api/alertas` (GET) | AlertasEndpoints.cs | ✅ portado |
| `/api/alertas/testar` (POST) | AlertasEndpoints.cs | ✅ portado (stub) |
| `/api/alertas/configurar` (POST) | AlertasEndpoints.cs | ✅ portado |
| `/api/onboarding/status` | OnboardingEndpoints.cs | ✅ portado |
| `/api/onboarding/termos` | OnboardingEndpoints.cs | ✅ portado |
| `/api/onboarding/shopee` | OnboardingEndpoints.cs | ✅ portado |
| `/api/onboarding/telegram` | OnboardingEndpoints.cs | ✅ portado |
| `/api/onboarding/validar` | OnboardingEndpoints.cs | ✅ portado |
| `/api/onboarding/excluir-conta` | OnboardingEndpoints.cs | ✅ portado |
| `/api/conversoes` | AnalyticsEndpoints.cs | ✅ portado |
| `/api/conversoes/reais` | AnalyticsEndpoints.cs | ✅ portado (proxy → analyzer) |
| `/api/estatisticas` | AnalyticsEndpoints.cs | ✅ portado (proxy → analyzer) |
| `/api/coletas` | AnalyticsEndpoints.cs | ✅ portado (proxy → analyzer) |
| `/api/resolver-link` | ResolverLinkEndpoints.cs | ✅ portado |

### Entidades adicionadas ao domínio/PostgreSQL:
- `TenantConfig` — credenciais + onboarding + alertas
- `Publicacao` — publicações agendadas/enviadas
- `Favorito` — produtos favoritos
- `Template` — templates de mensagem
- `Destino` — canais de publicação

### Rotas do Analyzer Python adicionadas:
- `/coletas` — histórico de coletas (BigQuery)
- `/conversoes` — conversões reais da Shopee (BigQuery)

### Migration EF Core criada:
- `AddPortedEntities` — adiciona as novas tabelas ao PostgreSQL

---

## 1. Bug grave: detecção de variações de preço não funciona

**Sintoma:** O monólito Go executou milhares de coletas mas nunca detectou nenhuma
variação de preço (quedas/altas). Os alertas nunca dispararam.

**Impacto:** A funcionalidade core de "Quedas" na página publicar sempre retorna
vazio. Alertas de preço (T-0005) nunca funcionaram.

**Hipóteses a investigar:**
1. Os snapshots estão gravando o mesmo produto com IDs diferentes a cada coleta (sem match para comparar)
2. O campo `preco` não é populado corretamente (sempre 0 ou sempre o mesmo valor)
3. A query de novidades/variações tem um bug no JOIN (nunca encontra o mesmo produto_id em dias diferentes)
4. O throttling/rotação faz com que nunca colete a mesma loja duas vezes no período da janela

**Ação proposta:**
- Investigar os dados no BigQuery (`snapshots` table) — existem produto_ids repetidos em datas diferentes?
- Se os dados estão corrompidos/inutilizáveis → **reset do BigQuery** (truncar tabelas de snapshots)
- Reimplementar a coleta no novo analyzer Python com validação de que o mesmo produto aparece em múltiplos snapshots

---

## 2. Reset do BigQuery

**Decisão:** Truncar as tabelas de snapshots no BigQuery e recomeçar coletas do zero com a nova arquitetura.

**Justificativa:**
- Dados coletados pelo monólito Go nunca produziram variações úteis
- O bug pode estar nos dados em si (IDs inconsistentes, preços faltando)
- Começar limpo permite validar que o pipeline novo (scheduler → collector → snapshots) funciona end-to-end

**Script (a executar no início da sessão):**
```sql
-- BigQuery: truncar snapshots para recomeçar
TRUNCATE TABLE `garimpo-500114.garimpei.snapshots`;
-- Manter conversoes (dados reais de vendas da Shopee)
-- Manter eventos (histórico, baixo volume)
```

---

## 3. Itens pendentes (próxima sessão)

### Prioridade Alta
- [ ] Investigar bug de variações (query BigQuery)
- [ ] Reset BigQuery (truncar snapshots)
- [ ] Aplicar migration `AddPortedEntities` no PostgreSQL de produção
- [ ] Deploy do C# API com os novos endpoints
- [ ] Configurar coleta no scheduler (crons para popular snapshots limpos)
- [ ] Validar pipeline de variações (coleta → snapshot → analyzer → quedas funciona)

### Prioridade Média
- [ ] Implementar envio real de alerta de teste via Telegram Bot API (AlertasEndpoints — atualmente stub)
- [ ] Implementar encriptação de credenciais no onboarding (ShopeeSecret, TelegramToken)
- [ ] T-0024: Testar WhatsApp Meta Cloud API
- [ ] T-0005: Alertas automáticos (disparar quando queda detectada)

### Prioridade Baixa
- [ ] Validar chamada real Shopee no onboarding/validar
- [ ] T-0007: Recomendação IA personalizada

---

## 4. Tarefas pendentes no backlog

| Task | Título | Prioridade |
|------|--------|-----------|
| T-0024 | Testar WhatsApp Meta Cloud API | Alta |
| T-0026 | Portar endpoints restantes | ✅ Concluído |
| T-0005 | Alertas configuráveis por usuário | Média |
| T-0002 | Persistir conversões no BigQuery | Média |
| T-0007 | Recomendação IA personalizada | Backlog |

---

## Ordem sugerida para a próxima sessão

1. **Aplicar migration** (PostgreSQL produção) + deploy
2. **Reset BigQuery** (truncar snapshots)
3. **Configurar coleta no scheduler** (recriar crons)
4. **Validar pipeline de variações** (coleta → snapshot → analyzer → quedas)
5. **Implementar alertas Telegram reais** (bot API)
6. **T-0024** (testar WhatsApp se sobrar tempo)
