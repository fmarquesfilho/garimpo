using Garimpei.Domain.Interfaces;
using Garimpei.Infrastructure.Persistence;
using Garimpei.Infrastructure.Tenancy;
using Collector.V1;
using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;

namespace Garimpei.Infrastructure;

public static class DependencyInjection
{
    public static IServiceCollection AddInfrastructure(
        this IServiceCollection services,
        IConfiguration configuration)
    {
        // Tenancy (scoped — one per request)
        services.AddScoped<TenantContext>();
        services.AddScoped<ITenantContext>(sp => sp.GetRequiredService<TenantContext>());

        // Persistence
        services.AddDbContext<AppDbContext>((sp, options) =>
            options.UseNpgsql(
                configuration.GetConnectionString("PostgreSQL"),
                npgsql => npgsql.MigrationsAssembly(typeof(AppDbContext).Assembly.FullName)));

        // gRPC clients (sidecars via localhost)
        var collectorAddr = configuration["Grpc:CollectorAddress"] ?? "http://localhost:50051";
        services.AddGrpcClient<CollectorService.CollectorServiceClient>(o =>
        {
            o.Address = new Uri(collectorAddr);
        });

        return services;
    }
}
