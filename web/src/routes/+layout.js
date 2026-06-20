// App estático no estilo SPA: tudo roda no navegador (fetch da API, localStorage
// do quadro). prerender gera o shell HTML; ssr off evita rodar no servidor o que
// só faz sentido no cliente.
export const prerender = true;
export const ssr = false;
