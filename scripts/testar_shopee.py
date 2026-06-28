#!/usr/bin/env python3
"""
Teste de fumaça da API de afiliados da Shopee (productOfferV2).

Mostra a resposta CRUA — ideal para depurar autenticação antes de confiar no
parsing do Go. Lê credenciais do ambiente (ou via flags).

Uso:
    export SHOPEE_APP_ID=...
    export SHOPEE_SECRET=...
    python3 tools/testar_shopee.py                       # lista geral (maior comissão)
    python3 tools/testar_shopee.py --keyword perfume     # busca por palavra
    python3 tools/testar_shopee.py --cat 100017           # filtra por categoria
    python3 tools/testar_shopee.py --dry-run             # só imprime a requisição assinada

Erros comuns (campo extensions.code):
    10020 Invalid Signature  -> confira AppId/Secret/Timestamp e a string assinada
    10030 Rate Limit         -> espere e reduza a frequência
    10035 No API Access      -> a conta ainda não tem a API liberada
    11001 Params Error       -> parâmetro inválido (ex.: listType/sortType)
"""
import argparse
import hashlib
import json
import os
import sys
import time
import urllib.request
import urllib.error


def build_query(args):
    parts = [
        f"listType: {args.list_type}",
        f"sortType: {args.sort_type}",
        "page: 1",
        f"limit: {args.limit}",
    ]
    if args.cat:
        parts.append(f"productCatId: {args.cat}")
    if args.keyword:
        parts.append(f'keyword: "{args.keyword}"')
    inner = ", ".join(parts)
    return (
        "{ productOfferV2(%s) { nodes { itemId productName offerLink priceMin "
        "sales ratingStar commissionRate shopName } pageInfo { hasNextPage } } }" % inner
    )


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--app-id", default=os.getenv("SHOPEE_APP_ID", ""))
    ap.add_argument("--secret", default=os.getenv("SHOPEE_SECRET", ""))
    ap.add_argument("--country", default="com.br", help="TLD do endpoint (com.br, sg, co.id...)")
    ap.add_argument("--keyword", default="")
    ap.add_argument("--cat", type=int, default=0, help="productCatId")
    ap.add_argument("--limit", type=int, default=5)
    ap.add_argument("--list-type", type=int, default=1, help="0=Recomendados,1=Maior comissão,2=Top perf")
    ap.add_argument("--sort-type", type=int, default=5, help="1=Relevância,2=Vendidos,...,5=Comissão")
    ap.add_argument("--dry-run", action="store_true", help="imprime a requisição sem enviar")
    args = ap.parse_args()

    if not args.app_id or not args.secret:
        sys.exit("Defina SHOPEE_APP_ID e SHOPEE_SECRET (ou use --app-id/--secret).")

    endpoint = f"https://open-api.affiliate.shopee.{args.country}/graphql"
    query = build_query(args)
    body = json.dumps({"query": query}, separators=(",", ":"))  # compacto, igual ao Go
    ts = str(int(time.time()))
    signature = hashlib.sha256((args.app_id + ts + body + args.secret).encode()).hexdigest()
    auth = f"SHA256 Credential={args.app_id}, Timestamp={ts}, Signature={signature}"

    if args.dry_run:
        print("POST", endpoint)
        print("Authorization:", auth)
        print("Body:", body)
        return

    req = urllib.request.Request(
        endpoint,
        data=body.encode(),
        headers={"Content-Type": "application/json", "Authorization": auth},
        method="POST",
    )
    try:
        with urllib.request.urlopen(req, timeout=20) as resp:
            status = resp.status
            payload = json.loads(resp.read().decode())
    except urllib.error.HTTPError as e:
        print(f"HTTP {e.code}")
        print(e.read().decode()[:1000])
        return
    except Exception as e:
        sys.exit(f"Falha de rede: {e}")

    print(f"HTTP {status}\n")

    if payload.get("errors"):
        print("ERROS retornados pela API:")
        for err in payload["errors"]:
            ext = err.get("extensions", {})
            print(f"  code={ext.get('code')}  message={err.get('message')}")
        print()

    nodes = (
        payload.get("data", {})
        .get("productOfferV2", {})
        .get("nodes", [])
    )
    if not nodes:
        print("Nenhum produto retornado. Resposta crua:")
        print(json.dumps(payload, ensure_ascii=False, indent=2)[:1500])
        return

    print(f"{len(nodes)} produto(s):\n")
    for n in nodes:
        comm = float(n.get("commissionRate") or 0) * 100
        print(f"  • {n.get('productName','?')[:60]}")
        print(
            f"      comissão={comm:.1f}%  preço={n.get('priceMin')}  "
            f"vendas={n.get('sales')}  nota={n.get('ratingStar')}  loja={n.get('shopName')}"
        )
        print(f"      {n.get('offerLink')}")


if __name__ == "__main__":
    main()
