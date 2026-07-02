using Collector.V1;
using Garimpei.Domain;
using Garimpei.Domain.Services;
using Garimpei.Domain.ValueObjects;
using Garimpei.Infrastructure.Sources;

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

        // /api/admin/me — verifica se o usuário logado é admin
        app.MapGet("/api/admin/me", (HttpContext context, IConfiguration config) =>
        {
            var email = context.User.FindFirst("email")?.Value
                ?? context.User.FindFirst(System.Security.Claims.ClaimTypes.Email)?.Value
                ?? "";

            var adminEmails = config["AdminEmails"] ?? "";
            var isAdmin = !string.IsNullOrEmpty(email)
                && adminEmails.Split(',', StringSplitOptions.RemoveEmptyEntries)
                    .Any(e => e.Trim().Equals(email, StringComparison.OrdinalIgnoreCase));

            return Results.Ok(new { admin = isAdmin, email });
        }).RequireAuthorization();

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

            var candidates = response.Products.Select(ProductMappings.ToCandidate).ToList();

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
                candidatos = ranked.Select(s => new
                {
                    id = s.Id,
                    nome = s.Name,
                    categoria = s.Category ?? "",
                    loja = s.ShopName ?? "",
                    loja_id = s.ShopId ?? "",
                    preco = s.Price,
                    preco_max = s.OriginalPrice,
                    desconto = s.DiscountPercent,
                    comissao = s.Commission,
                    vendas = s.Sales,
                    avaliacao = s.Rating,
                    link = s.Link ?? "",
                    link_produto = s.ProductLink ?? "",
                    imagem = s.ImageUrl ?? "",
                    score = s.Score,
                    componentes = new
                    {
                        comissao = s.Components.Commission,
                        valor_esperado = s.Components.ExpectedValue,
                        avaliacao = s.Components.Rating
                    },
                    suspeito = s.Suspicious,
                    oferta_expira = s.OfferExpiresAt ?? "",
                    marketplace = s.Marketplace
                }),
                total_bruto = response.TotalFound
            });
        });

        return app;
    }
}
