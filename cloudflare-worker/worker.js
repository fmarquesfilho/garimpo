/**
 * Garimpei Proxy — Cloudflare Worker
 *
 * Routing:
 *   /api/*   → C# Web App (Cloud Run garimpei-v2)
 *   /*       → Frontend SPA (Cloudflare Pages garimpei-web)
 *
 * Caching (L1):
 *   Cacheable GET routes are served from Workers Cache API when available.
 *   Cache-Tag enables tag-based purge via Cloudflare API.
 *   TTL: 5 minutes (max-age=300).
 *
 * Environment variables:
 *   V2_ENABLED: "true" to route /api/* to C# (default: "true")
 *   V2_ORIGIN:  URL do serviço C# no Cloud Run
 *   V1_ORIGIN:  URL do Go (fallback, descomissionar via T-0022)
 *   PAGES_URL:  URL do Cloudflare Pages (garimpei-web.pages.dev)
 *   CACHE_MAX_AGE: Cache TTL in seconds (default: "300")
 */

// Routes eligible for caching (GET only).
const CACHEABLE_ROUTES = [
  '/api/v2/curadoria/ranking',
  '/api/candidatos',
  '/api/lojas/novidades',
];

export default {
  async fetch(request, env) {
    const url = new URL(request.url);
    const path = url.pathname;

    const v2Enabled = env.V2_ENABLED === 'true';
    const v2Origin = env.V2_ORIGIN || '';
    const v1Origin = env.V1_ORIGIN || '';
    const pagesUrl = env.PAGES_URL || 'https://garimpei-web.pages.dev';
    const docsUrl = env.DOCS_URL || 'https://garimpei-docs.pages.dev';
    const cacheMaxAge = parseInt(env.CACHE_MAX_AGE || '300', 10);

    // API routes → Cloud Run (C#)
    if (path.startsWith('/api/')) {
      const origin = (v2Enabled && v2Origin) ? v2Origin : v1Origin;
      if (!origin) {
        return new Response('No backend configured', { status: 503 });
      }

      // Check if this is a cacheable GET request
      if (request.method === 'GET' && isCacheableRoute(path)) {
        return handleCachedRequest(request, url, origin, cacheMaxAge);
      }

      return proxyTo(request, url, origin, 'csharp');
    }

    // Docs site → Cloudflare Pages (Rspress)
    if (path.startsWith('/docs')) {
      // Strip /docs prefix — Rspress serves from root
      const docsPath = path.replace(/^\/docs/, '') || '/';
      const docsUrlObj = new URL(docsUrl);
      const targetUrl = new URL(docsPath + url.search, docsUrlObj);
      const docsRequest = new Request(targetUrl, {
        method: request.method,
        headers: request.headers,
      });
      docsRequest.headers.set('Host', docsUrlObj.hostname);
      const response = await fetch(docsRequest);
      const modifiedResponse = new Response(response.body, response);
      modifiedResponse.headers.set('X-Garimpei-Backend', 'docs');
      return modifiedResponse;
    }

    // Everything else → Cloudflare Pages (frontend SPA)
    return proxyTo(request, url, pagesUrl, 'pages');
  }
}

/**
 * Check if the given path matches a cacheable route.
 */
function isCacheableRoute(path) {
  return CACHEABLE_ROUTES.some(route => path.startsWith(route));
}

/**
 * Extract busca_id from query params or path for Cache-Tag.
 * Tries: ?keyword=..., ?busca_id=..., ?shop_id=...
 */
function extractBuscaTag(url) {
  const keyword = url.searchParams.get('keyword');
  if (keyword) return `busca:keyword-${keyword.toLowerCase().trim()}`;

  const buscaId = url.searchParams.get('busca_id');
  if (buscaId) return `busca:${buscaId}`;

  const shopId = url.searchParams.get('shop_id');
  if (shopId) return `busca:shop-${shopId}`;

  const shopIds = url.searchParams.get('shop_ids');
  if (shopIds) return `busca:shops-${shopIds}`;

  return null;
}

/**
 * Handle a cacheable GET request using Workers Cache API.
 */
async function handleCachedRequest(request, url, origin, cacheMaxAge) {
  const cache = caches.default;

  // Use the full URL as cache key
  const cacheKey = new Request(url.toString(), request);

  // Try cache hit
  let response = await cache.match(cacheKey);
  if (response) {
    // Cache HIT — return with status header
    const cached = new Response(response.body, response);
    cached.headers.set('X-Cache-Status', 'HIT');
    cached.headers.set('X-Garimpei-Backend', 'csharp');
    return cached;
  }

  // Cache MISS — fetch from origin
  response = await proxyToRaw(request, url, origin);

  // Only cache successful responses
  if (response.status === 200) {
    const buscaTag = extractBuscaTag(url);
    const cacheResponse = new Response(response.body, response);

    cacheResponse.headers.set('Cache-Control', `public, max-age=${cacheMaxAge}`);
    if (buscaTag) {
      cacheResponse.headers.set('Cache-Tag', buscaTag);
    }
    cacheResponse.headers.set('X-Cache-Status', 'MISS');
    cacheResponse.headers.set('X-Garimpei-Backend', 'csharp');

    // Store in cache (non-blocking)
    const cacheClone = cacheResponse.clone();
    // Ensure the response stored in cache has proper headers
    cache.put(cacheKey, cacheClone);

    return cacheResponse;
  }

  // Non-200 — pass through without caching
  const passthrough = new Response(response.body, response);
  passthrough.headers.set('X-Cache-Status', 'BYPASS');
  passthrough.headers.set('X-Garimpei-Backend', 'csharp');
  return passthrough;
}

/**
 * Proxy the request to the target origin (returns raw Response).
 */
async function proxyToRaw(request, url, origin) {
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

  return await fetch(newRequest);
}

/**
 * Proxy the request to the target origin.
 */
async function proxyTo(request, url, origin, backend) {
  const response = await proxyToRaw(request, url, origin);

  const modifiedResponse = new Response(response.body, response);
  modifiedResponse.headers.set('X-Garimpei-Backend', backend);
  return modifiedResponse;
}
