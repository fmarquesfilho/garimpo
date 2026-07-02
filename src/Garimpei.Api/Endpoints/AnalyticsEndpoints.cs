using Garimpei.Infrastructure.Persistence;
using Microsoft.EntityFrameworkCore;

/// <summary>
/// Analytics endpoints — conversões, estatísticas e coletas.
/// /api/conversoes, /api/conversoes/reais, /api/estatisticas, /api/coletas
/// </summary>
public static partial class EndpointExtensions
{
    public static WebApplication MapAnalyticsEndpoints(this WebApplication app)
    {
        // /api/conversoes — relatório de conversões (publicações por canal)
        app.MapGet("/api/conversoes", async (
            AppDbContext db,
            int? dias,
            CancellationToken ct) =>
        {
            var desde = DateTime.UtcNow.AddDays(-(dias ?? 30));

            var publicacoes = await db.Publicacoes
                .Where(p => p.CreatedAt >= desde && p.Status == "enviada")
                .ToListAsync(ct);

            var resumo = publicacoes
                .GroupBy(p => p.DestinoId ?? "sem-destino")
                .Select(g => new
                {
                    destino_id = g.Key,
                    publicacoes = g.Count(),
                    comissao_estimada = g.Sum(p => p.Comissao * (double)p.Preco),
                    ultimo_envio = g.Max(p => p.EnviadaEm)?.ToString("o") ?? ""
                })
                .ToList();

            return Results.Ok(new
            {
                dias_janela = dias ?? 30,
                total_publicacoes = publicacoes.Count,
                por_destino = resumo
            });
        }).RequireAuthorization().WithTags("Analytics");

        // /api/conversoes/reais — conversões reais da Shopee (proxy para scheduler/analyzer)
        app.MapGet("/api/conversoes/reais", async (
            HttpClient httpClient,
            IConfiguration config,
            int? dias,
            CancellationToken ct) =>
        {
            var analyzerUrl = config["Analyzer:BaseUrl"] ?? "http://localhost:8060";
            var url = $"{analyzerUrl}/conversoes?dias={dias ?? 30}";
            try
            {
                var response = await httpClient.GetFromJsonAsync<object>(url, ct);
                return Results.Ok(response);
            }
            catch
            {
                return Results.Ok(new
                {
                    fonte = "shopee-api",
                    status = "indisponível",
                    conversoes = Array.Empty<object>()
                });
            }
        }).RequireAuthorization().WithTags("Analytics");

        // /api/estatisticas — dashboard de snapshots por categoria
        app.MapGet("/api/estatisticas", async (
            HttpClient httpClient,
            IConfiguration config,
            int? dias,
            CancellationToken ct) =>
        {
            var analyzerUrl = config["Analyzer:BaseUrl"] ?? "http://localhost:8060";
            var url = $"{analyzerUrl}/estatisticas?dias={dias ?? 30}";
            try
            {
                var response = await httpClient.GetFromJsonAsync<object>(url, ct);
                return Results.Ok(response);
            }
            catch
            {
                return Results.Ok(new
                {
                    fonte = "bigquery",
                    dias_janela = dias ?? 30,
                    total_amostras = 0,
                    por_categoria = Array.Empty<object>()
                });
            }
        }).RequireAuthorization().WithTags("Analytics");

        // /api/coletas — histórico de coletas executadas
        app.MapGet("/api/coletas", async (
            HttpClient httpClient,
            IConfiguration config,
            int? dias,
            CancellationToken ct) =>
        {
            var analyzerUrl = config["Analyzer:BaseUrl"] ?? "http://localhost:8060";
            var url = $"{analyzerUrl}/coletas?dias={dias ?? 30}";
            try
            {
                var response = await httpClient.GetFromJsonAsync<object>(url, ct);
                return Results.Ok(response);
            }
            catch
            {
                return Results.Ok(new
                {
                    coletas = Array.Empty<object>(),
                    total = 0
                });
            }
        }).RequireAuthorization().WithTags("Analytics");

        return app;
    }
}
