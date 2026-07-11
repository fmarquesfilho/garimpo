"""Change Detection: timestamps de última atualização por seção do dashboard."""

from fastapi import APIRouter

from config import settings

router = APIRouter(tags=["Dashboard"])


@router.get("/dashboard/changes")
def get_dashboard_changes():
    if settings.mock_data:
        from datetime import datetime, timezone
        now = datetime.now(timezone.utc).isoformat()
        return {
            "saude_updated_at": now,
            "oportunidades_updated_at": now,
            "performance_updated_at": None,
        }

    import bq_client
    ds = f"`{settings.bq_project}.{settings.bq_dataset}`"

    saude_ts = None
    oportunidades_ts = None
    performance_ts = None

    # Saúde: MAX(coletado_em) from snapshots
    try:
        rows = bq_client.query(f"SELECT MAX(coletado_em) AS ts FROM {ds}.snapshots")
        if rows and rows[0].get("ts"):
            saude_ts = rows[0]["ts"].isoformat()
    except Exception:
        pass

    # Oportunidades: MAX of snapshot or publication timestamp
    oportunidades_ts = saude_ts  # same as saude (new data = new opportunities)
    try:
        rows = bq_client.query(f"SELECT MAX(criada_em) AS ts FROM {ds}.publicacoes")
        if rows and rows[0].get("ts"):
            pub_ts = rows[0]["ts"].isoformat()
            if oportunidades_ts is None or pub_ts > oportunidades_ts:
                oportunidades_ts = pub_ts
    except Exception:
        pass

    # Performance: MAX(convertido_em) from conversoes
    try:
        rows = bq_client.query(f"SELECT MAX(convertido_em) AS ts FROM {ds}.conversoes")
        if rows and rows[0].get("ts"):
            performance_ts = rows[0]["ts"].isoformat()
    except Exception:
        pass

    return {
        "saude_updated_at": saude_ts,
        "oportunidades_updated_at": oportunidades_ts,
        "performance_updated_at": performance_ts,
    }
