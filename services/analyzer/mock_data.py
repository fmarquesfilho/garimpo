"""Dados fictícios para testar o fluxo sem BigQuery."""

from datetime import datetime, timedelta, timezone

NOW = datetime.now(timezone.utc)
THREE_DAYS_AGO = NOW - timedelta(days=3)

KEYWORD = "perfumes-importados"

NOVIDADES_RESPONSE = {
    "busca_id": KEYWORD,
    "dias": 7,
    "produtos_novos": [
        {
            "produto_id": "SP-006",
            "nome": "Jean Paul Gaultier Le Male 125ml",
            "preco": 280.00,
            "comissao": 0.13,
            "vendas": 450,
            "nota": 4.5,
            "imagem": "https://cf.shopee.com.br/file/jpg-le-male.jpg",
            "link": "https://shopee.com.br/product/123456/SP-006",
            "loja": "ImportsPerfumaria",
        },
    ],
    "variacoes": [
        {
            "produto_id": "SP-001",
            "nome": "Perfume CK One 100ml EDT",
            "preco_anterior": 189.90,
            "preco_atual": 151.90,
            "variacao": -0.2001,
            "imagem": "https://cf.shopee.com.br/file/ck-one-100ml.jpg",
            "link": "https://shopee.com.br/product/123456/SP-001",
            "loja": "ImportsPerfumaria",
        },
        {
            "produto_id": "SP-002",
            "nome": "Dolce & Gabbana Light Blue 75ml",
            "preco_anterior": 299.00,
            "preco_atual": 194.00,
            "variacao": -0.3512,
            "imagem": "https://cf.shopee.com.br/file/dg-light-blue.jpg",
            "link": "https://shopee.com.br/product/123456/SP-002",
            "loja": "ImportsPerfumaria",
        },
        {
            "produto_id": "SP-004",
            "nome": "Carolina Herrera Good Girl 80ml",
            "preco_anterior": 420.00,
            "preco_atual": 462.00,
            "variacao": 0.1000,
            "imagem": "https://cf.shopee.com.br/file/ch-good-girl.jpg",
            "link": "https://shopee.com.br/product/123456/SP-004",
            "loja": "ImportsPerfumaria",
        },
    ],
    "total_novos": 1,
    "total_variacoes": 3,
}

QUEDAS_RESPONSE = {
    "dias": 7,
    "threshold": 0.15,
    "quedas": [
        {
            "produto_id": "SP-002",
            "nome": "Dolce & Gabbana Light Blue 75ml",
            "preco_anterior": 299.00,
            "preco_atual": 194.00,
            "variacao": -0.3512,
            "comissao": 0.10,
            "imagem": "https://cf.shopee.com.br/file/dg-light-blue.jpg",
            "link": "https://shopee.com.br/product/123456/SP-002",
            "loja": "ImportsPerfumaria",
        },
        {
            "produto_id": "SP-001",
            "nome": "Perfume CK One 100ml EDT",
            "preco_anterior": 189.90,
            "preco_atual": 151.90,
            "variacao": -0.2001,
            "comissao": 0.12,
            "imagem": "https://cf.shopee.com.br/file/ck-one-100ml.jpg",
            "link": "https://shopee.com.br/product/123456/SP-001",
            "loja": "ImportsPerfumaria",
        },
    ],
    "total": 2,
}

COLETAS_RESPONSE = {
    "coletas": [
        {
            "coletado_em": NOW.isoformat(),
            "keyword": KEYWORD,
            "produtos": 6,
        },
        {
            "coletado_em": THREE_DAYS_AGO.isoformat(),
            "keyword": KEYWORD,
            "produtos": 5,
        },
    ],
    "total": 2,
    "dias_janela": 30,
}

EVOLUCAO_RESPONSE = {
    "dias_janela": 30,
    "lojas": [
        {
            "loja": "ImportsPerfumaria",
            "produtos": 6,
            "preco_medio": 280.48,
            "variacao_media_pct": -0.09,
            "serie": [
                {"dia": THREE_DAYS_AGO.strftime("%Y-%m-%d"), "preco_medio": 300.78, "produtos": 5},
                {"dia": NOW.strftime("%Y-%m-%d"), "preco_medio": 280.48, "produtos": 6},
            ],
        }
    ],
    "resumo": {
        "total_lojas": 1,
        "total_produtos": 6,
        "preco_medio_global": 280.48,
        "variacao_media_global_pct": -0.09,
        "total_quedas": 2,
        "total_altas": 1,
    },
}

ESTATISTICAS_RESPONSE = {
    "dias_janela": 30,
    "categorias": [
        {
            "categoria": "perfumaria",
            "total_produtos": 6,
            "preco_medio": 280.48,
            "comissao_media": 0.115,
            "vendas_total": 14450,
        }
    ],
    "resumo": {
        "total_coletas": 2,
        "total_produtos_unicos": 6,
        "preco_medio_global": 280.48,
    },
}
