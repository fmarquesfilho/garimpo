using Garimpei.Domain.Entities;
using Garimpei.Infrastructure.Persistence;
using Microsoft.EntityFrameworkCore;

/// <summary>
/// Publicações endpoints — agendamento e histórico de publicações.
/// Compatibilidade: /api/publicacoes e /api/publicar (formato frontend).
/// </summary>
public static partial class EndpointExtensions
{
    public static WebApplication MapPublicacoesEndpoints(this WebApplication app)
    {
        // /api/publicacoes — lista e agenda publicações
        var publicacoes = app.MapGroup("/api/publicacoes")
            .RequireAuthorization()
            .WithTags("Publicações");

        publicacoes.MapGet("/", async (AppDbContext db, string? status, CancellationToken ct) =>
        {
            var query = db.Publicacoes.AsQueryable();

            if (!string.IsNullOrWhiteSpace(status))
                query = query.Where(p => p.Status == status);

            var lista = await query
                .OrderByDescending(p => p.CreatedAt)
                .Take(100)
                .ToListAsync(ct);

            return Results.Ok(new
            {
                publicacoes = lista.Select(p => new
                {
                    id = p.Id,
                    produto_id = p.ProdutoId,
                    nome = p.Nome,
                    categoria = p.Categoria ?? "",
                    preco = p.Preco,
                    comissao = p.Comissao,
                    link = p.Link ?? "",
                    imagem = p.Imagem ?? "",
                    estrategia = p.Estrategia ?? "",
                    destino_id = p.DestinoId ?? "",
                    template_id = p.TemplateId ?? "",
                    agendada_em = p.AgendadaEm?.ToString("o") ?? "",
                    status = p.Status,
                    detalhe = p.Detalhe ?? "",
                    criada_em = p.CreatedAt.ToString("o"),
                    enviada_em = p.EnviadaEm?.ToString("o") ?? ""
                }),
                total = lista.Count
            });
        });

        publicacoes.MapPost("/", async (AppDbContext db, AgendarPublicacaoRequest req, CancellationToken ct) =>
        {
            var pub = new Publicacao
            {
                ProdutoId = req.ProdutoId ?? req.Id ?? "",
                Nome = req.Nome ?? "Sem nome",
                Categoria = req.Categoria,
                Preco = req.Preco,
                Comissao = req.Comissao,
                Link = req.Link,
                Imagem = req.Imagem,
                Estrategia = req.Estrategia,
                DestinoId = req.DestinoId,
                TemplateId = req.TemplateId,
                AgendadaEm = req.AgendadaEm,
                Status = req.AgendadaEm.HasValue ? "agendada" : "pendente"
            };

            db.Publicacoes.Add(pub);
            await db.SaveChangesAsync(ct);

            return Results.Ok(new { id = pub.Id, status = pub.Status, criada_em = pub.CreatedAt.ToString("o") });
        });

        // /api/publicar — endpoint compat para publicar imediatamente (envia via gRPC publisher)
        app.MapPost("/api/publicar", async (
            AppDbContext db,
            Publisher.V1.PublisherService.PublisherServiceClient publisher,
            PublicarCompatRequest req,
            CancellationToken ct) =>
        {
            // Tentativa de envio via publisher gRPC
            try
            {
                var grpcRequest = new Publisher.V1.PublishRequest
                {
                    Channel = "telegram",
                    GroupId = req.DestinoId ?? "",
                    Content = new Publisher.V1.PublishContent
                    {
                        Title = req.Nome ?? "",
                        Description = req.Categoria ?? "",
                        ImageUrl = req.Imagem ?? "",
                        ProductUrl = req.Link ?? "",
                        Price = (double)req.Preco,
                        OriginalPrice = (double)req.Preco,
                        DiscountPercent = 0
                    }
                };

                var response = await publisher.PublishAsync(grpcRequest, cancellationToken: ct);

                // Registra a publicação
                var pub = new Publicacao
                {
                    ProdutoId = req.Id ?? "",
                    Nome = req.Nome ?? "",
                    Categoria = req.Categoria,
                    Preco = req.Preco,
                    Comissao = req.Comissao,
                    Link = req.Link,
                    Imagem = req.Imagem,
                    Estrategia = req.Estrategia,
                    DestinoId = req.DestinoId,
                    TemplateId = req.TemplateId,
                    Status = response.Success ? "enviada" : "erro",
                    Detalhe = response.Success ? response.MessageId : "Falha no envio",
                    EnviadaEm = response.Success ? DateTime.UtcNow : null
                };
                db.Publicacoes.Add(pub);
                await db.SaveChangesAsync(ct);

                return Results.Ok(new
                {
                    success = response.Success,
                    message_id = response.MessageId,
                    published_at = response.PublishedAt,
                    publicacao_id = pub.Id
                });
            }
            catch (Exception ex)
            {
                // Registra como erro
                var pub = new Publicacao
                {
                    ProdutoId = req.Id ?? "",
                    Nome = req.Nome ?? "",
                    Categoria = req.Categoria,
                    Preco = req.Preco,
                    Comissao = req.Comissao,
                    Link = req.Link,
                    Imagem = req.Imagem,
                    Estrategia = req.Estrategia,
                    DestinoId = req.DestinoId,
                    TemplateId = req.TemplateId,
                    Status = "erro",
                    Detalhe = ex.Message
                };
                db.Publicacoes.Add(pub);
                await db.SaveChangesAsync(ct);

                return Results.Ok(new
                {
                    success = false,
                    message_id = "",
                    published_at = "",
                    publicacao_id = pub.Id,
                    error = ex.Message
                });
            }
        }).RequireAuthorization().WithTags("Publicações");

        return app;
    }
}

public sealed record AgendarPublicacaoRequest
{
    public string? Id { get; init; }
    public string? ProdutoId { get; init; }
    public string? Nome { get; init; }
    public string? Categoria { get; init; }
    public decimal Preco { get; init; }
    public double Comissao { get; init; }
    public string? Link { get; init; }
    public string? Imagem { get; init; }
    public string? Estrategia { get; init; }
    public string? DestinoId { get; init; }
    public string? TemplateId { get; init; }
    public DateTime? AgendadaEm { get; init; }
}

public sealed record PublicarCompatRequest
{
    public string? Id { get; init; }
    public string? Nome { get; init; }
    public string? Categoria { get; init; }
    public decimal Preco { get; init; }
    public double Comissao { get; init; }
    public string? Link { get; init; }
    public string? Imagem { get; init; }
    public string? Estrategia { get; init; }
    public string? DestinoId { get; init; }
    public string? TemplateId { get; init; }
}
