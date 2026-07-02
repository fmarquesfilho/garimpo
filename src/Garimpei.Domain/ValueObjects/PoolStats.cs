namespace Garimpei.Domain.ValueObjects;

/// <summary>
/// Intermediate statistics for min-max normalization in the scoring engine.
/// Computed from the current product pool before scoring.
/// </summary>
public sealed record PoolStats
{
    public double MinCommission { get; init; }
    public double MaxCommission { get; init; }
    public double MinEV { get; init; }
    public double MaxEV { get; init; }
    public double MinRating { get; init; }
    public double MaxRating { get; init; }
    public double CommissionP75 { get; init; }
}
