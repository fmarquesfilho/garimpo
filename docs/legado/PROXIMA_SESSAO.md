# Próxima Sessão — Configurar Scheduler + Onboarding Lojas

**Data:** 2026-07-05
**Prioridade:** Alta

---

## Objetivo

Completar o fluxo de variação de preços tornando-o self-service:
1. Criar mecanismo de auto-sync (buscas ativas no PG → jobs no scheduler)
2. Permitir que o usuário adicione lojas pelo frontend e a coleta comece automaticamente
3. Testar publicação a partir de uma variação detectada

---

## Contexto (sessão 2026-07-04)

O fluxo de variação de preços está **funcionando em produção**:
- Scheduler coleta snapshots a cada ~8h (2 lojas + keywords)
- Analyzer detecta variações comparando snapshots no BigQuery
- Frontend exibe na aba "📉 Preços" da página /lojas
- Dashboard mostra contagem de quedas/altas

### Dados reais em produção

- 1576+ snapshots no BigQuery (2 lojas: Glory of Seoul, outra com 104 produtos)
- 5 variações de preço detectadas na loja-920292999 (2 quedas de 20%+)
- 11 variações na loja-457864097
- 3 produtos novos detectados

---

## O que falta

### 1. Sync buscas → scheduler (T-0028)
O scheduler tem jobs mas não se atualiza quando o usuário adiciona/remove lojas.
Precisa de um endpoint ou mecanismo que:
- Leia buscas ativas no PostgreSQL
- Crie/remova cron jobs no scheduler via gRPC `SetSchedule`
- Rode ao subir o serviço ou via trigger periódico

### 2. Onboarding de novas lojas
Quando o usuário adiciona uma loja pelo frontend (`POST /api/lojas`):
- A busca é criada no PG ✅
- Mas o scheduler não sabe dela
- Precisa triggar `SetSchedule` imediatamente

### 3. Publicação a partir de variação
O botão 📤 na aba Preços navega para /publicar com dados preenchidos.
Testar se o fluxo completo funciona (variação → publicar → Telegram).

### 4. Alertas automáticos de queda
O alerter (Go) está deployado mas não está wired ao fluxo de variações.
Precisa conectar: scheduler detecta queda → alerter notifica via Telegram.

---

## Resultados da sessão 2026-07-04

### Bug crítico corrigido
- Analyzer crashava: coluna `em` não existe no BigQuery (é `coletado_em`)
- Fix deployado via CI, endpoint `/api/lojas/novidades` agora retorna dados reais

### Features implementadas
- Mock mode no analyzer (`MOCK_DATA=true`) para dev local sem BigQuery
- Script `seed-local-test.py` para popular cenário fictício
- Endpoint `/evolucao` enriquecido com `resumo.total_quedas/total_altas`
- Endpoint `/estatisticas` retorna `total_amostras` para o dashboard
- Cloud Run service yaml atualizado (analyzer sidecar + ANALYZER_URL)

### Testes E2E adicionados
12 cenários Playwright para o fluxo de variação de preços:
- Seleção de loja → abas → tabela de variações
- Indicadores ↓/↑ para quedas/altas
- Botão publicar navega com dados
- Graceful degradation (analyzer offline)

### Documentação
- `docs/guias/fluxo-variacao-precos.md` — guia completo do fluxo
- `backlog/tasks/T-0044` — tarefa documentando o diagnóstico e fix
- Atualização do PROXIMA_SESSAO.md

### Tasks concluídas
- T-0044: Testar fluxo de variação de preços end-to-end

### Commits (5)
```
a4e52b7 feat(analyzer): add mock mode + seed script for local price variation testing
e7e8648 deploy: add analyzer sidecar to Cloud Run + wire ANALYZER_URL to scheduler
f2fab6e fix(analyzer): use coletado_em column name matching production BQ schema
8e68a52 docs: fluxo de variação de preços + task T-0044
9e6f5ca test(e2e): add price variation flow tests (12 scenarios)
9dd96ee feat(analyzer): enrich /evolucao with resumo (quedas/altas) + /estatisticas total_amostras
```
