namespace Garimpei.Domain.ValueObjects;

/// <summary>
/// A product scored by the curation engine with its components breakdown.
/// </summary>
public sealed record ScoredProduct
{
    public required string Id { get; init; }
    public required string Name { get; init; }
    public string? Category { get; init; }
    public string? ShopName { get; init; }
    public string? ShopId { get; init; }
    public string? Origin { get; init; }
    public decimal Price { get; init; }
    public decimal OriginalPrice { get; init; }
    public double DiscountPercent { get; init; }
    public double Commission { get; init; }
    public int Sales { get; init; }
    public double Rating { get; init; }
    public string? Link { get; init; }
    public string? ProductLink { get; init; }
    public string? ImageUrl { get; init; }
    public double Score { get; init; }
    public required ScoreComponents Components { get; init; }
    public bool Suspicious { get; init; }
    public string? OfferExpiresAt { get; init; }

    /// <summary>
    /// Marketplace from which this product was collected.
    /// </summary>
    public string Marketplace { get; init; } = Domain.Marketplaces.Shopee;
}

public sealed record ScoreComponents
{
    public double Commission { get; init; }
    public double ExpectedValue { get; init; }
    public double Rating { get; init; }
}
