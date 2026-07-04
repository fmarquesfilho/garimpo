# Próxima Sessão — Sync Scheduler + Onboarding

**Data:** 2026-07-05
**Prioridade:** Alta

---

## Objetivo

1. Auto-sync: buscas ativas no PG → cron jobs no scheduler (T-0028)
2. Onboarding: usuário adiciona loja → coleta começa automaticamente
3. Publicação a partir de variação (botão 📤 → Telegram)

---

## Estado atual (sessão 2026-07-04)

### Arquitetura (4 sidecars + Cloud Tasks)

```
Cloud Run: garimpei-api (C#) + collector (Go) + publisher (Go) + scheduler (Go) + analyzer (Python)
           2.5 vCPU, 1280Mi RAM | Scale 0→2

Cloud Tasks: queue price-alerts (1 msg/s, retry 5x, dedup por keyword+dia)
```

### Fluxo de alertas (validado em produção ✅)

```
Scheduler (pós-coleta) → Cloud Tasks (price-alerts)
  → C# /process-alert (proxy) → Scheduler HTTP :8054
  → Analyzer GET /quedas (BigQuery) → 9 drops
  → Publisher gRPC → Telegram → canal "Super Ofertas"
```

### Papéis (ADR-0023)

| Componente | Responsabilidade |
|---|---|
| Scheduler (Go) | Orquestração: quando coletar, quando alertar |
| C# API | CRUD + auth + proxy frontend + passthrough Cloud Tasks |
| Analyzer (Python) | Dados: queries BQ, detecção variações |
| Publisher (Go) | Entrega: Telegram, WhatsApp |
| Cloud Tasks | Barramento: rate limit, retry, durabilidade |

### O que foi feito nesta sessão

- ✅ Fix analyzer (`em` → `coletado_em`)
- ✅ Cloud Tasks queue `price-alerts` criada
- ✅ Alerter eliminado (-507 linhas)
- ✅ Orquestração migrada de C# para scheduler Go
- ✅ 12 testes E2E (Playwright)
- ✅ Documentação atualizada (02-arquitetura, ADR-0023, data ownership)
- ✅ Testes em produção: 2 lojas, 9 quedas detectadas, Telegram enviado

---

## Pendente

### T-0028: Auto-sync buscas → scheduler
Quando o usuário cadastra/remove loja no frontend, o scheduler precisa
criar/remover o cron job correspondente. Opções:
- C# chama scheduler via gRPC `SetSchedule` no POST/DELETE /api/lojas
- Scheduler consulta C# periodicamente (polling) para sincronizar

### Publicação manual a partir de variação
O botão 📤 na aba Preços navega para /publicar com dados.
Testar o fluxo completo: variação → publicar → publisher → Telegram.

### Segurança
- `/process-alert` está público (allUsers no Cloud Run)
- Mover para autenticação OIDC quando tiver múltiplos tenants
