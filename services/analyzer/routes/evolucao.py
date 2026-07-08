"""Evolução: série temporal de preço médio, segmentada por fonte (loja vs keyword)."""

from fastapi import APIRouter, Query

from config import settings

router = APIRouter(tags=["Evolução"])


@router.get("/evolucao")
def get_evolucao(
    dias: int = Query(30, ge=1, le=180),
):
    if settings.mock_data:
        from mock_data import EVOLUCAO_RESPONSE
        return {**EVOLUCAO_RESPONSE, "dias_janela": dias}

    import bq_client
    from google.cloud.bigquery import ScalarQueryParameter

    ds = f"`{settings.bq_project}.{settings.bq_dataset}`"

    sql = f"""
    SELECT
      DATE(coletado_em) AS dia,
      keyword AS loja,
      AVG(preco) AS preco_medio,
      COUNT(DISTINCT produto_id) AS produtos,
      MIN(preco) AS preco_min,
      MAX(preco) AS preco_max
    FROM {ds}.snapshots
    WHERE coletado_em >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL @dias DAY)
    GROUP BY dia, loja
    ORDER BY dia DESC, loja
    """

    rows = bq_client.query(sql, params=[
        ScalarQueryParameter("dias", "INT64", dias),
    ])

    # Agrupar por keyword
    agrupado: dict = {}
    for row in rows:
        loja = row["loja"]
        if loja not in agrupado:
            agrupado[loja] = []
        agrupado[loja].append({
            "dia": str(row["dia"]),
            "preco_medio": round(row["preco_medio"], 2),
            "produtos": row["produtos"],
            "preco_min": row["preco_min"],
            "preco_max": row["preco_max"],
        })

    # Calcular resumo global com contagem de quedas e altas
    resumo_sql = f"""
    WITH primeiros AS (
      SELECT produto_id, preco AS preco_primeiro, keyword,
        ROW_NUMBER() OVER (PARTITION BY produto_id ORDER BY coletado_em ASC) AS rn
      FROM {ds}.snapshots
      WHERE coletado_em >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL @dias DAY)
    ),
    ultimos AS (
      SELECT produto_id, preco AS preco_atual,
        ROW_NUMBER() OVER (PARTITION BY produto_id ORDER BY coletado_em DESC) AS rn
      FROM {ds}.snapshots
      WHERE coletado_em >= TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL @dias DAY)
    )
    SELECT
      CASE WHEN p.keyword LIKE 'loja-%' THEN 'loja' ELSE 'keyword' END AS fonte,
      COUNTIF(SAFE_DIVIDE(u.preco_atual - p.preco_primeiro, p.preco_primeiro) < -0.01) AS total_quedas,
      COUNTIF(SAFE_DIVIDE(u.preco_atual - p.preco_primeiro, p.preco_primeiro) > 0.01) AS total_altas,
      COUNT(*) AS total_produtos,
      AVG(u.preco_atual) AS preco_medio_global,
      AVG(SAFE_DIVIDE(u.preco_atual - p.preco_primeiro, p.preco_primeiro)) AS variacao_media_global_pct
    FROM primeiros p
    JOIN ultimos u ON p.produto_id = u.produto_id AND u.rn = 1
    WHERE p.rn = 1 AND p.preco_primeiro > 0
    GROUP BY fonte
    """

    resumo_rows = bq_client.query(resumo_sql, params=[
        ScalarQueryParameter("dias", "INT64", dias),
    ])

    # Montar resumo global (soma das fontes) e resumo_keywords
    resumo = {
        "total_lojas": 0,
        "total_produtos": 0,
        "preco_medio_global": 0,
        "variacao_media_global_pct": 0,
        "total_quedas": 0,
        "total_altas": 0,
    }
    resumo_keywords = {"total_quedas": 0, "total_altas": 0}

    total_produtos_all = 0
    preco_sum = 0.0
    variacao_sum = 0.0

    for r in resumo_rows:
        fonte = r.get("fonte", "")
        quedas = r.get("total_quedas", 0)
        altas = r.get("total_altas", 0)
        produtos = r.get("total_produtos", 0)
        preco_med = r.get("preco_medio_global", 0) or 0
        var_med = r.get("variacao_media_global_pct", 0) or 0

        resumo["total_quedas"] += quedas
        resumo["total_altas"] += altas
        total_produtos_all += produtos
        preco_sum += preco_med * produtos
        variacao_sum += var_med * produtos

        if fonte == "keyword":
            resumo_keywords["total_quedas"] = quedas
            resumo_keywords["total_altas"] = altas

    resumo["total_produtos"] = total_produtos_all
    if total_produtos_all > 0:
        resumo["preco_medio_global"] = round(preco_sum / total_produtos_all, 2)
        resumo["variacao_media_global_pct"] = round(variacao_sum / total_produtos_all, 4)

    # Separar em lojas vs keywords
    lojas_lista = []
    keywords_lista = []

    for nome, serie in agrupado.items():
        pontos = sorted(serie, key=lambda x: x["dia"])
        variacao_media = 0.0
        if len(pontos) >= 2 and pontos[0]["preco_medio"] > 0:
            variacao_media = (pontos[-1]["preco_medio"] - pontos[0]["preco_medio"]) / pontos[0]["preco_medio"]
        entry = {
            "busca_id": nome,
            "produtos": pontos[-1]["produtos"] if pontos else 0,
            "preco_medio": pontos[-1]["preco_medio"] if pontos else 0,
            "variacao_media_pct": round(variacao_media, 4),
            "pontos": [{"data": p["dia"], "preco_medio": p["preco_medio"]} for p in pontos],
        }
        if nome.startswith("loja-"):
            lojas_lista.append(entry)
        else:
            keywords_lista.append(entry)

    resumo["total_lojas"] = len(lojas_lista)

    return {
        "dias": dias,
        "lojas": lojas_lista,
        "keywords": keywords_lista,
        "resumo": resumo,
        "resumo_keywords": resumo_keywords,
        "total_lojas": len(lojas_lista),
    }
