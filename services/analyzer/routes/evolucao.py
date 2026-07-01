"""Evolução: série temporal de preço médio por loja."""

from fastapi import APIRouter, Query

from config import settings
import bq_client

router = APIRouter(tags=["Evolução"])


@router.get("/evolucao")
def get_evolucao(
    dias: int = Query(30, ge=1, le=180),
):
    """Série temporal de preço médio por dia, com resumo global e top variações."""
    ds = f"`{settings.bq_project}.{settings.bq_dataset}`"

    sql = f"""
    SELECT
      DATE(em) AS dia,
      keyword AS loja,
      AVG(preco) AS preco_medio,
      COUNT(DISTINCT produto_id) AS produtos,
      MIN(preco) AS preco_min,
      MAX(preco) AS preco_max
    FROM {ds}.snapshots
    WHERE em >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL @dias DAY)
    GROUP BY dia, loja
    ORDER BY dia DESC, loja
    """

    from google.cloud.bigquery import ScalarQueryParameter

    rows = bq_client.query(sql, params=[
        ScalarQueryParameter("dias", "INT64", dias),
    ])

    # Agrupar por loja
    lojas: dict = {}
    for row in rows:
        loja = row["loja"]
        if loja not in lojas:
            lojas[loja] = []
        lojas[loja].append({
            "dia": str(row["dia"]),
            "preco_medio": round(row["preco_medio"], 2),
            "produtos": row["produtos"],
            "preco_min": row["preco_min"],
            "preco_max": row["preco_max"],
        })

    return {
        "dias": dias,
        "lojas": lojas,
        "total_lojas": len(lojas),
    }
