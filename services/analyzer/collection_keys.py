"""Derivação determinística de collection_keys a partir de uma Busca.

Lógica idêntica em Go, Python, C# e TypeScript.
"""


def derive_collection_keys(
    shop_ids: list[int] | None,
    keywords: list[str] | None,
    categorias: list[str] | None = None,
) -> list[str]:
    """Computa collection_keys a partir de shop_ids, keywords e categorias.

    Regras:
      - Cada shop_id vira sua representação string
      - Cada keyword é trimmed e lowercased
      - Cada categoria é trimmed e lowercased (fallback para tipo=categoria)
      - Strings vazias após normalização são descartadas
      - Resultado é sorted lexicograficamente e sem duplicatas
    """
    seen: set[str] = set()
    keys: list[str] = []

    for sid in shop_ids or []:
        s = str(sid)
        if s not in seen:
            seen.add(s)
            keys.append(s)

    for kw in keywords or []:
        normalized = kw.strip().lower()
        if normalized and normalized not in seen:
            seen.add(normalized)
            keys.append(normalized)

    # Categorias are used as collection keys ONLY when shop_ids and keywords are both empty
    if not (shop_ids or []) and not (keywords or []):
        for cat in categorias or []:
            normalized = cat.strip().lower()
            if normalized and normalized not in seen:
                seen.add(normalized)
                keys.append(normalized)

    keys.sort()
    return keys
