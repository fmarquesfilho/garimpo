"""Quedas: produtos com variação negativa de preço significativa."""

from fastapi import APIRouter, Query

from config import settings
import bq_client

router = APIRouter(tags=["Quedas"])


@router.get("/quedas")
def get_quedas(
    dias: int = Query(7, ge=1, le=90),
    threshold: float = Query(0.15, ge=0.01, le=0.99, description="Variação mínima (ex: 0.15 = 15%)"),
    limit: int = Query(50, ge=1, le=200),
):
    """Produtos com queda de preço acima do threshold na janela de dias."""
    ds = f"`{settings.bq_project}.{settings.bq_dataset}`"

    sql = f"""
    WITH snapshots_janela AS (
      SELECT
        produto_id, nome, preco, comissao, imagem, link, loja,
        em,
        FIRST_VALUE(preco) OVER (PARTITION BY produto_id ORDER BY em ASC) AS preco_primeiro,
        LAST_VALUE(preco) OVER (PARTITION BY produto_id ORDER BY em ASC
          ROWS BETWEEN UNBOUNDED PRECEDING AND UNBOUNDED FOLLOWING) AS preco_atual
      FROM {ds}.snapshots
      WHERE em >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL @dias DAY)
    ),
    com_variacao AS (
      SELECT DISTINCT
        produto_id, nome, preco_primeiro, preco_atual, comissao, imagem, link, loja,
        SAFE_DIVIDE(preco_atual - preco_primeiro, preco_primeiro) AS variacao
      FROM snapshots_janela
      WHERE preco_primeiro > 0
    )
    SELECT *
    FROM com_variacao
    WHERE variacao <= -@threshold
    ORDER BY variacao ASC
    LIMIT @limit
    """

    from google.cloud.bigquery import ScalarQueryParameter

    rows = bq_client.query(sql, params=[
        ScalarQueryParameter("dias", "INT64", dias),
        ScalarQueryParameter("threshold", "FLOAT64", threshold),
        ScalarQueryParameter("limit", "INT64", limit),
    ])

    quedas = [
        {
            "produto_id": row["produto_id"],
            "nome": row["nome"],
            "preco_anterior": row["preco_primeiro"],
            "preco_atual": row["preco_atual"],
            "variacao": round(row["variacao"], 4),
            "comissao": row.get("comissao"),
            "imagem": row.get("imagem"),
            "link": row.get("link"),
            "loja": row.get("loja"),
        }
        for row in rows
    ]

    return {
        "dias": dias,
        "threshold": threshold,
        "quedas": quedas,
        "total": len(quedas),
    }
