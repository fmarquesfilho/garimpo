# Próxima Sessão — Ativar Cloud Tasks + Sync Scheduler

**Data:** 2026-07-05
**Prioridade:** Alta

---

## Objetivo

1. Criar a queue `price-alerts` no Cloud Tasks (setup único)
2. Deploy com as novas env vars no scheduler
3. Testar alertas end-to-end em produção (`./scripts/test-alerts.sh prod`)
4. Implementar auto-sync: buscas ativas no PG → jobs no scheduler

---

## Contexto (sessão 2026-07-04)

### Fluxo de variação de preços — funcionando ✅

```
Scheduler (cron) → Collector → BigQuery snapshots
                                    │
Frontend ← API C# ← Analyzer ←─────┘ (novidades, quedas, evolução)
```

**Produção verificada**: 1576+ snapshots, 5 variações detectadas, 2 quedas >20%.

### Alertas automáticos — implementado, pendente deploy ⬜

```
Scheduler → Cloud Tasks → C# /internal/alerts/check → Analyzer → Publisher → Telegram
```

**Testado localmente**: alerta de preço enviado com sucesso no Telegram (2 quedas formatadas em HTML).

### Alerter removido — 4 containers agora

Arquitetura simplificada de 6 → 4+1 containers:
- garimpei-api (C#) — ingress
- collector (Go) — coleta Shopee
- publisher (Go) — entrega Telegram/WhatsApp
- scheduler (Go) — cron, orquestração
- analyzer (Python) — analytics BigQuery

---

## O que falta para alertas em produção

### 1. Criar queue Cloud Tasks (uma vez)
```bash
./deploy/setup-cloud-tasks.sh
```

### 2. Deploy
O push da sessão atual inclui as env vars:
- `ALERT_TARGET_URL`, `ALERT_QUEUE_ID`, `ALERT_SA_EMAIL` no scheduler
- A queue precisa existir antes do deploy

### 3. Testar em produção
```bash
./scripts/test-alerts.sh prod loja-920292999
```

### 4. Sync buscas → scheduler (T-0028)
O scheduler tem jobs mas não se atualiza quando o usuário adiciona/remove lojas.

---

## Resultados da sessão 2026-07-04

### Bug crítico corrigido
- Analyzer crashava: coluna `em` → corrigido para `coletado_em`
- Deploy realizado, endpoint retorna dados reais

### Features implementadas
- Mock mode no analyzer (`MOCK_DATA=true`)
- `/evolucao` com `resumo.total_quedas/altas` para dashboard
- `/estatisticas` com `total_amostras`
- Cloud Tasks integration para alertas (Go + C#)
- Endpoint `POST /internal/alerts/check` (analyzer → publisher)
- Script `test-alerts.sh` (3 camadas: local, telegram, prod)

### Código morto removido (-507 linhas)
- `services/alerter/` — substituído por Cloud Tasks + publisher
- `internal/alerts/` — só era usado pelo alerter
- `garimpei-api-legacy` no docker-compose — monólito já não existe
- Referências ao alerter no CI e deploy YAMLs

### Testes
- 12 E2E Playwright (lojas-precos.spec.js) — todos passando
- 61 testes C# — todos passando
- Alerta testado localmente com envio real no Telegram ✅

### ADRs
- ADR-0023: Alertas via Cloud Tasks + Publisher (eliminar alerter)

### Tasks
- T-0044: Testar fluxo variação de preços (done)
