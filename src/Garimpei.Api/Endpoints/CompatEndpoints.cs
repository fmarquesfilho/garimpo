using Collector.V1;
using Garimpei.Domain.Services;

/// <summary>
/// Compatibility endpoints — serve /api/* routes during migration.
/// These mirror the Go legacy API shape so the frontend works without changes.
/// Will be removed when frontend migrates to /api/v2/*.
/// </summary>
public static partial class EndpointExtensions
{
    public static WebApplication MapCompatEndpoints(this WebApplication app)
    {
        // /api/health — public, no auth
        app.MapGet("/api/health", () => Results.Ok(new
        {
            status = "ok",
            fonte = "shopee",
            store = "postgresql",
            backend = "csharp-v2"
        }));

        // /api/candidatos — public (same as Go legacy)
        app.MapGet("/api/candidatos", async (
            CollectorService.CollectorServiceClient collector,
            string? keyword,
            int? top,
            double? comissao_min,
            int? vendas_min,
            double? nota_min,
            CancellationToken ct) =>
        {
            keyword ??= "";
            if (string.IsNullOrWhiteSpace(keyword))
                return Results.Ok(new { estrategia = "nicho", candidatos = Array.Empty<object>(), total_bruto = 0 });

            var response = await collector.FetchAsync(new FetchRequest
            {
                Keyword = keyword,
                Limit = Math.Min(top ?? 50, 100)
            }, cancellationToken: ct);

            var candidates = response.Products.Select(p => new ProductCandidate
            {
                Id = p.ItemId.ToString(),
                Name = p.Name,
                ShopName = p.ShopName,
                ShopId = p.ShopId.ToString(),
                Price = (decimal)p.Price,
                OriginalPrice = (decimal)p.OriginalPrice,
                DiscountPercent = p.DiscountPercent,
                Commission = 0,
                Sales = p.Sold,
                Rating = p.Rating,
                Link = p.ProductUrl,
                ImageUrl = p.ImageUrl
            }).ToList();

            var filter = new EligibilityFilter
            {
                MinCommission = comissao_min ?? 0.07,
                MinSales = vendas_min ?? 0,
                MinRating = nota_min ?? 0
            };

            var ranked = ScoringService.Rank(candidates, filter, top ?? 20);

            return Results.Ok(new
            {
                estrategia = "nicho",
                candidatos = ranked,
                total_bruto = response.TotalFound
            });
        });

        return app;
    }
}
