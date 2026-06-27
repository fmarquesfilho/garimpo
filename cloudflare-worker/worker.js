export default {
  async fetch(request) {
    const url = new URL(request.url);
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
