using Garimpei.Domain.Services;
using Garimpei.Domain.ValueObjects;
using Garimpei.Infrastructure.Sources;
using Cache.V1;
using Collector.V1;
using Microsoft.EntityFrameworkCore;

/// <summary>
/// Curadoria endpoints — serves the 4 primary sources for the publish page.
/// Now routes reads through Cache Sidecar (L2) with circuit breaker fallback.
/// </summary>
public static partial class EndpointExtensions
{
    public static RouteGroupBuilder MapCuradoriaEndpoints(this RouteGroupBuilder group)
    {
        var curadoria = group.MapGroup("/curadoria").WithTags("Curadoria");

        // GET /api/v2/curadoria/ranking?keyword=...&limit=20&comissao_min=0.07&vendas_min=0&nota_min=0
        curadoria.MapGet("/ranking", async (
            CacheService.CacheServiceClient cacheClient,
            CacheCircuitBreaker circuitBreaker,
            CollectorService.CollectorServiceClient collector,
            HttpContext httpContext,
            string? keyword,
            int? limit,
            double? comissao_min,
            int? vendas_min,
            double? nota_min,
            CancellationToken ct) =>
        {
            if (string.IsNullOrWhiteSpace(keyword))
                return Results.BadRequest(new { error = "keyword é obrigatório" });

            var ownerUid = httpContext.User.FindFirst("user_id")?.Value ?? "";
            var cacheSource = "l2-bypass";
            IReadOnlyList<Product> products;
            int totalFound;

            if (!circuitBreaker.IsOpen && !string.IsNullOrEmpty(ownerUid))
            {
                try
                {
                    var cacheResp = await cacheClient.GetAsync(new Cache.V1.GetRequest
                    {
                        CollectionKeys = { keyword },
                        BuscaId = $"busca-keyword-{keyword.ToLowerInvariant().Trim()}",
                        Marketplace = Collector.V1.Marketplace.Shopee,
                        OwnerUid = ownerUid,
                    }, cancellationToken: ct);

                    circuitBreaker.RecordSuccess();
                    cacheSource = cacheResp.CacheHit ? "l2-hit" : "l2-miss";
                    products = cacheResp.Products;
                    totalFound = cacheResp.Products.Count;
                }
                catch (Grpc.Core.RpcException)
                {
                    circuitBreaker.RecordFailure();
                    cacheSource = "l2-bypass";
                    var fallback = await collector.FetchAsync(new FetchRequest
                    {
                        Keyword = keyword,
                        Limit = limit ?? 50
                    }, cancellationToken: ct);
                    products = fallback.Products;
                    totalFound = fallback.TotalFound;
                }
            }
            else
            {
                var response = await collector.FetchAsync(new FetchRequest
                {
                    Keyword = keyword,
                    Limit = limit ?? 50
                }, cancellationToken: ct);
                products = response.Products;
                totalFound = response.TotalFound;
            }

            var candidates = products.Select(ToCandida‌te).ToList();

            var filter = new EligibilityFilter
            {
                MinCommission = comissao_min ?? 0.07,
                MinSales = vendas_min ?? 0,
                MinRating = nota_min ?? 0
            };

            var ranked = ScoringService.Rank(candidates, filter, limit ?? 20);

            httpContext.Response.Headers["X-Cache-Source"] = cacheSource;

            return Results.Ok(new
            {
                fonte = "collector-grpc",
                total_bruto = totalFound,
                candidatos = ranked
            });
        });

        // GET /api/v2/curadoria/ranking/shop?shop_id=123&limit=20
        curadoria.MapGet("/ranking/shop", async (
            CacheService.CacheServiceClient cacheClient,
            CacheCircuitBreaker circuitBreaker,
            CollectorService.CollectorServiceClient collector,
            HttpContext httpContext,
            long shop_id,
            int? limit,
            double? comissao_min,
            int? vendas_min,
            double? nota_min,
            CancellationToken ct) =>
        {
            if (shop_id == 0)
                return Results.BadRequest(new { error = "shop_id é obrigatório" });

            var ownerUid = httpContext.User.FindFirst("user_id")?.Value ?? "";
            var cacheSource = "l2-bypass";
            IReadOnlyList<Product> products;
            int totalFound;

            if (!circuitBreaker.IsOpen && !string.IsNullOrEmpty(ownerUid))
            {
                try
                {
                    var cacheResp = await cacheClient.GetAsync(new Cache.V1.GetRequest
                    {
                        CollectionKeys = { shop_id.ToString() },
                        BuscaId = $"busca-shop-{shop_id}",
                        Marketplace = Collector.V1.Marketplace.Shopee,
                        OwnerUid = ownerUid,
                    }, cancellationToken: ct);

                    circuitBreaker.RecordSuccess();
                    cacheSource = cacheResp.CacheHit ? "l2-hit" : "l2-miss";
                    products = cacheResp.Products;
                    totalFound = cacheResp.Products.Count;
                }
                catch (Grpc.Core.RpcException)
                {
                    circuitBreaker.RecordFailure();
                    cacheSource = "l2-bypass";
                    var fallback = await collector.FetchShopAsync(new FetchShopRequest
                    {
                        ShopId = shop_id,
                        Limit = limit ?? 50
                    }, cancellationToken: ct);
                    products = fallback.Products;
                    totalFound = fallback.TotalFound;
                }
            }
            else
            {
                var response = await collector.FetchShopAsync(new FetchShopRequest
                {
                    ShopId = shop_id,
                    Limit = limit ?? 50
                }, cancellationToken: ct);
                products = response.Products;
                totalFound = response.TotalFound;
            }

            var candidates = products.Select(ToCandida‌te).ToList();

            var filter = new EligibilityFilter
            {
                MinCommission = comissao_min ?? 0.07,
                MinSales = vendas_min ?? 0,
                MinRating = nota_min ?? 0
            };

            var ranked = ScoringService.Rank(candidates, filter, limit ?? 20);

            httpContext.Response.Headers["X-Cache-Source"] = cacheSource;

            return Results.Ok(new
            {
                fonte = "collector-grpc-shop",
                total_bruto = totalFound,
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
