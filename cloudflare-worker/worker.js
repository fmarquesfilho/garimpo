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

  const shopeeUrl = `https://shopee.com.br/api/v4/pdp/get_pc?item_id=${itemId}&shop_id=${shopId}`;

  try {
    const resp = await fetch(shopeeUrl, {
      headers: {
        'User-Agent': 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36',
        'Accept': 'application/json',
        'Referer': 'https://shopee.com.br/',
        'Accept-Language': 'pt-BR,pt;q=0.9',
      }
    });

    const data = await resp.json();

    // Extrai origem e marca dos atributos
    let origem = '';
    let marca = '';

    // Tenta múltiplos caminhos na estrutura de resposta
    const attrs = data?.data?.product?.attributes
      || data?.data?.item_basic?.attributes
      || data?.data?.item?.attributes
      || data?.item?.props
      || [];

    for (const attr of attrs) {
      const name = (attr.name || '').toLowerCase();
      const value = attr.value || '';
      if (!value) continue;

      if ((name.includes('origem') || name.includes('origin') || name.includes('país')
        || name.includes('envio de') || name.includes('fabricado')) && !origem) {
        origem = value;
      }
      if ((name.includes('marca') || name.includes('brand')) && !marca) {
        marca = value;
      }
    }

    // Fallback: campo brand direto
    if (!marca) {
      marca = data?.data?.product?.brand
        || data?.data?.item_basic?.brand
        || data?.data?.item?.brand
        || data?.item?.brand
        || '';
    }

    // Fallback: campo location
    if (!origem) {
      origem = data?.data?.item?.location
        || data?.item?.location
        || data?.item?.seller_info?.city
        || '';
    }

    return new Response(JSON.stringify({
      item_id: itemId,
      shop_id: shopId,
      origem: origem,
      marca: marca,
      fonte: 'cloudflare_proxy',
      raw_status: resp.status,
    }), {
      status: 200,
      headers: {
        'Content-Type': 'application/json',
        'Access-Control-Allow-Origin': '*',
        'Cache-Control': 'public, max-age=86400', // cache 24h na edge
      }
    });
  } catch (err) {
    return new Response(JSON.stringify({
      item_id: itemId,
      shop_id: shopId,
      origem: '',
      marca: '',
      fonte: 'erro',
      erro: err.message,
    }), {
      status: 200,
      headers: { 'Content-Type': 'application/json', 'Access-Control-Allow-Origin': '*' }
    });
  }
}
