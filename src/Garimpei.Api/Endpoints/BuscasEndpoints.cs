using Garimpei.Domain.Entities;
using Garimpei.Infrastructure.Persistence;
using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Logging;

/// <summary>
/// Buscas compat endpoints — /api/buscas (formato frontend).
/// Sincronização de perfis de busca para o scheduler.
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
                buscas = buscas.Select(b =>
                {
                    var hasShop = b.ShopIds is { Length: > 0 };
                    // Loja: keywords são o filtro (b.Keywords) e o nome é o shop name (b.Keyword).
                    // Keyword-only: keywords vêm do próprio b.Keyword (formato separado por vírgula).
                    var keywords = hasShop
                        ? (b.Keywords ?? Array.Empty<string>())
                        : b.Keyword.Split(',', StringSplitOptions.RemoveEmptyEntries | StringSplitOptions.TrimEntries);
                    return new
                    {
                        id = b.Id,
                        keywords,
                        shop_ids = b.ShopIds,
                        nome = hasShop ? b.Keyword : null,
                        cron = b.CronExpression,
                        ativo = b.Active,
                        criado_em = b.CreatedAt,
                        sort_by = b.SortBy,
                        limit = b.Limit
                    };
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
            // Se remover=true no query string, desativa
            var remover = context.Request.Query.ContainsKey("remover");

            if (remover)
            {
                // Busca por keyword para desativar
                var keyword = req.Keywords?.FirstOrDefault() ?? "";
                var existente = await db.Buscas
                    .FirstOrDefaultAsync(b => b.Keyword == keyword, ct);

                if (existente is not null)
                {
                    existente.Active = false;
                    existente.UpdatedAt = DateTime.UtcNow;
                    await db.SaveChangesAsync(ct);
                    // Pausa o job periódico no Scheduler (se houver)
                    await SchedulerJobs.PauseAsync(schedulerClient, existente, logger, ct);
                }

                return Results.Ok(new { status = "removida" });
            }

            // Salvar (upsert por keyword)
            var keywords = req.Keywords ?? (req.Keyword is not null ? [req.Keyword] : []);
            var kw = string.Join(",", keywords.Where(k => !string.IsNullOrWhiteSpace(k)));
            if (string.IsNullOrWhiteSpace(kw))
                return Results.BadRequest(new { error = "keyword é obrigatório" });

            // cron vazio = busca manual (sem agendamento). Só registra job se houver cron.
            var cron = string.IsNullOrWhiteSpace(req.Cron) ? null : req.Cron;

            // Busca por match exato ou pela primeira keyword
            var primeiraKw = keywords.FirstOrDefault() ?? kw;
            var busca = await db.Buscas
                .FirstOrDefaultAsync(b => b.Keyword == kw || b.Keyword == primeiraKw, ct);

            if (busca is null)
            {
                busca = new Busca
                {
                    Keyword = kw,
                    OwnerUid = "",
                    SortBy = req.SortBy ?? "relevance",
                    Limit = req.Limit ?? 50,
                    CronExpression = cron
                };
                db.Buscas.Add(busca);
            }
            else
            {
                busca.Active = true;
                busca.UpdatedAt = DateTime.UtcNow;
                busca.CronExpression = cron;
                // Atualiza keywords se mudaram
                if (busca.Keyword != kw)
                {
                    busca.Keyword = kw;
                }
            }

            await db.SaveChangesAsync(ct);

            // Todo agendamento passa pelo Scheduler (ADR-0023). Busca por palavra-chave
            // só vira job periódico quando o usuário define um cron; sem cron é manual.
            if (busca.CronExpression is not null)
                await SchedulerJobs.RegisterAsync(schedulerClient, busca, logger, ct);
            else
                await SchedulerJobs.PauseAsync(schedulerClient, busca, logger, ct);

            return Results.Ok(new { id = busca.Id, keywords, cron = busca.CronExpression, status = "salva" });
        }).RequireAuthorization().WithTags("Buscas");

        return app;
    }
}

public sealed record SyncBuscaRequest
{
    public string? Id { get; init; }
    public string? Keyword { get; init; }
    public string[]? Keywords { get; init; }
    public string? Cron { get; init; }
    public string? SortBy { get; init; }
    public int? Limit { get; init; }
}
