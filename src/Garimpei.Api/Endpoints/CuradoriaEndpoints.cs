using Garimpei.Domain.Services;
using Garimpei.Domain.ValueObjects;
using Garimpei.Infrastructure.Sources;
using Collector.V1;
using Microsoft.EntityFrameworkCore;

/// <summary>
/// Curadoria endpoints — serves the 4 primary sources for the publish page.
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
        curadoria.MapGet("/quedas", async (
            HttpClient httpClient,
            IConfiguration config,
            int? dias,
            double? threshold,
            int? limit,
            CancellationToken ct) =>
        {
            var analyzerUrl = config["Analyzer:BaseUrl"] ?? "http://localhost:8060";
            var url = $"{analyzerUrl}/quedas?dias={dias ?? 7}&threshold={threshold ?? 0.15}&limit={limit ?? 50}";
            var response = await httpClient.GetFromJsonAsync<object>(url, ct);
            return Results.Ok(response);
        });

        // GET /api/v2/curadoria/novos — recently detected products
        // DEPRECATED: Use /api/lojas/novidades instead (same Analyzer endpoint).
        // Kept for backward compat — both use busca_id UUID exact match.
        curadoria.MapGet("/novos", async (
            HttpClient httpClient,
            IConfiguration config,
            string? busca_id,
            int? dias,
            CancellationToken ct) =>
        {
            if (string.IsNullOrWhiteSpace(busca_id))
                return Results.BadRequest(new { error = "busca_id é obrigatório" });

            var analyzerUrl = config["Analyzer:BaseUrl"] ?? "http://localhost:8060";
            var url = $"{analyzerUrl}/novidades?busca_id={busca_id}&dias={dias ?? 7}";
            var response = await httpClient.GetFromJsonAsync<object>(url, ct);
            return Results.Ok(response);
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

    private static ProductCandidate ToCandida‌te(Product p) => ProductMappings.ToCandidate(p);
}
