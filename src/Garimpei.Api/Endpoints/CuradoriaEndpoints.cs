using Garimpei.Domain.Services;
using Garimpei.Domain.ValueObjects;
using Collector.V1;
using Microsoft.EntityFrameworkCore;

/// <summary>
/// Curadoria endpoints — serves the 4 primary sources for the publish page:
/// - Busca (curadoria): fetch via gRPC collector → rank by scoring
/// - Quedas: products with price drops (from snapshots)
/// - Novos: recently detected products (from snapshots)
/// - Favoritos: user-saved products (from PostgreSQL)
/// </summary>
public static partial class EndpointExtensions
{
    public static RouteGroupBuilder MapCuradoriaEndpoints(this RouteGroupBuilder group)
    {
        var curadoria = group.MapGroup("/curadoria").WithTags("Curadoria");

        // GET /api/v2/curadoria/ranking?keyword=...&limit=20&comissao_min=0.07&vendas_min=0&nota_min=0
        curadoria.MapGet("/ranking", async (
            CollectorService.CollectorServiceClient collector,
            string? keyword,
            int? limit,
            double? comissao_min,
            int? vendas_min,
            double? nota_min,
            CancellationToken ct) =>
        {
            if (string.IsNullOrWhiteSpace(keyword))
                return Results.BadRequest(new { error = "keyword é obrigatório" });

            var response = await collector.FetchAsync(new FetchRequest
            {
                Keyword = keyword,
                Limit = limit ?? 50
            }, cancellationToken: ct);

            var candidates = response.Products.Select(ToCandida‌te).ToList();

            var filter = new EligibilityFilter
            {
                MinCommission = comissao_min ?? 0.07,
                MinSales = vendas_min ?? 0,
                MinRating = nota_min ?? 0
            };

            var ranked = ScoringService.Rank(candidates, filter, limit ?? 20);

            return Results.Ok(new
            {
                fonte = "collector-grpc",
                total_bruto = response.TotalFound,
                candidatos = ranked
            });
        });

        // GET /api/v2/curadoria/ranking/shop?shop_id=123&limit=20
        curadoria.MapGet("/ranking/shop", async (
            CollectorService.CollectorServiceClient collector,
            long shop_id,
            int? limit,
            double? comissao_min,
            int? vendas_min,
            double? nota_min,
            CancellationToken ct) =>
        {
            if (shop_id == 0)
                return Results.BadRequest(new { error = "shop_id é obrigatório" });

            var response = await collector.FetchShopAsync(new FetchShopRequest
            {
                ShopId = shop_id,
                Limit = limit ?? 50
            }, cancellationToken: ct);

            var candidates = response.Products.Select(ToCandida‌te).ToList();

            var filter = new EligibilityFilter
            {
                MinCommission = comissao_min ?? 0.07,
                MinSales = vendas_min ?? 0,
                MinRating = nota_min ?? 0
            };

            var ranked = ScoringService.Rank(candidates, filter, limit ?? 20);

            return Results.Ok(new
            {
                fonte = "collector-grpc-shop",
                total_bruto = response.TotalFound,
                candidatos = ranked
            });
        });

        // GET /api/v2/curadoria/quedas — products with price drops
        // TODO: wire to snapshots repository when migrated to PG
        curadoria.MapGet("/quedas", () =>
        {
            // Will query PostgreSQL snapshots for price variations < -threshold
            return Results.Ok(new { fonte = "quedas", candidatos = Array.Empty<object>(), info = "migração pendente — dados em BigQuery via Go legado" });
        });

        // GET /api/v2/curadoria/novos — recently detected products
        // TODO: wire to snapshots repository when migrated to PG
        curadoria.MapGet("/novos", () =>
        {
            // Will query PostgreSQL snapshots for products with aparicoes == 1
            return Results.Ok(new { fonte = "novos", candidatos = Array.Empty<object>(), info = "migração pendente — dados em BigQuery via Go legado" });
        });

        // GET /api/v2/curadoria/favoritos — user's saved products
        curadoria.MapGet("/favoritos", async (
            Garimpei.Infrastructure.Persistence.AppDbContext db,
            CancellationToken ct) =>
        {
            var favoritos = await db.Products
                .Where(p => p.OwnerUid != null)
                .OrderByDescending(p => p.UpdatedAt)
                .Take(50)
                .ToListAsync(ct);

            return Results.Ok(new { fonte = "favoritos", candidatos = favoritos });
        });

        return group;
    }

    private static ProductCandidate ToCandida‌te(Product p) => new()
    {
        Id = p.ItemId.ToString(),
        Name = p.Name,
        ShopName = p.ShopName,
        ShopId = p.ShopId.ToString(),
        Price = (decimal)p.Price,
        OriginalPrice = (decimal)p.OriginalPrice,
        DiscountPercent = p.DiscountPercent,
        Commission = 0, // Shopee doesn't expose commission in this proto; will be enriched
        Sales = p.Sold,
        Rating = p.Rating,
        Link = p.ProductUrl,
        ImageUrl = p.ImageUrl
    };
}
