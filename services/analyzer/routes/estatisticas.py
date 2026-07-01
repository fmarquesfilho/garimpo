"""Estatísticas: resumo de mercado por categoria."""

from fastapi import APIRouter, Query

from config import settings
import bq_client

router = APIRouter(tags=["Estatísticas"])


@router.get("/estatisticas")
def get_estatisticas(
    dias: int = Query(30, ge=1, le=180),
):
    """Resumo por categoria: quantidade de produtos, preço médio, comissão média."""
    ds = f"`{settings.bq_project}.{settings.bq_dataset}`"

    sql = f"""
    WITH ultimos AS (
      SELECT
        produto_id, nome, preco, comissao, vendas, nota,
        ROW_NUMBER() OVER (PARTITION BY produto_id ORDER BY em DESC) AS rn
      FROM {ds}.snapshots
      WHERE em >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL @dias DAY)
    )
    SELECT
      COUNT(DISTINCT produto_id) AS total_produtos,
      AVG(preco) AS preco_medio,
      AVG(comissao) AS comissao_media,
      AVG(vendas) AS vendas_media,
      AVG(nota) AS nota_media,
      APPROX_QUANTILES(preco, 4)[OFFSET(2)] AS preco_mediana,
      APPROX_QUANTILES(comissao, 4)[OFFSET(2)] AS comissao_mediana
    FROM ultimos
    WHERE rn = 1
    """

    from google.cloud.bigquery import ScalarQueryParameter

    rows = bq_client.query(sql, params=[
        ScalarQueryParameter("dias", "INT64", dias),
    ])

    if not rows:
        return {"dias": dias, "resumo": None}

    row = rows[0]
    return {
        "dias": dias,
        "resumo": {
            "total_produtos": row.get("total_produtos", 0),
            "preco_medio": round(row.get("preco_medio", 0), 2),
            "comissao_media": round(row.get("comissao_media", 0), 4),
            "vendas_media": round(row.get("vendas_media", 0), 1),
            "nota_media": round(row.get("nota_media", 0), 2),
            "preco_mediana": row.get("preco_mediana"),
            "comissao_mediana": row.get("comissao_mediana"),
        },
    }
