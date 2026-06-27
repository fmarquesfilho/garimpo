export default {
  async fetch(request) {
    const url = new URL(request.url);

    // Proxy para API pública da Shopee (buscar origem do produto)
    // Rota: /shopee-proxy/pdp?item_id=X&shop_id=Y
    if (url.pathname === '/shopee-proxy/pdp') {
      return handleShopeeProxy(url);
    }

    // Proxy reverso padrão → Cloud Run
    url.hostname = 'garimpo-api-vj6afttbza-rj.a.run.app';

    const newRequest = new Request(url, {
      method: request.method,
      headers: request.headers,
      body: request.body,
    });
    newRequest.headers.set('Host', 'garimpo-api-vj6afttbza-rj.a.run.app');

    return fetch(newRequest);
  }
}

async function handleShopeeProxy(url) {
  const itemId = url.searchParams.get('item_id');
  const shopId = url.searchParams.get('shop_id');

  if (!itemId || !shopId) {
    return new Response(JSON.stringify({ erro: 'item_id e shop_id obrigatórios' }), {
      status: 400,
      headers: { 'Content-Type': 'application/json', 'Access-Control-Allow-Origin': '*' }
    });
  }

  const headers = {
    'User-Agent': 'Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)',
    'Accept': 'text/html,application/xhtml+xml',
    'Accept-Language': 'pt-BR,pt;q=0.9',
  };

  try {
    // Abordagem: buscar a página HTML do produto (renderizada server-side para SEO/Googlebot)
    // A Shopee serve HTML com dados embutidos para crawlers (User-Agent Googlebot)
    const productUrl = `https://shopee.com.br/product-i.${shopId}.${itemId}`;
    const resp = await fetch(productUrl, { headers, redirect: 'follow' });

    if (resp.status !== 200) {
      return jsonResponse({ item_id: itemId, shop_id: shopId, origem: '', marca: '', fonte: 'erro', erro: `html status ${resp.status}` });
    }

    const html = await resp.text();

    // Extrair dados de JSON-LD (structured data para SEO)
    let origem = '';
    let marca = '';

    // 1. Tentar JSON-LD <script type="application/ld+json">
    const ldMatches = html.match(/<script[^>]*type="application\/ld\+json"[^>]*>([\s\S]*?)<\/script>/gi);
    if (ldMatches) {
      for (const match of ldMatches) {
        const jsonStr = match.replace(/<script[^>]*>/, '').replace(/<\/script>/, '');
        try {
          const ld = JSON.parse(jsonStr);
          if (ld.brand?.name && !marca) marca = ld.brand.name;
          if (ld.brand && typeof ld.brand === 'string' && !marca) marca = ld.brand;
          if (ld.countryOfOrigin && !origem) origem = ld.countryOfOrigin;
        } catch {}
      }
    }

    // 2. Tentar meta tags (og:brand, product:brand)
    if (!marca) {
      const brandMeta = html.match(/<meta[^>]*property="product:brand"[^>]*content="([^"]+)"/i)
        || html.match(/<meta[^>]*name="brand"[^>]*content="([^"]+)"/i);
      if (brandMeta) marca = brandMeta[1];
    }

    // 3. Tentar extrair do __NEXT_DATA__ ou window.__INITIAL_STATE__ (dados SSR injetados)
    if (!origem || !marca) {
      const stateMatch = html.match(/window\.__INITIAL_STATE__\s*=\s*(\{[\s\S]*?\});?\s*<\/script>/);
      const nextMatch = html.match(/<script[^>]*id="__NEXT_DATA__"[^>]*>([\s\S]*?)<\/script>/);
      const dataStr = stateMatch?.[1] || nextMatch?.[1];
      if (dataStr) {
        try {
          const data = JSON.parse(dataStr);
          // Navega na estrutura para encontrar atributos
          const attrs = data?.item?.attributes || data?.props?.pageProps?.item?.attributes || [];
          for (const attr of attrs) {
            const name = (attr.name || '').toLowerCase();
            const value = attr.value || '';
            if (!value) continue;
            if ((name.includes('origem') || name.includes('origin') || name.includes('país')) && !origem) {
              origem = value;
            }
            if ((name.includes('marca') || name.includes('brand')) && !marca) {
              marca = value;
            }
          }
          if (!marca) {
            marca = data?.item?.brand || data?.props?.pageProps?.item?.brand || '';
          }
        } catch {}
      }
    }

    // 4. Tentar extrair "País de Origem" do HTML renderizado (tabela de especificações)
    if (!origem) {
      const origemMatch = html.match(/Pa[ií]s\s+de\s+Origem[^<]*<[^>]*>([^<]+)/i)
        || html.match(/Country\s+of\s+Origin[^<]*<[^>]*>([^<]+)/i)
        || html.match(/Origem[:\s]*<[^>]*>([^<]+)/i);
      if (origemMatch) origem = origemMatch[1].trim();
    }

    // 5. Marca como fallback do HTML
    if (!marca) {
      const marcaMatch = html.match(/Marca[:\s]*<[^>]*>([^<]+)/i)
        || html.match(/Brand[:\s]*<[^>]*>([^<]+)/i);
      if (marcaMatch) marca = marcaMatch[1].trim();
    }

    return jsonResponse({
      item_id: itemId,
      shop_id: shopId,
      origem: origem,
      marca: marca,
      fonte: 'cloudflare_html',
      raw_status: resp.status,
    });

  } catch (err) {
    return jsonResponse({
      item_id: itemId,
      shop_id: shopId,
      origem: '',
      marca: '',
      fonte: 'erro',
      erro: err.message,
    });
  }
}

function jsonResponse(body, status = 200) {
  return new Response(JSON.stringify(body), {
    status,
    headers: {
      'Content-Type': 'application/json',
      'Access-Control-Allow-Origin': '*',
      'Cache-Control': status === 200 && body.origem ? 'public, max-age=86400' : 'no-cache',
    }
  });
}
