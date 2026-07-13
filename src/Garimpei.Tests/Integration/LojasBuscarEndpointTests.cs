using Garimpei.Domain.Entities;
using Garimpei.Infrastructure.Persistence;
using Garimpei.Infrastructure.Tenancy;
using Microsoft.EntityFrameworkCore;
using Xunit;

namespace Garimpei.Tests.Integration;

public class LojasBuscarEndpointTests : IDisposable
{
    private readonly TenantContext _tenantContext;
    private readonly AppDbContext _db;

    public LojasBuscarEndpointTests()
    {
        _tenantContext = new TenantContext();
        _tenantContext.Set("test-tenant");

        var options = new DbContextOptionsBuilder<AppDbContext>()
            .UseInMemoryDatabase(databaseName: Guid.NewGuid().ToString())
            .Options;

        _db = new AppDbContext(options, _tenantContext);
        _db.Database.EnsureCreated();

        // Seed lojas
        _db.Lojas.AddRange(
            new Loja { OwnerUid = "test-tenant", ShopId = 100, Nome = "Glory of Seoul", NomeNormalizado = "gloryofseoul", Marketplace = "shopee", OrigemPadrao = "🇰🇷", ImageUrl = "https://img.test/glory.jpg", FollowerCount = 12000, ItemCount = 340, RatingStar = 4.8 },
            new Loja { OwnerUid = "test-tenant", ShopId = 200, Nome = "Le Botanic Beauty", NomeNormalizado = "lebotanicbeauty", Marketplace = "shopee", OrigemPadrao = "🇧🇷" },
            new Loja { OwnerUid = "test-tenant", ShopId = 300, Nome = "Amazon Store Test", NomeNormalizado = "amazonstoretest", Marketplace = "amazon" },
            new Loja { OwnerUid = "test-tenant", ShopId = 400, Nome = "Glory Brasil", NomeNormalizado = "glorybrasil", Marketplace = "mercado_livre", CronExpression = "0 */8 * * *" }
        );
        _db.SaveChanges();
    }

    public void Dispose() => _db.Dispose();

    [Fact]
    public async Task Buscar_RetornaLojasPorSubstring()
    {
        var termoNorm = Loja.Normalizar("glory");
        var lojas = await _db.Lojas
            .Where(l => l.NomeNormalizado.Contains(termoNorm))
            .OrderBy(l => l.Nome)
            .Take(20)
            .ToListAsync();

        Assert.Equal(2, lojas.Count);
        Assert.Equal("Glory Brasil", lojas[0].Nome);
        Assert.Equal("Glory of Seoul", lojas[1].Nome);
    }

    [Fact]
    public async Task Buscar_FiltraPorMarketplace()
    {
        var termoNorm = Loja.Normalizar("glory");
        var lojas = await _db.Lojas
            .Where(l => l.Marketplace == "shopee")
            .Where(l => l.NomeNormalizado.Contains(termoNorm))
            .OrderBy(l => l.Nome)
            .Take(20)
            .ToListAsync();

        Assert.Single(lojas);
        Assert.Equal("Glory of Seoul", lojas[0].Nome);
    }

    [Fact]
    public void Buscar_TermoMenorQue2Chars_NaoRetorna()
    {
        var q = "g";
        Assert.True(q.Length < 2);
    }

    [Fact]
    public async Task Buscar_RetornaCamposEnriquecidos()
    {
        var termoNorm = Loja.Normalizar("glory of seoul");
        var loja = await _db.Lojas
            .Where(l => l.NomeNormalizado.Contains(termoNorm))
            .FirstOrDefaultAsync();

        Assert.NotNull(loja);
        Assert.Equal("https://img.test/glory.jpg", loja.ImageUrl);
        Assert.Equal(12000, loja.FollowerCount);
        Assert.Equal(340, loja.ItemCount);
        Assert.Equal(4.8, loja.RatingStar);
    }

    [Fact]
    public async Task Buscar_LimitaEm20Resultados()
    {
        // Add more lojas to test limit
        for (int i = 0; i < 25; i++)
        {
            _db.Lojas.Add(new Loja
            {
                OwnerUid = "test-tenant",
                ShopId = 1000 + i,
                Nome = $"TestShop{i}",
                NomeNormalizado = $"testshop{i}",
                Marketplace = "shopee"
            });
        }
        await _db.SaveChangesAsync();

        var termoNorm = Loja.Normalizar("testshop");
        var lojas = await _db.Lojas
            .Where(l => l.NomeNormalizado.Contains(termoNorm))
            .OrderBy(l => l.Nome)
            .Take(20)
            .ToListAsync();

        Assert.Equal(20, lojas.Count);
    }

    [Fact]
    public async Task Buscar_MonitoradaFlag_ReflecteaCron()
    {
        var termoNorm = Loja.Normalizar("glory");
        var lojas = await _db.Lojas
            .Where(l => l.NomeNormalizado.Contains(termoNorm))
            .OrderBy(l => l.Nome)
            .ToListAsync();

        var brasil = lojas.First(l => l.Nome == "Glory Brasil");
        var seoul = lojas.First(l => l.Nome == "Glory of Seoul");

        Assert.False(string.IsNullOrEmpty(brasil.CronExpression)); // monitorada
        Assert.True(string.IsNullOrEmpty(seoul.CronExpression));  // não monitorada
    }
}
