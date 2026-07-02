using Garimpei.Domain.ValueObjects;

namespace Garimpei.Domain.Interfaces;

/// <summary>
/// Porta de entrada para busca de produtos candidatos.
/// Cada marketplace implementa esta interface — o código consumidor nunca sabe
/// qual marketplace está usando (Strategy Pattern via Keyed DI).
///
/// Adicionar um novo marketplace = criar nova classe + registrar com AddKeyedScoped.
/// Zero mudanças nos endpoints ou scoring.
/// </summary>
public interface IProductSource
{
    /// <summary>
    /// Identificador do marketplace servido por esta implementação.
    /// </summary>
    string MarketplaceId { get; }

    /// <summary>
    /// Busca produtos por keyword.
    /// </summary>
    Task<SourceResult> SearchAsync(SearchQuery query, CancellationToken ct = default);

    /// <summary>
    /// Busca produtos de uma loja específica.
    /// Pode retornar resultado vazio se o marketplace não suporta busca por loja.
    /// </summary>
    Task<SourceResult> FetchByShopAsync(string shopId, int limit, CancellationToken ct = default);

    /// <summary>
    /// Gera o link de afiliado para o produto dado.
    /// </summary>
    string GenerateAffiliateLink(string productUrl, string affiliateTag);
}

/// <summary>
/// Parâmetros de busca genéricos para qualquer marketplace.
/// </summary>
public sealed record SearchQuery
{
    public required string Keyword { get; init; }
    public int Limit { get; init; } = 50;
    public string? SortBy { get; init; }
    public string? OwnerUid { get; init; }
}

/// <summary>
/// Resultado padronizado de uma busca.
/// </summary>
public sealed record SourceResult
{
    public required IReadOnlyList<ProductCandidate> Products { get; init; }
    public int TotalFound { get; init; }
    public DateTime FetchedAt { get; init; } = DateTime.UtcNow;
}
