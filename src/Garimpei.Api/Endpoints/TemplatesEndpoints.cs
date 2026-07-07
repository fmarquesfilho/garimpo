using Garimpei.Domain.Entities;
using Garimpei.Infrastructure.Persistence;
using Microsoft.EntityFrameworkCore;

/// <summary>
/// Templates endpoints — CRUD de templates de mensagem + preview.
/// /api/templates (mesmo formato do frontend).
/// </summary>
public static partial class EndpointExtensions
{
    public static WebApplication MapTemplatesEndpoints(this WebApplication app)
    {
        var templates = app.MapGroup("/api/templates")
            .RequireAuthorization()
            .WithTags("Templates");

        templates.MapGet("/", async (AppDbContext db, CancellationToken ct) =>
        {
            var lista = await db.Templates
                .Where(t => t.Ativo)
                .OrderBy(t => t.Nome)
                .ToListAsync(ct);

            return Results.Ok(new
            {
                templates = lista.Select(t => new
                {
                    id = t.Id,
                    nome = t.Nome,
                    corpo = t.Corpo,
                    com_foto = t.ComFoto,
                    ativo = t.Ativo,
                    criado_em = t.CreatedAt
                }),
                total = lista.Count
            });
        });

        templates.MapPost("/", async (AppDbContext db, SalvarTemplateRequest req, CancellationToken ct) =>
        {
            Template template;
            if (req.Id is not null)
            {
                template = await db.Templates.FindAsync([req.Id.Value], ct)
                    ?? throw new InvalidOperationException("Template não encontrado");

                template.UpdatedAt = DateTime.UtcNow;
                if (req.ComFoto is not null) template.ComFoto = req.ComFoto.Value;
                if (req.Ativo is not null) template.Ativo = req.Ativo.Value;
            }
            else
            {
                template = new Template
                {
                    Nome = req.Nome ?? "Sem nome",
                    Corpo = req.Corpo ?? "",
                    ComFoto = req.ComFoto ?? false
                };
                db.Templates.Add(template);
            }

            await db.SaveChangesAsync(ct);
            return Results.Ok(new { id = template.Id, nome = template.Nome, status = "salvo" });
        });

        templates.MapDelete("/", async (AppDbContext db, string id, CancellationToken ct) =>
        {
            if (!Guid.TryParse(id, out var guid))
                return Results.BadRequest(new { error = "id inválido" });

            var template = await db.Templates.FindAsync([guid], ct);
            if (template is null) return Results.NotFound();

            template.Ativo = false;
            template.UpdatedAt = DateTime.UtcNow;
            await db.SaveChangesAsync(ct);

            return Results.Ok(new { status = "removido", id });
        });

        // Preview — renderiza template com dados de um produto
        app.MapPost("/api/templates/preview", async (AppDbContext db, TemplatePreviewRequest req, CancellationToken ct) =>
        {
            Template? template = null;
            if (req.TemplateId is not null)
            {
                template = await db.Templates.FindAsync([req.TemplateId.Value], ct);
            }

            var corpo = template?.Corpo ?? req.Corpo ?? "{{nome}} — {{preco}}\n{{link}}";
            var renderizado = corpo
                .Replace("{{nome}}", req.Nome?.Trim() ?? "")
                .Replace("{{preco}}", $"R$ {req.Preco:F2}")
                .Replace("{{categoria}}", req.Categoria ?? "")
                .Replace("{{estrategia}}", req.Estrategia ?? "")
                .Replace("{{link}}", req.Link ?? "");

            return Results.Ok(new { preview = renderizado, com_foto = template?.ComFoto ?? false });
        }).RequireAuthorization().WithTags("Templates");

        return app;
    }
}

public sealed record SalvarTemplateRequest
{
    public Guid? Id { get; init; }
    public string? Nome { get; init; }
    public string? Corpo { get; init; }
    public bool? ComFoto { get; init; }
    public bool? Ativo { get; init; }
}

public sealed record TemplatePreviewRequest
{
    public Guid? TemplateId { get; init; }
    public string? Corpo { get; init; }
    public string? Nome { get; init; }
    public decimal Preco { get; init; }
    public string? Categoria { get; init; }
    public string? Estrategia { get; init; }
    public string? Link { get; init; }
}
