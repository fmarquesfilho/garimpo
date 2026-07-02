using Garimpei.Domain.Entities;

namespace Garimpei.Domain.Services;

/// <summary>
/// Evaluates whether a coupon alert should be dispatched based on deduplication rules.
/// - 24h window per coupon_id × alert_rule_id
/// - Re-alert on "modified" only if discount_value increased
/// - Reset on rule edit (handled externally by clearing history)
/// </summary>
public static class CouponDeduplicationService
{
    /// <summary>
    /// Determines if a coupon should trigger an alert given existing history.
    /// </summary>
    /// <param name="history">Most recent alert history for this coupon+rule (null if none).</param>
    /// <param name="detectionStatus">Detection classification: "newly_discovered" or "modified".</param>
    /// <param name="currentDiscountValue">The coupon's current discount value.</param>
    /// <returns>True if alert should be sent.</returns>
    public static bool ShouldAlert(
        CouponAlertHistory? history,
        string detectionStatus,
        double currentDiscountValue)
    {
        // No previous alert → always send
        if (history is null)
            return true;

        // Within 24h dedup window?
        if (history.AlertedAt > DateTime.UtcNow.AddHours(-24))
        {
            // Modified coupon: re-alert only if discount increased
            if (detectionStatus == "modified")
                return currentDiscountValue > history.AlertedDiscountValue;

            // newly_discovered within window → skip (already alerted)
            return false;
        }

        // Outside window: only alert if modified since last alert
        if (detectionStatus == "modified")
            return currentDiscountValue != history.AlertedDiscountValue;

        return false;
    }

    /// <summary>
    /// Checks if a coupon matches an alert rule's criteria.
    /// </summary>
    public static bool MatchesRule(
        CouponAlertRule rule,
        string couponDiscountType,
        double couponDiscountValue,
        string couponMarketplace,
        string[] couponCategories)
    {
        // Discount type must match
        if (rule.DiscountType != couponDiscountType)
            return false;

        // Discount value must meet minimum threshold
        if (couponDiscountValue < rule.MinDiscount)
            return false;

        // Marketplace must be in rule's list
        var ruleMarketplaces = rule.GetMarketplaceList();
        if (!ruleMarketplaces.Contains(couponMarketplace, StringComparer.OrdinalIgnoreCase))
            return false;

        // Category check (optional — if rule has categories, coupon must overlap)
        var ruleCategories = rule.GetCategoryList();
        if (ruleCategories.Length > 0 && couponCategories.Length > 0)
        {
            var hasOverlap = ruleCategories.Any(rc =>
                couponCategories.Any(cc => cc.Equals(rc, StringComparison.OrdinalIgnoreCase)));
            if (!hasOverlap)
                return false;
        }

        return true;
    }
}
