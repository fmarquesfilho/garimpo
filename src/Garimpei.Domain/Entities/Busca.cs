using Garimpei.Domain.Interfaces;

namespace Garimpei.Domain.Entities;

public sealed class Busca : IOwnedEntity
{
    public Guid Id { get; init; } = Guid.NewGuid();
    public required string Keyword { get; set; }
    public required string OwnerUid { get; set; }
    public string SortBy { get; init; } = "relevance";
    public int Limit { get; init; } = 50;
    public bool Active { get; set; } = true;
    public DateTime CreatedAt { get; init; } = DateTime.UtcNow;
    public DateTime UpdatedAt { get; set; } = DateTime.UtcNow;

    /// <summary>
    /// Arrays de Shop IDs (ex: lojas monitoradas). Se nulo, trata-se de uma busca por palavra-chave genérica.
    /// </summary>
    public long[]? ShopIds { get; set; }

    /// <summary>
    /// URL original fornecida pelo usuário ao adicionar a loja (pode ser link de afiliada com tracking).
    /// Preservada para futura geração de links de conversão via generateShortLink.
    /// </summary>
    public string? SourceUrl { get; set; }

    /// <summary>
    /// Marketplaces to query for this search. Defaults to Shopee only.
    /// Stored as comma-separated string in the database (e.g. "shopee,amazon").
    /// </summary>
    public string Marketplaces { get; set; } = Domain.Marketplaces.Shopee;

    /// <summary>
    /// Returns the list of marketplace identifiers for this search.
    /// </summary>
    public string[] GetMarketplaceList() =>
        Marketplaces.Split(',', StringSplitOptions.RemoveEmptyEntries | StringSplitOptions.TrimEntries);
}
