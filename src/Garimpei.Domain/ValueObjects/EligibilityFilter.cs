namespace Garimpei.Domain.ValueObjects;

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
