using Garimpei.Domain.Interfaces;

namespace Garimpei.Domain.Entities;

public sealed class Product : IOwnedEntity
{
    public Guid Id { get; init; } = Guid.NewGuid();
    public required long ItemId { get; init; }
    public required long ShopId { get; init; }
    public required string Name { get; init; }
    public required decimal Price { get; set; }
    public decimal OriginalPrice { get; set; }
    public int Sold { get; set; }
    public double Rating { get; set; }
    public string? ImageUrl { get; set; }
    public string? ProductUrl { get; set; }
    public string? ShopName { get; set; }
    public double DiscountPercent { get; set; }
    public string OwnerUid { get; set; } = string.Empty;
    public DateTime CreatedAt { get; init; } = DateTime.UtcNow;
    public DateTime UpdatedAt { get; set; } = DateTime.UtcNow;
}
