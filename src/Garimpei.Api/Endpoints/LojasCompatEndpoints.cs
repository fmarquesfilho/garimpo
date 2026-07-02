using Garimpei.Domain.Entities;
using Garimpei.Infrastructure.Persistence;
using Microsoft.EntityFrameworkCore;

/// <summary>
/// Lojas compat endpoints — /api/lojas, /api/lojas/novidades, /api/lojas/evolucao.
/// Monitoramento de lojas: adicionar/remover + dados do analyzer Python.
/// </summary>
public static partial class EndpointExtensions
{
    public static WebApplication MapLojasCompatEndpoints(this WebApplication app)
    {
        // /api/lojas — listar/adicionar/remover lojas monitoradas (buscas com ShopIDs)
        app.MapGet("/api/lojas", async (AppDbContext db, CancellationToken ct) =>
        {
            var buscas = await db.Buscas
                .Where(b => b.Active)
                .OrderByDescending(b => b.CreatedAt)
                .ToListAsync(ct);

            return Results.Ok(new
            {
                lojas = buscas.Select(b => new
                {
                    id = b.Id,
                    keyword = b.Keyword,
                    ativo = b.Active,
                    criado_em = b.CreatedAt
                }),
                total = buscas.Count
            });
        }).RequireAuthorization().WithTags("Lojas");

        app.MapPost("/api/lojas", async (AppDbContext db, AdicionarLojaRequest req, CancellationToken ct) =>
        {
            // "input" pode ser URL da loja ou keyword
            var keyword = req.Input ?? "";

            var busca = new Busca
            {
                Keyword = keyword,
                OwnerUid = "",
                SortBy = "relevance",
                Limit = 50
            };

            db.Buscas.Add(busca);
            await db.SaveChangesAsync(ct);

            return Results.Ok(new { id = busca.Id, keyword, status = "adicionada" });
        }).RequireAuthorization().WithTags("Lojas");

        app.MapDelete("/api/lojas", async (AppDbContext db, string id, CancellationToken ct) =>
        {
            if (!Guid.TryParse(id, out var guid))
                return Results.BadRequest(new { error = "id inválido" });

            var busca = await db.Buscas.FindAsync([guid], ct);
            if (busca is null) return Results.NotFound();

            busca.Active = false;
            busca.UpdatedAt = DateTime.UtcNow;
            await db.SaveChangesAsync(ct);

            return Results.Ok(new { status = "removida", id });
        }).RequireAuthorization().WithTags("Lojas");

        // /api/lojas/novidades — novidades das lojas (produtos novos + variações preço)
        app.MapGet("/api/lojas/novidades", async (
            HttpClient httpClient,
            IConfiguration config,
            string? busca_id,
            int? dias,
            CancellationToken ct) =>
        {
            var analyzerUrl = config["Analyzer:BaseUrl"] ?? "http://localhost:8060";
            var url = $"{analyzerUrl}/novidades?busca_id={busca_id ?? ""}&dias={dias ?? 7}";
            try
            {
                var response = await httpClient.GetFromJsonAsync<object>(url, ct);
                return Results.Ok(response);
            }
            catch
            {
                // Analyzer offline — retorna vazio
                return Results.Ok(new
                {
                    busca_id = busca_id ?? "",
                    dias_janela = dias ?? 7,
                    produtos_novos = Array.Empty<object>(),
                    variacoes = Array.Empty<object>(),
                    total_atual = 0
                });
            }
        }).RequireAuthorization().WithTags("Lojas");

        // /api/lojas/evolucao — evolução de preço das lojas monitoradas
        app.MapGet("/api/lojas/evolucao", async (
            HttpClient httpClient,
            IConfiguration config,
            int? dias,
            CancellationToken ct) =>
        {
            var analyzerUrl = config["Analyzer:BaseUrl"] ?? "http://localhost:8060";
            var url = $"{analyzerUrl}/evolucao?dias={dias ?? 30}";
            try
            {
                var response = await httpClient.GetFromJsonAsync<object>(url, ct);
                return Results.Ok(response);
            }
            catch
            {
                return Results.Ok(new
                {
                    dias_janela = dias ?? 30,
                    lojas = Array.Empty<object>(),
                    resumo = new
                    {
                        total_lojas = 0,
                        total_produtos = 0,
                        preco_medio_global = 0.0,
                        variacao_media_global_pct = 0.0,
                        total_quedas = 0,
                        total_altas = 0
                    }
                });
            }
        }).RequireAuthorization().WithTags("Lojas");

        return app;
    }
}

public sealed record AdicionarLojaRequest
{
    public string? Input { get; init; }
    public string? Cron { get; init; }
    public string? OrigemPadrao { get; init; }
}
