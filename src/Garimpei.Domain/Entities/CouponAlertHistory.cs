using Garimpei.Domain.Interfaces;

namespace Garimpei.Domain.Entities;

/// <summary>
/// Tracks coupon alert dispatches for deduplication.
/// Records expire 72h after AlertedAt (48h beyond the 24h dedup window).
/// </summary>
public sealed class CouponAlertHistory : IOwnedEntity
{
    public Guid Id { get; init; } = Guid.NewGuid();
    public string OwnerUid { get; set; } = string.Empty;

    public string CouponId { get; set; } = string.Empty;
    public Guid AlertRuleId { get; set; }
    public double AlertedDiscountValue { get; set; }
    public DateTime AlertedAt { get; set; } = DateTime.UtcNow;

    /// <summary>Record expires 72h after AlertedAt for cleanup.</summary>
    public DateTime ExpiresAt { get; set; }
}
