/// <summary>
/// Resolver Link endpoint — resolve link curto da Shopee para URL final + dados do produto.
/// /api/resolver-link
/// </summary>
public static partial class EndpointExtensions
{
    public static WebApplication MapResolverLinkEndpoints(this WebApplication app)
    {
        app.MapPost("/api/resolver-link", async (
            HttpClient httpClient,
            ResolverLinkRequest req,
            CancellationToken ct) =>
        {
            if (string.IsNullOrWhiteSpace(req.Url))
                return Results.BadRequest(new { error = "url é obrigatório" });

            try
            {
                // Resolve o redirect da URL curta Shopee
                var request = new HttpRequestMessage(HttpMethod.Head, req.Url);
                request.Headers.Add("User-Agent", "Mozilla/5.0");
                var response = await httpClient.SendAsync(request, HttpCompletionOption.ResponseHeadersRead, ct);

                var finalUrl = response.RequestMessage?.RequestUri?.ToString() ?? req.Url;

                // Tenta extrair item_id e shop_id da URL final
                // Formato: https://shopee.com.br/product-name-i.SHOP_ID.ITEM_ID
                string? itemId = null;
                string? shopId = null;
                var parts = finalUrl.Split('-');
                if (parts.Length >= 2)
                {
                    var lastParts = parts[^1].Split('.');
                    if (lastParts.Length >= 2)
                    {
                        shopId = lastParts[^2].Replace("i", "");
                        itemId = lastParts[^1].Split('?')[0];
                    }
                }

                return Results.Ok(new
                {
                    url_original = req.Url,
                    url_final = finalUrl,
                    item_id = itemId ?? "",
                    shop_id = shopId ?? "",
                    resolvido = finalUrl != req.Url
                });
            }
            catch (Exception ex)
            {
                return Results.Ok(new
                {
                    url_original = req.Url,
                    url_final = req.Url,
                    item_id = "",
                    shop_id = "",
                    resolvido = false,
                    error = ex.Message
                });
            }
        }).RequireAuthorization().WithTags("Utilitários");

        return app;
    }
}

public sealed record ResolverLinkRequest
{
    public string? Url { get; init; }
}
