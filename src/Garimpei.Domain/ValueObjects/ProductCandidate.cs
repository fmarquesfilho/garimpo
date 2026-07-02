namespace Garimpei.Domain.ValueObjects;

/// <summary>
/// Input to the scoring engine — a raw product candidate from any source.
/// </summary>
public sealed record ProductCandidate
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
    public string? OfferExpiresAt { get; init; }
}
