using Garimpei.Domain.Interfaces;

namespace Garimpei.Domain.Entities;

public sealed class Busca : IOwnedEntity
{
    public Guid Id { get; init; } = Guid.NewGuid();
    public required string Keyword { get; init; }
    public required string OwnerUid { get; set; }
    public string SortBy { get; init; } = "relevance";
    public int Limit { get; init; } = 50;
    public bool Active { get; set; } = true;
    public DateTime CreatedAt { get; init; } = DateTime.UtcNow;
    public DateTime UpdatedAt { get; set; } = DateTime.UtcNow;

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
