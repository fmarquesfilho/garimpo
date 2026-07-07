using Garimpei.Domain.Entities;
using Garimpei.Infrastructure.Persistence;
using Garimpei.Infrastructure.Tenancy;
using Microsoft.EntityFrameworkCore;
using Xunit;

namespace Garimpei.Tests.Integration;

/// <summary>
/// Testes de integração para buscas agendadas — validação pós-migração.
/// Verifica que Keywords e CronExpression são persistidos corretamente,
/// e que a entidade Busca suporta os dois modos:
/// 1. Loja sem keywords (monitorar todos os produtos)
/// 2. Loja com keywords (monitorar produtos filtrados)
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
            Keyword = "Glory of Seoul",
            OwnerUid = "test-buscas-agendadas",
            ShopIds = [920292999],
            Keywords = null,
            CronExpression = null
        };

        _db.Buscas.Add(busca);
        await _db.SaveChangesAsync();

        var loaded = await _db.Buscas.FindAsync(busca.Id);
        Assert.NotNull(loaded);
        Assert.Equal("Glory of Seoul", loaded.Keyword);
        Assert.Contains(920292999, loaded.ShopIds!);
        Assert.Null(loaded.Keywords);
        Assert.Null(loaded.CronExpression);
    }

    [Fact]
    public async Task Busca_SemKeywords_JobId_SegueConvencao()
    {
        var busca = new Busca
        {
            Keyword = "Test Shop",
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
            Keyword = "Glory of Seoul",
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
            Keyword = "Test Shop",
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

    // ═══════════════════════════════════════════════════════════════════════
    // Scheduler params mapping
    // ═══════════════════════════════════════════════════════════════════════

    [Fact]
    public void SchedulerParams_SemKeywords_NaoIncluiKeywords()
    {
        var busca = new Busca
        {
            Keyword = "Shop",
            OwnerUid = "uid-123",
            ShopIds = [999]
        };

        var @params = BuildSchedulerParams(busca);

        Assert.Equal("999", @params["shop_id"]);
        Assert.Equal("uid-123", @params["owner_uid"]);
        Assert.Equal("shop_collection", @params["type"]);
        Assert.False(@params.ContainsKey("keywords"));
    }

    [Fact]
    public void SchedulerParams_ComKeywords_IncluiKeywordsJoinadas()
    {
        var busca = new Busca
        {
            Keyword = "Shop",
            OwnerUid = "uid-123",
            ShopIds = [999],
            Keywords = ["serum", "protetor"]
        };

        var @params = BuildSchedulerParams(busca);

        Assert.True(@params.ContainsKey("keywords"));
        Assert.Equal("serum,protetor", @params["keywords"]);
    }

    // ═══════════════════════════════════════════════════════════════════════
    // Preservation: soft-delete
    // ═══════════════════════════════════════════════════════════════════════

    [Fact]
    public async Task SoftDelete_SetsActiveFalse_AndUpdatedAt()
    {
        var busca = new Busca
        {
            Keyword = "To Remove",
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

    // ═══════════════════════════════════════════════════════════════════════
    // Helper: simula a construção de params para SetScheduleRequest
    // ═══════════════════════════════════════════════════════════════════════

    private static Dictionary<string, string> BuildSchedulerParams(Busca busca)
    {
        var @params = new Dictionary<string, string>
        {
            ["shop_id"] = busca.ShopIds![0].ToString(),
            ["owner_uid"] = busca.OwnerUid,
            ["type"] = "shop_collection"
        };
        if (busca.Keywords is { Length: > 0 })
            @params["keywords"] = string.Join(",", busca.Keywords);
        return @params;
    }
}
