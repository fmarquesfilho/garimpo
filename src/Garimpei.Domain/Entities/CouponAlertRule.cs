using Garimpei.Domain.Interfaces;

namespace Garimpei.Domain.Entities;

/// <summary>
/// User-defined rule for coupon alert triggers.
/// Coupons matching this rule generate notifications on the configured channel.
/// </summary>
public sealed class CouponAlertRule : IOwnedEntity
{
    public Guid Id { get; init; } = Guid.NewGuid();
    public string OwnerUid { get; set; } = string.Empty;

    /// <summary>Type of discount threshold: "percentage" or "fixed".</summary>
    public string DiscountType { get; set; } = "percentage";

    /// <summary>Minimum discount value to trigger alert (e.g. 10 for 10% or R$10).</summary>
    public double MinDiscount { get; set; }

    /// <summary>Target marketplaces (stored as comma-separated: "shopee,amazon").</summary>
    public string Marketplaces { get; set; } = "shopee";

    /// <summary>Optional target categories (stored as comma-separated, max 10).</summary>
    public string Categories { get; set; } = "";

    /// <summary>Notification channel: "telegram" or "whatsapp".</summary>
    public string Channel { get; set; } = "telegram";

    public bool IsActive { get; set; } = true;
    public DateTime CreatedAt { get; init; } = DateTime.UtcNow;
    public DateTime UpdatedAt { get; set; } = DateTime.UtcNow;

    public string[] GetMarketplaceList() =>
        Marketplaces.Split(',', StringSplitOptions.RemoveEmptyEntries | StringSplitOptions.TrimEntries);

    public string[] GetCategoryList() =>
        string.IsNullOrWhiteSpace(Categories)
            ? []
            : Categories.Split(',', StringSplitOptions.RemoveEmptyEntries | StringSplitOptions.TrimEntries);
}
