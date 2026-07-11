namespace Garimpei.Domain;

/// <summary>
/// Derivação determinística de collection_keys a partir de uma Busca.
/// Lógica idêntica em Go, Python, C# e TypeScript.
/// </summary>
public static class CollectionKeys
{
    /// <summary>
    /// Computa collection_keys a partir de shop_ids e keywords.
    /// <para>
    /// Regras:
    /// - Cada shop_id vira sua representação string
    /// - Cada keyword é trimmed e lowercased
    /// - Strings vazias após normalização são descartadas
    /// - Resultado é sorted lexicograficamente e sem duplicatas
    /// </para>
    /// </summary>
    public static string[] Derive(long[]? shopIds, string[]? keywords, string[]? categorias = null)
    {
        var seen = new HashSet<string>();
        var keys = new List<string>();

        if (shopIds is not null)
        {
            foreach (var id in shopIds)
            {
                var s = id.ToString();
                if (seen.Add(s))
                    keys.Add(s);
            }
        }

        if (keywords is not null)
        {
            foreach (var kw in keywords)
            {
                var normalized = kw.Trim().ToLowerInvariant();
                if (normalized.Length > 0 && seen.Add(normalized))
                    keys.Add(normalized);
            }
        }

        if (categorias is not null && (shopIds is null || shopIds.Length == 0) && (keywords is null || keywords.Length == 0))
        {
            foreach (var cat in categorias)
            {
                var normalized = cat.Trim().ToLowerInvariant();
                if (normalized.Length > 0 && seen.Add(normalized))
                    keys.Add(normalized);
            }
        }

        keys.Sort(StringComparer.Ordinal);
        return keys.ToArray();
    }
}
