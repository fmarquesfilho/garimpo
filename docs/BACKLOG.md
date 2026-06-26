# Product Backlog — Garimpei

Priorizado por valor de negócio. Atualizado em 26/06/2026.

---

## 🔴 Alta prioridade (próximas sessões)

### UX — Fluxo de curadoria incompleto
- **Problema observado:** Mileny busca produtos na Curadoria, vai para Publicar só para ver a imagem, e volta. O fluxo não permite ver detalhes sem sair da página.
- **Solução proposta:** modal/drawer de detalhes do produto (imagem, preço, comissão, link) direto na Curadoria. Botão "Publicar" dentro do modal.
- **Impacto:** reduz 3 cliques → 1 clique para publicar.

### Admin — Dashboard de monitoramento
- **Problema:** não há visibilidade sobre o que está funcionando (cron jobs, erros, consumo).
- **Necessidade:** painel admin com: status dos jobs (última execução, sucesso/falha), volume de dados coletados, erros recentes, consumo de recursos (BigQuery, Cloud Run).
- **Motivação:** dívida técnica — usando agentes se vai rápido mas se conhece menos sobre o que roda em produção.

### Alertas — Verificar eficácia
- **Pendência:** alertas foram configurados mas ainda não dispararam (precisa de 2+ coletas com variação). Monitorar se funcionam nas próximas 24h.
- **Ação:** se não disparar em 48h, investigar se o threshold (15%) é adequado para as lojas monitoradas.

---

## 🟡 Média prioridade (planning futuro)

### Arquitetura Multi-tenant
- **Visão:** transformar o Garimpei em SaaS — cada usuário pagante tem suas lojas, alertas, destinos e dados isolados.
- **Requisitos identificados:**
  - Isolamento de dados por tenant (BigQuery: partition por owner_uid ou dataset separado)
  - Cada tenant configura seus bots Telegram/WhatsApp
  - Billing por tenant (quantidade de lojas monitoradas, volume de coletas)
  - Onboarding self-service (cadastro → configura loja → recebe alertas)
- **Preocupações:** LGPD, custos por tenant, fair scheduling.

### LGPD
- **Contexto:** se abrir empresa no Brasil, precisa compliance com a Lei Geral de Proteção de Dados.
- **Itens a endereçar:**
  - Política de privacidade e termos de uso
  - Consentimento explícito para coleta de dados de uso
  - Direito de exclusão (apagar dados do usuário)
  - Encarregado de dados (DPO) — pode ser o próprio Fernando inicialmente
  - Logs de acesso a dados pessoais
  - Dados pessoais armazenados: email, nome, UID Firebase (mínimo)

### WhatsApp — Alternativa ao Maytapi
- **Problema:** Maytapi está tornando a conexão da Mileny lenta + sem notificações (abre sessão paralela que escuta tudo).
- **Alternativas a avaliar:**
  1. **WhatsApp Business API oficial** (via provedor BSP como 360Dialog, Twilio) — mais caro, mais estável, sem gambiarra de sessão
  2. **Evolution API** (open source, auto-hospedado) — grátis, controle total, mas precisa de VM dedicada
  3. **Segundo celular dedicado** — número separado só para automação (isola da conta pessoal)
- **Decisão:** avaliar custo × benefício de cada opção antes de implementar.
- **Ação imediata possível:** Mileny usar segundo celular com chip pré-pago para Maytapi.

### Arquitetura — Microserviços
- **Análise pendente:** o monólito atual (Go único no Cloud Run) funciona bem para o volume atual. Quando separar?
  - **Worker de coleta** — poderia ser um serviço separado (não compete com requests do frontend)
  - **Worker de alertas** — idem
  - **Frontend estático** — poderia ir para Firebase Hosting/Cloudflare Pages (CDN)
- **Trigger de migração:** quando o cold start do Cloud Run começar a afetar UX (>3s) ou quando a coleta demorar mais que o timeout de request.

---

## 🟢 Baixa prioridade (nice-to-have)

### Categorias dinâmicas da Shopee
- **Problema:** categorias hoje são estáticas (digitadas manualmente).
- **Solução:** buscar lista de categorias via API da Shopee e oferecer como autocomplete.
- **Complexidade:** a API de afiliados não expõe lista de categorias diretamente — precisaria de scrape ou endpoint não-oficial.

### Dashboard de conversões reais
- **Dependência:** integrar `conversionReport` da Shopee (requer webhook ou poll periódico).
- **Valor:** fecha o laço — saber qual publicação gerou venda.

### Notificação de deploy
- **Opções:** Telegram, email GitHub, WhatsApp, push do app GitHub.
- **Decisão:** não decidido ainda.

### Frontend — Refator completo com componentes UI
- **Status:** componentes existem (11) mas só 2 páginas usam.
- **Quando fazer:** quando surgir uma reescrita de página por outro motivo.

---

## ✅ Resolvido nesta sessão (26/06)

- [x] Monitoramento de lojas (formulário, rotação, throttling)
- [x] Alertas automáticos via Telegram (bot separado)
- [x] Evolução de preço nas estatísticas
- [x] Página de Oportunidades (feed unificado)
- [x] Menu drawer lateral
- [x] Domínio garimpei.app.br configurado
- [x] Paleta de cores harmonizada (WCAG AA)
- [x] Design tokens + Stylelint
- [x] Service layer para coleta (testável)
- [x] Fix: scheduler sync síncrono
- [x] Fix: logout + select_account
- [x] Fix: shopOfferV2 → productOfferV2
- [x] Fix: resolução de links curtos e slugs
- [x] OpenAPI spec + Swagger UI
- [x] Testes de regressão (9 cenários de produção)
