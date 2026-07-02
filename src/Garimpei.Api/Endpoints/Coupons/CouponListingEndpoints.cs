using Garimpei.Infrastructure.Persistence;
using Microsoft.EntityFrameworkCore;

namespace Garimpei.Api.Endpoints.Coupons;

public static class CouponListingEndpoints
{
    public static RouteGroupBuilder MapCouponListingEndpoints(this RouteGroupBuilder group)
    {
        var cupons = group.MapGroup("/cupons")
            .RequireAuthorization()
            .WithTags("Cupons");

        // GET /api/v2/cupons — list active coupons (proxies to analyzer or shows from BQ cache)
        cupons.MapGet("/", async (
            HttpClient httpClient,
            IConfiguration config,
            string? marketplace,
            string? category,
            int? limit,
            CancellationToken ct) =>
        {
            var analyzerUrl = config["Analyzer:BaseUrl"] ?? "http://localhost:8060";

            // Build query params
            var queryParts = new List<string> { $"limit={limit ?? 50}" };
            if (!string.IsNullOrWhiteSpace(marketplace))
                queryParts.Add($"marketplace={marketplace}");
            if (!string.IsNullOrWhiteSpace(category))
                queryParts.Add($"category={category}");

            var url = $"{analyzerUrl}/cupons-ativos?{string.Join("&", queryParts)}";

            try
            {
                var response = await httpClient.GetFromJsonAsync<object>(url, ct);
                return Results.Ok(response);
            }
            catch
            {
                // Fallback: return empty when analyzer is unavailable
                return Results.Ok(new { cupons = Array.Empty<object>(), total = 0, fonte = "unavailable" });
            }
        });

        // GET /api/v2/cupons/historico — coupon history analytics
        cupons.MapGet("/historico", async (
            HttpClient httpClient,
            IConfiguration config,
            string? marketplace,
            int? dias,
            double? desconto_min,
            int? limit,
            CancellationToken ct) =>
        {
            var analyzerUrl = config["Analyzer:BaseUrl"] ?? "http://localhost:8060";

            var queryParts = new List<string>
            {
                $"dias={dias ?? 30}",
                $"limit={limit ?? 100}"
            };
            if (!string.IsNullOrWhiteSpace(marketplace))
                queryParts.Add($"marketplace={marketplace}");
            if (desconto_min.HasValue)
                queryParts.Add($"desconto_min={desconto_min.Value}");

            var url = $"{analyzerUrl}/cupons-historico?{string.Join("&", queryParts)}";

            try
            {
                var response = await httpClient.GetFromJsonAsync<object>(url, ct);
                return Results.Ok(response);
            }
            catch
            {
                return Results.Ok(new { cupons = Array.Empty<object>(), total = 0, fonte = "unavailable" });
            }
        });

        return group;
    }
}
