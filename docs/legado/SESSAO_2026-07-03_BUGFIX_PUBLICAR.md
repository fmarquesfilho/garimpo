# Sessão 2026-07-03 — Bugfix: Página Publicar Travada em Carregando

## Problema

Após a migração para Bits UI (T-0037) e a refatoração arquitetural (ADR-0012),
a página `/publicar` ficava presa em "Carregando…" indefinidamente. O usuário
clicava "Publicar" em um produto e a página abria mas nunca renderizava o formulário.

## Causa Raiz

**`Context "Tooltip.Provider" not found`** — erro de runtime do Bits UI.

O componente `RichEditor` usa `Tooltip` (Bits UI) nos botões da toolbar (Negrito,
Itálico, Link). Em Bits UI v2, todo `Tooltip.Root` exige um `Tooltip.Provider`
ancestral na árvore de componentes. Como não havia Provider, o componente crasheava
silenciosamente ao renderizar, impedindo que o conteúdo pós-`carregando = false`
fosse exibido.

### Por que só afetava `/publicar`

- `Tooltip` é usado apenas no `RichEditor`
- `RichEditor` é usado apenas na página `/publicar`
- Outras páginas não usam `Tooltip` diretamente (usam `title=` nativo)

### Por que não foi detectado antes

- O `svelte-check` não detecta erros de contexto de runtime
- Os testes E2E existentes usavam mocks que bypasavam a auth → a página mostrava
  a landing page (Entrar com Google) e nunca renderizava o formulário completo
- O erro só se manifestava com autenticação real (layout renderiza `{@render children()}`)

## Diagnóstico

Reproduzido via Playwright + Firebase Auth Emulator:

```bash
FIREBASE_AUTH_EMULATOR_HOST=localhost:9099 npx playwright test publicar-debug.spec.js
```

Output do console do browser:
```
Context "Tooltip.Provider" not found
```

## Correções

| Arquivo | Mudança |
|---------|---------|
| `+layout.svelte` | Envolver todo o layout com `<Tooltip.Provider>` (Bits UI) |
| `firebase.js` | Timeout de 5s no `getIdToken()` para evitar hang caso token refresh falhe |
| `publicar/+page.svelte` | Timeout de 15s nas chamadas API + retry + safety de 20s |
| `publicar-store.js` | Novo: passagem de dados via `sessionStorage` ao invés de query string |
| `+page.svelte`, `lojas/+page.svelte` | Usar `prepararPublicacao()` para navegar |
| `publicar.spec.js` | Atualizar testes para usar sessionStorage |

## Refatoração: Query String → sessionStorage

**Antes:** `goto('/publicar?dados=${encodeURIComponent(JSON.stringify(produto))}')`
- URL com ~2KB de JSON encodado
- Limite de tamanho de URL (~8KB)
- URL feia e não-compartilhável
- Problemas de encoding com caracteres especiais

**Depois:** `goto(prepararPublicacao(produto))`
- URL limpa: `/publicar`
- Sem limite de tamanho
- sessionStorage limpo após leitura (one-shot)
- Módulo dedicado `$lib/publicar-store.js`

## Validação

- 141 testes unitários ✅
- 12 testes E2E publicar ✅
- 53 testes E2E totais ✅ (com Firebase Auth Emulator)
- svelte-check 0 erros ✅
- ESLint 0 warnings ✅
- Build estático OK ✅
