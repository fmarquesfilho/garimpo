using Collector.V1;
using Garimpei.Domain;
using Garimpei.Domain.Interfaces;
using Garimpei.Domain.ValueObjects;

namespace Garimpei.Infrastructure.Sources;

/// <summary>
/// Implementação de IProductSource para Amazon.
/// Delega para o collector-amazon via gRPC.
/// </summary>
public sealed class AmazonProductSource : IProductSource
{
    private readonly CollectorService.CollectorServiceClient _collector;

    public AmazonProductSource(CollectorService.CollectorServiceClient collector)
    {
        _collector = collector;
    }

    public string MarketplaceId => Marketplaces.Amazon;

    public async Task<SourceResult> SearchAsync(SearchQuery query, CancellationToken ct = default)
    {
        var response = await _collector.FetchAsync(new FetchRequest
        {
            Keyword = query.Keyword,
            Limit = query.Limit,
            SortBy = query.SortBy ?? "",
            OwnerUid = query.OwnerUid ?? "",
            Marketplace = Marketplace.Amazon
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
        // Amazon não suporta busca por loja.
        return new SourceResult { Products = [], TotalFound = 0 };
    }

    public string GenerateAffiliateLink(string productUrl, string affiliateTag)
    {
        // Amazon: append ?tag=PARTNER_TAG ao URL do produto.
        if (string.IsNullOrEmpty(affiliateTag))
            return productUrl;

        var separator = productUrl.Contains('?') ? "&" : "?";
        return $"{productUrl}{separator}tag={affiliateTag}";
    }
}
