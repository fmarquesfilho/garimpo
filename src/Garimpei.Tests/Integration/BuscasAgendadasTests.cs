using Garimpei.Domain.Entities;
using Garimpei.Infrastructure.Persistence;
using Garimpei.Infrastructure.Tenancy;
using Microsoft.EntityFrameworkCore;
using Xunit;

namespace Garimpei.Tests.Integration;

/// <summary>
/// Testes de integração para buscas agendadas — BuscaContract unificado.
/// Verifica persistência, SchedulerJobs.BuildRequest, e soft-delete.
/// Identidade: UUID (busca.Id). Zero dependência de campo Keyword legado.
/// </summary>
public class BuscasAgendadasTests : IDisposable
{
    private readonly AppDbContext _db;
    private readonly TenantContext _tenant;

    public BuscasAgendadasTests()
    {
        _tenant = new TenantContext();
        _tenant.Set("test-buscas-agendadas");

        var options = new DbContextOptionsBuilder<AppDbContext>()
            .UseInMemoryDatabase($"BuscasAgendadas_{Guid.NewGuid()}")
            .Options;

        _db = new AppDbContext(options, _tenant);
    }

    public void Dispose() => _db.Dispose();

    // ═══════════════════════════════════════════════════════════════════════
    // Busca sem Keywords (monitorar todos os produtos da loja)
    // ═══════════════════════════════════════════════════════════════════════

    [Fact]
    public async Task Busca_SemKeywords_PersisteSoComShopIds()
    {
        var busca = new Busca
        {
            OwnerUid = "test-buscas-agendadas",
            ShopIds = [920292999],
            ShopNames = new() { ["920292999"] = "Glory of Seoul" },
            Keywords = null,
            CronExpression = null
        };

        _db.Buscas.Add(busca);
        await _db.SaveChangesAsync();

        var loaded = await _db.Buscas.FindAsync(busca.Id);
        Assert.NotNull(loaded);
        Assert.Contains(920292999, loaded.ShopIds!);
        Assert.Null(loaded.Keywords);
        Assert.Null(loaded.CronExpression);
    }

    [Fact]
    public async Task Busca_SemKeywords_JobId_SegueConvencao()
    {
        var busca = new Busca
        {
            OwnerUid = "test-buscas-agendadas",
            ShopIds = [123456789]
        };

        _db.Buscas.Add(busca);
        await _db.SaveChangesAsync();

        var expectedJobId = $"busca-{busca.Id}";
        Assert.StartsWith("busca-", expectedJobId);
        Assert.Contains(busca.Id.ToString(), expectedJobId);
    }

    // ═══════════════════════════════════════════════════════════════════════
    // Busca com Keywords (monitorar produtos filtrados)
    // ═══════════════════════════════════════════════════════════════════════

    [Fact]
    public async Task Busca_ComKeywords_PersisteArrayCorreto()
    {
        var busca = new Busca
        {
            OwnerUid = "test-buscas-agendadas",
            ShopIds = [920292999],
            Keywords = ["serum", "protetor solar", "vitamina c"]
        };

        _db.Buscas.Add(busca);
        await _db.SaveChangesAsync();

        var loaded = await _db.Buscas.FindAsync(busca.Id);
        Assert.NotNull(loaded);
        Assert.NotNull(loaded.Keywords);
        Assert.Equal(3, loaded.Keywords.Length);
        Assert.Contains("serum", loaded.Keywords);
        Assert.Contains("protetor solar", loaded.Keywords);
        Assert.Contains("vitamina c", loaded.Keywords);
    }

    [Fact]
    public async Task Busca_ComCronCustomizado_Persiste()
    {
        var busca = new Busca
        {
            OwnerUid = "test-buscas-agendadas",
            ShopIds = [111222333],
            CronExpression = "0 */4 * * *"
        };

        _db.Buscas.Add(busca);
        await _db.SaveChangesAsync();

        var loaded = await _db.Buscas.FindAsync(busca.Id);
        Assert.NotNull(loaded);
        Assert.Equal("0 */4 * * *", loaded.CronExpression);
    }

    [Fact]
    public async Task Busca_Marketplaces_PersisteComoArray()
    {
        var busca = new Busca
        {
            OwnerUid = "test-buscas-agendadas",
            Keywords = ["perfume"],
            Marketplaces = ["shopee", "amazon"]
        };

        _db.Buscas.Add(busca);
        await _db.SaveChangesAsync();

        var loaded = await _db.Buscas.FindAsync(busca.Id);
        Assert.NotNull(loaded);
        Assert.Equal(2, loaded.Marketplaces.Length);
        Assert.Contains("shopee", loaded.Marketplaces);
        Assert.Contains("amazon", loaded.Marketplaces);
    }

    // ═══════════════════════════════════════════════════════════════════════
    // SchedulerJobs.BuildRequest (BuscaContract unificado)
    // ═══════════════════════════════════════════════════════════════════════

    [Fact]
    public void SchedulerJobs_LojaComKeywords_MontaMixed()
    {
        var busca = new Busca
        {
            OwnerUid = "uid-1",
            ShopIds = [920292999],
            Keywords = ["serum", "protetor"]
        };

        var req = SchedulerJobs.BuildRequest(busca, enabled: true);

        Assert.Equal($"busca-{busca.Id}", req.JobId);
        Assert.True(req.Enabled);
        Assert.Equal("mixed", req.Params["type"]);
        Assert.Equal("920292999", req.Params["shop_id"]);
        Assert.Equal("serum,protetor", req.Params["keywords"]);
        Assert.True(req.Params.ContainsKey("busca_id"));
        Assert.True(req.Params.ContainsKey("collection_keys"));
    }

    [Fact]
    public void SchedulerJobs_BuscaPalavraChave_MontaKeywordSearch()
    {
        var busca = new Busca
        {
            OwnerUid = "uid-2",
            Keywords = ["serum", "vitamina c"],
            CronExpression = "0 9 * * *"
        };

        var req = SchedulerJobs.BuildRequest(busca, enabled: true);

        Assert.Equal("keyword_search", req.Params["type"]);
        Assert.False(req.Params.ContainsKey("shop_id"));
        Assert.Equal("serum,vitamina c", req.Params["keywords"]);
        Assert.Equal("0 9 * * *", req.CronExpression);
        Assert.Equal(busca.Id.ToString(), req.Params["busca_id"]);
    }

    [Fact]
    public void SchedulerJobs_SemCron_UsaDefault()
    {
        var busca = new Busca
        {
            OwnerUid = "uid-3",
            ShopIds = [123]
        };

        var req = SchedulerJobs.BuildRequest(busca, enabled: true);

        Assert.Equal(SchedulerJobs.DefaultCron, req.CronExpression);
    }

    [Fact]
    public void SchedulerJobs_LojaSemKeywords_NaoIncluiKeywords()
    {
        var busca = new Busca
        {
            OwnerUid = "uid-4",
            ShopIds = [999]
        };

        var req = SchedulerJobs.BuildRequest(busca, enabled: true);

        Assert.Equal("shop_collection", req.Params["type"]);
        Assert.False(req.Params.ContainsKey("keywords"));
    }

    // ═══════════════════════════════════════════════════════════════════════
    // Soft-delete
    // ═══════════════════════════════════════════════════════════════════════

    [Fact]
    public async Task SoftDelete_SetsActiveFalse_AndUpdatedAt()
    {
        var busca = new Busca
        {
            OwnerUid = "test-buscas-agendadas",
            ShopIds = [111]
        };

        _db.Buscas.Add(busca);
        await _db.SaveChangesAsync();

        busca.Active = false;
        busca.UpdatedAt = DateTime.UtcNow;
        await _db.SaveChangesAsync();

        var loaded = await _db.Buscas.IgnoreQueryFilters().FirstAsync(b => b.Id == busca.Id);
        Assert.False(loaded.Active);
        Assert.True(loaded.UpdatedAt > loaded.CreatedAt);
    }
}
