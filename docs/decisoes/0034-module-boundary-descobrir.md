# ADR-0034: Module Boundary — Descobrir

**Status:** Aceita  
**Data:** 2026-07-13

## Contexto

O módulo Descobrir (BuscaEngine + Omnibox + lógica de busca) cresceu para 13 arquivos JS + 5 componentes Svelte. Sem boundaries, qualquer código pode importar internals diretamente — criando acoplamento acidental que quebra testes e dificulta refatorações.

## Decisão

Implementar **logical boundary** via barrel (`$lib/descobrir/index.js`) + **ESLint enforcement** (`no-restricted-imports`).

### Arquitetura

```
src/lib/descobrir/
  index.js              ← API pública (barrel)

src/lib/
  busca-engine.svelte.js    ← internal (classe + OTel spans)
  busca-engine-state.js     ← internal
  busca-engine-effects.js   ← internal
  busca-engine-omnibox.js   ← internal
  busca-engine-lojas.js     ← internal
  busca-engine-persistencia.js ← internal
  busca-config.js           ← internal (lê rules/busca-rules.json)
  descobrir-logic.js        ← internal
  descobrir.js              ← internal (API calls)
  busca-unificada-logic.js  ← internal
  omnibox-intencao.js       ← internal
  omnibox-parser.js         ← internal
  omnibox-sugestoes.js      ← internal
```

### Regras

1. **Código externo** (rotas, outros módulos) importa SOMENTE de `$lib/descobrir`
2. **Código interno** (os 13 arquivos acima + componentes do módulo) pode se referenciar livremente por path relativo
3. **Testes** do módulo (`src/tests/busca-*`, `src/tests/omnibox-*`, `src/tests/descobrir-*`) podem importar internals (acesso de teste)
4. **ESLint** bloqueia violações em lint e CI — compile-time enforcement

### Alternativas descartadas

- **`eslint-plugin-boundaries`** — overkill para um único módulo; +1 dependência
- **Physical relocation** (`_internal/` subdirectory) — quebrou Vitest $app/environment resolution e OTel imports; complexidade sem benefício proporcional
- **Subdirectory `package.json` com `exports`** — behavior inconsistente com SvelteKit $lib alias em dev

## Consequências

- Imports externos violando a boundary falham no lint (CI bloqueia merge)
- Refatorações internas (renomear, split, merge) não quebram consumidores externos
- OTel spans (`engine.*`) permanecem intactos — a engine usa `@opentelemetry/api` diretamente
- Testes continuam importando internals para white-box testing
