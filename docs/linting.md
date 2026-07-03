# Linting & Qualidade — Frontend Garimpo

## Visão Geral das Ferramentas

| Ferramenta | O que verifica | Comando |
|---|---|---|
| `svelte-check` | Tipos, props incorretos, CSS unused, a11y (compile-time) | `npm run check` |
| `stylelint` | CSS duplicado, hex hardcoded, propriedades inválidas | `npm run lint:css` |
| `eslint` + `eslint-plugin-svelte` | Code smells JS/Svelte, regras a11y, no-unused-vars | `npm run lint:js` |
| `knip` | Dead code, imports e exports não usados | `npm run lint:dead` |
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

# Lint JS/Svelte (code smells, unused vars, a11y rules)
npm run lint:js

# Detectar dead code e imports não usados
npm run lint:dead

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
- **Svelte-specific**: no-dupe-style-properties, require-each-key (desativada), no-at-html-tags
- **JS quality**: no-unused-vars, no-undef (desativada para runes)

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

O lint completo roda no CI via mise. Ordem recomendada:

```yaml
steps:
  - npm run check        # svelte-check (tipos + a11y compile-time)
  - npm run lint:css     # stylelint (protege tokens)
  - npm run lint:js      # eslint
  - npm run build        # build production
  - npm run test:unit    # vitest + axe-core
  - npm run test         # playwright + @axe-core/playwright
```

## Fluxo de Desenvolvimento

Ao trabalhar nos componentes UI:

1. **`npm run check:watch`** — feedback contínuo de tipos e a11y
2. **Ao commitar** — rodar `npm run lint` (check + css + js)
3. **Antes de PR** — `npm run lint && npm run build && npm run test:unit`

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

- **Antes de cada PR de migração** — para medir progresso
- **Ao planejar a próxima sprint** — para priorizar quais componentes atacar
- **Após cada fase concluída** — para atualizar o score de cobertura
