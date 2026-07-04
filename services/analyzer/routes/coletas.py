"""Coletas: histórico de coletas executadas."""

from fastapi import APIRouter, Query

from config import settings

router = APIRouter(tags=["Coletas"])


@router.get("/coletas")
def get_coletas(
    dias: int = Query(30, ge=1, le=180),
):
    if settings.mock_data:
        from mock_data import COLETAS_RESPONSE
        return {**COLETAS_RESPONSE, "dias_janela": dias}

    import bq_client
    """Histórico de coletas executadas (snapshots agrupados por execução)."""
    ds = bq_client.dataset_ref()

    sql = f"""
    SELECT
      DATE(coletado_em) AS data,
      keyword,
      COUNT(DISTINCT produto_id) AS produtos,
      MIN(coletado_em) AS coletado_em
    FROM {ds}.snapshots
    WHERE coletado_em >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL @dias DAY)
    GROUP BY data, keyword
    ORDER BY data DESC
    LIMIT 100
    """

    from google.cloud.bigquery import ScalarQueryParameter

    rows = bq_client.query(sql, params=[
        ScalarQueryParameter("dias", "INT64", dias),
    ])

    coletas = [
        {
            "coletado_em": str(row.get("coletado_em", "")),
            "keyword": row.get("keyword", ""),
            "produtos": row.get("produtos", 0),
        }
        for row in rows
    ]

    return {"coletas": coletas, "total": len(coletas), "dias_janela": dias}
