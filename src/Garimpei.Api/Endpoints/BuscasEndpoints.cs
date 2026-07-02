/// <summary>
/// V2 Buscas endpoints — intentionally empty.
/// The actual /api/v2/buscas CRUD is handled by LojasEndpoints.cs
/// (MapLojasEndpoints maps to the same /buscas group with full EF Core implementation).
/// This file kept for future MediatR-based refactoring.
/// </summary>
public static partial class EndpointExtensions
{
    public static RouteGroupBuilder MapBuscasEndpoints(this RouteGroupBuilder group)
    {
        // Handled by MapLojasEndpoints — same /buscas group
        return group;
    }
}
