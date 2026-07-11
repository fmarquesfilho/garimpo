"""Saúde das coletas: última execução, atrasos, keywords paradas."""

from datetime import datetime, timezone

from fastapi import APIRouter

from config import settings

router = APIRouter(tags=["Saúde"])


@router.get("/coletas/saude")
def get_saude_coletas():
    if settings.mock_data:
        return {
            "ultima_coleta": "2026-07-10T18:00:00Z",
            "minutos_desde_ultima": 120,
            "status": "ok",
            "coletas_24h": 6,
            "coletas_esperadas_24h": 9,
            "keywords_atrasadas": [],
        }

    import bq_client
    ds = f"`{settings.bq_project}.{settings.bq_dataset}`"

    # Última coleta e contagem 24h
    sql = f"""
    SELECT
      MAX(coletado_em) AS ultima_coleta,
      COUNTIF(coletado_em >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 24 HOUR)) AS coletas_24h
    FROM {ds}.snapshots
    """

    try:
        rows = bq_client.query(sql)
    except Exception:
        return {
            "ultima_coleta": None,
            "minutos_desde_ultima": None,
            "status": "sem_dados",
            "coletas_24h": 0,
            "coletas_esperadas_24h": 0,
            "keywords_atrasadas": [],
        }

    if not rows or rows[0].get("ultima_coleta") is None:
        return {
            "ultima_coleta": None,
            "minutos_desde_ultima": None,
            "status": "sem_dados",
            "coletas_24h": 0,
            "coletas_esperadas_24h": 0,
            "keywords_atrasadas": [],
        }

    row = rows[0]
    ultima = row["ultima_coleta"]
    agora = datetime.now(timezone.utc)
    minutos = int((agora - ultima).total_seconds() / 60) if ultima else None
    status_val = "atrasado" if minutos and minutos > 360 else "ok"

    # Keywords ativas (vistas nos últimos 7 dias) vs sem coleta nas últimas 24h
    sql_kw = f"""
    WITH ativas AS (
      SELECT DISTINCT keyword
      FROM {ds}.snapshots
      WHERE coletado_em >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 7 DAY)
    ),
    recentes AS (
      SELECT DISTINCT keyword
      FROM {ds}.snapshots
      WHERE coletado_em >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 24 HOUR)
    )
    SELECT a.keyword
    FROM ativas a
    LEFT JOIN recentes r ON a.keyword = r.keyword
    WHERE r.keyword IS NULL
    ORDER BY a.keyword
    LIMIT 20
    """

    try:
        kw_rows = bq_client.query(sql_kw)
        atrasadas = [r["keyword"] for r in kw_rows]
    except Exception:
        atrasadas = []

    # Esperadas = keywords distintas nos últimos 7 dias (1 coleta/dia mínimo)
    esperadas = row.get("coletas_24h", 0) + len(atrasadas)

    return {
        "ultima_coleta": ultima.isoformat() if ultima else None,
        "minutos_desde_ultima": minutos,
        "status": status_val,
        "coletas_24h": row.get("coletas_24h", 0),
        "coletas_esperadas_24h": esperadas,
        "keywords_atrasadas": atrasadas,
    }
