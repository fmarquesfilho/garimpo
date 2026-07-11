# ADR-0015: Multi-tenancy com credenciais por tenant e compartilhamento

## Status

Aceito (2026-07-02)

## Contexto

O Garimpei opera como SaaS multi-tenant: cada usuário (tenant) configura suas
próprias credenciais de API (Shopee, Telegram, WhatsApp) via onboarding. O admin
(Fernando) e sua esposa são os dois primeiros tenants.

Cada tenant tem suas próprias credenciais Shopee (AppID + Secret). Porém, há
cenários onde um tenant quer compartilhar suas credenciais com outro — por exemplo,
a esposa tem a conta de afiliada e compartilha o acesso com o admin.

O monólito Go usava env vars globais. Na nova arquitetura, credenciais são
per-tenant no PostgreSQL (`TenantConfig`).

## Decisão

### Credenciais são per-tenant por padrão

Cada tenant configura suas credenciais no onboarding:
- **Shopee**: AppID + Secret (obrigatório, step 2)
- **Telegram**: Bot Token + Chat ID (opcional, step 3)
- **WhatsApp**: Phone Number ID + Access Token Meta (opcional, step 3)

### Compartilhamento de credenciais Shopee

Um tenant pode **compartilhar** suas credenciais Shopee com outro tenant:
- A esposa configura AppID + Secret no onboarding dela
- O admin pode referenciar as credenciais dela em vez de ter as próprias
- Implementação: campo `SharedFromUid` no `TenantConfig` — se preenchido, o sistema usa as credenciais do tenant referenciado para coletas

### Canais (Telegram/WhatsApp) são sempre individuais

Cada tenant configura seus próprios bots e canais. Não há compartilhamento
de canais — um bot Telegram pertence a um tenant e publica nos grupos dele.

### Fluxo de compartilhamento

1. Tenant A (esposa) completa onboarding com credenciais Shopee
2. Tenant B (admin) no step 2 do onboarding pode escolher:
   - "Usar minhas próprias credenciais" → preenche AppID + Secret
   - "Usar credenciais compartilhadas" → informa email/uid do Tenant A
3. Sistema valida que Tenant A existe e aceita compartilhamento
4. Coletas do Tenant B usam AppID + Secret do Tenant A

### Tokens no publisher (gRPC)

Os tokens de Telegram/WhatsApp são passados do C# API para o publisher Go
via campos no `PublishRequest` gRPC. Cada publicação usa os tokens do tenant
que está publicando — não existem tokens globais.

## Consequências

### Positivas
- Cada tenant opera de forma independente (sem interferência)
- Compartilhamento é explícito e auditável
- Admin pode operar sem ter conta de afiliado própria
- Publisher não precisa de acesso ao banco — recebe tokens no request

### Negativas
- Tokens em texto plano no PostgreSQL (TODO: encriptar com AES-256)
- Se esposa revogar credenciais, admin para de funcionar (acoplamento)
- Rate limits da Shopee são compartilhados entre quem divide credenciais

### Riscos
- Tokens não encriptados no banco (mitigação: campo `*Enc` preparado para encriptação)
- Compartilhamento pode gerar confusão ("quem está usando minha cota?")

## Implementação

| Componente | Estado |
|-----------|--------|
| `TenantConfig` entity (per-tenant) | ✅ Implementado |
| Onboarding 4 steps (Shopee + Telegram + WhatsApp) | ✅ Implementado |
| Frontend /configurar (wizard completo) | ✅ Implementado |
| Compartilhamento `SharedFromUid` | 🔮 Planejado (T-0031) |
| Encriptação de tokens | 🔮 Planejado |
| Publisher recebe tokens via gRPC | 🔮 Planejado (T-0027) |
