using Garimpei.Domain.Entities;
using Garimpei.Infrastructure.Persistence;
using Microsoft.EntityFrameworkCore;

/// <summary>
/// Lojas/Buscas endpoints — CRUD for monitored shops and searches.
/// </summary>
public static partial class EndpointExtensions
{
    public static RouteGroupBuilder MapLojasEndpoints(this RouteGroupBuilder group)
    {
        var lojas = group.MapGroup("/buscas").WithTags("Buscas & Lojas");

        lojas.MapGet("/", async (AppDbContext db, CancellationToken ct) =>
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
                    keywords = b.Keyword.Split(',', StringSplitOptions.RemoveEmptyEntries | StringSplitOptions.TrimEntries),
                    shop_ids = b.ShopIds,
                    sort_by = b.SortBy,
                    limit = b.Limit,
                    active = b.Active,
                    created_at = b.CreatedAt
                }),
                total = buscas.Count
            });
        });

        lojas.MapPost("/", async (AppDbContext db, CreateBuscaRequest req, CancellationToken ct) =>
        {
            var busca = new Busca
            {
                Keyword = req.Keyword,
                OwnerUid = "", // auto-set by TenantContext on SaveChanges
                SortBy = req.SortBy ?? "relevance",
                Limit = req.Limit ?? 50,
            };

            db.Buscas.Add(busca);
            await db.SaveChangesAsync(ct);

            return Results.Created($"/api/v2/buscas/{busca.Id}", new { id = busca.Id, keyword = busca.Keyword });
        });

        lojas.MapDelete("/{id:guid}", async (Guid id, AppDbContext db, CancellationToken ct) =>
        {
            var busca = await db.Buscas.FindAsync([id], ct);
            if (busca is null) return Results.NotFound();

            busca.Active = false;
            busca.UpdatedAt = DateTime.UtcNow;
            await db.SaveChangesAsync(ct);

            return Results.Ok(new { status = "removed", id });
        });

        return group;
    }
}

public sealed record CreateBuscaRequest
{
    public required string Keyword { get; init; }
    public string? SortBy { get; init; }
    public int? Limit { get; init; }
}
