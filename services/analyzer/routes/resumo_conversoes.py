"""Resumo de conversões: comissão total, canais, melhor canal."""

from fastapi import APIRouter, Query

from config import settings

router = APIRouter(tags=["Conversões"])


@router.get("/conversoes/resumo")
def get_resumo_conversoes(
    dias: int = Query(30, ge=1, le=180),
):
    if settings.mock_data:
        return {
            "dias": dias,
            "comissao_total": 0,
            "conversoes": 0,
            "produtos_distintos": 0,
            "por_canal": [],
            "melhor_canal": None,
            "status": "sem_dados",
        }

    import bq_client
    from google.cloud.bigquery import ScalarQueryParameter

    ds = f"`{settings.bq_project}.{settings.bq_dataset}`"

    sql = f"""
    SELECT
      IFNULL(canal, 'desconhecido') AS canal,
      SUM(comissao * preco) AS comissao_total,
      COUNT(*) AS conversoes,
      COUNT(DISTINCT produto_id) AS produtos_distintos
    FROM {ds}.conversoes
    WHERE convertido_em >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL @dias DAY)
    GROUP BY canal
    ORDER BY comissao_total DESC
    """

    try:
        rows = bq_client.query(sql, params=[
            ScalarQueryParameter("dias", "INT64", dias),
        ])
    except Exception:
        return {
            "dias": dias,
            "comissao_total": 0,
            "conversoes": 0,
            "produtos_distintos": 0,
            "por_canal": [],
            "melhor_canal": None,
            "status": "sem_dados",
        }

    if not rows:
        return {
            "dias": dias,
            "comissao_total": 0,
            "conversoes": 0,
            "produtos_distintos": 0,
            "por_canal": [],
            "melhor_canal": None,
            "status": "sem_dados",
        }

    por_canal = []
    comissao_total = 0.0
    conversoes_total = 0
    produtos_total = 0
    melhor_canal = None

    for r in rows:
        canal_comissao = round(r.get("comissao_total", 0) or 0, 2)
        canal_conv = r.get("conversoes", 0)
        por_canal.append({
            "canal": r["canal"],
            "comissao": canal_comissao,
            "conversoes": canal_conv,
        })
        comissao_total += canal_comissao
        conversoes_total += canal_conv
        produtos_total += r.get("produtos_distintos", 0)

    if por_canal:
        melhor_canal = por_canal[0]["canal"]  # already sorted DESC

    return {
        "dias": dias,
        "comissao_total": round(comissao_total, 2),
        "conversoes": conversoes_total,
        "produtos_distintos": produtos_total,
        "por_canal": por_canal,
        "melhor_canal": melhor_canal,
        "status": "ok",
    }
