# Próxima Sessão — Planejamento (Sprint 2026-S27)

## Status atual (pós-sessão 2026-07-01/02)

### Concluído ✅
- [x] T-0026: Endpoints portados para C# (todos que o frontend usa)
- [x] BigQuery resetado (8 tabelas truncadas, 0 linhas)
- [x] PostgreSQL produção: migration aplicada (9 tabelas, 0 linhas)
- [x] Scheduler: in-memory, sem state persistido (reinicia vazio)
- [x] CI: 3 scripts de drift (api-contract, config-consistency, schema-sync)
- [x] CI: 13 fitness functions (NetArchTest) validando Clean Architecture
- [x] Docs: arquitetura e qualidade atualizados
- [x] READMEs: root, src, web atualizados
- [x] Dataset BigQuery corrigido em todo codebase (garimpei → garimpo)

### Bloqueios para fase de testes
1. **Publisher usa tokens globais** — não lê tokens do tenant (T-0027)
2. **Scheduler sem crons** — após reset, precisa reconfigurar coletas (T-0028)
3. **Deploy não feito** — imagem com endpoints novos ainda não está em produção (T-0029)

---

## Sprint S27 — Tasks

| Task | Título | Prioridade | Estimativa |
|------|--------|-----------|------------|
| T-0027 | Publisher multi-tenant: tokens do tenant | Alta | M |
| T-0028 | Configurar coleta no scheduler | Alta | M |
| T-0029 | Deploy nova API em produção | Alta | P |

---

## T-0027: Publisher multi-tenant tokens

### Problema
O publisher Go lê `TELEGRAM_BOT_TOKEN` e `TELEGRAM_CHAT_ID` de env vars.
Cada tenant configura seus tokens via onboarding (step 3), mas esses tokens
ficam no PostgreSQL (`TenantConfig`) e nunca chegam ao publisher.

### Solução proposta

1. **Expandir o proto `publisher.v1.PublishRequest`:**
   ```protobuf
   message PublishRequest {
     string channel = 1;
     string group_id = 2;
     PublishContent content = 3;
     // NEW: per-tenant credentials (optional, overrides env vars)
     string bot_token = 4;
     string chat_id = 5;
   }
   ```

2. **C# API ao publicar:** ler `TenantConfig` do tenant e passar `bot_token` + `chat_id` no request gRPC

3. **Publisher Go:** se `bot_token` vem no request → cria sender efêmero com esse token; se vazio → fallback para env var

4. **Mesma abordagem para WhatsApp (Meta Cloud API):**
   ```protobuf
   string whatsapp_token = 6;       // Meta access token
   string phone_number_id = 7;      // Meta phone number ID
   ```

### Frontend
- Página `/canais` já permite cadastrar destinos (nome, tipo, config)
- Adicionar campo de token no formulário quando tipo=telegram
- Para WhatsApp: campo de phone_number_id e access_token Meta

---

## T-0028: Configurar coleta no scheduler

### Problema
BigQuery está vazio. Precisa popular snapshots com coletas regulares para
que variações de preço sejam detectáveis (mesmo produto_id em 2+ dias).

### Plano
1. Usuário cria buscas via interface (/lojas ou /configurar)
2. C# API chama `scheduler.SetSchedule` via gRPC para cada busca com cron
3. Scheduler executa coleta no horário → collector busca → grava snapshot no BigQuery
4. Após 2+ ciclos, analyzer detecta variações

### Validação
```sql
-- Rodar após 2 dias de coleta:
SELECT produto_id, COUNT(*) AS aparicoes
FROM `garimpo-500114.garimpo.snapshots`
GROUP BY produto_id
HAVING aparicoes > 1
LIMIT 10;
-- Se retorna linhas → pipeline funciona
```

---

## T-0029: Deploy em produção

1. Build imagem: `docker build -f src/Garimpei.Api/Dockerfile src/`
2. Push para Artifact Registry
3. `gcloud run services replace deploy/cloud-run-deploy-now.yaml`
4. Smoke test: `curl https://garimpei.app.br/api/health`

---

## Ordem de execução

1. **T-0027** — Expandir proto + publisher + C# API (tokens multi-tenant)
2. **T-0029** — Deploy em produção
3. **T-0028** — Configurar coletas, validar pipeline

---

## Melhorias identificadas (pós-testes)

- [ ] Encriptação real de tokens no PostgreSQL (atualmente plain text nos campos `*Enc`)
- [ ] Validação real de credenciais Shopee no onboarding/validar (atualmente stub)
- [ ] Telegram Bot API: envio real de alertas de teste
- [ ] WhatsApp Meta Cloud API: envio real (T-0024)
- [ ] T-0005: Alertas automáticos (scheduler → alerter → Telegram quando queda detectada)
