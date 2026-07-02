namespace Garimpei.Domain.Interfaces;

/// <summary>
/// Port for coupon collection per marketplace.
/// Registered via AddKeyedScoped — same Strategy pattern as IProductSource.
/// </summary>
public interface ICouponSource
{
    string MarketplaceId { get; }
    Task<CouponSourceResult> FetchCouponsAsync(string ownerUid, CancellationToken ct = default);
}

public sealed record CouponSourceResult
{
    public required IReadOnlyList<CouponCandidate> Coupons { get; init; }
    public int TotalFound { get; init; }
    public DateTime FetchedAt { get; init; } = DateTime.UtcNow;
}

public sealed record CouponCandidate
{
    public required string CouponId { get; init; }
    public required string Marketplace { get; init; }
    public string? Code { get; init; }
    public required string DiscountType { get; init; }
    public required double DiscountValue { get; init; }
    public double MinSpend { get; init; }
    public DateTime? StartTime { get; init; }
    public DateTime? EndTime { get; init; }
    public List<string> ApplicableCategories { get; init; } = [];
}
