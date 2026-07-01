using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Design;

namespace Garimpei.Infrastructure.Persistence;

/// <summary>
/// Factory para uso com dotnet ef migrations (design-time).
/// </summary>
public class DesignTimeDbContextFactory : IDesignTimeDbContextFactory<AppDbContext>
{
    public AppDbContext CreateDbContext(string[] args)
    {
        var options = new DbContextOptionsBuilder<AppDbContext>()
            .UseNpgsql("Host=localhost;Port=5432;Database=garimpei;Username=garimpei;Password=garimpei_dev")
            .Options;

        return new AppDbContext(options);
    }
}
