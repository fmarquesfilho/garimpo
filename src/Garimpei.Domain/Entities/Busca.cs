using Garimpei.Domain.Interfaces;

namespace Garimpei.Domain.Entities;

public sealed class Busca : IOwnedEntity
{
    public Guid Id { get; init; } = Guid.NewGuid();
    public required string Keyword { get; init; }
    public required string OwnerUid { get; set; }
    public string SortBy { get; init; } = "relevance";
    public int Limit { get; init; } = 50;
    public bool Active { get; set; } = true;
    public DateTime CreatedAt { get; init; } = DateTime.UtcNow;
    public DateTime UpdatedAt { get; set; } = DateTime.UtcNow;
}
