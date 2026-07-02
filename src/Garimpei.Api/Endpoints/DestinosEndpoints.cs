using Garimpei.Domain.Entities;
using Garimpei.Infrastructure.Persistence;
using Microsoft.EntityFrameworkCore;

/// <summary>
/// Destinos endpoints — CRUD de canais de publicação (Telegram, WhatsApp, etc.).
/// Compatibilidade: /api/destinos (mesmo formato do frontend).
/// </summary>
public static partial class EndpointExtensions
{
    public static WebApplication MapDestinosEndpoints(this WebApplication app)
    {
        var destinos = app.MapGroup("/api/destinos")
            .RequireAuthorization()
            .WithTags("Destinos");

        destinos.MapGet("/", async (AppDbContext db, CancellationToken ct) =>
        {
            var lista = await db.Destinos
                .Where(d => d.Ativo)
                .OrderBy(d => d.Nome)
                .ToListAsync(ct);

            return Results.Ok(new
            {
                destinos = lista.Select(d => new
                {
                    id = d.Id,
                    nome = d.Nome,
                    tipo = d.Tipo,
                    config = d.Config,
                    ativo = d.Ativo
                }),
                total = lista.Count
            });
        });

        destinos.MapPost("/", async (AppDbContext db, SalvarDestinoRequest req, CancellationToken ct) =>
        {
            Destino destino;
            if (req.Id is not null)
            {
                // Update
                destino = await db.Destinos.FindAsync([req.Id.Value], ct)
                    ?? throw new InvalidOperationException("Destino não encontrado");

                destino.UpdatedAt = DateTime.UtcNow;
                if (req.Config is not null) destino.Config = req.Config;
                if (req.Ativo is not null) destino.Ativo = req.Ativo.Value;
            }
            else
            {
                destino = new Destino
                {
                    Nome = req.Nome ?? "Sem nome",
                    Tipo = req.Tipo ?? "telegram",
                    Config = req.Config ?? ""
                };
                db.Destinos.Add(destino);
            }

            await db.SaveChangesAsync(ct);
            return Results.Ok(new { id = destino.Id, nome = destino.Nome, status = "salvo" });
        });

        destinos.MapDelete("/", async (AppDbContext db, string id, CancellationToken ct) =>
        {
            if (!Guid.TryParse(id, out var guid))
                return Results.BadRequest(new { error = "id inválido" });

            var destino = await db.Destinos.FindAsync([guid], ct);
            if (destino is null) return Results.NotFound();

            destino.Ativo = false;
            destino.UpdatedAt = DateTime.UtcNow;
            await db.SaveChangesAsync(ct);

            return Results.Ok(new { status = "removido", id });
        });

        return app;
    }
}

public sealed record SalvarDestinoRequest
{
    public Guid? Id { get; init; }
    public string? Nome { get; init; }
    public string? Tipo { get; init; }
    public string? Config { get; init; }
    public bool? Ativo { get; init; }
}
