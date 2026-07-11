using Garimpei.Domain.Entities;
using Xunit;

namespace Garimpei.Tests;

public class SchedulerJobsTests
{
    [Fact]
    public void BuildRequest_KeywordSearch_IncludesBuscaIdAndCollectionKeys()
    {
        var busca = new Busca
        {
            Id = Guid.Parse("11111111-1111-1111-1111-111111111111"),
            Keyword = "serum",
            Keywords = ["serum"],
            OwnerUid = "user-123",
            Marketplaces = "shopee"
        };

        var req = SchedulerJobs.BuildRequest(busca, enabled: true);

        Assert.True(req.Params.ContainsKey("busca_id"));
        Assert.Equal("11111111-1111-1111-1111-111111111111", req.Params["busca_id"]);
        Assert.True(req.Params.ContainsKey("collection_keys"));
        Assert.Equal("serum", req.Params["collection_keys"]);
        Assert.Equal("keyword_search", req.Params["type"]);
    }

    [Fact]
    public void BuildRequest_ShopCollection_IncludesBuscaIdAndCollectionKeys()
    {
        var busca = new Busca
        {
            Id = Guid.Parse("22222222-2222-2222-2222-222222222222"),
            Keyword = "Glory of Seoul",
            Keywords = [],
            ShopIds = [920292999],
            OwnerUid = "user-123",
            Marketplaces = "shopee"
        };

        var req = SchedulerJobs.BuildRequest(busca, enabled: true);

        Assert.Equal("22222222-2222-2222-2222-222222222222", req.Params["busca_id"]);
        Assert.Equal("920292999", req.Params["collection_keys"]);
        Assert.Equal("shop_collection", req.Params["type"]);
        Assert.Equal("920292999", req.Params["shop_id"]);
    }

    [Fact]
    public void BuildRequest_Mixed_TypeIsMixed()
    {
        var busca = new Busca
        {
            Id = Guid.Parse("33333333-3333-3333-3333-333333333333"),
            Keyword = "serum vitamina c",
            Keywords = ["serum vitamina c"],
            ShopIds = [920292999],
            OwnerUid = "user-123",
            Marketplaces = "shopee"
        };

        var req = SchedulerJobs.BuildRequest(busca, enabled: true);

        Assert.Equal("33333333-3333-3333-3333-333333333333", req.Params["busca_id"]);
        Assert.Equal("mixed", req.Params["type"]);
        Assert.Contains("920292999", req.Params["collection_keys"]);
        Assert.Contains("serum vitamina c", req.Params["collection_keys"]);
    }

    [Fact]
    public void BuildRequest_AlwaysHasBuscaId()
    {
        var busca = new Busca
        {
            Id = Guid.NewGuid(),
            Keyword = "",
            Keywords = ["test"],
            OwnerUid = "owner",
            Marketplaces = "shopee"
        };

        var req = SchedulerJobs.BuildRequest(busca, enabled: true);

        Assert.True(req.Params.ContainsKey("busca_id"));
        Assert.NotEmpty(req.Params["busca_id"]);
        Assert.True(req.Params.ContainsKey("collection_keys"));
        Assert.NotEmpty(req.Params["collection_keys"]);
    }
}
