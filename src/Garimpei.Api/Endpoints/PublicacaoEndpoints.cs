using Publisher.V1;

/// <summary>
/// Publication endpoints — publish offers to Telegram/WhatsApp via publisher gRPC sidecar.
/// </summary>
public static partial class EndpointExtensions
{
    public static RouteGroupBuilder MapPublicacaoEndpoints(this RouteGroupBuilder group)
    {
        var pub = group.MapGroup("/publicar").WithTags("Publicação");

        pub.MapPost("/", async (
            PublisherService.PublisherServiceClient publisher,
            PublishOfferRequest request,
            CancellationToken ct) =>
        {
            var grpcRequest = new PublishRequest
            {
                Channel = request.Channel ?? "telegram",
                GroupId = request.GroupId ?? "",
                Content = new PublishContent
                {
                    Title = request.Title,
                    Description = request.Description ?? "",
                    ImageUrl = request.ImageUrl ?? "",
                    ProductUrl = request.ProductUrl ?? "",
                    Price = request.Price,
                    OriginalPrice = request.OriginalPrice,
                    DiscountPercent = request.DiscountPercent,
                }
            };

            var response = await publisher.PublishAsync(grpcRequest, cancellationToken: ct);

            return Results.Ok(new
            {
                success = response.Success,
                message_id = response.MessageId,
                published_at = response.PublishedAt,
                channel = request.Channel ?? "telegram"
            });
        });

        pub.MapGet("/destinos", async (
            PublisherService.PublisherServiceClient publisher,
            string? channel,
            CancellationToken ct) =>
        {
            var response = await publisher.ListGroupsAsync(
                new ListGroupsRequest { Channel = channel ?? "" },
                cancellationToken: ct);

            return Results.Ok(new { destinos = response.Groups });
        });

        return group;
    }
}

public sealed record PublishOfferRequest
{
    public required string Title { get; init; }
    public string? Description { get; init; }
    public string? ImageUrl { get; init; }
    public string? ProductUrl { get; init; }
    public double Price { get; init; }
    public double OriginalPrice { get; init; }
    public double DiscountPercent { get; init; }
    public string? Channel { get; init; }
    public string? GroupId { get; init; }
}
