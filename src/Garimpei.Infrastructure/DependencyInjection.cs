using Garimpei.Domain;
using Garimpei.Domain.Interfaces;
using Garimpei.Infrastructure.Persistence;
using Garimpei.Infrastructure.Sources;
using Garimpei.Infrastructure.Tenancy;
using Collector.V1;
using Publisher.V1;
using Scheduler.V1;
using Cache.V1;
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

        // Persistence — PostgreSQL via EF Core
        services.AddDbContext<AppDbContext>((sp, options) =>
            options.UseNpgsql(
                configuration.GetConnectionString("PostgreSQL"),
                npgsql => npgsql.MigrationsAssembly(typeof(AppDbContext).Assembly.FullName)));

        // ─── gRPC clients (sidecars via localhost) ──────────────────────────
        // Unified collector handles all marketplaces on a single port (ADR-0018).
        var collectorAddr = configuration["Grpc:CollectorAddress"] ?? "http://localhost:50051";
        services.AddGrpcClient<CollectorService.CollectorServiceClient>(o =>
        {
            o.Address = new Uri(collectorAddr);
        });

        var publisherAddr = configuration["Grpc:PublisherAddress"] ?? "http://localhost:50052";
        services.AddGrpcClient<PublisherService.PublisherServiceClient>(o =>
        {
            o.Address = new Uri(publisherAddr);
        });

        var schedulerAddr = configuration["Grpc:SchedulerAddress"] ?? "http://localhost:50054";
        services.AddGrpcClient<SchedulerService.SchedulerServiceClient>(o =>
        {
            o.Address = new Uri(schedulerAddr);
        });

        // Cache sidecar (L2) — localhost:50055 within same Cloud Run pod
        var cacheAddr = configuration["Grpc:CacheAddress"] ?? "http://localhost:50055";
        services.AddGrpcClient<CacheService.CacheServiceClient>(o =>
        {
            o.Address = new Uri(cacheAddr);
        });

        // Circuit breaker for cache sidecar
        services.AddSingleton<CacheCircuitBreaker>();

        // ─── Keyed IProductSource (Strategy Pattern via .NET Keyed Services) ─
        // Adicionar um novo marketplace = adicionar uma linha aqui. Zero mudanças nos endpoints.
        services.AddKeyedScoped<IProductSource, ShopeeProductSource>(Marketplaces.Shopee);
        services.AddKeyedScoped<IProductSource, AmazonProductSource>(Marketplaces.Amazon);

        return services;
    }
}

