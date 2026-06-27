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
- **Status:** ✅ Implementado (27/06)
- **O que foi feito:**
  - Endpoint `GET /api/conversoes/reais` consulta conversionReport da Shopee em tempo real
  - Aba Desempenho mostra: produto, loja, comissão, status, canal, data
  - Resumo visual: comissão total, conversões, confirmadas, pendentes
  - Seletor de período (7/30/90 dias) + botão sincronizar
  - Endpoint protegido por auth (Bearer token, não COLETA_TOKEN)
- **Pendente:** persistir conversões no BigQuery para histórico (hoje é consulta on-demand)

### SPEC NECESSÁRIA: Redefinir página de Estatísticas
- **Problema:** seção "Mercado por categoria" não reflete os fluxos reais. Categorias são rótulos manuais, não dados reais da Shopee.
- **O que deveria mostrar:** evolução de preço das lojas (já existe), performance por publicação (precisa de conversões), volume de coletas.
- **Ação:** criar spec em `.kiro/specs/statistics-redesign/`

### UX — Fluxo de curadoria incompleto
- **Status:** ✅ Resolvido (28/06). Página Descobrir unificada com ProductCard consistente, feed de fontes múltiplas, favoritos.
- **Pendente:** modal de detalhes do produto (imagem ampliada + dados completos) sem sair da página.

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
- **Status:** ✅ MVP implementado (27/06)
- **O que foi feito:**
  - Package `internal/tenant` com Config, Store interface, MemoryStore
  - Criptografia AES-256-GCM para secrets (env var `ENCRYPTION_KEY`)
  - 6 endpoints de onboarding: `/api/onboarding/{status,termos,shopee,telegram,validar,excluir-conta}`
  - Validação real de credenciais Shopee (chamada de teste à API)
  - Página `/configurar` no frontend: wizard de 4 steps com instruções passo-a-passo
  - Exclusão de conta (LGPD) com confirmação dupla
- **Pendente para beta testers entrarem:**
  - Implementar `BigQueryTenantStore` (persistir configs — hoje é MemoryStore)
  - `ScopedStore`: filtrar dados por `owner_uid` em todas as queries
  - Resolver credenciais do tenant no middleware (usar tokens do tenant nas coletas)
  - Billing por tenant (futuro, não MVP)

### LGPD
- **Status:** ✅ Parcialmente implementado (27/06)
- **Itens implementados:**
  - Termos de uso com aceite explícito e timestamp (step 1 do onboarding)
  - Direito de exclusão: endpoint `POST /api/onboarding/excluir-conta`
  - Secrets criptografados (nunca em plaintext no banco)
  - Dados pessoais mínimos: email, UID Firebase
- **Pendente:**
  - Política de privacidade (página estática com texto legal)
  - Logs de acesso a dados pessoais
  - DPO: Fernando como encarregado inicial

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
- **Status:** ✅ Concluído (28/06). ProductCard unificado, 34 componentes totais, todos os arquivos ≤400 linhas.
- **Resultado:** 4 cards distintos → 1 ProductCard com 3 layouts. 6 componentes mortos removidos.

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

## ✅ Resolvido nesta sessão (28/06)

- [x] Refatoração: bigquery_store.go (1121→7 arquivos ≤400 linhas cada)
- [x] Refatoração: lojas.go split em lojas.go + shopee_resolver.go
- [x] Refatoração: +layout.svelte → NavDrawer + LandingHero
- [x] CI: check-file-size.sh (max 400 linhas, bloqueia deploy)
- [x] Timeout em todas as chamadas à API Shopee (20-30s client-side)
- [x] Fix: aba Desempenho em loop infinito de loading
- [x] UX: botão ✕ no input de busca + ESC para limpar
- [x] UX: Publicações movida para seção Monitoramento no menu
- [x] Alertas de produtos novos desabilitados (preparado para config futura)
- [x] ProductCard unificado com 3 layouts (full/compact/feed)
- [x] Favoritos: backend (BigQuery) + frontend (store reativo com sync servidor)
- [x] Favoritos: indicação visual ★/☆ reativa no ProductCard
- [x] Página Descobrir unificada: feed com filtros de fonte (Busca/Quedas/Novos/Favoritos)
- [x] Busca universal filtra por nome de produto OU nome de loja
- [x] Buscas salvas: atalhos na Descobrir + gestão completa em /lojas
- [x] Modelo Busca expandido: fontes[], categorias[], dias_janela
- [x] GerenciarBuscas: UI com fontes, categorias, dias_janela, agendamento
- [x] Validação flexível: busca aceita keyword OU loja OU categoria OU fonte
- [x] Cache: backend 5 min (novidades), frontend 2 min (oportunidades)
- [x] Snapshot enriquecido: imagem, link, loja gravados na coleta
- [x] Rota /oportunidades → redirect 301 para / (unificada)
- [x] Dead code: 6 componentes removidos, ~1400 linhas eliminadas
- [x] 20 testes Vitest para lógica da Descobrir (62 total, <2s)
- [x] 17 cenários de busca agendada documentados e testados (NormalizarBusca)
- [x] golangci-lint limpo, 0 issues
- [x] Documentação: BUSCAS_AGENDADAS.md, TESTES_DESCOBRIR.md, MELHORIAS_28_06.md, REFATORACAO.md

- [x] Origem do produto: campo `Origin` no domínio, adaptadores, badge no card
- [x] Origem do produto: fallback `origem_padrao` por loja monitorada
- [x] Origem do produto: CLI de introspecção GraphQL (`cmd/shopee-introspect`)
- [x] Multi-tenant: package `internal/tenant` com Config, Store, criptografia AES-256-GCM
- [x] Multi-tenant: 6 endpoints de onboarding (termos, shopee, telegram, validar, excluir)
- [x] Multi-tenant: página `/configurar` no frontend (wizard 4 steps)
- [x] Multi-tenant: validação real de credenciais Shopee via chamada de teste
- [x] LGPD: aceite de termos com timestamp + exclusão de conta
- [x] Schema evolution automática: `EnsureSchema` adiciona colunas novas sem migração manual
- [x] Spec documentada: `.kiro/specs/product-origin/` (requisitos detalhados)
- [x] Spec documentada: `.kiro/specs/multi-tenant-beta/` (design de alto nível)
- [x] Documentação atualizada: BACKLOG, ENTIDADES, APIS

---

## 📝 Itens para próxima sessão (documentados 28/06)

### Feature: Alertas configuráveis
- **Problema:** Alertas hoje são hardcoded (chat_id fixo, só variação de preço). Mileny quer separar alertas de preço e novidades em destinos diferentes.
- **Proposta:** Página `/alertas` (ou seção em `/configurar`) com tipos: variação preço, produto novo, conversão. Cada tipo com destino, threshold e frequência.
- **Documentado em:** docs/MELHORIAS_28_06.md item 7
- **Complexidade:** Alta (~4-6h)

### Feature: Scroll infinito na busca
- **Problema:** Busca retorna max 20 resultados. Mileny pode querer ver mais sem recarregar.
- **Solução:** Lazy-load ao scroll, paginação server-side (offset/cursor).
- **Complexidade:** Média (~2h)

### Feature: Modal de detalhes do produto
- **Problema:** Para ver imagem ampliada ou dados completos, Mileny precisa abrir link externo.
- **Solução:** Click no card abre modal overlay com imagem grande + todos os dados + ações.
- **Complexidade:** Média (~1.5h)

### UX: Filtro de busca accent-insensitive
- **Problema:** "sérum" não encontra "Serum" (sem acento). Dados da Shopee são inconsistentes.
- **Solução:** Normalizar acentos no filtro client-side (`.normalize('NFD').replace(...)`)
- **Complexidade:** Baixa (~15 min)

### Feature: Origem do produto (Coréia/Japão)
- **Regra de domínio da Mileny:** precisa saber se produto é de origem Coréia/Japão (muitos são falsificados). A Shopee mostra um campo "Origem" no produto e na loja.
- **Status:** ✅ Implementado (27/06) — via `origem_padrao` por loja monitorada
- **Solução implementada:**
  1. Campo `origem_padrao` na Busca — ao adicionar loja, o usuário marca "Coreia"/"Japão" e todos os produtos herdam
  2. Badge visual no CandidateCard: 🇰🇷 Coreia / 🇯🇵 Japão / 🇨🇳 China
  3. Motor de coleta aplica `origem_padrao` a todos os produtos da loja
  4. Endpoint `/api/produto/origem` consulta `origem_padrao` da loja monitorada
- **Limitação confirmada:** a API de afiliados da Shopee NÃO expõe país de origem. A API pública v4 bloqueia IPs de datacenter (403). Documentado em `docs/SHOPEE_INTROSPECT_RESULT.md`.
- **Alternativa futura:** proxy residencial (Bright Data ~$3/mês) desbloqueia acesso à API pública. Código preparado mas não ativo.

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
