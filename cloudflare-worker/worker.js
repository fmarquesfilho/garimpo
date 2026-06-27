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

  // Tenta múltiplos endpoints da Shopee (fallback chain)
  const endpoints = [
    `https://shopee.com.br/api/v4/item/get?itemid=${itemId}&shopid=${shopId}`,
    `https://shopee.com.br/api/v4/pdp/get_pc?item_id=${itemId}&shop_id=${shopId}`,
  ];

  const headers = {
    'User-Agent': 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36',
    'Accept': 'application/json',
    'Referer': 'https://shopee.com.br/',
    'Accept-Language': 'pt-BR,pt;q=0.9,en-US;q=0.8,en;q=0.7',
    'sec-fetch-dest': 'empty',
    'sec-fetch-mode': 'cors',
    'sec-fetch-site': 'same-origin',
  };

  let lastError = '';
  let lastStatus = 0;

  for (const shopeeUrl of endpoints) {
    try {
      const resp = await fetch(shopeeUrl, { headers });
      lastStatus = resp.status;

      if (resp.status !== 200) {
        lastError = `status ${resp.status} from ${shopeeUrl.split('?')[0]}`;
        continue;
      }

      const data = await resp.json();

      // Extrai origem e marca dos atributos
      let origem = '';
      let marca = '';

      // Formato /api/v4/item/get: data.item_basic.attributes ou data.item.attributes
      const attrs = data?.data?.product?.attributes
        || data?.data?.item_basic?.attributes
        || data?.data?.item?.attributes
        || data?.item?.attributes
        || data?.item_basic?.attributes
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
          || data?.item_basic?.brand
          || '';
      }

      // Fallback: campo location
      if (!origem) {
        origem = data?.data?.item?.location
          || data?.item?.location
          || data?.item_basic?.shop_location
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
          'Cache-Control': 'public, max-age=86400',
        }
      });
    } catch (err) {
      lastError = err.message;
    }
  }

  // Todos os endpoints falharam
  return new Response(JSON.stringify({
    item_id: itemId,
    shop_id: shopId,
    origem: '',
    marca: '',
    fonte: 'erro',
    erro: lastError,
    raw_status: lastStatus,
  }), {
    status: 200,
    headers: { 'Content-Type': 'application/json', 'Access-Control-Allow-Origin': '*' }
  });
}
