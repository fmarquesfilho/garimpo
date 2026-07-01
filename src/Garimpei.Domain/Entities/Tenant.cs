namespace Garimpei.Domain.Entities;

public sealed class Tenant
{
    public Guid Id { get; init; } = Guid.NewGuid();
    public required string OwnerUid { get; init; }
    public required string Name { get; init; }
    public string? Email { get; set; }
    public bool Active { get; set; } = true;
    public DateTime CreatedAt { get; init; } = DateTime.UtcNow;
}
