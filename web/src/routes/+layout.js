// App estático no estilo SPA: tudo roda no navegador (fetch da API, localStorage
// do quadro). prerender gera o shell HTML; ssr off evita rodar no servidor o que
// só faz sentido no cliente.
export const prerender = true;
export const ssr = false;

// OpenTelemetry: propaga traceparent em todo fetch() para o backend.
// Side-effect import — ativa instrumentação no boot.
import '$lib/telemetry.js';
