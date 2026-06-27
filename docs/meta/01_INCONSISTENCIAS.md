# Auditoria de Inconsistências — Documentação Garimpei

Comparação cruzada dos 18 documentos (estado em 2026-06-27). Cada item traz:
onde diverge, por que importa, e a correção sugerida. A coluna **Resolvido por**
aponta a decisão canônica em `00_LEIA-ME.md` quando existe.

Severidade: 🔴 confunde quem lê / induz a erro técnico · 🟡 desatualizado mas
inofensivo · 🟢 cosmético.

---

## 🔴 Alta

### I1 — Nome do produto: "Garimpo" vs "Garimpei"
- **Onde:** `MODELO.md`, `DEPLOY.md`, `DEPLOY_GCP.md`, `COLETA.md`, `MANUAL.md`,
  `JORNADA.md` usam **"Garimpo"**. `REFATORACAO.md`, `ANALISE_ESTATICA.md`,
  `ENTIDADES.md`, `openapi.yaml` (title "Garimpei API"), domínio `garimpei.app.br`
  e o `BACKLOG.md` (linha "Rebrand: Garimpo → Garimpei") usam **"Garimpei"**.
- **Por que importa:** leitor novo não sabe se são o mesmo produto. Afeta o
  branding em material para beta testers.
- **Correção:** produto = **Garimpei**; manter `garimpo`/`garimpo-api`/dataset
  `garimpo` como nomes internos. Busca-e-substitui controlado nos docs conceituais.
- **Resolvido por:** decisão canônica "Nome do produto".

### I2 — Estratégias "nicho vs diversificada": viva ou descontinuada?
- **Onde:** `MODELO.md` §3 e `JORNADA.md` (etapas 3–4, "modo Comparar") tratam as
  duas estratégias como experimento **ativo** (70/30, comparação). `ENTIDADES.md`,
  `MANUAL.md` e `BACKLOG.md` dizem que **diversificada foi descontinuada** e só
  "nicho" roda.
- **Por que importa:** `MODELO.md` é o documento-arcabouço que alguém leria
  primeiro; ele vende uma feature que não existe mais na UI.
- **Correção:** rebaixar a comparação de estratégias a **visão futura** (selo 🔮),
  condicionada a dados de conversão. Manter a explicação conceitual, marcando que
  hoje o ranking usa só "nicho".
- **Resolvido por:** "Estratégias de ranking".

### I3 — Quadro Kanban: feature do produto ou removido?
- **Onde:** `MODELO.md` §5 (quadro com WIP), `JORNADA.md` (etapa 6 "Produção
  (Quadro)"), `DEPLOY_GCP.md` (checklist: "A aba Quadro persiste os cards").
  `BACKLOG.md` e `MELHORIAS_28_06.md` dizem que o **Kanban foi removido**.
- **Por que importa:** o checklist de verificação de deploy manda testar uma aba
  que não existe — quebra o runbook.
- **Correção:** remover Quadro do `DEPLOY_GCP.md` (checklist) e marcar como visão
  em `MODELO.md`/`JORNADA.md`. (Nota: o Kanban **de gestão do projeto** volta em
  outra forma no `03_BACKLOG_COMO_CODIGO.md` — não confundir com o do produto.)
- **Resolvido por:** "Quadro Kanban (do produto)".

### I4 — Canal de publicação: Instagram vs Telegram/WhatsApp
- **Onde:** `APIS.md` (seção 2 inteira), `MODELO.md` (VSM: "Publicar no Instagram"),
  `JORNADA.md` assumem **Instagram** como canal. `MANUAL.md`, `ENTIDADES.md`,
  `COLETA.md`, `openapi.yaml` (tags Publicação/Destinos) implementam **Telegram +
  WhatsApp**; não há integração Instagram no produto.
- **Por que importa:** desenha a operação da Mileny em torno de um canal não
  implementado; confunde escopo de release.
- **Correção:** marcar todo o conteúdo de Instagram como 🔮 Planejado. Atualizar o
  VSM de `MODELO.md` para Telegram/WhatsApp como canal real, Instagram como alvo.
- **Resolvido por:** "Canal de publicação".

### I5 — Arquitetura de deploy: OCI vs GCP
- **Onde:** `DEPLOY.md` recomenda **OCI Free Tier** (VM ARM, nginx, systemd,
  encurtador Go+Postgres, `ci.yml`/`deploy.yml` via SSH/rsync). Todos os outros
  (`DEPLOY_GCP.md`, `COLETA.md`, `ENTIDADES.md`, `ANALISE_ESTATICA.md`,
  `MELHORIAS_28_06.md`) assumem **GCP** (Cloud Run + BigQuery, `deploy-gcp.yml`).
- **Por que importa:** dois runbooks contraditórios; menção a Postgres que não
  existe na realidade (tudo é BigQuery). Risco real de seguir o caminho errado.
- **Correção:** arquivar `DEPLOY.md` em `docs/legado/` com aviso no topo, ou
  removê-lo. GCP é a verdade. Se OCI for um plano B real, vira uma seção "Alternativa
  de hospedagem" dentro do doc de deploy único.
- **Resolvido por:** "Deploy".

### I6 — Modelo de "categoria": singular vs plural
- **Onde:** `ENTIDADES.md` define `BUSCA.categoria string` (singular, "rótulo
  opcional"). `BUSCAS_AGENDADAS.md` e `TESTES_DESCOBRIR.md` usam `categorias[]`
  (array, filtro por OR, múltiplas). `BACKLOG.md` (28/06) confirma "Modelo Busca
  expandido: ... categorias[]".
- **Por que importa:** é divergência de **schema**, não cosmética. Quem
  implementar contra o ENTIDADES quebra o que a UI espera.
- **Correção:** atualizar o ER para `categorias ARRAY<STRING>`. Quando o ER passar
  a ser **gerado do schema** (ver `02`), isso deixa de divergir por construção.
- **Resolvido por:** "Categorias".

---

## 🟡 Média

### I7 — Campo `fontes` na Busca: existe ou está por vir?
- **Onde:** `BUSCAS_AGENDADAS.md` lista `fontes` como **o que falta** ("Campo
  `fontes` — para saber quais tipos de resultado..."). `TESTES_DESCOBRIR.md` e o
  `BACKLOG.md` (28/06: "Modelo Busca expandido: fontes[]") tratam como **existente**.
  `ENTIDADES.md` **não tem** o campo.
- **Por que importa:** o ER (que deveria ser a referência de schema) está atrás da
  implementação.
- **Correção:** adicionar `fontes ARRAY<STRING>` e `config_novos`/`dias_janela` ao
  ER. Idem: gerar o ER do schema elimina o drift.

### I8 — Alertas: configuração por env var vs por banco
- **Onde:** `MANUAL.md` e `DEPLOY_GCP.md` documentam alertas via **env vars**
  (`ALERTAS_TELEGRAM_CHAT_ID`, `ALERTAS_THRESHOLD`...). `MELHORIAS_28_06.md` (item 7)
  e `ANALISE_ESTATICA.md` (`alertas.go` com "/alertas config, teste, update")
  apontam migração para **config no banco** (tabela `alertas_config`).
- **Por que importa:** dois modelos mentais para a mesma feature; o operador pode
  configurar no lugar errado.
- **Correção:** documentar o estado atual (env var) como vigente e a config por
  banco como 🔮. Quando migrar, atualizar `MANUAL`/`DEPLOY` juntos.

### I9 — Alertas de "produtos novos": ligados ou desligados?
- **Onde:** `MANUAL.md` afirma "Também notifica produtos novos detectados (🆕)".
  `MELHORIAS_28_06.md` e `BACKLOG.md` (28/06) dizem **desabilitados** ("preparado
  para config futura").
- **Por que importa:** expectativa errada do operador sobre o que chega no Telegram.
- **Correção:** ajustar `MANUAL.md` para "preparado, desabilitado por ora".
- **Resolvido por:** "Alertas de produtos novos".

### I10 — Janela de "novos": 7 dias fixo vs configurável
- **Onde:** `MANUAL.md` diz "últimos **7 dias**" (fixo). `BUSCAS_AGENDADAS.md` e
  `TESTES_DESCOBRIR.md` (#35) usam `dias_janela` **configurável** (1–30, default 7).
- **Por que importa:** menor, mas o teste #35 falharia contra a doc do MANUAL.
- **Correção:** MANUAL passa a "padrão 7 dias, configurável por busca".

### I11 — Contagem de testes de frontend
- **Onde:** `ANALISE_ESTATICA.md` diz frontend **26** (SeletorGrupo 10 +
  CandidateCard 16). `REFATORACAO.md` diz **34** ("npx vitest run ✔ 34"). `BACKLOG.md`
  (28/06) diz **62** ("20 testes Vitest... 62 total").
- **Por que importa:** três números para a mesma métrica; nenhum é claramente o
  vigente. Sintoma de número escrito à mão que envelhece.
- **Correção:** **não** versionar contagem de testes em prosa. Gerar do relatório
  do Vitest na CI (badge ou linha em `ANALISE_ESTATICA`), ou simplesmente remover o
  número e dizer "ver job `test-web` na CI".

---

## 🟢 Baixa / cosmético

### I12 — Nomes de componentes em trânsito (CandidateCard / CardOportunidade / ProductCard)
- **Onde:** `MELHORIAS_28_06.md` e `ANALISE_ESTATICA.md` ainda citam `CandidateCard`
  e `CardOportunidade`; a decisão (e o `BACKLOG.md` 28/06) é **`ProductCard`
  unificado** com 3 layouts. `MELHORIAS` chega a sugerir manter `CandidateCard` como
  alias por compatibilidade de testes.
- **Por que importa:** pouco — mas a doc de qualidade lista testes de um componente
  que virou layout de outro.
- **Correção:** ao consolidar, falar de `ProductCard` e citar os layouts; mencionar
  os nomes antigos só uma vez, como "ex-".

### Observações adicionais (não-bloqueantes)
- **`openapi.yaml`** lista `/api/onboarding/{status,termos,shopee,...}`; o
  `BACKLOG.md` fala em 6 endpoints incluindo `/validar` e `/excluir-conta`. Conferir
  se o spec cobre todos — quando o spec for a SSOT e for validado na CI, isso
  resolve sozinho.
- **`MANUAL.md`** ainda descreve a aba **Quadro** ("persiste os cards") nas dúvidas
  frequentes — mesmo problema do I3, em outro arquivo.
- **`COLETA.md`** menciona salvar buscas "na tela de Curadoria (campo nome do
  perfil + cron)"; pós-unificação a gestão de buscas vive em `/lojas` e os atalhos na
  Descobrir. Atualizar a referência de tela.

---

## Como usar esta auditoria

1. Itens 🔴 entram como **tarefas de documentação** no backlog (`03`), epic
   `docs-migration`.
2. Ao migrar cada doc para a estrutura nova (`02`), resolver os itens que tocam
   aquele arquivo e marcar aqui como `[x]`.
3. Os itens de **drift de schema** (I6, I7) e **métrica escrita à mão** (I11)
   só "ficam resolvidos de verdade" quando a informação passa a ser **gerada** —
   por isso o `02` prioriza geração automática de ER e de números da CI.
