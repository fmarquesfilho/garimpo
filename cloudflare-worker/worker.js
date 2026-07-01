/**
 * Garimpei Proxy — Cloudflare Worker
 *
 * Routing:
 *   /api/*   → C# Web App (Cloud Run garimpei-v2)
 *   /*       → Frontend SPA (Cloudflare Pages garimpei-web)
 *
 * Environment variables:
 *   V2_ENABLED: "true" to route /api/* to C# (default: "true")
 *   V2_ORIGIN:  URL do serviço C# no Cloud Run
 *   V1_ORIGIN:  URL do Go legado (fallback, pode ser removido)
 *   PAGES_URL:  URL do Cloudflare Pages (garimpei-web.pages.dev)
 */

export default {
  async fetch(request, env) {
    const url = new URL(request.url);
    const path = url.pathname;

    const v2Enabled = env.V2_ENABLED === 'true';
    const v2Origin = env.V2_ORIGIN || '';
    const v1Origin = env.V1_ORIGIN || '';
    const pagesUrl = env.PAGES_URL || 'https://garimpei-web.pages.dev';
    const docsUrl = env.DOCS_URL || 'https://garimpei-docs.pages.dev';

    // API routes → Cloud Run (C#)
    if (path.startsWith('/api/')) {
      const origin = (v2Enabled && v2Origin) ? v2Origin : v1Origin;
      if (!origin) {
        return new Response('No backend configured', { status: 503 });
      }
      return proxyTo(request, url, origin, 'csharp');
    }

    // Docs site → Cloudflare Pages (Starlight)
    if (path.startsWith('/docs')) {
      return proxyTo(request, url, docsUrl, 'docs');
    }

    // Everything else → Cloudflare Pages (frontend SPA)
    return proxyTo(request, url, pagesUrl, 'pages');
  }
}

/**
 * Proxy the request to the target origin.
 */
async function proxyTo(request, url, origin, backend) {
  const target = new URL(origin);
  url.hostname = target.hostname;
  url.port = target.port;
  url.protocol = target.protocol;

  const newRequest = new Request(url, {
    method: request.method,
    headers: request.headers,
    body: request.body,
  });
  newRequest.headers.set('Host', target.hostname);

  const response = await fetch(newRequest);

  const modifiedResponse = new Response(response.body, response);
  modifiedResponse.headers.set('X-Garimpei-Backend', backend);
  return modifiedResponse;
}
