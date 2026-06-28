---
title: Variáveis de ambiente
description: Lista de variáveis de ambiente extraídas do código-fonte.
---

:::caution[Arquivo gerado]
Este documento é gerado por `scripts/gen-env-doc.sh`. Não edite manualmente.
Rode `make docs-env` para regenerar.
:::

<!-- Conteúdo gerado automaticamente -->

| Variável | Descrição | Obrigatória |
|---|---|---|
| `PORT` | Porta HTTP do servidor | Não (padrão: 8080) |
| `GCP_PROJECT` | ID do projeto GCP | Sim |
| `BIGQUERY_DATASET` | Dataset BigQuery | Não (padrão: garimpo) |
| `SHOPEE_APP_ID` | App ID da API de afiliados Shopee | Sim |
| `SHOPEE_SECRET` | Secret da API Shopee (criptografado no Secret Manager) | Sim |
| `TELEGRAM_TOKEN` | Token do bot Telegram | Não |
| `TELEGRAM_CHAT_ID` | Chat ID para notificações | Não |
| `GARIMPO_TOKEN` | Token para autenticação de coleta (X-Garimpo-Token) | Sim |
| `ENCRYPTION_KEY` | Chave AES para criptografia de credenciais de tenant | Sim |
| `FIREBASE_AUTH_DISABLED` | Desabilita autenticação (dev only) | Não |
| `WHATSAPP_PRODUCT_ID` | Product ID Maytapi (WhatsApp) | Não |
| `WHATSAPP_TOKEN` | Token Maytapi | Não |
| `LOG_LEVEL` | Nível de log (debug, info, warn, error) | Não (padrão: info) |
