using Garimpei.Domain.Entities;
using Garimpei.Infrastructure.Persistence;
using Microsoft.EntityFrameworkCore;

/// <summary>
/// Publicações endpoints — agendamento e histórico de publicações.
/// /api/publicacoes e /api/publicar (formato frontend).
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

        publicacoes.MapPost("/", async (
            AppDbContext db,
            Publisher.V1.PublisherService.PublisherServiceClient publisher,
            Collector.V1.CollectorService.CollectorServiceClient collector,
            Scheduler.V1.SchedulerService.SchedulerServiceClient scheduler,
            ILogger<AppDbContext> logger,
            AgendarPublicacaoRequest req,
            CancellationToken ct) =>
        {
            var productUrl = req.Link ?? "";

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

            // Se envio imediato (sem agendamento), publica via gRPC
            if (!req.AgendadaEm.HasValue && !string.IsNullOrWhiteSpace(req.DestinoId))
            {
                string resolvedGroupId = "";
                if (Guid.TryParse(req.DestinoId, out var destinoGuid))
                {
                    var destino = await db.Destinos.FindAsync([destinoGuid], ct);
                    resolvedGroupId = destino?.Config ?? req.DestinoId;
                }
                else
                {
                    resolvedGroupId = req.DestinoId;
                }

                // Gera link de afiliada com sub_ids para rastreamento de conversão
                if (!string.IsNullOrWhiteSpace(productUrl))
                {
                    try
                    {
                        var subIds = new[] {
                            resolvedGroupId.Replace("@", ""),
                            req.Estrategia ?? "manual",
                            DateTime.UtcNow.ToString("yyyyMMdd")
                        };
                        var linkResp = await collector.GenerateAffiliateLinkAsync(
                            new Collector.V1.GenerateAffiliateLinkRequest
                            {
                                OriginalUrl = productUrl,
                                Marketplace = Collector.V1.Marketplace.Shopee,
                                SubIds = { subIds }
                            }, cancellationToken: ct);

                        if (!string.IsNullOrEmpty(linkResp.ShortLink))
                            productUrl = linkResp.ShortLink;
                    }
                    catch (Exception ex)
                    {
                        logger.LogWarning(ex, "GenerateAffiliateLink falhou, usando link original");
                    }
                }

                pub.Link = productUrl; // persiste o link de afiliada gerado

                try
                {
                    var grpcRequest = new Publisher.V1.PublishRequest
                    {
                        Channel = "telegram",
                        GroupId = resolvedGroupId,
                        Content = new Publisher.V1.PublishContent
                        {
                            Title = req.Nome ?? "",
                            Description = req.LegendaCustom ?? "",
                            ImageUrl = req.Imagem ?? "",
                            ProductUrl = productUrl,
                            Price = (double)req.Preco,
                            OriginalPrice = (double)req.Preco,
                            DiscountPercent = 0
                        }
                    };

                    var response = await publisher.PublishAsync(grpcRequest, cancellationToken: ct);
                    pub.Status = response.Success ? "enviada" : "erro";
                    pub.Detalhe = response.Success ? response.MessageId : "Falha no envio";
                    pub.EnviadaEm = response.Success ? DateTime.UtcNow : null;
                }
                catch (Exception ex)
                {
                    pub.Status = "erro";
                    pub.Detalhe = ex.Message;
                }
            }

            db.Publicacoes.Add(pub);
            await db.SaveChangesAsync(ct);

            // Se agendada, registra job no Scheduler para disparo no horário
            if (req.AgendadaEm.HasValue && pub.Status == "agendada")
            {
                try
                {
                    // Converte agendada_em para cron one-shot (minuto exato)
                    var dt = req.AgendadaEm.Value.ToUniversalTime();
                    var cronExpr = $"{dt.Minute} {dt.Hour} {dt.Day} {dt.Month} *";

                    var setReq = new Scheduler.V1.SetScheduleRequest
                    {
                        JobId = $"pub-{pub.Id}",
                        CronExpression = cronExpr,
                        Enabled = true
                    };
                    setReq.Params.Add("type", "scheduled_publish");
                    setReq.Params.Add("publicacao_id", pub.Id.ToString());
                    setReq.Params.Add("owner_uid", pub.OwnerUid);

                    await scheduler.SetScheduleAsync(setReq, cancellationToken: ct);
                }
                catch (Exception ex)
                {
                    logger.LogWarning(ex, "Falha ao agendar publicação {PubId} no Scheduler", pub.Id);
                }
            }

            return Results.Ok(new { publicacao = new { id = pub.Id, status = pub.Status, detalhe = pub.Detalhe ?? "", criada_em = pub.CreatedAt.ToString("o") } });
        });

        // /api/publicar — publicar imediatamente (envia via gRPC publisher)
        app.MapPost("/api/publicar", async (
            AppDbContext db,
            Publisher.V1.PublisherService.PublisherServiceClient publisher,
            Collector.V1.CollectorService.CollectorServiceClient collector,
            ILogger<AppDbContext> logger,
            PublicarRequest req,
            CancellationToken ct) =>
        {
            // Resolve o destino: se DestinoId informado, busca o Config (chat_id real)
            // O publisher Go usa o Config diretamente — não conhece os UUIDs do PostgreSQL.
            string resolvedGroupId = "";
            if (!string.IsNullOrWhiteSpace(req.DestinoId))
            {
                if (Guid.TryParse(req.DestinoId, out var destinoGuid))
                {
                    var destino = await db.Destinos.FindAsync([destinoGuid], ct);
                    resolvedGroupId = destino?.Config ?? req.DestinoId;
                }
                else
                {
                    // Se não for UUID, assume que já é o chat_id diretamente
                    resolvedGroupId = req.DestinoId;
                }
            }

            // Gera link de afiliada com tracking via Collector (generateShortLink)
            var productUrl = req.Link ?? "";
            if (!string.IsNullOrWhiteSpace(productUrl))
            {
                try
                {
                    var subIds = new[] {
                        resolvedGroupId.Replace("@", ""),  // canal
                        req.Estrategia ?? "manual",        // estratégia
                        DateTime.UtcNow.ToString("yyyyMMdd") // data
                    };
                    var linkResp = await collector.GenerateAffiliateLinkAsync(
                        new Collector.V1.GenerateAffiliateLinkRequest
                        {
                            OriginalUrl = productUrl,
                            Marketplace = Collector.V1.Marketplace.Shopee,
                            SubIds = { subIds }
                        }, cancellationToken: ct);

                    if (!string.IsNullOrEmpty(linkResp.ShortLink))
                        productUrl = linkResp.ShortLink;
                }
                catch (Exception ex)
                {
                    // Fallback: usa link original se GenerateAffiliateLink falhar
                    logger.LogWarning(ex, "GenerateAffiliateLink falhou, usando link original");
                }
            }

            // Tentativa de envio via publisher gRPC
            try
            {
                var grpcRequest = new Publisher.V1.PublishRequest
                {
                    Channel = "telegram",
                    GroupId = resolvedGroupId,
                    Content = new Publisher.V1.PublishContent
                    {
                        Title = req.Nome ?? "",
                        Description = req.Categoria ?? "",
                        ImageUrl = req.Imagem ?? "",
                        ProductUrl = productUrl,
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
                    Link = productUrl, // link de afiliada gerado (ou original como fallback)
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
    [System.Text.Json.Serialization.JsonPropertyName("id")]
    public string? Id { get; init; }
    [System.Text.Json.Serialization.JsonPropertyName("produto_id")]
    public string? ProdutoId { get; init; }
    [System.Text.Json.Serialization.JsonPropertyName("nome")]
    public string? Nome { get; init; }
    [System.Text.Json.Serialization.JsonPropertyName("categoria")]
    public string? Categoria { get; init; }
    [System.Text.Json.Serialization.JsonPropertyName("preco")]
    public decimal Preco { get; init; }
    [System.Text.Json.Serialization.JsonPropertyName("comissao")]
    public double Comissao { get; init; }
    [System.Text.Json.Serialization.JsonPropertyName("link")]
    public string? Link { get; init; }
    [System.Text.Json.Serialization.JsonPropertyName("imagem")]
    public string? Imagem { get; init; }
    [System.Text.Json.Serialization.JsonPropertyName("estrategia")]
    public string? Estrategia { get; init; }
    [System.Text.Json.Serialization.JsonPropertyName("destino_id")]
    public string? DestinoId { get; init; }
    [System.Text.Json.Serialization.JsonPropertyName("template_id")]
    public string? TemplateId { get; init; }
    [System.Text.Json.Serialization.JsonPropertyName("agendada_em")]
    public DateTime? AgendadaEm { get; init; }
    [System.Text.Json.Serialization.JsonPropertyName("legenda_custom")]
    public string? LegendaCustom { get; init; }
}

public sealed record PublicarRequest
{
    [System.Text.Json.Serialization.JsonPropertyName("id")]
    public string? Id { get; init; }
    [System.Text.Json.Serialization.JsonPropertyName("nome")]
    public string? Nome { get; init; }
    [System.Text.Json.Serialization.JsonPropertyName("categoria")]
    public string? Categoria { get; init; }
    [System.Text.Json.Serialization.JsonPropertyName("preco")]
    public decimal Preco { get; init; }
    [System.Text.Json.Serialization.JsonPropertyName("comissao")]
    public double Comissao { get; init; }
    [System.Text.Json.Serialization.JsonPropertyName("link")]
    public string? Link { get; init; }
    [System.Text.Json.Serialization.JsonPropertyName("imagem")]
    public string? Imagem { get; init; }
    [System.Text.Json.Serialization.JsonPropertyName("estrategia")]
    public string? Estrategia { get; init; }
    [System.Text.Json.Serialization.JsonPropertyName("destino_id")]
    public string? DestinoId { get; init; }
    [System.Text.Json.Serialization.JsonPropertyName("template_id")]
    public string? TemplateId { get; init; }
}
