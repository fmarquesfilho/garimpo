using Garimpei.Domain.Entities;
using Garimpei.Domain.Services;
using Garimpei.Infrastructure.Persistence;
using Microsoft.EntityFrameworkCore;

namespace Garimpei.Api.Endpoints.Coupons;

public static class CouponAlertEvaluationEndpoints
{
    public static WebApplication MapCouponAlertEvaluationEndpoints(this WebApplication app)
    {
        // Internal endpoint — no auth (called by Python analyzer on internal network)
        app.MapPost("/internal/coupon-alerts/evaluate", async (
            AppDbContext db,
            CouponDetectionPayload payload,
            CancellationToken ct) =>
        {
            if (string.IsNullOrWhiteSpace(payload.OwnerUid) || payload.Detections is null || payload.Detections.Count == 0)
                return Results.Ok(new { alerts_sent = 0 });

            // Load active rules for this tenant (bypass query filter using owner_uid directly)
            var rules = await db.CouponAlertRules
                .IgnoreQueryFilters()
                .Where(r => r.OwnerUid == payload.OwnerUid && r.IsActive)
                .ToListAsync(ct);

            if (rules.Count == 0)
                return Results.Ok(new { alerts_sent = 0 });

            // Filter to actionable detections only
            var actionable = payload.Detections
                .Where(d => d.DetectionStatus is "newly_discovered" or "modified")
                .ToList();

            if (actionable.Count == 0)
                return Results.Ok(new { alerts_sent = 0 });

            var alertsSent = 0;

            foreach (var detection in actionable)
            {
                var couponCategories = detection.ApplicableCategories?.ToArray() ?? [];

                foreach (var rule in rules)
                {
                    // Check rule match
                    if (!CouponDeduplicationService.MatchesRule(
                        rule,
                        detection.DiscountType,
                        detection.DiscountValue,
                        payload.Marketplace,
                        couponCategories))
                    {
                        continue;
                    }

                    // Check deduplication
                    var history = await db.CouponAlertHistories
                        .IgnoreQueryFilters()
                        .Where(h => h.CouponId == detection.CouponId
                                 && h.AlertRuleId == rule.Id
                                 && h.OwnerUid == payload.OwnerUid)
                        .OrderByDescending(h => h.AlertedAt)
                        .FirstOrDefaultAsync(ct);

                    if (!CouponDeduplicationService.ShouldAlert(history, detection.DetectionStatus, detection.DiscountValue))
                        continue;

                    // Record alert in history
                    db.CouponAlertHistories.Add(new CouponAlertHistory
                    {
                        OwnerUid = payload.OwnerUid,
                        CouponId = detection.CouponId,
                        AlertRuleId = rule.Id,
                        AlertedDiscountValue = detection.DiscountValue,
                        AlertedAt = DateTime.UtcNow,
                        ExpiresAt = DateTime.UtcNow.AddHours(72)
                    });

                    alertsSent++;
                    // TODO(T-0045): dispatch via Alerter gRPC (Task 12)
                }
            }

            if (alertsSent > 0)
                await db.SaveChangesAsync(ct);

            return Results.Ok(new { alerts_sent = alertsSent, owner_uid = payload.OwnerUid });
        });

        return app;
    }
}

public sealed record CouponDetectionPayload
{
    public required string OwnerUid { get; init; }
    public required string Marketplace { get; init; }
    public List<CouponDetectionItem>? Detections { get; init; }
}

public sealed record CouponDetectionItem
{
    public required string CouponId { get; init; }
    public required string DetectionStatus { get; init; }
    public required string DiscountType { get; init; }
    public required double DiscountValue { get; init; }
    public string? EndTime { get; init; }
    public List<string>? ApplicableCategories { get; init; }
}
