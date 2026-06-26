# Product Backlog — Garimpei

Priorizado por valor de negócio. Atualizado em 26/06/2026.

---

## 🔴 Alta prioridade (próximas sessões)

### SPEC NECESSÁRIA: Modelagem de entidades (Busca/Coleta/Loja)
- **Problema:** uma "Busca" pode ser por keyword OU por loja, dispara processos distintos, mas usa a mesma struct. Isso gera confusão (ex: campo "keyword" preenchido com "loja-457...").
- **Decisão necessária:** separar em entidades distintas? Usar composição? Interface?
- **Impacto:** afeta BigQuery schema, API, scheduler, frontend.
- **Questão de UX:** faz sentido monitorar múltiplas lojas numa mesma "busca"? (Mileny cadastra 1 loja por vez — manter 1:1 simplifica.)
- **Ação:** criar spec em `.kiro/specs/entity-model/`

### SPEC NECESSÁRIA: Rastreamento de conversões (fechar o ciclo)
- **Problema:** aba "Desempenho" nas publicações não mostra conversões reais. A API da Shopee tem `conversionReport` que retorna vendas reais por `subId`.
- **Impacto:** fundamental para estatística e estratégia — saber o que realmente converteu.
- **Questão técnica:** precisa de poll periódico (webhook não existe na API de afiliados). O subId já é gerado em cada publicação.
- **Ação:** criar spec em `.kiro/specs/conversions-tracking/`

### SPEC NECESSÁRIA: Redefinir página de Estatísticas
- **Problema:** seção "Mercado por categoria" não reflete os fluxos reais. Categorias são rótulos manuais, não dados reais da Shopee.
- **O que deveria mostrar:** evolução de preço das lojas (já existe), performance por publicação (precisa de conversões), volume de coletas.
- **Ação:** criar spec em `.kiro/specs/statistics-redesign/`

### UX — Fluxo de curadoria incompleto
- **Status:** ✅ Parcialmente resolvido (imagem full-width no card, filtros colapsáveis, título simplificado).
- **Pendente:** modal de detalhes do produto (imagem ampliada + dados completos) sem sair da página. Cenário BDD já definido em BDD_STRATEGY.md.

### UX — Página de Coletas precisa de redesign
- **Problema:** mostra "buscas agendadas" como cards e "resumo por keyword" como tabela — ambos confusos (mostra loja-920... como keyword, categoria com traço).
- **Solução:** colapsar cards em tabela expansível, mover para área de "logs do usuário" (não admin, mas visibilidade operacional).
- **Ação:** redesenhar após spec de entidades.

### Regra de negócio — Simplificar filtros backend
- **Problema:** API aplica filtros (comissão mín, vendas mín, nota mín) que não estão explícitos na interface simplificada. A Mileny não sabe que produtos estão sendo excluídos.
- **Princípio:** se não está na UI, não deveria filtrar. Ou mostrar claramente "X produtos excluídos por filtros".
- **Ação:** revisar elegibilidade no backend, alinhar com o que a UI mostra.

### "Descobrir novos" — precisa de explicação na UI
- **O que faz:** reserva ~20% dos resultados para produtos fora do topo (exploração).
- **Problema:** Mileny não entende o que isso significa.
- **Solução:** tooltip ou texto explicativo: "Mostra produtos que normalmente não aparecem no topo — ajuda a encontrar oportunidades escondidas."

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

### Quadro Kanban
- **Status:** ❌ Removido (26/06). Não estava sendo usado pela Mileny.
- **Decisão:** se surgir necessidade de fluxo de trabalho visual no futuro, reavaliar.

### Nicho vs Diversificada (estratégias de ranking)
- **Status:** removido da UI (26/06). Backend ainda suporta ambas.
- **Decisão futura:** quando houver dados de conversão (spec pendente), reavaliar se vale mostrar comparação de performance por estratégia.
- **Dívida técnica:** código de Strategy pattern no backend permanece mas não é usado pelo frontend simplificado.

---

## ✅ Resolvido nesta sessão (26/06)

- [x] Monitoramento de lojas (formulário, rotação, throttling)
- [x] Alertas automáticos via Telegram (bot separado)
- [x] Evolução de preço nas estatísticas
- [x] Página de Oportunidades (feed unificado)
- [x] Menu drawer lateral
- [x] Domínio garimpei.app.br configurado (Cloudflare Worker)
- [x] Paleta de cores harmonizada (WCAG AA)
- [x] Design tokens + Stylelint (0 cores hex nas páginas)
- [x] Service layer para coleta (testável, 145→62 linhas no handler)
- [x] Fix: scheduler sync síncrono (jobs não eram criados)
- [x] Fix: logout + select_account
- [x] Fix: shopOfferV2 → productOfferV2
- [x] Fix: resolução de links curtos, slugs e URLs de produto
- [x] Fix: nome da loja (API v4 Shopee) em vez de "loja-457..."
- [x] Fix: keyword no snapshot para estatísticas funcionarem
- [x] Fix: publicações preservam título após envio
- [x] OpenAPI spec + Swagger UI em /api/docs
- [x] Testes de regressão (9 cenários de produção)
- [x] 11 novos testes E2E (total: 43)
- [x] Curadoria simplificada (sem jargão nicho/diversificada)
- [x] FilterBar colapsável (busca proeminente, filtros escondidos)
- [x] CandidateCard com imagem full-width
- [x] Rebrand: Garimpo → Garimpei
- [x] Dependências atualizadas (Vite 8, Go 1.26, Node 24)
- [x] Migração BigQuery (colunas novas)
- [x] Coletas manuais disparadas para popular dados iniciais
- [x] Quadro removido (não era usado)
- [x] Lojas movida para Configurações no menu
- [x] Enter funciona na busca
- [x] "Mercado por categoria" removido (conceito obsoleto)
- [x] Categoria opcional (não obriga "cosméticos")
- [x] Buscas salvas filtram apenas keywords (lojas ficam em /lojas)
- [x] Tag "1 loja" removida dos cards
- [x] Endpoint POST /api/conversoes/sync (conversionReport Shopee)
- [x] ESLint + knip adicionados ao CI
- [x] 3 specs documentadas (entidades, conversões, estatísticas)
- [x] Página Estatísticas reformulada (resumo operacional para Mileny)
- [x] Badges de oportunidades com tooltips e labels claros
- [x] Aba Desempenho em Publicações: explicação do ciclo de conversões
- [x] Coletas movida para seção Admin no menu
- [x] Max-width padronizado (900px em todas as páginas)
- [x] Diagrama ER (docs/ENTIDADES.md) + teste de sincronização
- [x] Estratégia "diversificada" descontinuada do service layer
- [x] ESLint + knip + Stylelint: 0 erros em tudo
- [x] 6 novos testes de regressão (conversões, buscas, service)
- [x] Scoring neutro (remove bonus por categoria — cosméticos não tem mais vantagem)
- [x] Categoria real da API Shopee (productCatIds → nomes via mapeamento)
- [x] Card mostra nome da loja + imagem clicável (abre produto na Shopee)
- [x] Botão Publicar com cor suave (não compete visualmente)
- [x] Removidos: checkbox "descobrir novos", select "resultados"
- [x] Feed com 20 resultados (era 9)
- [x] Endpoint conversionReport corrigido (campos reais da API)
- [x] Dashboard de Estatísticas compacto (single page, sem scroll)
- [x] 4 novos componentes: MetricCard, MiniChart, DashPanel, RankList
- [x] Quadro removido + Lojas movida para Configurações
- [x] Enter funciona na busca
- [x] Categorias não são mais carimbadas manualmente nos produtos


---

## 📝 Itens para próxima sessão (documentados 27/06)

### Feature: Origem do produto (Coréia/Japão)
- **Regra de domínio da Mileny:** precisa saber se produto é de origem Coréia/Japão (muitos são falsificados). A Shopee mostra um campo "Origem" no produto e na loja.
- **Limitação descoberta:** a API de afiliados (GraphQL) **não expõe** o campo de origem do produto. Campos disponíveis: productName, shopName, shopId, productCatIds, shopType, preço, vendas, comissão, imagem, link. A API pública v4 (`/api/v4/item/get`) que mostra o campo "Origem" exige cookie de sessão autenticada — não pode ser chamada server-side.
- **Ação:** pesquisar se existe endpoint alternativo ou se a Shopee expõe isso em algum outro lugar. Alternativas: (1) scraping com sessão, (2) Mileny marca manualmente quais lojas são verificadas, (3) nova versão da API de afiliados pode expor no futuro.
- **Status:** documentado como limitação técnica. Feature bloqueada até encontrar fonte de dados.

### Feature: Categorias dinâmicas da API Shopee
- **Status:** ✅ Implementado (27/06)
- **Solução:** mapeamento de ~20 IDs de categoria nível 1 da Shopee Brasil (empírico via productCatIds)
- **Como funciona:** API retorna `productCatIds` → `NomeCategoriaPrincipal()` traduz para nome legível
- **Próximo:** expandir mapeamento se novos IDs aparecerem nos dados; pesquisar se há endpoint oficial de árvore de categorias

### UX: Feed infinito na busca
- **Problema de Mileny:** retornava poucos produtos (6-9). Deveria funcionar como feed.
- **Status:** ✅ Parcialmente resolvido — default aumentado para 20 resultados.
- **Pendente:** scroll infinito (paginação — carregar mais ao rolar). Adicionar ao backlog da próxima sessão.

### UX: Modo debug para desenvolvedor
- **Necessidade:** Fernando quer ver quais filtros estão sendo aplicados, o que a API retorna, e ter um botão "copiar como cURL" para testar no Postman.
- **Implementação:** badge discreto "🔧" que mostra os params enviados + resposta crua + cURL.

### Publicações: Informações úteis sobre conversões
- **O que Mileny quer ver:** quando alguém compra pelo link, saber de onde veio (qual sub_id converteu, quais canais, quais produtos venderam).
- **Dependência:** spec `conversions-tracking` (endpoint já criado, falta persistir + mostrar no frontend).
