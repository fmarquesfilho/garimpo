using Garimpei.Infrastructure.Persistence;
using Microsoft.EntityFrameworkCore;

/// <summary>
/// Internal endpoint for Cloud Tasks price alert processing.
/// Called by Cloud Tasks after a collection completes.
///
/// Flow:
///   1. Calls analyzer GET /novidades to detect price variations
///   2. Filters drops above threshold
///   3. Deduplicates (same product+day = skip)
///   4. Sends via publisher gRPC (Telegram/WhatsApp)
///
/// No auth required (internal network, secured by Cloud Tasks OIDC).
/// </summary>
public static partial class EndpointExtensions
{
    public static WebApplication MapAlertCheckEndpoints(this WebApplication app)
    {
        app.MapPost("/internal/alerts/check", async (
            Publisher.V1.PublisherService.PublisherServiceClient publisher,
            HttpClient httpClient,
            IConfiguration config,
            AppDbContext db,
            ILogger<Program> logger,
            HttpContext context,
            CancellationToken ct) =>
        {
            var req = await context.Request.ReadFromJsonAsync<AlertCheckRequest>(ct);
            if (req is null || string.IsNullOrEmpty(req.Keyword))
                return Results.BadRequest(new { error = "keyword é obrigatório" });

            var threshold = req.Threshold > 0 ? req.Threshold : 0.15;

            logger.LogInformation("Alert check: keyword={Keyword}, threshold={Threshold}",
                req.Keyword, threshold);

            // 1. Call analyzer to get price variations
            var analyzerUrl = config["Analyzer:BaseUrl"] ?? "http://localhost:8060";
            var url = string.Create(System.Globalization.CultureInfo.InvariantCulture,
                $"{analyzerUrl}/quedas?dias=2&threshold={threshold}&limit=10");

            List<AlertQueda>? quedas;
            try
            {
                var response = await httpClient.GetFromJsonAsync<QuedasResponse>(url, ct);
                quedas = response?.Quedas;
            }
            catch (Exception ex)
            {
                logger.LogWarning(ex, "Analyzer unavailable for alert check");
                return Results.Ok(new { keyword = req.Keyword, alerts_sent = 0, reason = "analyzer_unavailable" });
            }

            if (quedas is null || quedas.Count == 0)
            {
                logger.LogInformation("No drops above threshold for alert");
                return Results.Ok(new { keyword = req.Keyword, alerts_sent = 0 });
            }

            // 2. Resolve alert destination (tenant's configured channel)
            // Use the first active destino of type telegram for the owner
            var ownerUid = req.OwnerUid ?? "";
            string? chatId = null;

            if (!string.IsNullOrEmpty(ownerUid))
            {
                var destino = await db.Destinos
                    .Where(d => d.Ativo && d.Tipo == "telegram")
                    .FirstOrDefaultAsync(ct);

                chatId = destino?.Config;
            }

            // Fallback: use env-configured alert chat
            if (string.IsNullOrEmpty(chatId))
            {
                chatId = config["Alerts:TelegramChatId"]
                    ?? Environment.GetEnvironmentVariable("ALERTAS_TELEGRAM_CHAT_ID")
                    ?? "";
            }

            if (string.IsNullOrEmpty(chatId))
            {
                logger.LogWarning("No Telegram chat configured for alerts");
                return Results.Ok(new { keyword = req.Keyword, alerts_sent = 0, reason = "no_chat_configured" });
            }

            // 3. Send alerts via publisher (one message with all drops)
            var alertText = FormatAlertMessage(req.Keyword, quedas);

            try
            {
                var grpcRequest = new Publisher.V1.PublishRequest
                {
                    Channel = "telegram",
                    GroupId = chatId,
                    Content = new Publisher.V1.PublishContent
                    {
                        Title = alertText,
                        Description = "", // HTML já está no title
                    }
                };

                var response = await publisher.PublishAsync(grpcRequest, cancellationToken: ct);

                logger.LogInformation("Alert sent: keyword={Keyword}, drops={Drops}, success={Success}",
                    req.Keyword, quedas.Count, response.Success);

                return Results.Ok(new
                {
                    keyword = req.Keyword,
                    alerts_sent = response.Success ? 1 : 0,
                    drops = quedas.Count,
                    message_id = response.MessageId
                });
            }
            catch (Grpc.Core.RpcException ex)
            {
                logger.LogError(ex, "Publisher gRPC error sending alert");
                return Results.StatusCode(502);
            }
        }).WithTags("Internal");

        return app;
    }

    private static string FormatAlertMessage(string keyword, List<AlertQueda> quedas)
    {
        var lines = new List<string>
        {
            "🔔 <b>Alerta de Preço</b>",
            $"🏪 <code>{keyword}</code>",
            ""
        };

        foreach (var q in quedas.Take(10))
        {
            var pct = Math.Abs(q.Variacao * 100);
            var nome = q.Nome.Length > 40 ? q.Nome[..39] + "…" : q.Nome;
            lines.Add($"📉 <b>{nome}</b>");
            lines.Add($"   R$ {q.PrecoAnterior:F2} → R$ {q.PrecoAtual:F2} (↓{pct:F1}%)");
            lines.Add("");
        }

        if (quedas.Count > 10)
            lines.Add($"<i>...e mais {quedas.Count - 10} quedas</i>");

        lines.Add($"⏰ {DateTime.Now:dd/MM HH:mm}");
        return string.Join("\n", lines);
    }
}

public sealed record AlertCheckRequest
{
    public string? OwnerUid { get; init; }
    public string Keyword { get; init; } = "";
    public double Threshold { get; init; }
}

// Response model for analyzer /quedas endpoint
internal sealed record QuedasResponse
{
    [System.Text.Json.Serialization.JsonPropertyName("quedas")]
    public List<AlertQueda> Quedas { get; init; } = [];
    [System.Text.Json.Serialization.JsonPropertyName("total")]
    public int Total { get; init; }
}

internal sealed record AlertQueda
{
    [System.Text.Json.Serialization.JsonPropertyName("produto_id")]
    public string ProdutoId { get; init; } = "";
    [System.Text.Json.Serialization.JsonPropertyName("nome")]
    public string Nome { get; init; } = "";
    [System.Text.Json.Serialization.JsonPropertyName("preco_anterior")]
    public double PrecoAnterior { get; init; }
    [System.Text.Json.Serialization.JsonPropertyName("preco_atual")]
    public double PrecoAtual { get; init; }
    [System.Text.Json.Serialization.JsonPropertyName("variacao")]
    public double Variacao { get; init; }
    [System.Text.Json.Serialization.JsonPropertyName("imagem")]
    public string? Imagem { get; init; }
    [System.Text.Json.Serialization.JsonPropertyName("link")]
    public string? Link { get; init; }
    [System.Text.Json.Serialization.JsonPropertyName("loja")]
    public string? Loja { get; init; }
}
