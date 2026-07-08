# Impacto da Migração UI — Qualidade de Código e Dívida Técnica

**Data:** 2026-07-05
**Escopo:** Adoção de Bits UI + Design Tokens no frontend Svelte 5 + Tema Garimpo Quente
**Referência:** T-0037, ADR-0001

---

## Resumo Executivo

A migração introduziu uma biblioteca de componentes padronizada (10 primitivos + compostos Bits UI), formalizou design tokens, e eliminou 100% dos hex colors hardcoded. Além disso, revitalizou toda a estética visual do sistema aplicando o tema **Garimpo Quente** (cores mais quentes, terracota, ouro e sépia para o dark mode). O resultado é uma base de código mais consistente, esteticamente apurada, acessível, e protegida por tooling automatizado contra regressão.

---

## Dívida Técnica — Antes vs Depois

### Antes da migração

| Problema | Impacto | Severidade |
|---|---|---|
| 50 hex colors hardcoded em componentes | Visual inconsistente, impossível mudar paleta sem buscar em 15+ arquivos | Alta |
| Zero ARIA/keyboard nav em interações (select, modal, tabs) | Inacessível para ~15% dos usuários (WCAG fail) | Alta |
| 6 utility classes de feedback (`.msg-erro`, `.msg-sucesso`) misturadas com lógica | Duplicação, inconsistência visual entre páginas | Média |
| Cada componente reinventava botões, inputs, badges com estilos inline | ~75 `<button>` com 75 estilos diferentes, zero padronização | Alta |
| Tokens informais em `:root` sem proteção | Nada impedia developers de usar `#7a5a1e` ao invés de `var(--ouro-escuro)` | Média |
| Zero type-checking nos componentes | Props erradas passavam silenciosamente | Média |
| Sem auditoria de cobertura | Impossível medir progresso de padronização | Baixa |

### Depois da migração

| Melhoria | Métrica | Proteção |
|---|---|---|
| Hex colors eliminados | 50 → **0** | Stylelint `color-no-hex` **bloqueia CI** |
| Design tokens formalizados | 82 LOC em `tokens.css` | Arquivo único, importado globalmente |
| Componentes acessíveis (ARIA + keyboard) | 5 compostos Bits UI em uso | axe-core + @axe-core/playwright |
| Props type-checked | 0 erros em `components/ui/` | svelte-check no CI |
| Dead code zero | knip "Excellent, no issues" | knip no CI |
| Cobertura auditável | `npm run audit:ui` | mise task + relatório automático |

---

## Métricas Quantitativas

### Redução de padrões legados

| Padrão | Antes | Depois | Redução | Proteção |
|---|---|---|---|---|
| Hex colors hardcoded | 50 | **0** | −100% | Stylelint bloqueia |
| `<button>` inline | 75 | 51 | −32% | Audit reporta |
| `<select>` nativo | 8 | 5 | −37% | Audit reporta |
| `.msg-erro/sucesso` class | 6 | 1 | −83% | Audit reporta |
| Badge utility class | 25 | 16 | −36% | Audit reporta |
| `.btn` utility class | ~40 | 26 | −35% | Audit reporta |
| Compostos Bits UI sem uso | 5/5 | **0/5** | −100% | Audit reporta |

### Cobertura de componentes padronizados

| Componente UI | Consumidores | Padrão substituído |
|---|---|---|
| Button | 9 | `<button class="btn-...">` ad-hoc |
| Alert | 5 | `.msg-erro`, `.msg-sucesso`, divs inline |
| Card | 4 | `.card`, `.painel`, divs com border/radius |
| Input | 3 | `<input>` com estilos scoped repetidos |
| Badge | 3 | `.badge-*` utility classes |
| Select (Bits UI) | 3 | `<select>` nativo sem keyboard nav |
| Tabs (Bits UI) | 2 | TabBar custom sem ARIA |
| Dialog (Bits UI) | 1 | `confirm()` nativo sem estilo |
| DropdownMenu (Bits UI) | 1 | Botões separados editar/remover |
| Tooltip (Bits UI) | 1 | `title=` sem estilo acessível |

### Volume de mudança

| Métrica | Valor |
|---|---|
| Commits | 15 |
| Arquivos alterados | 57 |
| Linhas adicionadas | +2.824 |
| Linhas removidas | −717 |
| Saldo líquido | +2.107 (componentes + docs + tooling) |

---

## Qualidade de Código — Indicadores

### Consistência

| Indicador | Antes | Depois |
|---|---|---|
| Fonte de verdade para cores | Nenhuma (cada arquivo define) | `tokens.css` (82 LOC) |
| API de componentes | Ad-hoc (cada um com props diferentes) | Padronizada: `variant`, `size`, `...rest` |
| Estilo de hover/focus | Inconsistente (alguns sem, alguns com) | Padrão: 150ms ease, `--ouro` outline |
| Acessibilidade | Zero a11y checks | svelte-check warnings + axe-core |

### Manutenibilidade

| Indicador | Antes | Depois |
|---|---|---|
| Para mudar a cor de acento primária | Buscar em 15+ arquivos | Mudar 1 valor em `tokens.css` |
| Para adicionar um novo botão | Copiar CSS de outro componente | `<Button variant="primary">` |
| Para garantir keyboard nav num select | Implementar manualmente (~50 LOC) | `<Select>` (Bits UI cuida) |
| Para auditar padrões legados | Manualmente com grep | `npm run audit:ui` (10s) |
| Para detectar prop errada | Runtime (bug em produção) | svelte-check (compile-time) |

### Proteção contra regressão

| Guarda | O que protege | Quando roda |
|---|---|---|
| `stylelint color-no-hex` | Ninguém introduz hex em componentes | CI + `npm run lint:css` |
| `svelte-check` | Props incorretas, CSS morto, tipos | CI + `npm run check` |
| `knip` | Imports/exports não usados | CI + `npm run lint:dead` |
| `audit:ui --strict` | Hex colors (exit 1 se > 0) | CI |
| `eslint --max-warnings=0` | Unused vars, a11y warnings | CI + `npm run lint:js` |
| `vitest` (174 tests) | Lógica de negócio | CI + `npm run test:unit` |

---

## Dívida Técnica Residual

### Aceita (baixo risco, escopo limitado)

| Item | Razão para aceitar | Próximo passo |
|---|---|---|
| 51 `<button>` inline | Toggles estilizados e buttons em layouts multi-variante (ProductCard) onde `<Button>` não agrega valor | T-0038 avalia caso a caso |
| 28 `<input>` inline | BuscaUnificada autocomplete requer input customizado | Manter, documentar como exceção |
| 5 `<select>` nativo | Selects simples em forms (tipo canal, dias_janela) | T-0038 fase 2a |
| 85 erros svelte-check em src/ | Pre-existentes (tipos em libs, test fixtures) — 0 em `components/ui/` | Resolver incrementalmente |
| `title=` sem Tooltip em 9 arquivos | Tooltips informativos em spans/badges — acessibilidade já ok via texto | T-0038 fase 2b |

### Eliminada

| Item | Como foi resolvido |
|---|---|
| Hex colors hardcoded | 100% eliminados + Stylelint previne reintrodução |
| Ausência de type-checking | svelte-check no CI (0 erros em componentes UI) |
| Dead code pós-migração | knip + remoção manual (EmptyState, CSS morto) |
| Falta de documentação | ADR + docs/componentes.md + docs/linting.md |
| `confirm()` para ações destrutivas | Dialog acessível (canais) |
| Tabs sem ARIA | Bits UI Tabs com keyboard nav (lojas, publicacoes) |
| Select sem keyboard nav | Bits UI Select acessível (publicar, BuscaUnificada) |

---

## ROI da Migração

### Tempo investido
- ~1 sessão de desenvolvimento intensivo
- 15 commits, 57 arquivos tocados

### Retorno imediato
- **Acessibilidade**: 5 padrões de interação agora WCAG-compliant (Select, Tabs, Dialog, Tooltip, DropdownMenu)
- **Velocidade de desenvolvimento**: novo componente = `<Button variant="x">` ao invés de 20+ LOC de CSS
- **Confiança no refactor**: stylelint + svelte-check + audit pegam regressão antes do push
- **Onboarding**: `docs/componentes.md` documenta toda a API disponível

### Retorno de longo prazo
- **Mudança de paleta** agora é 1 arquivo (tokens.css) — não 50+ buscas
- **Novos desenvolvedores** não precisam reinventar buttons/modais
- **A11y compliance** é automática para quem usa os compostos
- **Bundle size** controlado — Bits UI tree-shakes, sem CSS-in-JS

---

## Unificação Descobrir + Lojas → Garimpar (/)

**Data:** 2026-07-09
**Referência:** T-0039

### Motivação

As páginas "Descobrir" (`/`) e "Lojas" (`/lojas`) compartilhavam grid de produtos, ações
(publicar, favoritar) e layout. A separação gerava navegação desnecessária e duplicação de
lógica de filtragem. A unificação simplifica a experiência: tudo em uma tela.

### Mudanças realizadas

| Aspecto | Antes | Depois |
|---|---|---|
| Páginas | 2 (`/` + `/lojas`) | **1** (`/` — "Garimpar") |
| Rota `/lojas` | Página dedicada | Removida (404) |
| Seleção de fonte | Navegar entre páginas | `ToggleGroup type="multiple"` com badges |
| Configuração (loja, buscas, alertas) | Dividida entre páginas | BuscaUnificada (integra filtros, lojas e agendamento) + `⚙️ Configuração` colapsável para alertas |
| NavDrawer | "Descobrir" + "Lojas" | "Garimpar" (link único) |
| `ListaProdutosLoja.svelte` | Componente ativo | **Deletado** (dead code) |

### Impacto em código

| Métrica | Valor |
|---|---|
| Linhas substituídas pelo ToggleGroup | ~75 (botões raw de seleção de fonte) |
| Componentes removidos | 1 (`ListaProdutosLoja`) |
| Componentes novos/estendidos | `ToggleGroup` (mode multiple + badges), `Collapsible` |
| Rotas removidas | 1 (`/lojas`) |
| Testes E2E atualizados | Sim (navegação e fluxos de loja migrados para `/`) |

### Backend

Nenhuma alteração. Os endpoints `/api/lojas`, `/api/buscas` e `/api/candidatos` continuam
inalterados — a mudança foi exclusivamente de apresentação no frontend.

---

## BuscaUnificada — Integração de filtros, lojas e agendamento

**Data:** 2026-07-12
**Referência:** T-0040

### Motivação

A página Garimpar usava três componentes separados para configurar buscas: `FilterBar` (filtros de comissão e keywords), `FormAdicionarLoja` (cadastro de lojas monitoradas) e `GerenciarBuscas` (palavras-chave, fontes, janela de tempo). Essa fragmentação obrigava o usuário a navegar entre seções e gerava acoplamento de estado entre componentes. A unificação em `BuscaUnificada` consolida toda a experiência de configuração em um formulário coeso.

### Mudanças realizadas

| Aspecto | Antes | Depois |
|---|---|---|
| Componentes de busca/config | 3 (`FilterBar`, `FormAdicionarLoja`, `GerenciarBuscas`) | **1** (`BuscaUnificada`) |
| Arquivo da página `+page.svelte` | ~150 linhas com lógica de orquestração | ~50 linhas (grid + BuscaUnificada) |
| Gerenciamento de estado | Disperso entre componentes | `.svelte.js` module (`criarEstado`, `criarDerivados`, `criarHandlers`) |
| Lógica pura | Embutida nos templates | `busca-unificada-logic.js` (testável isoladamente) |
| Seleção de lojas | Formulário dedicado, uma loja por vez | Seletor plural multi-marketplace integrado |
| Filtros avançados | Colapsável separado | Seção colapsável dentro do componente |
| Buscas salvas | Lista em GerenciarBuscas | Chips clicáveis inline |

### Padrão `.svelte.js`

Adotado para separar lógica reativa do template:

- **`BuscaUnificada.svelte.js`** — estado reativo com runes (`$state`, `$derived`), handlers de eventos
- **`busca-unificada-logic.js`** — funções puras (sem dependência de Svelte): `configToPayload`, `payloadToConfig`, `gerarResumo`, `contarFiltrosAtivos`, `cronLabel`, `gerarLabelBusca`

ESLint atualizado: arquivos `.svelte.js` recebem `max-lines-per-function: 150` (vs 50 padrão) dado que encapsulam state factories.

### Backend — campos adicionados

| Entidade/Endpoint | Campo novo | Tipo |
|---|---|---|
| `Busca` (entity) | `ComissaoMin` | `decimal?` |
| `Busca` (entity) | `VendasMin` | `int?` |
| `Busca` (entity) | `Categorias` | `string[]?` |
| `Busca` (entity) | `Fontes` | `string[]?` |
| `POST /api/buscas` | `shop_ids[]` | array de IDs |
| `POST /api/buscas` | `comissao_min` | decimal |
| `POST /api/buscas` | `vendas_min` | int |
| `POST /api/buscas` | `categorias[]` | array de strings |
| `POST /api/buscas` | `fontes[]` | array de strings |
| `POST /api/buscas` | `marketplaces` | string |

### Impacto em código

| Métrica | Valor |
|---|---|
| Componentes removidos | 3 (`FilterBar`, `FormAdicionarLoja`, `GerenciarBuscas`) |
| Componentes criados | 1 (`BuscaUnificada` + `.svelte.js` + logic) |
| Testes passando | 174 |
| Linhas na página principal | ~50 (antes ~150) |

---

## Próximos Passos (T-0038)

1. Migrar os 5 `<select>` nativos restantes → `<Select>`
2. Expandir Dialog para confirmações em lojas e configurar
3. Adicionar Tooltip nos selos informativos (ProductCard, ScoreMeter)
4. Avaliar se os 51 buttons restantes justificam migração ou são exceções documentadas
