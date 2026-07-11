/// <summary>
/// Proxy endpoints for new Analyzer v2 routes.
/// All require authorization and proxy to the Python Analyzer service.
/// </summary>
public static partial class EndpointExtensions
{
    public static WebApplication MapAnalyzerProxyEndpoints(this WebApplication app)
    {
        // /api/coletas/saude — collection health status
        app.MapGet("/api/coletas/saude", async (
            HttpClient httpClient,
            IConfiguration config,
            CancellationToken ct) =>
        {
            var analyzerUrl = config["Analyzer:BaseUrl"] ?? "http://localhost:8060";
            try
            {
                var response = await httpClient.GetFromJsonAsync<object>(
                    $"{analyzerUrl}/coletas/saude", ct);
                return Results.Ok(response);
            }
            catch
            {
                return Results.Ok(new
                {
                    ultima_coleta = (string?)null,
                    minutos_desde_ultima = (int?)null,
                    status = "indisponivel",
                    coletas_24h = 0,
                    coletas_esperadas_24h = 0,
                    keywords_atrasadas = Array.Empty<string>()
                });
            }
        }).RequireAuthorization().WithTags("Analytics");

        // /api/oportunidades/agora — opportunities (drops + new + high-value)
        app.MapGet("/api/oportunidades/agora", async (
            HttpClient httpClient,
            IConfiguration config,
            int? dias,
            CancellationToken ct) =>
        {
            var analyzerUrl = config["Analyzer:BaseUrl"] ?? "http://localhost:8060";
            try
            {
                var response = await httpClient.GetFromJsonAsync<object>(
                    $"{analyzerUrl}/oportunidades/agora?dias={dias ?? 7}", ct);
                return Results.Ok(response);
            }
            catch
            {
                return Results.Ok(new
                {
                    dias = dias ?? 7,
                    quedas = Array.Empty<object>(),
                    novos = Array.Empty<object>(),
                    alto_valor = Array.Empty<object>(),
                    total_quedas = 0,
                    total_novos = 0,
                    total_alto_valor = 0,
                    filtro_publicacoes = false
                });
            }
        }).RequireAuthorization().WithTags("Analytics");

        // /api/conversoes/resumo — revenue summary
        app.MapGet("/api/conversoes/resumo", async (
            HttpClient httpClient,
            IConfiguration config,
            int? dias,
            CancellationToken ct) =>
        {
            var analyzerUrl = config["Analyzer:BaseUrl"] ?? "http://localhost:8060";
            try
            {
                var response = await httpClient.GetFromJsonAsync<object>(
                    $"{analyzerUrl}/conversoes/resumo?dias={dias ?? 30}", ct);
                return Results.Ok(response);
            }
            catch
            {
                return Results.Ok(new
                {
                    dias = dias ?? 30,
                    comissao_total = 0,
                    conversoes = 0,
                    produtos_distintos = 0,
                    por_canal = Array.Empty<object>(),
                    melhor_canal = (string?)null,
                    status = "indisponivel"
                });
            }
        }).RequireAuthorization().WithTags("Analytics");

        // /api/alertas/eficacia — alert efficacy metrics
        app.MapGet("/api/alertas/eficacia", async (
            HttpClient httpClient,
            IConfiguration config,
            int? dias,
            CancellationToken ct) =>
        {
            var analyzerUrl = config["Analyzer:BaseUrl"] ?? "http://localhost:8060";
            try
            {
                var response = await httpClient.GetFromJsonAsync<object>(
                    $"{analyzerUrl}/alertas/eficacia?dias={dias ?? 30}", ct);
                return Results.Ok(response);
            }
            catch
            {
                return Results.Ok(new
                {
                    dias = dias ?? 30,
                    quedas_detectadas = 0,
                    alertas_enviados = 0,
                    conversoes_atribuidas = 0,
                    taxa_deteccao = (double?)null,
                    taxa_conversao = (double?)null,
                    melhor_keyword = (string?)null,
                    conversoes_disponiveis = false
                });
            }
        }).RequireAuthorization().WithTags("Analytics");

        // /api/dashboard/changes — lightweight change detection for smart polling
        app.MapGet("/api/dashboard/changes", async (
            HttpClient httpClient,
            IConfiguration config,
            HttpContext context,
            CancellationToken ct) =>
        {
            var analyzerUrl = config["Analyzer:BaseUrl"] ?? "http://localhost:8060";
            try
            {
                var response = await httpClient.GetFromJsonAsync<object>(
                    $"{analyzerUrl}/dashboard/changes", ct);
                context.Response.Headers["Cache-Control"] = "no-store";
                return Results.Ok(response);
            }
            catch
            {
                context.Response.Headers["Cache-Control"] = "no-store";
                return Results.Ok(new
                {
                    saude_updated_at = (string?)null,
                    oportunidades_updated_at = (string?)null,
                    performance_updated_at = (string?)null
                });
            }
        }).RequireAuthorization().WithTags("Analytics");

        return app;
    }
}
