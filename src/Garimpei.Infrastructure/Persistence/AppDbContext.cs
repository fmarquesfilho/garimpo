using Garimpei.Domain.Entities;
using Garimpei.Domain.Interfaces;
using Microsoft.EntityFrameworkCore;

namespace Garimpei.Infrastructure.Persistence;

public class AppDbContext : DbContext
{
    private readonly ITenantContext _tenantContext;

    public AppDbContext(DbContextOptions<AppDbContext> options, ITenantContext tenantContext)
        : base(options)
    {
        _tenantContext = tenantContext;
    }

    // Design-time constructor (migrations — no tenant context needed)
    public AppDbContext(DbContextOptions<AppDbContext> options)
        : base(options)
    {
        _tenantContext = new NullTenantContext();
    }

    public DbSet<Product> Products => Set<Product>();
    public DbSet<Busca> Buscas => Set<Busca>();
    public DbSet<Tenant> Tenants => Set<Tenant>();
    public DbSet<Favorito> Favoritos => Set<Favorito>();
    public DbSet<Destino> Destinos => Set<Destino>();
    public DbSet<Template> Templates => Set<Template>();
    public DbSet<Publicacao> Publicacoes => Set<Publicacao>();
    public DbSet<TenantConfig> TenantConfigs => Set<TenantConfig>();
    public DbSet<CouponAlertRule> CouponAlertRules => Set<CouponAlertRule>();
    public DbSet<CouponAlertHistory> CouponAlertHistories => Set<CouponAlertHistory>();

    protected override void OnModelCreating(ModelBuilder modelBuilder)
    {
        modelBuilder.Entity<Product>(entity =>
        {
            entity.HasKey(e => e.Id);
            entity.HasIndex(e => e.OwnerUid);
            entity.HasIndex(e => new { e.ItemId, e.ShopId }).IsUnique();
            entity.HasQueryFilter(e => e.OwnerUid == _tenantContext.OwnerUid);
        });

        modelBuilder.Entity<Busca>(entity =>
        {
            entity.HasKey(e => e.Id);
            entity.HasIndex(e => e.OwnerUid);
            entity.HasQueryFilter(e => e.OwnerUid == _tenantContext.OwnerUid);
        });

        modelBuilder.Entity<Tenant>(entity =>
        {
            entity.HasKey(e => e.Id);
            entity.HasIndex(e => e.OwnerUid).IsUnique();
        });

        modelBuilder.Entity<Favorito>(entity =>
        {
            entity.HasKey(e => e.Id);
            entity.HasIndex(e => e.OwnerUid);
            entity.HasIndex(e => new { e.OwnerUid, e.ProdutoId }).IsUnique();
            entity.HasQueryFilter(e => e.OwnerUid == _tenantContext.OwnerUid);
        });

        modelBuilder.Entity<Destino>(entity =>
        {
            entity.HasKey(e => e.Id);
            entity.HasIndex(e => e.OwnerUid);
            entity.HasQueryFilter(e => e.OwnerUid == _tenantContext.OwnerUid);
        });

        modelBuilder.Entity<Template>(entity =>
        {
            entity.HasKey(e => e.Id);
            entity.HasIndex(e => e.OwnerUid);
            entity.HasQueryFilter(e => e.OwnerUid == _tenantContext.OwnerUid);
        });

        modelBuilder.Entity<Publicacao>(entity =>
        {
            entity.HasKey(e => e.Id);
            entity.HasIndex(e => e.OwnerUid);
            entity.HasIndex(e => e.Status);
            entity.HasQueryFilter(e => e.OwnerUid == _tenantContext.OwnerUid);
        });

        modelBuilder.Entity<TenantConfig>(entity =>
        {
            entity.HasKey(e => e.Id);
            entity.HasIndex(e => e.OwnerUid).IsUnique();
            entity.HasQueryFilter(e => e.OwnerUid == _tenantContext.OwnerUid);
        });

        modelBuilder.Entity<CouponAlertRule>(entity =>
        {
            entity.HasKey(e => e.Id);
            entity.HasIndex(e => e.OwnerUid);
            entity.HasQueryFilter(e => e.OwnerUid == _tenantContext.OwnerUid);
        });

        modelBuilder.Entity<CouponAlertHistory>(entity =>
        {
            entity.HasKey(e => e.Id);
            entity.HasIndex(e => e.OwnerUid);
            entity.HasIndex(e => new { e.CouponId, e.AlertRuleId, e.AlertedAt });
            entity.HasQueryFilter(e => e.OwnerUid == _tenantContext.OwnerUid);
        });
    }

    /// <summary>
    /// Automatically sets OwnerUid on new IOwnedEntity entries before saving.
    /// </summary>
    public override Task<int> SaveChangesAsync(CancellationToken cancellationToken = default)
    {
        SetOwnerUidOnNewEntities();
        return base.SaveChangesAsync(cancellationToken);
    }

    public override int SaveChanges()
    {
        SetOwnerUidOnNewEntities();
        return base.SaveChanges();
    }

    private void SetOwnerUidOnNewEntities()
    {
        if (!_tenantContext.IsResolved) return;

        var entries = ChangeTracker.Entries<IOwnedEntity>()
            .Where(e => e.State == EntityState.Added && string.IsNullOrEmpty(e.Entity.OwnerUid));

        foreach (var entry in entries)
        {
            entry.Entity.OwnerUid = _tenantContext.OwnerUid;
        }
    }

    /// <summary>
    /// Null object for design-time/migration scenarios.
    /// </summary>
    private sealed class NullTenantContext : ITenantContext
    {
        public string OwnerUid => string.Empty;
        public bool IsResolved => false;
    }
}
