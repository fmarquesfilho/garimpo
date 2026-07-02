using Garimpei.Domain;
using Garimpei.Domain.Entities;
using Garimpei.Domain.Interfaces;
using Garimpei.Infrastructure.Persistence;
using Microsoft.EntityFrameworkCore;

namespace Garimpei.Api.Endpoints.Coupons;

public static class CouponRulesEndpoints
{
    public static RouteGroupBuilder MapCouponRulesEndpoints(this RouteGroupBuilder group)
    {
        var rules = group.MapGroup("/cupons/regras")
            .RequireAuthorization()
            .WithTags("Cupons");

        rules.MapGet("/", async (AppDbContext db, CancellationToken ct) =>
        {
            var list = await db.CouponAlertRules
                .OrderByDescending(r => r.CreatedAt)
                .ToListAsync(ct);

            return Results.Ok(new { regras = list, total = list.Count });
        });

        rules.MapPost("/", async (AppDbContext db, ITenantContext tenant, CreateCouponRuleRequest req, CancellationToken ct) =>
        {
            // Validate max 20 active rules
            var activeCount = await db.CouponAlertRules.CountAsync(r => r.IsActive, ct);
            if (activeCount >= 20)
                return Results.Conflict(new { error = "Máximo de 20 regras ativas atingido" });

            // Validate fields
            if (req.MinDiscount <= 0)
                return Results.BadRequest(new { error = "min_discount deve ser > 0" });
            if (req.DiscountType == "percentage" && (req.MinDiscount < 1 || req.MinDiscount > 99))
                return Results.BadRequest(new { error = "min_discount deve ser entre 1 e 99 para percentage" });
            if (string.IsNullOrWhiteSpace(req.Marketplaces))
                return Results.BadRequest(new { error = "pelo menos um marketplace é obrigatório" });

            var categories = req.Categories ?? "";
            var catList = categories.Split(',', StringSplitOptions.RemoveEmptyEntries | StringSplitOptions.TrimEntries);
            if (catList.Length > 10)
                return Results.BadRequest(new { error = "máximo de 10 categorias por regra" });

            // Validate marketplace values
            foreach (var mkt in req.Marketplaces.Split(',', StringSplitOptions.RemoveEmptyEntries | StringSplitOptions.TrimEntries))
            {
                if (!Marketplaces.IsValid(mkt))
                    return Results.BadRequest(new { error = $"marketplace inválido: {mkt}" });
            }

            var rule = new CouponAlertRule
            {
                OwnerUid = tenant.OwnerUid,
                DiscountType = req.DiscountType ?? "percentage",
                MinDiscount = req.MinDiscount,
                Marketplaces = req.Marketplaces,
                Categories = categories,
                Channel = req.Channel ?? "telegram"
            };

            db.CouponAlertRules.Add(rule);
            await db.SaveChangesAsync(ct);

            return Results.Created($"/api/v2/cupons/regras/{rule.Id}", rule);
        });

        rules.MapPut("/{id:guid}", async (Guid id, AppDbContext db, UpdateCouponRuleRequest req, CancellationToken ct) =>
        {
            var rule = await db.CouponAlertRules.FindAsync([id], ct);
            if (rule is null)
                return Results.NotFound(new { error = "regra não encontrada" });

            if (req.MinDiscount is > 0)
                rule.MinDiscount = req.MinDiscount.Value;
            if (!string.IsNullOrWhiteSpace(req.DiscountType))
                rule.DiscountType = req.DiscountType;
            if (!string.IsNullOrWhiteSpace(req.Marketplaces))
                rule.Marketplaces = req.Marketplaces;
            if (req.Categories is not null)
                rule.Categories = req.Categories;
            if (!string.IsNullOrWhiteSpace(req.Channel))
                rule.Channel = req.Channel;

            rule.UpdatedAt = DateTime.UtcNow;

            // Reset dedup state when rule is edited (R9-AC5)
            var history = await db.CouponAlertHistories
                .Where(h => h.AlertRuleId == id)
                .ToListAsync(ct);
            db.CouponAlertHistories.RemoveRange(history);

            await db.SaveChangesAsync(ct);
            return Results.Ok(rule);
        });

        rules.MapDelete("/{id:guid}", async (Guid id, AppDbContext db, CancellationToken ct) =>
        {
            var rule = await db.CouponAlertRules.FindAsync([id], ct);
            if (rule is null)
                return Results.NotFound(new { error = "regra não encontrada" });

            db.CouponAlertRules.Remove(rule);
            await db.SaveChangesAsync(ct);
            return Results.Ok(new { status = "deleted" });
        });

        rules.MapPatch("/{id:guid}/toggle", async (Guid id, AppDbContext db, CancellationToken ct) =>
        {
            var rule = await db.CouponAlertRules.FindAsync([id], ct);
            if (rule is null)
                return Results.NotFound(new { error = "regra não encontrada" });

            rule.IsActive = !rule.IsActive;
            rule.UpdatedAt = DateTime.UtcNow;
            await db.SaveChangesAsync(ct);

            return Results.Ok(new { id = rule.Id, is_active = rule.IsActive });
        });

        return group;
    }
}

public sealed record CreateCouponRuleRequest
{
    public string? DiscountType { get; init; }
    public double MinDiscount { get; init; }
    public required string Marketplaces { get; init; }
    public string? Categories { get; init; }
    public string? Channel { get; init; }
}

public sealed record UpdateCouponRuleRequest
{
    public string? DiscountType { get; init; }
    public double? MinDiscount { get; init; }
    public string? Marketplaces { get; init; }
    public string? Categories { get; init; }
    public string? Channel { get; init; }
}
