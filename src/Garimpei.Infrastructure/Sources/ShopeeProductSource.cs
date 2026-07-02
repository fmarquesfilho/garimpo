using Collector.V1;
using Garimpei.Domain;
using Garimpei.Domain.Interfaces;
using Garimpei.Domain.ValueObjects;

namespace Garimpei.Infrastructure.Sources;

/// <summary>
/// Implementação de IProductSource para Shopee.
/// Delega para o collector-shopee via gRPC.
/// </summary>
public sealed class ShopeeProductSource : IProductSource
{
    private readonly CollectorService.CollectorServiceClient _collector;

    public ShopeeProductSource(CollectorService.CollectorServiceClient collector)
    {
        _collector = collector;
    }

    public string MarketplaceId => Marketplaces.Shopee;

    public async Task<SourceResult> SearchAsync(SearchQuery query, CancellationToken ct = default)
    {
        var response = await _collector.FetchAsync(new FetchRequest
        {
            Keyword = query.Keyword,
            Limit = query.Limit,
            SortBy = query.SortBy ?? "",
            OwnerUid = query.OwnerUid ?? "",
            Marketplace = Marketplace.Shopee
        }, cancellationToken: ct);

        return new SourceResult
        {
            Products = response.Products.Select(ProductMappings.ToCandidate).ToList(),
            TotalFound = response.TotalFound,
            FetchedAt = DateTime.UtcNow
        };
    }

    public async Task<SourceResult> FetchByShopAsync(string shopId, int limit, CancellationToken ct = default)
    {
        if (!long.TryParse(shopId, out var numericId))
            return new SourceResult { Products = [], TotalFound = 0 };

        var response = await _collector.FetchShopAsync(new FetchShopRequest
        {
            ShopId = numericId,
            Limit = limit,
            Marketplace = Marketplace.Shopee
        }, cancellationToken: ct);

        return new SourceResult
        {
            Products = response.Products.Select(ProductMappings.ToCandidate).ToList(),
            TotalFound = response.TotalFound,
            FetchedAt = DateTime.UtcNow
        };
    }

    public string GenerateAffiliateLink(string productUrl, string affiliateTag)
    {
        // Shopee: links de afiliado já vêm prontos da API (offerLink).
        // Se precisar gerar manualmente, seria via API call.
        return productUrl;
    }
}
