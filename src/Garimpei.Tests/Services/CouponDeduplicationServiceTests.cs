using Garimpei.Domain.Entities;
using Garimpei.Domain.Services;
using Xunit;

namespace Garimpei.Tests.Services;

public class CouponDeduplicationServiceTests
{
    // ═════════════════════════════════════════════════════════════════════
    // ShouldAlert
    // ═════════════════════════════════════════════════════════════════════

    [Fact]
    public void ShouldAlert_NoHistory_ReturnsTrue()
    {
        var result = CouponDeduplicationService.ShouldAlert(null, "newly_discovered", 20.0);
        Assert.True(result);
    }

    [Fact]
    public void ShouldAlert_NewlyDiscovered_WithinWindow_ReturnsFalse()
    {
        var history = new CouponAlertHistory
        {
            AlertedAt = DateTime.UtcNow.AddHours(-12), // within 24h window
            AlertedDiscountValue = 20.0
        };

        var result = CouponDeduplicationService.ShouldAlert(history, "newly_discovered", 20.0);
        Assert.False(result);
    }

    [Fact]
    public void ShouldAlert_Modified_WithinWindow_DiscountIncreased_ReturnsTrue()
    {
        var history = new CouponAlertHistory
        {
            AlertedAt = DateTime.UtcNow.AddHours(-6),
            AlertedDiscountValue = 15.0
        };

        // Discount went from 15% to 25% → should re-alert
        var result = CouponDeduplicationService.ShouldAlert(history, "modified", 25.0);
        Assert.True(result);
    }

    [Fact]
    public void ShouldAlert_Modified_WithinWindow_DiscountDecreased_ReturnsFalse()
    {
        var history = new CouponAlertHistory
        {
            AlertedAt = DateTime.UtcNow.AddHours(-6),
            AlertedDiscountValue = 25.0
        };

        // Discount went from 25% to 20% → don't re-alert
        var result = CouponDeduplicationService.ShouldAlert(history, "modified", 20.0);
        Assert.False(result);
    }

    [Fact]
    public void ShouldAlert_Modified_OutsideWindow_Changed_ReturnsTrue()
    {
        var history = new CouponAlertHistory
        {
            AlertedAt = DateTime.UtcNow.AddHours(-30), // outside 24h window
            AlertedDiscountValue = 15.0
        };

        var result = CouponDeduplicationService.ShouldAlert(history, "modified", 20.0);
        Assert.True(result);
    }

    [Fact]
    public void ShouldAlert_Modified_OutsideWindow_SameValue_ReturnsFalse()
    {
        var history = new CouponAlertHistory
        {
            AlertedAt = DateTime.UtcNow.AddHours(-30),
            AlertedDiscountValue = 20.0
        };

        // Same discount, outside window → no re-alert
        var result = CouponDeduplicationService.ShouldAlert(history, "modified", 20.0);
        Assert.False(result);
    }

    // ═════════════════════════════════════════════════════════════════════
    // MatchesRule
    // ═════════════════════════════════════════════════════════════════════

    [Fact]
    public void MatchesRule_AllCriteriaMatch_ReturnsTrue()
    {
        var rule = new CouponAlertRule
        {
            DiscountType = "percentage",
            MinDiscount = 10.0,
            Marketplaces = "shopee,amazon",
            Categories = "electronics,fashion"
        };

        var result = CouponDeduplicationService.MatchesRule(
            rule,
            couponDiscountType: "percentage",
            couponDiscountValue: 20.0,
            couponMarketplace: "shopee",
            couponCategories: ["electronics", "home"]);

        Assert.True(result);
    }

    [Fact]
    public void MatchesRule_WrongDiscountType_ReturnsFalse()
    {
        var rule = new CouponAlertRule
        {
            DiscountType = "percentage",
            MinDiscount = 10.0,
            Marketplaces = "shopee"
        };

        var result = CouponDeduplicationService.MatchesRule(
            rule,
            couponDiscountType: "fixed",
            couponDiscountValue: 15.0,
            couponMarketplace: "shopee",
            couponCategories: []);

        Assert.False(result);
    }

    [Fact]
    public void MatchesRule_DiscountBelowMinimum_ReturnsFalse()
    {
        var rule = new CouponAlertRule
        {
            DiscountType = "percentage",
            MinDiscount = 20.0,
            Marketplaces = "shopee"
        };

        var result = CouponDeduplicationService.MatchesRule(
            rule,
            couponDiscountType: "percentage",
            couponDiscountValue: 15.0, // below 20% minimum
            couponMarketplace: "shopee",
            couponCategories: []);

        Assert.False(result);
    }

    [Fact]
    public void MatchesRule_WrongMarketplace_ReturnsFalse()
    {
        var rule = new CouponAlertRule
        {
            DiscountType = "percentage",
            MinDiscount = 5.0,
            Marketplaces = "amazon"
        };

        var result = CouponDeduplicationService.MatchesRule(
            rule,
            couponDiscountType: "percentage",
            couponDiscountValue: 30.0,
            couponMarketplace: "shopee", // rule only targets amazon
            couponCategories: []);

        Assert.False(result);
    }

    [Fact]
    public void MatchesRule_NoCategoryOverlap_ReturnsFalse()
    {
        var rule = new CouponAlertRule
        {
            DiscountType = "percentage",
            MinDiscount = 5.0,
            Marketplaces = "shopee",
            Categories = "electronics,tech"
        };

        var result = CouponDeduplicationService.MatchesRule(
            rule,
            couponDiscountType: "percentage",
            couponDiscountValue: 30.0,
            couponMarketplace: "shopee",
            couponCategories: ["fashion", "beauty"]); // no overlap

        Assert.False(result);
    }

    [Fact]
    public void MatchesRule_NoCategoriesInRule_MatchesAnyCouponCategory()
    {
        var rule = new CouponAlertRule
        {
            DiscountType = "percentage",
            MinDiscount = 5.0,
            Marketplaces = "shopee",
            Categories = "" // no category filter
        };

        var result = CouponDeduplicationService.MatchesRule(
            rule,
            couponDiscountType: "percentage",
            couponDiscountValue: 30.0,
            couponMarketplace: "shopee",
            couponCategories: ["anything"]);

        Assert.True(result); // no category restriction
    }

    [Fact]
    public void MatchesRule_FixedDiscount_MatchesCorrectly()
    {
        var rule = new CouponAlertRule
        {
            DiscountType = "fixed",
            MinDiscount = 10.0, // R$10 minimum
            Marketplaces = "amazon"
        };

        var result = CouponDeduplicationService.MatchesRule(
            rule,
            couponDiscountType: "fixed",
            couponDiscountValue: 25.0, // R$25 > R$10
            couponMarketplace: "amazon",
            couponCategories: []);

        Assert.True(result);
    }
}
