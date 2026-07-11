"""Oportunidades: quedas ativas + novos + alto-valor não publicados."""

from fastapi import APIRouter, Query

from config import settings

router = APIRouter(tags=["Oportunidades"])


@router.get("/oportunidades/agora")
def get_oportunidades(
    dias: int = Query(7, ge=1, le=30),
):
    if settings.mock_data:
        return {
            "dias": dias,
            "quedas": [],
            "novos": [],
            "alto_valor": [],
            "total_quedas": 0,
            "total_novos": 0,
            "total_alto_valor": 0,
            "filtro_publicacoes": False,
        }

    import bq_client
    from google.cloud.bigquery import ScalarQueryParameter

    ds = f"`{settings.bq_project}.{settings.bq_dataset}`"

    # 1. Active price drops (top 10 by magnitude)
    sql_quedas = f"""
    WITH primeiros AS (
      SELECT produto_id, nome, preco AS preco_primeiro, imagem, link, loja,
        ROW_NUMBER() OVER (PARTITION BY produto_id ORDER BY coletado_em ASC) AS rn
      FROM {ds}.snapshots
      WHERE coletado_em >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL @dias DAY)
    ),
    ultimos AS (
      SELECT produto_id, preco AS preco_atual,
        ROW_NUMBER() OVER (PARTITION BY produto_id ORDER BY coletado_em DESC) AS rn
      FROM {ds}.snapshots
      WHERE coletado_em >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL @dias DAY)
    )
    SELECT
      p.produto_id, p.nome, p.preco_primeiro AS preco_anterior, u.preco_atual,
      SAFE_DIVIDE(u.preco_atual - p.preco_primeiro, p.preco_primeiro) AS variacao,
      p.loja, p.imagem, p.link
    FROM primeiros p
    JOIN ultimos u ON p.produto_id = u.produto_id AND u.rn = 1
    WHERE p.rn = 1 AND p.preco_primeiro > 0
      AND SAFE_DIVIDE(u.preco_atual - p.preco_primeiro, p.preco_primeiro) <= -0.10
    ORDER BY variacao ASC
    LIMIT 10
    """

    quedas = []
    try:
        rows = bq_client.query(sql_quedas, params=[
            ScalarQueryParameter("dias", "INT64", dias),
        ])
        quedas = [
            {
                "produto_id": r["produto_id"],
                "nome": r["nome"],
                "preco_anterior": round(r["preco_anterior"], 2),
                "preco_atual": round(r["preco_atual"], 2),
                "variacao": round(r["variacao"], 4),
                "loja": r.get("loja", ""),
                "imagem": r.get("imagem", ""),
                "link": r.get("link", ""),
            }
            for r in rows
        ]
    except Exception:
        pass

    # 2. New products (detected in last 48h, only 1 appearance)
    sql_novos = f"""
    SELECT produto_id, nome, preco, comissao, loja, imagem, link,
      MIN(coletado_em) AS detectado_em
    FROM {ds}.snapshots
    WHERE coletado_em >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 2 DAY)
    GROUP BY produto_id, nome, preco, comissao, loja, imagem, link
    HAVING COUNT(*) = 1
    ORDER BY detectado_em DESC
    LIMIT 10
    """

    novos = []
    try:
        rows = bq_client.query(sql_novos)
        novos = [
            {
                "produto_id": r["produto_id"],
                "nome": r["nome"],
                "preco": r.get("preco", 0),
                "comissao": r.get("comissao", 0),
                "loja": r.get("loja", ""),
                "detectado_em": str(r.get("detectado_em", "")),
            }
            for r in rows
        ]
    except Exception:
        pass

    # 3. High-value unpublished (commission > P75 AND sales > median)
    alto_valor = []
    filtro_pub = False
    sql_alto = f"""
    WITH stats AS (
      SELECT
        APPROX_QUANTILES(comissao, 4)[OFFSET(3)] AS p75_comissao,
        APPROX_QUANTILES(vendas, 2)[OFFSET(1)] AS mediana_vendas
      FROM {ds}.snapshots
      WHERE coletado_em >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL @dias DAY)
    ),
    candidatos AS (
      SELECT DISTINCT s.produto_id, s.nome, s.preco, s.comissao, s.vendas, s.loja, s.imagem, s.link
      FROM {ds}.snapshots s, stats
      WHERE s.coletado_em >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL @dias DAY)
        AND s.comissao >= stats.p75_comissao
        AND s.vendas >= stats.mediana_vendas
    )
    SELECT c.*
    FROM candidatos c
    LEFT JOIN {ds}.publicacoes p ON c.produto_id = p.produto_id
    WHERE p.produto_id IS NULL
    ORDER BY c.comissao * c.preco DESC
    LIMIT 5
    """

    try:
        rows = bq_client.query(sql_alto, params=[
            ScalarQueryParameter("dias", "INT64", dias),
        ])
        alto_valor = [
            {
                "produto_id": r["produto_id"],
                "nome": r["nome"],
                "preco": r.get("preco", 0),
                "comissao": r.get("comissao", 0),
                "vendas": r.get("vendas", 0),
                "loja": r.get("loja", ""),
            }
            for r in rows
        ]
        filtro_pub = True
    except Exception:
        # publicacoes table may not exist — try without the join
        sql_alto_nojoin = f"""
        WITH stats AS (
          SELECT
            APPROX_QUANTILES(comissao, 4)[OFFSET(3)] AS p75_comissao,
            APPROX_QUANTILES(vendas, 2)[OFFSET(1)] AS mediana_vendas
          FROM {ds}.snapshots
          WHERE coletado_em >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL @dias DAY)
        )
        SELECT DISTINCT s.produto_id, s.nome, s.preco, s.comissao, s.vendas, s.loja
        FROM {ds}.snapshots s, stats
        WHERE s.coletado_em >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL @dias DAY)
          AND s.comissao >= stats.p75_comissao
          AND s.vendas >= stats.mediana_vendas
        ORDER BY s.comissao * s.preco DESC
        LIMIT 5
        """
        try:
            rows = bq_client.query(sql_alto_nojoin, params=[
                ScalarQueryParameter("dias", "INT64", dias),
            ])
            alto_valor = [
                {
                    "produto_id": r["produto_id"],
                    "nome": r["nome"],
                    "preco": r.get("preco", 0),
                    "comissao": r.get("comissao", 0),
                    "vendas": r.get("vendas", 0),
                    "loja": r.get("loja", ""),
                }
                for r in rows
            ]
        except Exception:
            pass
        filtro_pub = False

    return {
        "dias": dias,
        "quedas": quedas,
        "novos": novos,
        "alto_valor": alto_valor,
        "total_quedas": len(quedas),
        "total_novos": len(novos),
        "total_alto_valor": len(alto_valor),
        "filtro_publicacoes": filtro_pub,
    }
