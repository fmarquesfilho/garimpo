# ADR 0023 — Alertas de preço via Cloud Tasks + Publisher (eliminar alerter)

**Status:** aceite  
**Data:** 2026-07-04  

## Contexto

O sistema monitora lojas e detecta variações de preço. Quando uma queda
significativa é encontrada, o usuário deve ser notificado automaticamente
via Telegram (e futuramente WhatsApp).

Existiam dois caminhos paralelos para enviar mensagens no Telegram:

1. **Publisher** (Go, gRPC :50052) — usado para publicações manuais do usuário
2. **Alerter** (Go, gRPC :50053) — usado para alertas automáticos de preço

Ambos tinham implementações independentes de `sendMessage` e `sendPhoto` contra
a Telegram Bot API. Essa duplicação gerava:
- Dois tokens de bot configurados separadamente
- Dois conjuntos de retry/error handling
- Impossibilidade de reutilizar o canal WhatsApp para alertas
- Confusão sobre quem é responsável pela entrega

Além disso, o scheduler não chamava o alerter — a integração nunca foi conectada.

## Decisão

### Eliminar o alerter como serviço separado

A responsabilidade é redistribuída:

| Papel | Antes | Depois |
|-------|-------|--------|
| "O que mudou" (detecção) | Alerter (BigQuery direto) | Analyzer Python |
| "Devo alertar?" (decisão) | Alerter (threshold fixo) | C# API (regras no PG) |
| "Como entregar" (envio) | Alerter (Telegram direto) | Publisher (gRPC) |
| "Quando verificar" (trigger) | Não existia | Cloud Tasks (durável) |

### Usar Cloud Tasks como fila de alertas

Após cada coleta bem-sucedida, o scheduler cria uma Cloud Task que chama
`POST /internal/alerts/check` na API C#. O Cloud Tasks garante:

- **Rate limiting**: 1 dispatch/s (respeita Telegram: 1 msg/s por chat)
- **Durabilidade**: tasks sobrevivem restarts, deploys, scale-to-zero
- **Deduplicação**: task name = `alert-{keyword}-{YYYY-MM-DD}` (1 por dia)
- **Retry**: 5 tentativas com backoff exponencial 10s→300s
- **Auth**: OIDC token em cada entrega

### Fluxo completo

```
Scheduler (após coleta)
    │ CreateTask (Cloud Tasks API)
    ▼
Queue: price-alerts (1 msg/s, retry 5x)
    │ HTTPS + OIDC
    ▼
C# API: POST /internal/alerts/check
    │ 1. GET /quedas no Analyzer (variações BQ)
    │ 2. Aplica regras do tenant (PG: threshold, canais)
    │ 3. Deduplicação (mesmo produto+dia = skip)
    ▼
Publisher (gRPC :50052) → Telegram / WhatsApp
```

### Responsabilidades por componente

| Componente | Responsabilidade |
|---|---|
| **Scheduler** | Orquestra coletas, enfileira alerta após sucesso |
| **Cloud Tasks** | Fila durável com rate limiting e retry |
| **C# API** | Decisão: aplica regras, dedup, escolhe canal |
| **Analyzer** | Dados: retorna variações detectadas no BigQuery |
| **Publisher** | Entrega: envia para Telegram/WhatsApp |

## Alternativas avaliadas

### 1. Canal em memória (Go channel)

- Simples mas perde tudo no restart
- Sem retry, sem deduplicação
- **Descartado**: não durável

### 2. Pub/Sub

- Desacoplamento total, fan-out nativo
- Rate limiting não é nativo (implementação manual)
- **Descartado**: overkill para 1 produtor → 1 consumidor

### 3. PostgreSQL como queue (transactional outbox)

- Zero dependência nova, transacional
- Polling (delay 1min), precisa de worker
- **Boa opção** mas Cloud Tasks é mais simples para o caso

### 4. Manter alerter separado + conectar ao scheduler

- Manteria código duplicado (2 implementações Telegram)
- Alerter precisaria de SnapshotRepo (BigQuery direto, viola ownership)
- **Descartado**: duplicação + violação de fronteira

## Consequências

### Se aceitar

- -1 sidecar no Cloud Run (alerter removido futuramente)
- Canal de entrega unificado (Telegram/WhatsApp via publisher)
- Alertas passam a suportar WhatsApp sem nenhum código novo
- Regras de alerta por tenant (PG) em vez de threshold fixo
- Zero code novo no alerter (deprecated, pode ser removido)
- Cloud Tasks no free tier (1M tasks/mês, custo $0)

### Custo/risco

- +1 serviço GCP (Cloud Tasks queue) — setup único
- Task delivery depende do Cloud Run estar acessível (mitigado: retry 5x)
- Scheduler precisa de credenciais Cloud Tasks (service account já existente)

## Migração

1. ✅ Implementar `internal/taskqueue` (Cloud Tasks client Go)
2. ✅ Criar `POST /internal/alerts/check` no C# (analyzer → publisher)
3. ⬜ Criar queue `price-alerts` via `deploy/setup-cloud-tasks.sh`
4. ⬜ Deploy com env vars no scheduler
5. ⬜ Testar end-to-end (script `test-alerts.sh`)
6. ⬜ Após confirmar, remover alerter do `cloud-run-service.yaml`
7. ⬜ Remover `services/alerter/` e `internal/alerts/` do repo
