using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Design;

namespace Garimpei.Infrastructure.Persistence;

/// <summary>
/// Factory para uso com dotnet ef migrations (design-time).
/// Usa a env var CONNECTION_STRING se disponível, senão fallback para localhost.
/// </summary>
public class DesignTimeDbContextFactory : IDesignTimeDbContextFactory<AppDbContext>
{
    public AppDbContext CreateDbContext(string[] args)
    {
        var connectionString = Environment.GetEnvironmentVariable("CONNECTION_STRING")
            ?? "Host=localhost;Port=5432;Database=garimpei;Username=garimpei;Password=garimpei_dev";

        var options = new DbContextOptionsBuilder<AppDbContext>()
            .UseNpgsql(connectionString)
            .Options;

        return new AppDbContext(options);
    }
}
