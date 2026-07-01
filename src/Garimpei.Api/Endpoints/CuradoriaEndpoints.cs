using MediatR;

public static partial class EndpointExtensions
{
    public static RouteGroupBuilder MapCuradoriaEndpoints(this RouteGroupBuilder group)
    {
        var curadoria = group.MapGroup("/curadoria").WithTags("Curadoria");

        curadoria.MapGet("/ranking", async (IMediator mediator, CancellationToken ct) =>
        {
            // TODO: implement via MediatR query (scoring + ranking)
            return Results.Ok(Array.Empty<object>());
        });

        return group;
    }
}
