# Linting & Qualidade — Frontend Garimpo

## Visão Geral das Ferramentas

| Ferramenta | O que verifica | Comando |
|---|---|---|
| `svelte-check` | Tipos, props incorretos, CSS unused, a11y (compile-time) | `npm run check` |
| `stylelint` | CSS duplicado, hex hardcoded, propriedades inválidas | `npm run lint:css` |
| `eslint` + `eslint-plugin-svelte` | Code smells, complexidade, XSS, a11y | `npm run lint:js` |
| `knip` | Dead code, imports e exports não usados | `npm run lint:dead` |
| `npm audit` | Vulnerabilidades conhecidas (CVEs) em dependências | `npm audit --audit-level=high` |
| `Semgrep` | Security SAST (XSS, injection, secrets, patterns) | `npm run lint:security` |
| `axe-core` / `vitest-axe` | Acessibilidade runtime (WCAG 2.1 AA) em testes unitários | `npm run test:unit` |
| `@axe-core/playwright` | Acessibilidade e2e contra DOM real renderizado | `npm run test` |

## Scripts Disponíveis

```bash
# Verifica tipos, props e a11y compile-time (todos os .svelte)
npm run check

# Watch mode para desenvolvimento
npm run check:watch

# Lint CSS (hex colors, duplicatas, formatação)
npm run lint:css

# Lint JS/Svelte (code smells, complexity, XSS warnings, a11y)
npm run lint:js

# Security SAST (Semgrep — patterns para XSS, injection, secrets)
npm run lint:security

# Detectar dead code e imports não usados
npm run lint:dead

# Vulnerabilidades em dependências (CVEs high/critical)
npm audit --audit-level=high

# Todos os linters juntos (check + css + js)
npm run lint

# Auditoria de cobertura da biblioteca UI
npm run audit:ui

# Testes unitários com axe-core
npm run test:unit

# Testes e2e com acessibilidade
npm run test
```

## Detalhamento por Ferramenta

### svelte-check

Analisa estaticamente todos os arquivos `.svelte`:

- **Tipos**: verifica que props passadas a componentes existem e têm o tipo correto
- **CSS unused**: detecta selectores CSS scoped que não correspondem a nenhum elemento
- **Acessibilidade**: warnings do compilador Svelte para padrões a11y (missing alt, missing label, etc.)
- **Props incorretos**: detecta props que não existem em um componente importado
- **Bindings**: valida que `bind:value` é usado em props `$bindable`

**Config**: `tsconfig.json` no root de `web/` estende `.svelte-kit/tsconfig.json`.

### Stylelint

Configurado em `.stylelintrc.json`:

- **`color-no-hex: true`** — bloqueia cores hex hardcoded em `.svelte` e `.css`. Exceção: `tokens.css` e `app.css` (onde os tokens são definidos).
- **`declaration-block-no-duplicate-properties`** — detecta propriedades CSS duplicadas
- **`selector-pseudo-class-no-unknown`** — com exceção para `:global` (Svelte)
- **Custom syntax**: `postcss-html` para parsear `<style>` dentro de `.svelte`

**Escopo expandido**: agora lint roda em `src/**/*.svelte` e `src/**/*.css` (antes era só `src/routes/`).

#### Proteção de tokens

A regra `color-no-hex` garante que **nenhum componente** pode usar cores hex diretamente — forçando o uso de `var(--token)`. Isso protege a consistência do design system após a migração.

### ESLint + eslint-plugin-svelte

Usa `flat/recommended` do eslint-plugin-svelte que inclui:

- **Regras a11y**: missing labels, interactive elements sem keyboard access, role conflicts
- **Svelte-specific**: no-dupe-style-properties, require-each-key (desativada), no-at-html-tags (warn)
- **JS quality**: no-unused-vars, no-undef (desativada para runes)
- **Complexidade**: `complexity` (max 15), `max-depth` (max 4), `max-lines-per-function` (max 80), `max-params` (max 4)
- **Svelte blocks**: `svelte/max-lines-per-block` (script: 120, style: 150, template: 250)
- **XSS**: `svelte/no-at-html-tags` como warning (sinaliza uso de `{@html}` para review)

### Semgrep (Security SAST)

Análise estática de segurança via pattern matching. Roda no CI via container Docker e localmente via `npm run lint:security`.

Detecta:
- XSS (innerHTML, {@html} com dados não sanitizados)
- Prototype pollution
- Open redirects
- Hardcoded secrets/credentials
- Insecure crypto patterns
- SQL/NoSQL injection patterns

**CI**: Job `security` roda em paralelo com os outros jobs.
**Local**: `npm run lint:security` (requer `semgrep` instalado ou usa `npx`).

### npm audit (Dependency Vulnerabilities)

Verifica se as dependências têm CVEs conhecidos reportados no npm advisory database.

- `--audit-level=high` — reporta apenas high e critical
- Roda no CI antes do build
- Local: `npm audit` para ver tudo, `npm audit fix` para corrigir automáticas

### Complexity Rules (ESLint)

As regras de complexidade funcionam como **guardrails preventivos** — não bloqueiam imediatamente mas sinalizam código que precisa de refactor:

| Regra | Limite | O que mede |
|---|---|---|
| `complexity` | max 15 | Cyclomatic complexity (branches) |
| `max-depth` | max 4 | Nível de nesting (if dentro de if...) |
| `max-lines-per-function` | max 80 | Tamanho da função |
| `max-params` | max 4 | Parâmetros de função |
| `svelte/max-lines-per-block` | script:120, style:150, template:250 | Tamanho dos blocos Svelte |

O `--max-warnings=6` no CI permite os 6 hotspots conhecidos mas bloqueia qualquer warning novo.

### knip (Dead Code)

Detecta:
- Exports não importados por ninguém
- Dependências listadas no package.json mas não usadas
- Arquivos órfãos sem importadores

### axe-core (Acessibilidade Runtime)

Duas formas de uso:

#### Em testes unitários (vitest-axe)

```javascript
import { render } from '@testing-library/svelte';
import { axe } from 'vitest-axe';
import Button from '$lib/components/ui/Button.svelte';

test('Button não tem violações a11y', async () => {
  const { container } = render(Button, { props: { children: () => {} } });
  const results = await axe(container);
  expect(results).toHaveNoViolations();
});
```

#### Em testes e2e (@axe-core/playwright)

```javascript
import { test, expect } from '@playwright/test';
import AxeBuilder from '@axe-core/playwright';

test('página Descobrir sem violações a11y', async ({ page }) => {
  await page.goto('/');
  const results = await new AxeBuilder({ page })
    .withTags(['wcag2a', 'wcag2aa'])
    .analyze();
  expect(results.violations).toHaveLength(0);
});
```

## Integração com CI

O lint completo roda no CI via GitHub Actions (job `frontend`). Ordem de execução:

```yaml
# .github/workflows/ci.yml — job: frontend
steps:
  - npm ci
  - npm audit --audit-level=high     # dependency vulnerabilities
  - npx svelte-kit sync              # generate types
  - npm run check                    # svelte-check (tipos + a11y)
  - npm run build                    # build production
  - npm run lint:css                 # stylelint (protege tokens)
  - npm run lint:js                  # eslint (complexity + XSS + quality)
  - mise run check:ui-coverage -- --strict  # bloqueia se hex hardcoded
  - npx vitest run                   # unit tests + axe-core + contrast
  - npx playwright test              # e2e + @axe-core/playwright

# .github/workflows/ci.yml — job: security
steps:
  - semgrep scan --config auto --config p/javascript --error  # SAST
```

### Mise Tasks

O projeto usa [mise](https://mise.jdx.dev) para orquestrar tarefas. Tasks relevantes:

| Task | Comando | Comportamento no CI |
|---|---|---|
| `check:ui-coverage` | `mise run check:ui-coverage` | Relatório de cobertura UI. Com `--strict`: **falha se hex colors > 0** |
| `check:docs-drift` | `mise run check:docs-drift` | **Falha se docs/gerado/ stale**, sync quebrado, ou links quebrados no index |
| `check:file-size` | `mise run check:file-size` | Bloqueia arquivos > 400 linhas |
| `check:config-consistency` | `mise run check:config-consistency` | Verifica configs sincronizadas |
| `checks` | `mise run checks` | Roda **todos** os checks (contracts, API, config, schema, docs-drift, etc.) |

### Modos da auditoria UI

```bash
# Relatório completo (não bloqueia CI)
mise run check:ui-coverage

# Modo strict — falha se hex colors encontrados (usado no CI)
mise run check:ui-coverage -- --strict
```

O modo `--strict` bloqueia o merge se algum componente introduzir hex colors hardcoded. Os demais indicadores (buttons inline, utility classes) são informativos — servem para medir progresso sprint a sprint.

## Fluxo de Desenvolvimento

Ao trabalhar nos componentes UI:

1. **`npm run check:watch`** — feedback contínuo de tipos e a11y
2. **Ao commitar** — rodar `npm run lint` (check + css + js)
3. **Antes de push** — `npm run ci:local` (check + css + js + knip + build + unit tests)
4. **Medir progresso** — `npm run audit:ui`

### `npm run ci:local` — Validação local completa

Executa todos os checks que o CI faria, na mesma ordem:

```bash
npm run ci:local
# Equivale a: check → lint:css → lint:js → lint:dead → build → test:unit
```

Rode **sempre antes de push**. Se passa local, passa no CI.

## O que cada ferramenta pega na migração

| Problema | Ferramenta que detecta |
|---|---|
| Hex color hardcoded num componente | Stylelint (`color-no-hex`) |
| Prop inválida passada a `<Button>` | svelte-check |
| CSS selector sem match (sobra de migração) | svelte-check (css_unused_selector) |
| Componente sem keyboard nav | eslint-plugin-svelte (a11y rules) |
| Violação WCAG no DOM renderizado | axe-core (unit) / @axe-core/playwright (e2e) |
| Import de componente deletado | svelte-check + build failure |
| Export sem consumidor (legacy) | knip |
| Prop sem `$bindable` usada com `bind:` | svelte-check |
| Compostos Bits UI não adotados | `npm run audit:ui` |
| `<select>` nativo ao invés de `<Select>` | `npm run audit:ui` |
| Padrões reimplementados (modais, tabs) | `npm run audit:ui` |
| Utility classes legadas (.btn, .badge) | `npm run audit:ui` |
| **Função muito complexa (cyclomatic > 15)** | ESLint `complexity` |
| **Script/style block muito longo** | ESLint `svelte/max-lines-per-block` |
| **XSS via `{@html}` não sanitizado** | ESLint `svelte/no-at-html-tags` + Semgrep |
| **Dependência com CVE high/critical** | `npm audit` |
| **Injection patterns, secrets hardcoded** | Semgrep SAST |

## Auditoria de Cobertura UI (`npm run audit:ui`)

Script dedicado que mede o progresso da migração para a biblioteca de componentes:

```bash
npm run audit:ui
```

Gera relatório com:
- **Componentes adotados** — quantas vezes cada componente UI é importado
- **Compostos Bits UI não usados** — Dialog, Select, Tabs, Tooltip, DropdownMenu disponíveis mas sem consumidores
- **Elementos nativos com equivalente** — `<button>`, `<input>`, `<select>` que poderiam ser `<Button>`, `<Input>`, `<Select>`
- **Padrões reimplementados** — modais, tabs, tooltips, dropdowns que reimplementam o que Bits UI oferece
- **Utility classes legadas** — `.badge`, `.msg-erro`, `.btn` que deveriam ser componentes
- **Hex colors** — cores hardcoded (meta: zero)

### Exemplo de output

```
✗ COMPOSTOS BITS UI DISPONÍVEIS MAS NÃO USADOS
  ⚠  Dialog         → 0 consumidores
  ⚠  Select         → 0 consumidores
  ⚠  Tabs           → 0 consumidores

⚡ ELEMENTOS NATIVOS COM EQUIVALENTE UI
  <button> inline:  57  (equivalente: <Button>)
  <input> inline:   28  (equivalente: <Input>)
  <select> nativo:   8  (equivalente: <Select>)

── RESUMO ──
  Elementos nativos com equivalente:  93
  Utility classes legadas:            49
  Hex colors hardcoded:                0
```

### Quando rodar

- **CI (automático)**: `mise run check:ui-coverage -- --strict` roda em todo PR e push para main
- **Antes de cada PR de migração** — para medir progresso
- **Ao planejar a próxima sprint** — para priorizar quais componentes atacar
- **Após cada fase concluída** — para atualizar o score de cobertura
