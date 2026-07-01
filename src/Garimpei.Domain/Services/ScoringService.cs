using Garimpei.Domain.ValueObjects;

namespace Garimpei.Domain.Services;

/// <summary>
/// Curation scoring engine. Scores products using weighted normalization:
///   score = 0.45 * norm(commission) + 0.35 * norm(EV) + 0.20 * norm(rating)
///
/// EV (expected value) = commission * price * sales30d
/// Normalization is min-max relative to the current pool.
/// </summary>
public static class ScoringService
{
    /// <summary>
    /// Rank products by scoring. Applies eligibility filter first, then scores survivors.
    /// </summary>
    public static IReadOnlyList<ScoredProduct> Rank(
        IEnumerable<ProductCandidate> candidates,
        EligibilityFilter filter,
        int top = 20)
    {
        var eligible = candidates.Where(filter.Passes).ToList();
        if (eligible.Count == 0) return [];

        var stats = ComputeStats(eligible);

        var scored = eligible
            .Select(p => Score(p, stats))
            .OrderByDescending(s => s.Score)
            .Take(top)
            .ToList();

        return scored;
    }

    private static ScoredProduct Score(ProductCandidate p, PoolStats stats)
    {
        var nComm = MinMax(p.Commission, stats.MinCommission, stats.MaxCommission);
        var ev = p.Commission * (double)p.Price * p.Sales;
        var nEv = MinMax(ev, stats.MinEV, stats.MaxEV);
        var nRating = MinMax(p.Rating, stats.MinRating, stats.MaxRating);

        var components = new ScoreComponents
        {
            Commission = 0.45 * nComm,
            ExpectedValue = 0.35 * nEv,
            Rating = 0.20 * nRating
        };

        var score = components.Commission + components.ExpectedValue + components.Rating;

        var suspicious = p.Commission >= stats.CommissionP75
                         && stats.CommissionP75 > 0
                         && (p.Sales == 0 || p.Rating == 0);

        return new ScoredProduct
        {
            Id = p.Id,
            Name = p.Name,
            Category = p.Category,
            ShopName = p.ShopName,
            ShopId = p.ShopId,
            Origin = p.Origin,
            Price = p.Price,
            OriginalPrice = p.OriginalPrice,
            DiscountPercent = p.DiscountPercent,
            Commission = p.Commission,
            Sales = p.Sales,
            Rating = p.Rating,
            Link = p.Link,
            ProductLink = p.ProductLink,
            ImageUrl = p.ImageUrl,
            Score = score,
            Components = components,
            Suspicious = suspicious,
            OfferExpiresAt = p.OfferExpiresAt
        };
    }

    private static PoolStats ComputeStats(List<ProductCandidate> products)
    {
        var commissions = products.Select(p => p.Commission).ToList();
        var evs = products.Select(p => p.Commission * (double)p.Price * p.Sales).ToList();
        var ratings = products.Select(p => p.Rating).ToList();

        commissions.Sort();

        return new PoolStats
        {
            MinCommission = commissions[0],
            MaxCommission = commissions[^1],
            MinEV = evs.Min(),
            MaxEV = evs.Max(),
            MinRating = ratings.Min(),
            MaxRating = ratings.Max(),
            CommissionP75 = Percentile(commissions, 0.75)
        };
    }

    private static double MinMax(double value, double min, double max)
        => max == min ? 0.5 : (value - min) / (max - min);

    private static double Percentile(List<double> sorted, double p)
    {
        if (sorted.Count == 0) return 0;
        var idx = (int)Math.Ceiling(p * sorted.Count) - 1;
        return sorted[Math.Clamp(idx, 0, sorted.Count - 1)];
    }

    private sealed record PoolStats
    {
        public double MinCommission { get; init; }
        public double MaxCommission { get; init; }
        public double MinEV { get; init; }
        public double MaxEV { get; init; }
        public double MinRating { get; init; }
        public double MaxRating { get; init; }
        public double CommissionP75 { get; init; }
    }
}

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

/// <summary>
/// Eligibility filter — products below thresholds are excluded before scoring.
/// </summary>
public sealed record EligibilityFilter
{
    public double MinCommission { get; init; } = 0.07;
    public int MinSales { get; init; } = 0;
    public double MinRating { get; init; } = 0;

    public bool Passes(ProductCandidate p)
    {
        if (p.Commission < MinCommission) return false;
        if (MinSales > 0 && p.Sales < MinSales) return false;
        if (MinRating > 0 && p.Rating < MinRating) return false;
        return true;
    }
}
