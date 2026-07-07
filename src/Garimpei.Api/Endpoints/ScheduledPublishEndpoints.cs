using Garimpei.Infrastructure.Persistence;
using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Logging;

/// <summary>
/// Internal endpoint for processing scheduled publications.
/// Called by the Scheduler Go service when a publication's scheduled time arrives.
/// No auth required — internal network only (Cloud Run sidecar → ingress container).
/// </summary>
public static partial class EndpointExtensions
{
    public static WebApplication MapScheduledPublishEndpoints(this WebApplication app)
    {
        app.MapPost("/internal/publish-scheduled", async (
            AppDbContext db,
            Publisher.V1.PublisherService.PublisherServiceClient publisher,
            Collector.V1.CollectorService.CollectorServiceClient collector,
            ILogger<AppDbContext> logger,
            ScheduledPublishRequest req,
            CancellationToken ct) =>
        {
            if (string.IsNullOrWhiteSpace(req.PublicacaoId))
                return Results.BadRequest(new { error = "publicacao_id é obrigatório" });

            if (!Guid.TryParse(req.PublicacaoId, out var pubGuid))
                return Results.BadRequest(new { error = "publicacao_id inválido" });

            // Busca a publicação agendada (sem query filter de tenant — interno)
            var pub = await db.Publicacoes
                .IgnoreQueryFilters()
                .FirstOrDefaultAsync(p => p.Id == pubGuid, ct);

            if (pub is null)
                return Results.NotFound(new { error = "publicação não encontrada" });

            if (pub.Status != "agendada")
                return Results.Ok(new { status = pub.Status, detail = "já processada" });

            // Resolve destino
            string resolvedGroupId = "";
            if (!string.IsNullOrWhiteSpace(pub.DestinoId))
            {
                if (Guid.TryParse(pub.DestinoId, out var destinoGuid))
                {
                    var destino = await db.Destinos
                        .IgnoreQueryFilters()
                        .FirstOrDefaultAsync(d => d.Id == destinoGuid, ct);
                    resolvedGroupId = destino?.Config ?? pub.DestinoId;
                }
                else
                {
                    resolvedGroupId = pub.DestinoId;
                }
            }

            if (string.IsNullOrWhiteSpace(resolvedGroupId))
            {
                pub.Status = "erro";
                pub.Detalhe = "destino não configurado";
                await db.SaveChangesAsync(ct);
                return Results.Ok(new { status = "erro", detail = "destino não configurado" });
            }

            // Gera link de afiliada com tracking
            var productUrl = pub.Link ?? "";
            if (!string.IsNullOrWhiteSpace(productUrl))
            {
                try
                {
                    var subIds = new[] {
                        resolvedGroupId.Replace("@", ""),
                        pub.Estrategia ?? "agendada",
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
                    logger.LogWarning(ex, "GenerateAffiliateLink falhou para pub agendada {PubId}", pub.Id);
                }
            }

            // Publica via Publisher gRPC
            try
            {
                var grpcRequest = new Publisher.V1.PublishRequest
                {
                    Channel = "telegram",
                    GroupId = resolvedGroupId,
                    Content = new Publisher.V1.PublishContent
                    {
                        Title = pub.Nome,
                        Description = "",
                        ImageUrl = pub.Imagem ?? "",
                        ProductUrl = productUrl,
                        Price = (double)pub.Preco,
                        OriginalPrice = (double)pub.Preco,
                        DiscountPercent = 0
                    }
                };

                var response = await publisher.PublishAsync(grpcRequest, cancellationToken: ct);

                pub.Status = response.Success ? "enviada" : "erro";
                pub.Detalhe = response.Success ? response.MessageId : "Falha no envio";
                pub.EnviadaEm = response.Success ? DateTime.UtcNow : null;
                pub.Link = productUrl; // persiste link de afiliada gerado
            }
            catch (Exception ex)
            {
                pub.Status = "erro";
                pub.Detalhe = ex.Message;
                logger.LogError(ex, "Falha ao enviar publicação agendada {PubId}", pub.Id);
            }

            await db.SaveChangesAsync(ct);

            logger.LogInformation("Publicação agendada processada: {PubId} → {Status}", pub.Id, pub.Status);

            return Results.Ok(new { status = pub.Status, detail = pub.Detalhe ?? "" });
        });

        return app;
    }
}

public sealed record ScheduledPublishRequest
{
    public string? PublicacaoId { get; init; }
}
