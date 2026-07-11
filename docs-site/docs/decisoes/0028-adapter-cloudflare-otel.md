# ADR-0028: Migrar adapter-static → adapter-cloudflare + Observabilidade OTel

**Data:** 2026-07-10
**Status:** Aceita
**Decisores:** Fernando

## Contexto

O frontend Garimpei usava `@sveltejs/adapter-static` (SPA pura deployada no Cloudflare Pages). Desde SvelteKit 2.31 (agosto 2025), o framework tem suporte nativo a OpenTelemetry (traces automáticos de handle, load, form actions) — mas requer um adapter com server component.

Simultaneamente, o backend precisava de observabilidade unificada (traces distribuídos correlacionando todos os 5 services). Sem trace propagation do frontend, o trace começa no C# API sem contexto da experiência do usuário.

## Decisão

1. **Migrar de `adapter-static` para `adapter-cloudflare`** — habilita SSR (opcional), `instrumentation.server.js`, e traces nativos do SvelteKit.
2. **Manter `ssr = false`** nas rotas (preserva comportamento SPA atual) — a migração é transparente para o usuário.
3. **Ativar `experimental.tracing.server` e `experimental.instrumentation.server`** — traces automáticos no Worker.
4. **Migrar de npm para bun** — instalação 60x mais rápida (68ms vs 30s+).

## Consequências

### Positivas
- Traces distribuídos completos: Browser → Worker → C# API → Go services → Python
- `src/instrumentation.server.js` com OTel nativo (W3C propagation automática)
- Deploy continua via wrangler (mesmo CI) — Cloudflare Workers Static Assets
- Futuro: SSR seletivo em rotas que se beneficiem (SEO, performance)
- Bun: CI e dev local significativamente mais rápidos

### Negativas / Riscos
- `experimental` flags podem mudar sem aviso (aceito: features estão estáveis desde 2025)
- adapter-cloudflare requer Wrangler CLI (já usado no CI)
- Cloudflare Workers tem limites de CPU time (10ms na free tier, 50ms na paid) — observável em requests SSR complexos

### Neutras
- O output final (static assets + Worker entry point) é equivalente ao anterior para SPA
- Testes E2E locais continuam funcionando (preview server via bun)
- package-lock.json substituído por bun.lock

## Alternativas Consideradas

1. **Manter adapter-static + OTel browser SDK manual** — funciona para propagação mas perde traces server-side nativos do SvelteKit.
2. **adapter-node** — requer hosting Node.js (conflita com Cloudflare Pages).
3. **Sentry + SDK proprietário** — vendor lock-in, custo crescente com volume.
