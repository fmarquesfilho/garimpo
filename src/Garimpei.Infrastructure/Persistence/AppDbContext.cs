using Garimpei.Domain.Entities;
using Microsoft.EntityFrameworkCore;

namespace Garimpei.Infrastructure.Persistence;

public class AppDbContext(DbContextOptions<AppDbContext> options) : DbContext(options)
{
    public DbSet<Product> Products => Set<Product>();
    public DbSet<Busca> Buscas => Set<Busca>();
    public DbSet<Tenant> Tenants => Set<Tenant>();

    protected override void OnModelCreating(ModelBuilder modelBuilder)
    {
        modelBuilder.Entity<Product>(entity =>
        {
            entity.HasKey(e => e.Id);
            entity.HasIndex(e => e.OwnerUid);
            entity.HasIndex(e => new { e.ItemId, e.ShopId }).IsUnique();
            entity.HasQueryFilter(e => EF.Property<string>(e, "OwnerUid") != null);
        });

        modelBuilder.Entity<Busca>(entity =>
        {
            entity.HasKey(e => e.Id);
            entity.HasIndex(e => e.OwnerUid);
        });

        modelBuilder.Entity<Tenant>(entity =>
        {
            entity.HasKey(e => e.Id);
            entity.HasIndex(e => e.OwnerUid).IsUnique();
        });
    }
}
