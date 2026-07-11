# Configurar WhatsApp — Meta Business Cloud API

Guia para configurar o envio de mensagens via API oficial da Meta (substitui Maytapi).

## Pré-requisitos

1. Conta no [Meta for Developers](https://developers.facebook.com/)
2. Um app Meta do tipo "Business"
3. Número de telefone verificado (pode usar o número de teste da Meta)

## Passo a passo

### 1. Criar app Meta Business

1. Acesse [developers.facebook.com/apps](https://developers.facebook.com/apps/)
2. Clique em "Criar app" → tipo **Business**
3. Adicione o produto **WhatsApp** ao app

### 2. Obter Phone Number ID

1. No painel do app, vá em **WhatsApp > API Setup**
2. Anote o **Phone number ID** (ex: `113456789012345`)
3. Este é o valor de `WHATSAPP_PHONE_NUMBER_ID`

### 3. Gerar Access Token permanente

O token temporário da UI expira em 24h. Para produção:

1. Vá em **Business Settings > System Users**
2. Crie um System User (admin)
3. Atribua permissão `whatsapp_business_messaging` e `whatsapp_business_management`
4. Gere um token permanente para esse System User
5. Este é o valor de `WHATSAPP_ACCESS_TOKEN`

### 4. Configurar variáveis de ambiente

**Local (dev):**
```bash
export WHATSAPP_PHONE_NUMBER_ID="seu_phone_number_id"
export WHATSAPP_ACCESS_TOKEN="seu_token_permanente"
```

**Cloud Run (produção):**
```bash
# Criar secrets no Secret Manager
echo -n "seu_phone_number_id" | gcloud secrets create WHATSAPP_PHONE_NUMBER_ID --data-file=-
echo -n "seu_token_permanente" | gcloud secrets create WHATSAPP_ACCESS_TOKEN --data-file=-
```

Os secrets já estão referenciados no `deploy/cloud-run-service.yaml`.

### 5. Adicionar número de teste

Para enviar mensagens, o destinatário precisa estar na lista de testes OU o app precisa estar em produção:

1. Em **WhatsApp > API Setup > To**, adicione o número do grupo/destinatário
2. O destinatário recebe um código de verificação

### 6. Testar envio

**Via curl (direto na API):**
```bash
curl -X POST "https://graph.facebook.com/v25.0/${WHATSAPP_PHONE_NUMBER_ID}/messages" \
  -H "Authorization: Bearer ${WHATSAPP_ACCESS_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "messaging_product": "whatsapp",
    "to": "5511999999999",
    "type": "text",
    "text": { "body": "✅ Teste Garimpei — WhatsApp Meta Cloud API funcionando!" }
  }'
```

**Via aplicação (Go monólito com variáveis configuradas):**
1. Adicione um destino do tipo "whatsapp" com o número do grupo
2. Publique uma oferta pelo frontend → deve chegar no WhatsApp

### 7. Enviar imagem com caption

```bash
curl -X POST "https://graph.facebook.com/v25.0/${WHATSAPP_PHONE_NUMBER_ID}/messages" \
  -H "Authorization: Bearer ${WHATSAPP_ACCESS_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "messaging_product": "whatsapp",
    "to": "5511999999999",
    "type": "image",
    "image": {
      "link": "https://cf.shopee.com.br/file/exemplo.jpg",
      "caption": "✨ *Produto Teste*\n💸 *R$ 49,90*\n\n🛒 https://shope.ee/xxx"
    }
  }'
```

## Variáveis removidas (Maytapi)

As seguintes variáveis **não são mais usadas** e podem ser removidas do deploy:

| Variável antiga | Substituída por |
|-----------------|----------------|
| `WHATSAPP_API_KEY` | `WHATSAPP_ACCESS_TOKEN` |
| `WHATSAPP_PRODUCT_ID` | *(removido — não há equivalente)* |
| `WHATSAPP_PHONE_ID` | `WHATSAPP_PHONE_NUMBER_ID` |

## Troubleshooting

| Erro | Causa | Solução |
|------|-------|---------|
| 401 OAuthException | Token expirado/inválido | Regenerar token permanente via System User |
| 400 "recipient not in allowed list" | Número não está na lista de teste | Adicionar em API Setup > To |
| 400 "message template required" | Tentou mensagem proativa sem template | Para mensagens proativas (fora de 24h), criar template aprovado |
| 131030 "rate limit" | Muitas mensagens em pouco tempo | Respeitar limites: 80msg/s (tier padrão) |

## Referências

- [Meta WhatsApp Cloud API docs](https://developers.facebook.com/docs/whatsapp/cloud-api/)
- [Send messages guide](https://developers.facebook.com/docs/whatsapp/cloud-api/guides)
- [ADR-0013](../decisoes/0013-whatsapp-meta-cloud-api.md)
