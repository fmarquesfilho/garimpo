"""Eficácia dos alertas: quedas detectadas → alertas enviados → conversões."""

from fastapi import APIRouter, Query

from config import settings

router = APIRouter(tags=["Alertas"])


@router.get("/alertas/eficacia")
def get_eficacia_alertas(
    dias: int = Query(30, ge=1, le=180),
):
    if settings.mock_data:
        return {
            "dias": dias,
            "quedas_detectadas": 0,
            "alertas_enviados": 0,
            "conversoes_atribuidas": 0,
            "taxa_deteccao": None,
            "taxa_conversao": None,
            "melhor_keyword": None,
            "conversoes_disponiveis": False,
        }

    import bq_client
    from google.cloud.bigquery import ScalarQueryParameter

    ds = f"`{settings.bq_project}.{settings.bq_dataset}`"

    # 1. Price drops detected (variacao <= -15%)
    sql_quedas = f"""
    WITH primeiros AS (
      SELECT produto_id, preco AS p1, keyword,
        ROW_NUMBER() OVER (PARTITION BY produto_id ORDER BY coletado_em ASC) AS rn
      FROM {ds}.snapshots
      WHERE coletado_em >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL @dias DAY)
    ),
    ultimos AS (
      SELECT produto_id, preco AS p2,
        ROW_NUMBER() OVER (PARTITION BY produto_id ORDER BY coletado_em DESC) AS rn
      FROM {ds}.snapshots
      WHERE coletado_em >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL @dias DAY)
    )
    SELECT
      COUNT(*) AS quedas_detectadas,
      COUNT(DISTINCT p.keyword) AS keywords_com_queda
    FROM primeiros p
    JOIN ultimos u ON p.produto_id = u.produto_id AND u.rn = 1
    WHERE p.rn = 1 AND p.p1 > 0
      AND SAFE_DIVIDE(u.p2 - p.p1, p.p1) <= -0.15
    """

    quedas_detectadas = 0
    try:
        rows = bq_client.query(sql_quedas, params=[
            ScalarQueryParameter("dias", "INT64", dias),
        ])
        if rows:
            quedas_detectadas = rows[0].get("quedas_detectadas", 0)
    except Exception:
        pass

    if quedas_detectadas == 0:
        return {
            "dias": dias,
            "quedas_detectadas": 0,
            "alertas_enviados": 0,
            "conversoes_atribuidas": 0,
            "taxa_deteccao": None,
            "taxa_conversao": None,
            "melhor_keyword": None,
            "conversoes_disponiveis": False,
        }

    # 2. Alerts sent (heuristic: publications with estrategia containing 'alert' or 'coleta')
    sql_alertas = f"""
    SELECT COUNT(*) AS alertas_enviados
    FROM {ds}.publicacoes
    WHERE criada_em >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL @dias DAY)
      AND (LOWER(estrategia) LIKE '%alert%' OR LOWER(estrategia) LIKE '%coleta%' OR LOWER(detalhe) LIKE '%alerta%')
    """

    alertas_enviados = 0
    try:
        rows = bq_client.query(sql_alertas, params=[
            ScalarQueryParameter("dias", "INT64", dias),
        ])
        if rows:
            alertas_enviados = rows[0].get("alertas_enviados", 0)
    except Exception:
        pass

    # 3. Conversions attributed to dropped products
    conversoes_atribuidas = 0
    melhor_keyword = None
    conversoes_disponiveis = False

    sql_conv = f"""
    WITH drops AS (
      SELECT DISTINCT p.produto_id, p.keyword
      FROM (
        SELECT produto_id, keyword, preco AS p1,
          ROW_NUMBER() OVER (PARTITION BY produto_id ORDER BY coletado_em ASC) AS rn
        FROM {ds}.snapshots
        WHERE coletado_em >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL @dias DAY)
      ) p
      JOIN (
        SELECT produto_id, preco AS p2,
          ROW_NUMBER() OVER (PARTITION BY produto_id ORDER BY coletado_em DESC) AS rn
        FROM {ds}.snapshots
        WHERE coletado_em >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL @dias DAY)
      ) u ON p.produto_id = u.produto_id AND u.rn = 1
      WHERE p.rn = 1 AND p.p1 > 0
        AND SAFE_DIVIDE(u.p2 - p.p1, p.p1) <= -0.15
    )
    SELECT
      COUNT(*) AS conversoes,
      APPROX_TOP_COUNT(d.keyword, 1)[OFFSET(0)].value AS melhor_keyword
    FROM {ds}.conversoes c
    JOIN drops d ON c.produto_id = d.produto_id
    WHERE c.convertido_em >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL @dias DAY)
    """

    try:
        rows = bq_client.query(sql_conv, params=[
            ScalarQueryParameter("dias", "INT64", dias),
        ])
        if rows:
            conversoes_atribuidas = rows[0].get("conversoes", 0)
            melhor_keyword = rows[0].get("melhor_keyword")
        conversoes_disponiveis = True
    except Exception:
        # conversoes table doesn't exist
        conversoes_disponiveis = False

    taxa_deteccao = round(alertas_enviados / quedas_detectadas * 100, 1) if quedas_detectadas > 0 else None
    taxa_conversao = round(conversoes_atribuidas / alertas_enviados * 100, 1) if alertas_enviados > 0 else None

    return {
        "dias": dias,
        "quedas_detectadas": quedas_detectadas,
        "alertas_enviados": alertas_enviados,
        "conversoes_atribuidas": conversoes_atribuidas,
        "taxa_deteccao": taxa_deteccao,
        "taxa_conversao": taxa_conversao,
        "melhor_keyword": melhor_keyword,
        "conversoes_disponiveis": conversoes_disponiveis,
    }
