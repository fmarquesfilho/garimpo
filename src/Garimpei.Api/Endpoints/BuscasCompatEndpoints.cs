using Garimpei.Domain.Entities;
using Garimpei.Infrastructure.Persistence;
using Microsoft.EntityFrameworkCore;

/// <summary>
/// Buscas compat endpoints — /api/buscas (formato frontend).
/// Sincronização de perfis de busca para o scheduler.
/// </summary>
public static partial class EndpointExtensions
{
    public static WebApplication MapBuscasCompatEndpoints(this WebApplication app)
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
                    keywords = new[] { b.Keyword },
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
                }

                return Results.Ok(new { status = "removida" });
            }

            // Salvar (upsert por keyword)
            var kw = req.Keywords?.FirstOrDefault() ?? req.Keyword ?? "";
            if (string.IsNullOrWhiteSpace(kw))
                return Results.BadRequest(new { error = "keyword é obrigatório" });

            var busca = await db.Buscas
                .FirstOrDefaultAsync(b => b.Keyword == kw, ct);

            if (busca is null)
            {
                busca = new Busca
                {
                    Keyword = kw,
                    OwnerUid = "",
                    SortBy = req.SortBy ?? "relevance",
                    Limit = req.Limit ?? 50
                };
                db.Buscas.Add(busca);
            }
            else
            {
                busca.Active = true;
                busca.UpdatedAt = DateTime.UtcNow;
            }

            await db.SaveChangesAsync(ct);
            return Results.Ok(new { id = busca.Id, keyword = kw, status = "salva" });
        }).RequireAuthorization().WithTags("Buscas");

        return app;
    }
}

public sealed record SyncBuscaRequest
{
    public string? Id { get; init; }
    public string? Keyword { get; init; }
    public string[]? Keywords { get; init; }
    public string? SortBy { get; init; }
    public int? Limit { get; init; }
}
