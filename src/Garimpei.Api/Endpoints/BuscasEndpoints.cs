using Garimpei.Domain;
using Garimpei.Domain.Entities;
using Garimpei.Infrastructure.Persistence;
using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Logging;

/// <summary>
/// Buscas endpoints — /api/buscas (formato frontend).
/// Sincronização de perfis de busca para o scheduler.
/// Identidade por UUID (BuscaContract). Zero dependência de campo Keyword legado.
/// </summary>
public static partial class EndpointExtensions
{
    public static WebApplication MapBuscasEndpoints(this WebApplication app)
    {
        app.MapGet("/api/buscas", async (AppDbContext db, CancellationToken ct) =>
        {
            var buscas = await db.Buscas
                .Where(b => b.Active)
                .OrderByDescending(b => b.CreatedAt)
                .ToListAsync(ct);

            return Results.Ok(new
            {
                buscas = buscas.Select(b => new
                {
                    id = b.Id,
                    keywords = b.Keywords ?? Array.Empty<string>(),
                    shop_ids = b.ShopIds,
                    shop_names = b.ShopNames,
                    cron = b.CronExpression,
                    comissao_min = b.ComissaoMin,
                    vendas_min = b.VendasMin,
                    categorias = b.Categorias,
                    fontes = b.Fontes,
                    marketplaces = b.Marketplaces,
                    ativo = b.Active,
                    criado_em = b.CreatedAt,
                    sort_by = b.SortBy,
                    limit = b.Limit
                }),
                total = buscas.Count
            });
        }).RequireAuthorization().WithTags("Buscas");

        app.MapPost("/api/buscas", async (
            AppDbContext db,
            Scheduler.V1.SchedulerService.SchedulerServiceClient schedulerClient,
            ILogger<AppDbContext> logger,
            HttpContext context,
            SyncBuscaRequest req,
            CancellationToken ct) =>
        {
            // Se remover=true no query string, desativa por UUID
            var remover = context.Request.Query.ContainsKey("remover");

            if (remover)
            {
                if (req.Id is null || !Guid.TryParse(req.Id, out var buscaGuid))
                    return Results.BadRequest(new { error = "id (UUID) é obrigatório para remoção" });

                var existente = await db.Buscas.FindAsync([buscaGuid], ct);
                if (existente is not null)
                {
                    existente.Active = false;
                    existente.UpdatedAt = DateTime.UtcNow;
                    await db.SaveChangesAsync(ct);
                    await SchedulerJobs.PauseAsync(schedulerClient, existente, logger, ct);
                }

                return Results.Ok(new { status = "removida" });
            }

            // Salvar (upsert por UUID)
            var keywordsArray = (req.Keywords ?? Array.Empty<string>())
                .Where(k => !string.IsNullOrWhiteSpace(k))
                .ToArray();

            var marketplaces = req.Marketplaces is { Length: > 0 }
                ? req.Marketplaces
                : [Garimpei.Domain.Marketplaces.Shopee];

            // cron vazio = busca manual (sem agendamento)
            var cron = string.IsNullOrWhiteSpace(req.Cron) ? null : req.Cron;

            // Busca existente por ID
            Busca? busca = null;
            if (req.Id is not null && Guid.TryParse(req.Id, out var reqGuid))
            {
                busca = await db.Buscas.FindAsync([reqGuid], ct);
            }

            if (busca is null)
            {
                // Validação: precisa ter ao menos keywords, shop_ids, ou categorias
                if (keywordsArray.Length == 0 && req.ShopIds is not { Length: > 0 } && req.Categorias is not { Length: > 0 })
                    return Results.BadRequest(new { error = "keywords, shop_ids, ou categorias é obrigatório" });

                busca = new Busca
                {
                    Keywords = keywordsArray,
                    OwnerUid = "",
                    SortBy = req.SortBy ?? "relevance",
                    Limit = req.Limit ?? 50,
                    CronExpression = cron,
                    ShopIds = req.ShopIds,
                    ShopNames = req.ShopNames,
                    ComissaoMin = req.ComissaoMin,
                    VendasMin = req.VendasMin,
                    Categorias = req.Categorias,
                    Fontes = req.Fontes,
                    Marketplaces = marketplaces
                };
                db.Buscas.Add(busca);
            }
            else
            {
                busca.Active = true;
                busca.UpdatedAt = DateTime.UtcNow;
                busca.CronExpression = cron;
                busca.Keywords = keywordsArray;
                if (req.ShopIds is { Length: > 0 }) busca.ShopIds = req.ShopIds;
                if (req.ShopNames is not null) busca.ShopNames = req.ShopNames;
                if (req.ComissaoMin is not null) busca.ComissaoMin = req.ComissaoMin;
                if (req.VendasMin is not null) busca.VendasMin = req.VendasMin;
                if (req.Categorias is not null) busca.Categorias = req.Categorias;
                if (req.Fontes is not null) busca.Fontes = req.Fontes;
                if (req.Marketplaces is { Length: > 0 }) busca.Marketplaces = req.Marketplaces;
            }

            await db.SaveChangesAsync(ct);

            if (busca.CronExpression is not null)
                await SchedulerJobs.RegisterAsync(schedulerClient, busca, logger, ct);
            else
                await SchedulerJobs.PauseAsync(schedulerClient, busca, logger, ct);

            return Results.Ok(new { id = busca.Id, keywords = busca.Keywords, cron = busca.CronExpression, status = "salva" });
        }).RequireAuthorization().WithTags("Buscas");

        return app;
    }
}

public sealed record SyncBuscaRequest
{
    public string? Id { get; init; }
    public string[]? Keywords { get; init; }
    public long[]? ShopIds { get; init; }
    public Dictionary<string, string>? ShopNames { get; init; }
    public string? Cron { get; init; }
    public string? SortBy { get; init; }
    public int? Limit { get; init; }
    public decimal? ComissaoMin { get; init; }
    public int? VendasMin { get; init; }
    public string[]? Categorias { get; init; }
    public string[]? Fontes { get; init; }
    public string[]? Marketplaces { get; init; }
}
