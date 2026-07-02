using Collector.V1;
using Garimpei.Domain;
using Garimpei.Domain.ValueObjects;

namespace Garimpei.Infrastructure.Sources;

/// <summary>
/// Mapeamento centralizado: proto Product → domain ProductCandidate.
/// Elimina duplicação entre CompatEndpoints, CuradoriaEndpoints e qualquer
/// outro consumer do gRPC collector.
/// </summary>
public static class ProductMappings
{
    /// <summary>
    /// Converte um proto Product (gRPC) para um ProductCandidate (domínio).
    /// </summary>
    public static ProductCandidate ToCandidate(Product p) => new()
    {
        Id = p.ItemId != 0 ? p.ItemId.ToString() : p.Name.GetHashCode().ToString("x8"),
        Name = p.Name,
        Category = p.Category,
        ShopName = p.ShopName,
        ShopId = p.ShopId != 0 ? p.ShopId.ToString() : null,
        Price = (decimal)p.Price,
        OriginalPrice = (decimal)p.OriginalPrice,
        DiscountPercent = p.DiscountPercent,
        Commission = p.Commission,
        Sales = p.Sold,
        Rating = p.Rating,
        Link = p.Link,
        ProductLink = p.ProductUrl,
        ImageUrl = p.ImageUrl,
        Marketplace = MapMarketplace(p.Marketplace)
    };

    /// <summary>
    /// Converte o enum proto Marketplace para a string de domínio.
    /// </summary>
    public static string MapMarketplace(Marketplace m) => m switch
    {
        Marketplace.Shopee => Marketplaces.Shopee,
        Marketplace.Amazon => Marketplaces.Amazon,
        Marketplace.Mercadolivre => Marketplaces.MercadoLivre,
        _ => Marketplaces.Shopee
    };
}
