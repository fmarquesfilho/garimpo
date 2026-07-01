/**
 * Garimpei Proxy — Cloudflare Worker
 *
 * Routing split para coexistência Go (legado) + C# (v2):
 *   /api/v2/*  → C# Web App (Cloud Run garimpei-v2)
 *   /api/*     → Go monólito (Cloud Run garimpo-api)
 *   /*         → Go monólito (frontend + docs)
 *
 * Feature flags via env vars (Wrangler secrets/vars):
 *   V2_ENABLED: "true" para habilitar routing para C# (default: "false")
 *   V2_ORIGIN:  URL do serviço C# no Cloud Run
 *   V1_ORIGIN:  URL do serviço Go legado
 *
 * Rollback instantâneo: set V2_ENABLED=false no Cloudflare dashboard.
 */

export default {
  async fetch(request, env) {
    const url = new URL(request.url);
    const path = url.pathname;

    const v2Enabled = env.V2_ENABLED === 'true';
    const v2Origin = env.V2_ORIGIN || '';
    const v1Origin = env.V1_ORIGIN || 'https://garimpo-api-vj6afttbza-rj.a.run.app';

    // Route /api/v2/* to C# if enabled
    if (v2Enabled && v2Origin && path.startsWith('/api/v2/')) {
      return proxyTo(request, url, v2Origin);
    }

    // Route ALL /api/* to C# if v2 is enabled (migration complete)
    if (v2Enabled && v2Origin && path.startsWith('/api/')) {
      return proxyTo(request, url, v2Origin);
    }

    // Everything else goes to Go legacy (frontend, docs)
    return proxyTo(request, url, v1Origin);
  }
}

/**
 * Proxy the request to the target origin, preserving method/headers/body.
 */
async function proxyTo(request, url, origin) {
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

  // Add routing header for debugging
  const modifiedResponse = new Response(response.body, response);
  modifiedResponse.headers.set('X-Garimpei-Backend', origin.includes('v2') ? 'csharp' : 'go');
  return modifiedResponse;
}
