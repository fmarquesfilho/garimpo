using Collector.V1;
using Garimpei.Domain;
using Garimpei.Domain.Services;
using Garimpei.Domain.ValueObjects;
using Garimpei.Infrastructure.Sources;

/// <summary>
/// Core endpoints — /api/health, /api/admin/me, /api/candidatos, /api/categorias.
/// </summary>
public static partial class EndpointExtensions
{
    public static WebApplication MapCoreEndpoints(this WebApplication app)
    {
        // /api/health — public, no auth
        app.MapGet("/api/health", () => Results.Ok(new
        {
            status = "ok",
            fonte = "shopee",
            store = "postgresql",
            backend = "csharp-v2",
            quality = new
            {
                codacy = "https://app.codacy.com",
                lint_go = "golangci-lint (0 issues)",
                lint_python = "ruff (0 issues)",
                lint_csharp = "NetArchTest (13 rules)",
                tests_csharp = 51,
                pre_push_checks = 9
            }
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

            var codacyUrl = config["Codacy:DashboardUrl"] ?? "https://app.codacy.com";

            return Results.Ok(new
            {
                admin = isAdmin,
                email,
                tools = new
                {
                    codacy_dashboard = codacyUrl,
                    github_actions = "https://github.com/fmarquesfilho/garimpo/actions",
                    pull_requests = "https://github.com/fmarquesfilho/garimpo/pulls"
                }
            });
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

        // /api/categorias — lista de categorias por marketplace (público)
        app.MapGet("/api/categorias", () =>
        {
            var shopee = new[]
            {
                new { id = 100001, nome = "Alimentos", slug = "alimentos" },
                new { id = 100009, nome = "Celulares", slug = "celulares" },
                new { id = 100011, nome = "Roupas Femininas", slug = "roupas-femininas" },
                new { id = 100012, nome = "Calçados", slug = "calcados" },
                new { id = 100013, nome = "Acessórios Celular", slug = "acessorios-celular" },
                new { id = 100017, nome = "Roupas Masculinas", slug = "roupas-masculinas" },
                new { id = 100535, nome = "Celulares & Tablets", slug = "celulares-tablets" },
                new { id = 100630, nome = "Beleza", slug = "beleza" },
                new { id = 100631, nome = "Saúde & Bem-estar", slug = "saude-bem-estar" },
                new { id = 100632, nome = "Brinquedos & Bebês", slug = "brinquedos-bebes" },
                new { id = 100633, nome = "Acessórios & Bolsas", slug = "acessorios-bolsas" },
                new { id = 100636, nome = "Casa & Decoração", slug = "casa-decoracao" },
                new { id = 100637, nome = "Moda", slug = "moda" },
                new { id = 100640, nome = "Perfumaria", slug = "perfumaria" },
                new { id = 100643, nome = "Papelaria & Livros", slug = "papelaria-livros" },
                new { id = 100644, nome = "Áudio & Eletrônicos", slug = "audio-eletronicos" },
                new { id = 100658, nome = "Manicure & Pedicure", slug = "manicure-pedicure" },
                new { id = 100659, nome = "Cuidados com o Cabelo", slug = "cuidados-cabelo" },
                new { id = 100663, nome = "Maquiagem", slug = "maquiagem" },
                new { id = 100664, nome = "Cuidados com a Pele", slug = "cuidados-pele" }
            };

            return Results.Ok(new
            {
                marketplaces = new[]
                {
                    new { marketplace = "shopee", categorias = shopee }
                }
            });
        });

        return app;
    }
}
