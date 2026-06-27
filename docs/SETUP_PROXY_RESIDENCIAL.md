# Setup do Proxy Residencial (Bright Data)

A Shopee bloqueia chamadas à API pública a partir de IPs de datacenter (Cloud Run, Cloudflare Workers). Para obter o País de Origem e Marca dos produtos automaticamente, usamos um proxy residencial que roteia as requests por IPs domésticos brasileiros.

## Custo estimado

- ~20 produtos por busca × ~100KB cada = 2MB/busca
- ~10 buscas/dia = 20MB/dia
- ~600MB/mês = **~$3.50/mês** com Bright Data

## Passo a passo

### 1. Criar conta no Bright Data

1. Acesse [brightdata.com](https://brightdata.com) e crie uma conta
2. O free trial dá $5 de crédito sem cartão (suficiente para ~1 mês de uso)
3. No dashboard, vá em **Proxies & Scraping → Residential Proxies**
4. Crie uma nova zona (zone) — pode chamar "garimpei"

### 2. Obter credenciais

Na página da zona residencial, você verá:

- **Host:** `brd.superproxy.io`
- **Port:** `33335`
- **Username:** algo como `brd-customer-XXXXX-zone-garimpei`
- **Password:** a senha gerada

### 3. Montar a URL do proxy

O formato é:
```
http://USERNAME:PASSWORD@brd.superproxy.io:33335
```

Para geo-targeting Brasil, adicione `-country-br` ao username:
```
http://brd-customer-XXXXX-zone-garimpei-country-br:SENHA@brd.superproxy.io:33335
```

### 4. Configurar no Cloud Run

Adicione como secret no GCP e vincule ao Cloud Run:

```bash
# Criar o secret
printf 'http://brd-customer-XXXXX-zone-garimpei-country-br:SENHA@brd.superproxy.io:33335' | \
  gcloud secrets create RESIDENTIAL_PROXY_URL --data-file=-

# Vincular ao Cloud Run
gcloud run services update garimpo-api \
  --region southamerica-east1 \
  --update-secrets RESIDENTIAL_PROXY_URL=RESIDENTIAL_PROXY_URL:latest
```

### 5. Atualizar o workflow de deploy

Adicionar ao `--update-secrets` no `deploy-gcp.yml`:
```yaml
RESIDENTIAL_PROXY_URL=RESIDENTIAL_PROXY_URL:latest
```

### 6. Testar

Após deploy, acesse no navegador (logado como admin):
```
https://garimpei.app.br/api/produto/origem?item_id=23198188943&shop_id=258316442
```

Resposta esperada:
```json
{
  "item_id": "23198188943",
  "shop_id": "258316442",
  "origem": "Coreia",
  "marca": "SKIN1004",
  "fonte": "proxy_residencial"
}
```

## Fallback

Se `RESIDENTIAL_PROXY_URL` não estiver configurado:
- O endpoint retorna `{"fonte": "erro", "erro": "RESIDENTIAL_PROXY_URL não configurado"}`
- Os badges não aparecem nos cards
- O fallback `origem_padrao` por loja monitorada continua funcionando

## Segurança

- A URL do proxy contém credenciais — sempre como secret (nunca em variáveis de ambiente expostas)
- O proxy é usado apenas para requests à `shopee.com.br` (não para outros fins)
- Consumo monitorável no dashboard do Bright Data

## Alternativas testadas e descartadas

| Abordagem | Resultado |
|-----------|-----------|
| API de afiliados (GraphQL) | Campo de origem não existe |
| API pública v4 direto do Cloud Run | 403 (IP datacenter bloqueado) |
| Cloudflare Worker como proxy | 403 (IP CDN também bloqueado) |
| Worker com User-Agent Googlebot | 403 (bloqueio por IP, não User-Agent) |
| Proxy residencial (Bright Data) | ✅ Funciona |
