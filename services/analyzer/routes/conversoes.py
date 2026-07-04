"""Conversões: relatório de conversões reais da Shopee (se disponível)."""

from fastapi import APIRouter, Query

from config import settings

router = APIRouter(tags=["Conversões"])


@router.get("/conversoes")
def get_conversoes(
    dias: int = Query(30, ge=1, le=180),
):
    if settings.mock_data:
        return {"conversoes": [], "total": 0, "dias_janela": dias}

    import bq_client
    """Conversões reais da Shopee (tabela conversoes, preenchida via webhook)."""
    ds = bq_client.dataset_ref()

    sql = f"""
    SELECT
      sub_id,
      produto_id,
      nome,
      canal,
      comissao,
      preco,
      convertido_em
    FROM {ds}.conversoes
    WHERE convertido_em >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL @dias DAY)
    ORDER BY convertido_em DESC
    LIMIT 100
    """

    from google.cloud.bigquery import ScalarQueryParameter

    try:
        rows = bq_client.query(sql, params=[
            ScalarQueryParameter("dias", "INT64", dias),
        ])
    except Exception:
        # Tabela pode não existir ainda
        return {
            "fonte": "shopee-api",
            "status": "sem_dados",
            "conversoes": [],
            "total": 0,
            "dias_janela": dias,
        }

    conversoes = [
        {
            "sub_id": row.get("sub_id", ""),
            "produto_id": row.get("produto_id", ""),
            "nome": row.get("nome", ""),
            "canal": row.get("canal", ""),
            "comissao": row.get("comissao", 0),
            "preco": row.get("preco", 0),
            "convertido_em": str(row.get("convertido_em", "")),
        }
        for row in rows
    ]

    return {
        "fonte": "shopee-api",
        "conversoes": conversoes,
        "total": len(conversoes),
        "dias_janela": dias,
    }
