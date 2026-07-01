using MediatR;

public static partial class EndpointExtensions
{
    public static RouteGroupBuilder MapBuscasEndpoints(this RouteGroupBuilder group)
    {
        var buscas = group.MapGroup("/buscas").WithTags("Buscas");

        buscas.MapGet("/", async (IMediator mediator, CancellationToken ct) =>
        {
            // TODO: implement via MediatR query
            return Results.Ok(Array.Empty<object>());
        });

        buscas.MapPost("/", async (IMediator mediator, CancellationToken ct) =>
        {
            // TODO: implement via MediatR command
            return Results.Created();
        });

        return group;
    }
}
