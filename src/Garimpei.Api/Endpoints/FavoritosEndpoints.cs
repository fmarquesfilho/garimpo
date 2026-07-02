using Garimpei.Domain.Entities;
using Garimpei.Domain.Interfaces;
using Garimpei.Infrastructure.Persistence;
using Microsoft.EntityFrameworkCore;

/// <summary>
/// Favoritos endpoints — CRUD de produtos favoritos do usuário.
/// Compatibilidade: /api/favoritos (mesmo formato do frontend).
/// </summary>
public static partial class EndpointExtensions
{
    public static WebApplication MapFavoritosEndpoints(this WebApplication app)
    {
        var favoritos = app.MapGroup("/api/favoritos")
            .RequireAuthorization()
            .WithTags("Favoritos");

        favoritos.MapGet("/", async (AppDbContext db, CancellationToken ct) =>
        {
            var lista = await db.Favoritos
                .Where(f => f.Ativo)
                .OrderByDescending(f => f.CreatedAt)
                .ToListAsync(ct);

            return Results.Ok(new
            {
                favoritos = lista.Select(f => new
                {
                    produto_id = f.ProdutoId,
                    nome = f.Nome,
                    preco = f.Preco,
                    comissao = f.Comissao,
                    link = f.Link ?? "",
                    imagem = f.Imagem ?? "",
                    loja = f.Loja ?? "",
                    categoria = f.Categoria ?? "",
                    origem = f.Origem ?? "",
                    salvo_em = f.CreatedAt,
                    ativo = f.Ativo
                }),
                total = lista.Count
            });
        });

        favoritos.MapPost("/", async (AppDbContext db, ITenantContext tenant, SalvarFavoritoRequest req, CancellationToken ct) =>
        {
            // Upsert: se já existe, reativa
            var existente = await db.Favoritos
                .IgnoreQueryFilters()
                .FirstOrDefaultAsync(f => f.ProdutoId == req.ProdutoId
                    && f.OwnerUid == tenant.OwnerUid, ct);

            if (existente is not null)
            {
                existente.Ativo = true;
                existente.Preco = req.Preco;
                existente.Comissao = req.Comissao;
                existente.UpdatedAt = DateTime.UtcNow;
            }
            else
            {
                db.Favoritos.Add(new Favorito
                {
                    ProdutoId = req.ProdutoId,
                    Nome = req.Nome,
                    Preco = req.Preco,
                    Comissao = req.Comissao,
                    Link = req.Link,
                    Imagem = req.Imagem,
                    Loja = req.Loja,
                    Categoria = req.Categoria,
                    Origem = req.Origem
                });
            }

            await db.SaveChangesAsync(ct);
            return Results.Ok(new { status = "salvo", produto_id = req.ProdutoId });
        });

        favoritos.MapDelete("/", async (AppDbContext db, string produto_id, CancellationToken ct) =>
        {
            var favorito = await db.Favoritos
                .FirstOrDefaultAsync(f => f.ProdutoId == produto_id, ct);

            if (favorito is null) return Results.NotFound();

            favorito.Ativo = false;
            favorito.UpdatedAt = DateTime.UtcNow;
            await db.SaveChangesAsync(ct);

            return Results.Ok(new { status = "removido", produto_id });
        });

        return app;
    }
}

public sealed record SalvarFavoritoRequest
{
    public required string ProdutoId { get; init; }
    public required string Nome { get; init; }
    public decimal Preco { get; init; }
    public double Comissao { get; init; }
    public string? Link { get; init; }
    public string? Imagem { get; init; }
    public string? Loja { get; init; }
    public string? Categoria { get; init; }
    public string? Origem { get; init; }
}
