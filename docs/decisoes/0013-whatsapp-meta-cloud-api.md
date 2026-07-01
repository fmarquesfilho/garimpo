# ADR 0013 — Migrar WhatsApp de Maytapi para Meta Cloud API

**Status:** aceite  
**Data:** 2026-07-01  

## Contexto

O Garimpei usa a Maytapi (intermediário third-party) para enviar mensagens via WhatsApp.
Isso funciona, mas tem limitações:

1. **Custo desnecessário** — Maytapi cobra por mensagem + assinatura, enquanto a Meta Cloud API é gratuita (paga-se apenas por conversas iniciadas pelo negócio)
2. **Camada extra** — dependência de um intermediário que pode sair do ar, mudar preços, ou ser descontinuado
3. **Sem suporte a templates** — a Meta exige templates aprovados para mensagens proativas; Maytapi abstrai isso mas limita o controle
4. **Compliance** — a API oficial da Meta segue os termos de uso do WhatsApp diretamente

Conversando com a Mileny, decidimos migrar para a API oficial.

## Decisão

Substituir o `WhatsAppSender` (Maytapi) por um novo sender que usa a **Meta WhatsApp Business Cloud API** diretamente.

### Endpoint

```
POST https://graph.facebook.com/v25.0/{PHONE_NUMBER_ID}/messages
Authorization: Bearer {WHATSAPP_ACCESS_TOKEN}
Content-Type: application/json
```

### Variáveis de ambiente (novas)

| Variável | Descrição |
|----------|-----------|
| `WHATSAPP_ACCESS_TOKEN` | Token permanente da Meta (System User token) |
| `WHATSAPP_PHONE_NUMBER_ID` | ID do número de telefone no Meta Business |
| `WHATSAPP_BUSINESS_ACCOUNT_ID` | WABA ID (para futuras funcionalidades) |

### Variáveis removidas

| Variável | Motivo |
|----------|--------|
| `WHATSAPP_API_KEY` | Era o token da Maytapi |
| `WHATSAPP_PRODUCT_ID` | ID do produto Maytapi |
| `WHATSAPP_PHONE_ID` | ID do telefone na Maytapi (diferente do Phone Number ID da Meta) |

## Impacto

- `internal/publish/whatsapp.go` — reescrever para usar Graph API da Meta
- `services/publisher/server.go` — atualizar factory (troca `NovoWhatsAppSenderFromEnv`)
- `internal/publish/novo.go` — atualizar factory
- Variáveis de ambiente no deploy

## Consequências

### Se aceitar

- Comunicação direta com a Meta sem intermediário
- Custo mais baixo (sem markup do Maytapi)
- Acesso a features oficiais (templates, botões interativos, read receipts)
- Mais controle sobre retry e rate limiting

### Se rejeitar

- Continua dependente da Maytapi
- Risco de indisponibilidade por problemas no intermediário

## Referência

- [Meta WhatsApp Cloud API docs](https://developers.facebook.com/docs/whatsapp/cloud-api/)
- [Send messages guide](https://developers.facebook.com/docs/whatsapp/cloud-api/guides)
