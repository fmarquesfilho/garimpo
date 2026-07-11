"""Quedas: produtos com variação negativa de preço significativa."""

from fastapi import APIRouter, Query

from config import settings

router = APIRouter(tags=["Quedas"])


@router.get("/quedas")
def get_quedas(
    busca_id: str = Query("", description="ID da busca (opcional, filtra por busca)"),
    dias: int = Query(7, ge=1, le=90),
    threshold: float = Query(0.15, ge=0.01, le=0.99, description="Variação mínima (ex: 0.15 = 15%)"),
    limit: int = Query(50, ge=1, le=200),
):
    if settings.mock_data:
        from mock_data import QUEDAS_RESPONSE
        return {**QUEDAS_RESPONSE, "dias": dias, "threshold": threshold}

    import bq_client
    """Produtos com queda de preço acima do threshold na janela de dias."""
    ds = f"`{settings.bq_project}.{settings.bq_dataset}`"

    busca_filter = "AND busca_id = @busca_id" if busca_id else ""

    sql = f"""
    WITH snapshots_janela AS (
      SELECT
        produto_id, nome, preco, comissao, imagem, link, loja,
        coletado_em,
        FIRST_VALUE(preco) OVER (PARTITION BY produto_id ORDER BY coletado_em ASC) AS preco_primeiro,
        LAST_VALUE(preco) OVER (PARTITION BY produto_id ORDER BY coletado_em ASC
          ROWS BETWEEN UNBOUNDED PRECEDING AND UNBOUNDED FOLLOWING) AS preco_atual
      FROM {ds}.snapshots
      WHERE coletado_em >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL @dias DAY)
        {busca_filter}
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

    params = [
        ScalarQueryParameter("dias", "INT64", dias),
        ScalarQueryParameter("threshold", "FLOAT64", threshold),
        ScalarQueryParameter("limit", "INT64", limit),
    ]
    if busca_id:
        params.append(ScalarQueryParameter("busca_id", "STRING", busca_id))

    rows = bq_client.query(sql, params=params)

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
