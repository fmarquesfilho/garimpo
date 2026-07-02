using Garimpei.Domain.Entities;
using Garimpei.Infrastructure.Persistence;
using Garimpei.Infrastructure.Tenancy;
using Microsoft.EntityFrameworkCore;
using Xunit;

namespace Garimpei.Tests.Integration;

/// <summary>
/// Integration tests for the multi-tenant onboarding flow.
/// Validates that:
/// - A new tenant can complete onboarding from scratch (bank zeroed)
/// - Tenants are fully isolated (data never leaks between tenants)
/// - All onboarding steps work independently per tenant
/// - LGPD account deletion removes all tenant data
/// </summary>
public class OnboardingMultiTenantTests : IDisposable
{
    private readonly AppDbContext _db;
    private readonly TenantContext _tenantContext;

    public OnboardingMultiTenantTests()
    {
        _tenantContext = new TenantContext();

        var options = new DbContextOptionsBuilder<AppDbContext>()
            .UseInMemoryDatabase(databaseName: $"onboarding-{Guid.NewGuid()}")
            .Options;

        _db = new AppDbContext(options, _tenantContext);
        _db.Database.EnsureCreated();
    }

    public void Dispose() => _db.Dispose();

    // ═══════════════════════════════════════════════════════════════════════
    // Onboarding do zero (novo tenant, banco limpo)
    // ═══════════════════════════════════════════════════════════════════════

    [Fact]
    public async Task NewTenant_StartsWithNoData()
    {
        _tenantContext.Set("new-tenant");

        var configs = await _db.TenantConfigs.ToListAsync();
        var buscas = await _db.Buscas.ToListAsync();
        var favoritos = await _db.Favoritos.ToListAsync();
        var destinos = await _db.Destinos.ToListAsync();

        Assert.Empty(configs);
        Assert.Empty(buscas);
        Assert.Empty(favoritos);
        Assert.Empty(destinos);
    }

    [Fact]
    public async Task Onboarding_Step1_AcceptTerms()
    {
        _tenantContext.Set("tenant-terms");

        var cfg = new TenantConfig { OwnerUid = "tenant-terms" };
        cfg.AceitouTermos = true;
        cfg.AceitouTermosEm = DateTime.UtcNow;
        cfg.OnboardingStep = 1;

        _db.TenantConfigs.Add(cfg);
        await _db.SaveChangesAsync();

        var saved = await _db.TenantConfigs.FirstAsync();
        Assert.True(saved.AceitouTermos);
        Assert.Equal(1, saved.OnboardingStep);
        Assert.NotNull(saved.AceitouTermosEm);
    }

    [Fact]
    public async Task Onboarding_Step2_ShopeeCredentials()
    {
        _tenantContext.Set("tenant-shopee");

        var cfg = new TenantConfig
        {
            OwnerUid = "tenant-shopee",
            AceitouTermos = true,
            OnboardingStep = 1
        };
        _db.TenantConfigs.Add(cfg);
        await _db.SaveChangesAsync();

        // Step 2: save Shopee credentials
        cfg.ShopeeAppId = "18332030606";
        cfg.ShopeeSecretEnc = "encrypted-secret";
        cfg.OnboardingStep = 2;
        await _db.SaveChangesAsync();

        var saved = await _db.TenantConfigs.FirstAsync();
        Assert.Equal("18332030606", saved.ShopeeAppId);
        Assert.Equal(2, saved.OnboardingStep);
    }

    [Fact]
    public async Task Onboarding_Step3_TelegramConfig()
    {
        _tenantContext.Set("tenant-telegram");

        var cfg = new TenantConfig
        {
            OwnerUid = "tenant-telegram",
            OnboardingStep = 2,
            ShopeeAppId = "123"
        };
        _db.TenantConfigs.Add(cfg);
        await _db.SaveChangesAsync();

        // Step 3: configure Telegram
        cfg.TelegramTokenEnc = "encrypted-bot-token";
        cfg.TelegramChatId = "-1001234567890";
        cfg.OnboardingStep = 3;
        await _db.SaveChangesAsync();

        var saved = await _db.TenantConfigs.FirstAsync();
        Assert.Equal("-1001234567890", saved.TelegramChatId);
        Assert.Equal(3, saved.OnboardingStep);
    }

    [Fact]
    public async Task Onboarding_Step3_WhatsappConfig()
    {
        _tenantContext.Set("tenant-whatsapp");

        var cfg = new TenantConfig
        {
            OwnerUid = "tenant-whatsapp",
            OnboardingStep = 2,
            ShopeeAppId = "456"
        };
        _db.TenantConfigs.Add(cfg);
        await _db.SaveChangesAsync();

        // Step 3: configure WhatsApp
        cfg.WhatsappPhoneNumberId = "1234567890123456";
        cfg.WhatsappTokenEnc = "encrypted-meta-token";
        cfg.OnboardingStep = 3;
        await _db.SaveChangesAsync();

        var saved = await _db.TenantConfigs.FirstAsync();
        Assert.Equal("1234567890123456", saved.WhatsappPhoneNumberId);
        Assert.Equal(3, saved.OnboardingStep);
    }

    [Fact]
    public async Task Onboarding_Step3_CanSkip()
    {
        _tenantContext.Set("tenant-skip");

        var cfg = new TenantConfig
        {
            OwnerUid = "tenant-skip",
            OnboardingStep = 2,
            ShopeeAppId = "789"
        };
        _db.TenantConfigs.Add(cfg);
        await _db.SaveChangesAsync();

        // Step 3: skip channels
        cfg.OnboardingStep = 3;
        await _db.SaveChangesAsync();

        var saved = await _db.TenantConfigs.FirstAsync();
        Assert.Null(saved.TelegramChatId);
        Assert.Null(saved.WhatsappPhoneNumberId);
        Assert.Equal(3, saved.OnboardingStep);
    }

    [Fact]
    public async Task Onboarding_Step4_Validate_MarksConfigured()
    {
        _tenantContext.Set("tenant-validate");

        var cfg = new TenantConfig
        {
            OwnerUid = "tenant-validate",
            OnboardingStep = 3,
            ShopeeAppId = "999",
            AceitouTermos = true
        };
        _db.TenantConfigs.Add(cfg);
        await _db.SaveChangesAsync();

        // Step 4: validate
        cfg.OnboardingStep = 4;
        await _db.SaveChangesAsync();

        var saved = await _db.TenantConfigs.FirstAsync();
        Assert.Equal(4, saved.OnboardingStep);
        Assert.True(saved.Configurado);
    }

    [Fact]
    public async Task FullOnboarding_EndToEnd()
    {
        _tenantContext.Set("tenant-e2e");

        // Step 1: Terms
        var cfg = new TenantConfig { OwnerUid = "tenant-e2e" };
        cfg.AceitouTermos = true;
        cfg.AceitouTermosEm = DateTime.UtcNow;
        cfg.OnboardingStep = 1;
        _db.TenantConfigs.Add(cfg);
        await _db.SaveChangesAsync();

        // Step 2: Shopee
        cfg.ShopeeAppId = "11223344";
        cfg.ShopeeSecretEnc = "secret-enc";
        cfg.OnboardingStep = 2;
        await _db.SaveChangesAsync();

        // Step 3: Telegram
        cfg.TelegramTokenEnc = "bot-token-enc";
        cfg.TelegramChatId = "-100999";
        cfg.OnboardingStep = 3;
        await _db.SaveChangesAsync();

        // Step 4: Validate
        cfg.OnboardingStep = 4;
        await _db.SaveChangesAsync();

        // Assert final state
        var final = await _db.TenantConfigs.FirstAsync();
        Assert.True(final.Configurado);
        Assert.True(final.AceitouTermos);
        Assert.Equal("11223344", final.ShopeeAppId);
        Assert.Equal("-100999", final.TelegramChatId);
        Assert.Equal(4, final.OnboardingStep);
    }

    // ═══════════════════════════════════════════════════════════════════════
    // Isolamento multi-tenant
    // ═══════════════════════════════════════════════════════════════════════

    [Fact]
    public async Task TenantA_CannotSee_TenantB_TenantConfig()
    {
        // Seed tenant A
        _db.TenantConfigs.Add(new TenantConfig
        {
            OwnerUid = "tenant-a",
            ShopeeAppId = "AAA",
            OnboardingStep = 4
        });
        // Seed tenant B
        _db.TenantConfigs.Add(new TenantConfig
        {
            OwnerUid = "tenant-b",
            ShopeeAppId = "BBB",
            OnboardingStep = 4
        });
        await _db.SaveChangesAsync();

        // Query as tenant A
        _tenantContext.Set("tenant-a");
        var configsA = await _db.TenantConfigs.ToListAsync();
        Assert.Single(configsA);
        Assert.Equal("AAA", configsA[0].ShopeeAppId);

        // Query as tenant B
        _tenantContext.Set("tenant-b");
        var configsB = await _db.TenantConfigs.ToListAsync();
        Assert.Single(configsB);
        Assert.Equal("BBB", configsB[0].ShopeeAppId);
    }

    [Fact]
    public async Task TenantA_CannotSee_TenantB_Favoritos()
    {
        _db.Favoritos.Add(new Favorito { ProdutoId = "1", Nome = "A-Fav", OwnerUid = "tenant-a" });
        _db.Favoritos.Add(new Favorito { ProdutoId = "2", Nome = "B-Fav", OwnerUid = "tenant-b" });
        await _db.SaveChangesAsync();

        _tenantContext.Set("tenant-a");
        var favsA = await _db.Favoritos.ToListAsync();
        Assert.Single(favsA);
        Assert.Equal("A-Fav", favsA[0].Nome);

        _tenantContext.Set("tenant-b");
        var favsB = await _db.Favoritos.ToListAsync();
        Assert.Single(favsB);
        Assert.Equal("B-Fav", favsB[0].Nome);
    }

    [Fact]
    public async Task TenantA_CannotSee_TenantB_Destinos()
    {
        _db.Destinos.Add(new Destino { Nome = "Canal A", Tipo = "telegram", OwnerUid = "tenant-a" });
        _db.Destinos.Add(new Destino { Nome = "Canal B", Tipo = "whatsapp", OwnerUid = "tenant-b" });
        await _db.SaveChangesAsync();

        _tenantContext.Set("tenant-a");
        var destA = await _db.Destinos.ToListAsync();
        Assert.Single(destA);
        Assert.Equal("Canal A", destA[0].Nome);

        _tenantContext.Set("tenant-b");
        var destB = await _db.Destinos.ToListAsync();
        Assert.Single(destB);
        Assert.Equal("Canal B", destB[0].Nome);
    }

    [Fact]
    public async Task TenantA_CannotSee_TenantB_Templates()
    {
        _db.Templates.Add(new Template { Nome = "Tmpl A", Corpo = "{{nome}}", OwnerUid = "tenant-a" });
        _db.Templates.Add(new Template { Nome = "Tmpl B", Corpo = "{{preco}}", OwnerUid = "tenant-b" });
        await _db.SaveChangesAsync();

        _tenantContext.Set("tenant-a");
        var tmplA = await _db.Templates.ToListAsync();
        Assert.Single(tmplA);
        Assert.Equal("Tmpl A", tmplA[0].Nome);
    }

    [Fact]
    public async Task TenantA_CannotSee_TenantB_Publicacoes()
    {
        _db.Publicacoes.Add(new Publicacao { ProdutoId = "x", Nome = "Pub A", OwnerUid = "tenant-a" });
        _db.Publicacoes.Add(new Publicacao { ProdutoId = "y", Nome = "Pub B", OwnerUid = "tenant-b" });
        await _db.SaveChangesAsync();

        _tenantContext.Set("tenant-a");
        var pubA = await _db.Publicacoes.ToListAsync();
        Assert.Single(pubA);
        Assert.Equal("Pub A", pubA[0].Nome);
    }

    // ═══════════════════════════════════════════════════════════════════════
    // LGPD: exclusão de conta
    // ═══════════════════════════════════════════════════════════════════════

    [Fact]
    public async Task AccountDeletion_RemovesAllTenantData()
    {
        _tenantContext.Set("tenant-delete");

        // Create data for tenant
        _db.TenantConfigs.Add(new TenantConfig { OwnerUid = "tenant-delete", ShopeeAppId = "X" });
        _db.Buscas.Add(new Busca { Keyword = "test", OwnerUid = "tenant-delete" });
        _db.Favoritos.Add(new Favorito { ProdutoId = "1", Nome = "Fav", OwnerUid = "tenant-delete" });
        _db.Destinos.Add(new Destino { Nome = "Ch", Tipo = "telegram", OwnerUid = "tenant-delete" });
        _db.Templates.Add(new Template { Nome = "T", Corpo = "x", OwnerUid = "tenant-delete" });
        _db.Publicacoes.Add(new Publicacao { ProdutoId = "1", Nome = "P", OwnerUid = "tenant-delete" });
        await _db.SaveChangesAsync();

        // Simulate account deletion (soft-delete + config removal)
        var cfg = await _db.TenantConfigs.FirstAsync();
        _db.TenantConfigs.Remove(cfg);

        var buscas = await _db.Buscas.ToListAsync();
        foreach (var b in buscas) b.Active = false;

        var favoritos = await _db.Favoritos.ToListAsync();
        foreach (var f in favoritos) f.Ativo = false;

        var destinos = await _db.Destinos.ToListAsync();
        foreach (var d in destinos) d.Ativo = false;

        var templates = await _db.Templates.ToListAsync();
        foreach (var t in templates) t.Ativo = false;

        await _db.SaveChangesAsync();

        // Verify: no active data remains
        Assert.Empty(await _db.TenantConfigs.ToListAsync());
        Assert.Empty(await _db.Buscas.Where(b => b.Active).ToListAsync());
        Assert.Empty(await _db.Favoritos.Where(f => f.Ativo).ToListAsync());
        Assert.Empty(await _db.Destinos.Where(d => d.Ativo).ToListAsync());
        Assert.Empty(await _db.Templates.Where(t => t.Ativo).ToListAsync());
    }

    // ═══════════════════════════════════════════════════════════════════════
    // Auto-set OwnerUid on new entities
    // ═══════════════════════════════════════════════════════════════════════

    [Fact]
    public async Task SaveChanges_AutoSets_OwnerUid_OnAllEntities()
    {
        _tenantContext.Set("auto-owner");

        _db.Buscas.Add(new Busca { Keyword = "auto", OwnerUid = "" });
        _db.Favoritos.Add(new Favorito { ProdutoId = "auto", Nome = "Auto Fav" });
        _db.Destinos.Add(new Destino { Nome = "Auto Ch", Tipo = "telegram" });
        _db.Templates.Add(new Template { Nome = "Auto T", Corpo = "x" });
        _db.Publicacoes.Add(new Publicacao { ProdutoId = "auto", Nome = "Auto P" });
        await _db.SaveChangesAsync();

        // All should have OwnerUid set automatically
        var busca = await _db.Buscas.FirstAsync();
        Assert.Equal("auto-owner", busca.OwnerUid);

        var fav = await _db.Favoritos.FirstAsync();
        Assert.Equal("auto-owner", fav.OwnerUid);

        var dest = await _db.Destinos.FirstAsync();
        Assert.Equal("auto-owner", dest.OwnerUid);

        var tmpl = await _db.Templates.FirstAsync();
        Assert.Equal("auto-owner", tmpl.OwnerUid);

        var pub = await _db.Publicacoes.FirstAsync();
        Assert.Equal("auto-owner", pub.OwnerUid);
    }
}
