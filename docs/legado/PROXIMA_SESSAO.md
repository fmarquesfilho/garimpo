# Próxima Sessão — Testar Fluxo de Variação de Preços

**Data:** 2026-07-04
**Prioridade:** Alta (feature core do produto)

---

## Objetivo

Validar o fluxo completo de detecção de variações de preço: coleta de snapshots → comparação → detecção de quedas/altas → exibição na UI (página Lojas, aba "📉 Preços") → publicação de ofertas com preço antigo vs atual.

---

## Contexto

O sistema monitora lojas Shopee via coletas agendadas (scheduler → collector). Cada coleta grava um snapshot no BigQuery com os preços atuais. O frontend compara snapshots para detectar variações e exibe na aba "Preços" da página Lojas.

### Fluxo esperado

```
Scheduler (cron) → Collector (gRPC FetchShop) → BigQuery (snapshot)
                                                       │
Frontend (GET /api/lojas/novidades) ←── API C# ←──────┘
    │                                    (compara snapshots, detecta diffs)
    ▼
Aba "📉 Preços" mostra variações
    │ clique "📤 Publicar"
    ▼
Publicação com preço anterior vs atual
```

### Serviços envolvidos

| Serviço | Porta | Responsabilidade |
|---------|-------|-----------------|
| Scheduler Go | :50054 | Orquestra coletas por cron |
| Collector Go | :50051 | FetchShop → busca produtos da loja na Shopee |
| API C# | :5000 | Endpoint /api/lojas/novidades (compara snapshots) |
| BigQuery | — | Armazena snapshots históricos |
| Frontend | :5173 | Exibe variações na aba Preços |

---

## O que testar

### 1. Coleta de snapshots funciona?
- O scheduler está disparando coletas?
- O collector grava no BigQuery?
- Verificar com: `./scripts/prod-api.sh GET /api/coletas`

### 2. Endpoint de novidades retorna dados?
- `./scripts/prod-api.sh GET '/api/lojas/novidades?busca_id=<id>&dias=7'`
- Espera: `{ produtos_novos: [...], variacoes: [...], dias_janela: 7 }`

### 3. Frontend exibe variações?
- Navegar para /lojas → selecionar loja → aba "📉 Preços"
- Verificar se a tabela mostra: Produto, Antes, Agora, Variação%, Detectado

### 4. Publicar a partir de variação funciona?
- Clicar 📤 na linha da variação
- Deve navegar para /publicar com `preco_atual` e `nome` preenchidos
- Enviar → mensagem no Telegram

### 5. Se não há snapshots (primeira coleta)
- Como o sistema se comporta com zero histórico?
- O scheduler precisa rodar pelo menos 2x para ter diff

---

## Setup necessário

### Dev local (contra produção)
```bash
cd web && VITE_API_BASE=https://garimpei.app.br npm run dev
# Login via Google no browser
```

### Testar APIs diretamente
```bash
./scripts/prod-api.sh GET /api/coletas
./scripts/prod-api.sh GET '/api/lojas/novidades?busca_id=<ID_DA_BUSCA>&dias=30'
./scripts/prod-api.sh GET /api/buscas   # lista perfis de busca com shop_ids
```

### Verificar BigQuery (snapshots)
```bash
# Via banco PostgreSQL (buscas salvas)
PG_URL=<...>  # ver scripts/prod-api.sh para construir
psql "$PG_URL" -c 'SELECT "Id", "Keywords", "ShopIds" FROM "Buscas" LIMIT 5;'
```

---

## Resultados da sessão anterior (2026-07-03)

### Bugs corrigidos
- ✅ Página publicar travada em "Carregando..." (Tooltip.Provider faltando)
- ✅ Select mostrando UUID ao invés de nome (selectedLabel derivado)
- ✅ Envio Telegram falhando silenciosamente (sendPhoto fallback + JsonPropertyName)
- ✅ Cloud Run não criava nova revision (SHA tags no deploy)
- ✅ CORS para dev local

### Migração UI concluída
- ✅ shadcn-svelte + Tailwind CSS v4 em todos os 50 componentes
- ✅ Prettier com Tailwind class sorting
- ✅ Dark mode corrigido (contraste)
- ✅ Layout padronizado (space-y-8)

### Infra/DX
- ✅ pre-push ~20s (sem E2E)
- ✅ E2E disponível via `mise run test:e2e`
- ✅ `scripts/prod-api.sh` para debug produção
- ✅ PostgreSQL auto-start no `mise run test:csharp`
- ✅ CI otimizado (mise@v4, buf binary direto, deploy com SHA)

### Tasks concluídas
- T-0035: Pós-migração bugs
- T-0037: Bits UI + tokens
- T-0038: UI fase 2
- T-0040: shadcn-svelte + Tailwind
- T-0041: format:check CI
- T-0042: Dark mode + layout
- T-0043: Deploy Cloud Run + JSON binding

### Pendente (backlog)
- T-0034: Descobrir filtros (badge drift)
- T-0027: Publisher multi-tenant tokens
- T-0028: Configurar coleta no scheduler (popular snapshots)
- T-0002: Persistir conversões BigQuery

---

## Decisões a tomar nesta sessão

1. O scheduler está configurado e rodando em produção? Ou precisa ser ativado?
2. Há snapshots no BigQuery ou preciso popular manualmente?
3. O endpoint `/api/lojas/novidades` está implementado no C# ou ainda aponta para o Go legado?
4. As buscas com `shop_ids` estão configuradas no banco?
