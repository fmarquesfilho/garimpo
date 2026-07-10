// SPA mode: toda lógica roda no navegador. SSR desabilitado para manter
// compatibilidade com o fluxo atual (auth Firebase client-side, api.js fetch).
// Com adapter-cloudflare, o Worker serve o shell e as pages como static assets.
export const ssr = false;

// OpenTelemetry: propaga traceparent em todo fetch() para o backend.
// Side-effect import — ativa instrumentação no boot.
import '$lib/telemetry.js';
