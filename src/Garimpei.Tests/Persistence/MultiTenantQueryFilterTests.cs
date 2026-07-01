using Garimpei.Domain.Entities;
using Garimpei.Domain.Interfaces;
using Garimpei.Infrastructure.Persistence;
using Garimpei.Infrastructure.Tenancy;
using Microsoft.EntityFrameworkCore;
using Xunit;

namespace Garimpei.Tests.Persistence;

public class MultiTenantQueryFilterTests : IDisposable
{
    private readonly TenantContext _tenantContext;
    private readonly AppDbContext _db;

    public MultiTenantQueryFilterTests()
    {
        _tenantContext = new TenantContext();

        var options = new DbContextOptionsBuilder<AppDbContext>()
            .UseInMemoryDatabase(databaseName: Guid.NewGuid().ToString())
            .Options;

        _db = new AppDbContext(options, _tenantContext);
        _db.Database.EnsureCreated();
    }

    public void Dispose()
    {
        _db.Dispose();
    }

    [Fact]
    public void Product_Implements_IOwnedEntity()
    {
        IOwnedEntity entity = new Product
        {
            ItemId = 1,
            ShopId = 1,
            Name = "Test",
            Price = 10m
        };

        entity.OwnerUid = "owner-xyz";

        Assert.Equal("owner-xyz", entity.OwnerUid);
    }

    [Fact]
    public void Busca_Implements_IOwnedEntity()
    {
        IOwnedEntity entity = new Busca
        {
            Keyword = "test",
            OwnerUid = ""
        };

        entity.OwnerUid = "owner-abc";

        Assert.Equal("owner-abc", entity.OwnerUid);
    }

    [Fact]
    public void Tenant_Has_Required_Properties()
    {
        var tenant = new Tenant
        {
            OwnerUid = "uid-123",
            Name = "My Store"
        };

        Assert.Equal("uid-123", tenant.OwnerUid);
        Assert.Equal("My Store", tenant.Name);
        Assert.NotEqual(Guid.Empty, tenant.Id);
        Assert.True(tenant.Active);
    }

    [Fact]
    public async Task Products_Are_Filtered_By_OwnerUid()
    {
        // Seed products directly bypassing the query filter
        _db.Products.Add(new Product { ItemId = 1, ShopId = 1, Name = "P1", Price = 10, OwnerUid = "tenant-a" });
        _db.Products.Add(new Product { ItemId = 2, ShopId = 2, Name = "P2", Price = 20, OwnerUid = "tenant-b" });
        await _db.SaveChangesAsync();

        // Set tenant context to tenant-a
        _tenantContext.Set("tenant-a");

        var products = await _db.Products.ToListAsync();

        Assert.Single(products);
        Assert.Equal("P1", products[0].Name);
    }

    [Fact]
    public async Task Buscas_Are_Filtered_By_OwnerUid()
    {
        _db.Buscas.Add(new Busca { Keyword = "shoes", OwnerUid = "tenant-a" });
        _db.Buscas.Add(new Busca { Keyword = "bags", OwnerUid = "tenant-b" });
        await _db.SaveChangesAsync();

        _tenantContext.Set("tenant-a");

        var buscas = await _db.Buscas.ToListAsync();

        Assert.Single(buscas);
        Assert.Equal("shoes", buscas[0].Keyword);
    }

    [Fact]
    public async Task SaveChangesAsync_AutoSets_OwnerUid_On_New_IOwnedEntity()
    {
        _tenantContext.Set("auto-owner");

        var product = new Product { ItemId = 99, ShopId = 99, Name = "Auto", Price = 5m };
        _db.Products.Add(product);
        await _db.SaveChangesAsync();

        Assert.Equal("auto-owner", product.OwnerUid);
    }

    [Fact]
    public async Task TenantA_Cannot_See_TenantB_Products()
    {
        // Seed as tenant-a
        _tenantContext.Set("tenant-a");
        _db.Products.Add(new Product { ItemId = 10, ShopId = 10, Name = "A-Product", Price = 100m });
        await _db.SaveChangesAsync();

        // Seed as tenant-b (need to bypass filter for seeding)
        _db.Products.Add(new Product { ItemId = 20, ShopId = 20, Name = "B-Product", Price = 200m, OwnerUid = "tenant-b" });
        await _db.SaveChangesAsync();

        // Query as tenant-a — should only see A's product
        var productsA = await _db.Products.ToListAsync();
        Assert.Single(productsA);
        Assert.Equal("A-Product", productsA[0].Name);

        // Switch to tenant-b
        _tenantContext.Set("tenant-b");

        var productsB = await _db.Products.ToListAsync();
        Assert.Single(productsB);
        Assert.Equal("B-Product", productsB[0].Name);
    }
}
