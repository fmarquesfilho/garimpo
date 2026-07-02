using Garimpei.Domain;
using Garimpei.Domain.Interfaces;
using Garimpei.Infrastructure.Persistence;
using Garimpei.Infrastructure.Sources;
using Garimpei.Infrastructure.Tenancy;
using Collector.V1;
using Publisher.V1;
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

        // ─── gRPC clients (sidecars via localhost) ──────────────────────────
        var collectorShopeeAddr = configuration["Grpc:CollectorAddress"] ?? "http://localhost:50051";
        services.AddGrpcClient<CollectorService.CollectorServiceClient>("shopee-collector", o =>
        {
            o.Address = new Uri(collectorShopeeAddr);
        });

        var collectorAmazonAddr = configuration["Grpc:CollectorAmazonAddress"] ?? "http://localhost:50055";
        services.AddGrpcClient<CollectorService.CollectorServiceClient>("amazon-collector", o =>
        {
            o.Address = new Uri(collectorAmazonAddr);
        });

        // Legacy non-keyed client for backward compat (endpoints that still inject directly)
        services.AddGrpcClient<CollectorService.CollectorServiceClient>(o =>
        {
            o.Address = new Uri(collectorShopeeAddr);
        });

        var publisherAddr = configuration["Grpc:PublisherAddress"] ?? "http://localhost:50052";
        services.AddGrpcClient<PublisherService.PublisherServiceClient>(o =>
        {
            o.Address = new Uri(publisherAddr);
        });

        // ─── Keyed IProductSource (Strategy Pattern via .NET Keyed Services) ─
        // Adicionar um novo marketplace = adicionar uma linha aqui. Zero mudanças nos endpoints.
        services.AddKeyedScoped<IProductSource, ShopeeProductSource>(Marketplaces.Shopee);
        services.AddKeyedScoped<IProductSource, AmazonProductSource>(Marketplaces.Amazon);

        return services;
    }
}
