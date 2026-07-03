# Backlog como Código — Garimpei

Problema atual: `BACKLOG.md` e `MELHORIAS_28_06.md` são listas longas, manuais, com
status misturado (✅/❌/🟡), datas no título e itens repetidos entre sessões. Difícil
de atualizar, fácil de envelhecer. Você trabalha sozinho e quer **um backlog dentro
do repo que seja fácil de mexer e que gere, a cada build, um quadro com o estado
atual e o próximo sprint**.

Solução: **um arquivo por tarefa em YAML**, validado por schema, com um **gerador
em Go** que produz, no build, (a) um quadro Kanban e (b) um roadmap Now/Next/Later.

> Por que não GitHub Projects? Ele é ótimo, mas vive fora do repo (não versiona com
> o código, não entra no PR, não gera artefato no build). Como você quer
> *in-repo + gerado no build*, arquivos YAML + gerador ganham. Dá para sincronizar
> com Issues depois, se quiser (seção 6).

---

## 1. Por que "tarefa por arquivo" e não um YAML gigante

- **Diffs limpos:** mexer numa tarefa muda um arquivo, não um monólito.
- **Sem conflito de merge** quando várias tarefas mudam.
- **Histórico git por tarefa** (quando nasceu, quando virou done).
- O custo (muitos arquivos) é resolvido pelo gerador, que agrega tudo.

---

## 2. Estrutura no repositório

```
backlog/
  schema.json               ← JSON Schema que valida cada tarefa (CI)
  epics.yaml                ← lista de epics (id, título, cor)
  tasks/
    T-0001-modelo-entidades.yaml
    T-0002-conversoes-persistir.yaml
    T-0003-docs-migration.yaml
    ...
  arquivo/                  ← tarefas done há mais de N dias (mantém tasks/ enxuto)
cmd/
  gen-board/                ← gerador Go (board + roadmap)
docs/gerado/
  BOARD.md                  ← quadro Kanban (gerado, não editar)
  ROADMAP.md                ← Now/Next/Later (gerado, não editar)
  board.svg                 ← versão visual do quadro (gerado)
```

## 3. Formato de uma tarefa

`backlog/tasks/T-0002-conversoes-persistir.yaml`:

```yaml
id: T-0002
titulo: Persistir conversões da Shopee no BigQuery
epic: conversoes
status: next            # backlog | next | doing | review | done | blocked
prioridade: alta        # alta | media | baixa
estimativa: M           # P | M | G  (ou pontos: 1,2,3,5,8)
sprint: 2026-S27        # sprint atual (ano-Sxx) ou vazio
valor: >
  Hoje /api/conversoes/reais consulta a Shopee on-demand. Persistir fecha o laço
  receita↔curadoria e destrava análise histórica (ver docs 07-dados-e-ia).
criterios:
  - Tabela `conversoes` populada via job periódico (validatedReport)
  - Idempotência por conversion_id (sem duplicar)
  - Cruzamento sub_id ↔ publicacao.detalhe testado
depende_de: [T-0001]    # ids de outras tarefas
tags: [backend, bigquery, atribuicao]
criada_em: 2026-06-27
atualizada_em: 2026-06-27
```

Campos mínimos obrigatórios: `id`, `titulo`, `status`, `prioridade`. O resto é
opcional — manter baixo o atrito de criar tarefa.

`backlog/epics.yaml`:

```yaml
- id: conversoes
  titulo: Fechar o ciclo de atribuição
  cor: "#2e7d32"
- id: multi-tenant
  titulo: SaaS multi-tenant + LGPD
  cor: "#1565c0"
- id: docs-migration
  titulo: Migração da documentação v2
  cor: "#6a1b9a"
- id: dados-ia
  titulo: Análises e IA integrada
  cor: "#ef6c00"
```

## 4. O gerador (`cmd/gen-board`)

Um binário Go (~150 linhas, fica no padrão do projeto) que:

1. Lê `epics.yaml` + todos os `tasks/*.yaml`.
2. Valida contra `schema.json` (falha o build se inválido — campo errado, status
   desconhecido, `depende_de` apontando para id inexistente).
3. Emite três saídas em `docs/gerado/`:

**`BOARD.md`** — Kanban do sprint atual (colunas = status):

```markdown
# Quadro — Sprint 2026-S27  (gerado em 2026-06-27)

| 📋 Backlog | ⏭️ Next | 🔨 Doing | 👀 Review | ✅ Done |
|---|---|---|---|---|
| T-0005 Scroll infinito (M) | **T-0002** Persistir conversões (M) | T-0003 Migração docs (G) | — | T-0001 Modelo entidades (M) |
| T-0007 Modal produto (M) | T-0009 ScopedStore por owner_uid (G) | | | T-0010 Favoritos sync (M) |

**Bloqueadas:** T-0008 (aguarda decisão de provedor WhatsApp)
**WIP em Doing:** 1/2 ✅
```

**`ROADMAP.md`** — Now/Next/Later (horizontes, sem datas — alinhado ao `04`):

```markdown
# Roadmap  (gerado em 2026-06-27)

## 🔵 Now (sprint atual — máx. 5)
- **Persistir conversões** · conversoes · M
- **ScopedStore por owner_uid** · multi-tenant · G

## 🟡 Next (1–3 sprints)
- Scroll infinito na busca · M
- Modal de detalhes do produto · M
- Alertas configuráveis por usuário · G

## ⚪ Later (radar)
- IA integrada (assistente de insights) · ver doc 07
- Billing por tenant
- WhatsApp via Evolution API (avaliar)
```

**`board.svg`** — mesma informação do BOARD.md em SVG, para colar no README do
repo ou no site de docs (renderiza em qualquer lugar, inclusive no GitHub).

### Regras úteis que o gerador aplica

- **Limite de WIP:** se `Doing` tiver mais que N (config, ex. 2), emite ⚠️ no board
  e (opcional) falha o build — força foco, no espírito do Kanban do `MODELO.md`.
- **Now ≤ 5 itens:** se passar, avisa (false precision / espalhamento).
- **Sprint atual:** lido de `backlog/sprint.txt` (uma linha, ex. `2026-S27`) ou da
  task com maior `sprint`. Tarefas `done` fora do sprint atual não entram no board.
- **Métricas no rodapé:** contagem por status, throughput do último sprint
  (quantas foram para `done`), idade média das `doing` (sinaliza tarefa travada).

## 5. Plugar no build

No `Makefile` (já referenciado em `02`):

```makefile
docs-board:
	go run ./cmd/gen-board
	@echo "board e roadmap gerados em docs/gerado/"
```

Na CI, junto do `docs-check`:

```yaml
- name: Validar e gerar backlog
  run: |
    go run ./cmd/gen-board
    git diff --exit-code docs/gerado/BOARD.md docs/gerado/ROADMAP.md \
      || (echo "Board desatualizado: rode 'mise run docs-board' e commite"; exit 1)
```

Efeito: **a cada build, o quadro reflete o YAML.** Para mover uma tarefa de coluna,
você edita uma linha (`status: next` → `status: doing`) e commita — o board
atualiza sozinho. Sem manter listas à mão.

## 6. Fluxo diário (solo dev)

```
Nova ideia            → cria T-XXXX.yaml com status: backlog (30s)
Começou a tarefa      → status: doing  (1 linha)
Terminou              → status: done   + atualizada_em
Planeja o sprint      → seta sprint: e status: next nos escolhidos
Build/commit          → BOARD.md e ROADMAP.md regeneram
```

Atalho opcional: um script `mise run backlog:create "Titulo" --epic dados-ia` que cria o
YAML pré-preenchido com o próximo id e a data — reduz o atrito a um comando.

## 7. Sincronização com GitHub (opcional, Later)

Quando entrar mais gente ou quiser feedback dos beta testers:
- **Espelhar para Issues:** um GitHub Action lê `tasks/*.yaml` e cria/atualiza
  Issues com label = epic e status (via Projects v2 API). O YAML continua sendo a
  fonte; as Issues são a vitrine pública.
- **Feedback dos beta testers vira tarefa:** uma Issue com label `feedback` é
  convertida num `T-XXXX.yaml` (script). Mantém o backlog como SSOT e abre um canal
  de entrada — exatamente o "caminho de retorno curto" que o `DEPLOY.md` sugere.

## 8. Migração do backlog atual

1. Quebrar `BACKLOG.md` + os "Itens para próxima sessão" do `MELHORIAS_28_06.md` em
   tasks YAML. Os "✅ Resolvido nesta sessão" viram `status: done` em `arquivo/` (ou
   nem migram — o git já é o histórico).
2. Os itens 🔴 do `01_INCONSISTENCIAS.md` viram tasks do epic `docs-migration`.
3. Escrever `schema.json` + `cmd/gen-board`, ligar na CI.
4. Apagar `BACKLOG.md` da raiz de `docs/` (ou deixar um stub apontando para
   `docs/gerado/BOARD.md`).
