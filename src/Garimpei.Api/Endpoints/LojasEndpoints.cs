using Garimpei.Domain.Entities;
using Garimpei.Infrastructure.Persistence;
using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Logging;

/// <summary>
/// Lojas endpoints — /api/lojas (CRUD + ResolveShop + Scheduler), /api/lojas/novidades, /api/lojas/evolucao.
/// Monitoramento de lojas: adicionar/remover + agendamento + dados do analyzer Python.
/// </summary>
public static partial class EndpointExtensions
{
    public static WebApplication MapLojasEndpoints(this WebApplication app)
    {
        // /api/lojas/registro — lista o registro central de lojas do tenant
        app.MapGet("/api/lojas/registro", async (AppDbContext db, CancellationToken ct) =>
        {
            var lojas = await db.Lojas
                .OrderByDescending(l => l.CreatedAt)
                .Select(l => new
                {
                    // id = ShopId numérico (chave de escopo da busca), NÃO o Guid PK.
                    // Ver design.md §11 e a revisão do store workflow (bug de escopo).
                    id = l.ShopId.ToString(),
                    nome = l.Nome,
                    nome_normalizado = l.NomeNormalizado,
                    marketplace = l.Marketplace,
                    cron = l.CronExpression,
                    origem = l.OrigemPadrao,
                    monitorada = !string.IsNullOrEmpty(l.CronExpression)
                })
                .ToListAsync(ct);

            return Results.Ok(new { lojas, total = lojas.Count });
        }).RequireAuthorization().WithTags("Lojas");

        // /api/lojas/resolver — resolve via Collector e faz upsert no registro de Lojas
        app.MapPost("/api/lojas/resolver", async (
            AppDbContext db,
            Collector.V1.CollectorService.CollectorServiceClient collectorClient,
            ResolverLojaRequest req,
            CancellationToken ct) =>
        {
            if (string.IsNullOrWhiteSpace(req.Input))
                return Results.BadRequest(new { error = "O input da loja é obrigatório." });

            var marketplaceInput = req.Marketplace?.ToLowerInvariant();
            var marketplace = marketplaceInput switch
            {
                "amazon" => Collector.V1.Marketplace.Amazon,
                "mercadolivre" or "ml" => Collector.V1.Marketplace.Mercadolivre,
                "shopee" => Collector.V1.Marketplace.Shopee,
                null or "" => Collector.V1.Marketplace.Shopee,
                _ => Collector.V1.Marketplace.Unspecified
            };

            if (marketplace == Collector.V1.Marketplace.Unspecified)
                return Results.BadRequest(new { error = $"Marketplace '{req.Marketplace}' não suportado." });

            try
            {
                var resolveResp = await collectorClient.ResolveShopAsync(new Collector.V1.ResolveShopRequest
                {
                    UsernameOrUrl = req.Input,
                    Marketplace = marketplace
                }, cancellationToken: ct);

                if (resolveResp.ShopId > 0)
                {
                    var mktStr = marketplace.ToString().ToLowerInvariant();
                    var nomeNormalizado = Loja.Normalizar(resolveResp.ShopName);

                    var existingLoja = await db.Lojas.FirstOrDefaultAsync(l =>
                        l.ShopId == resolveResp.ShopId &&
                        l.Marketplace == mktStr, ct);

                    if (existingLoja != null)
                    {
                        if (existingLoja.Nome != resolveResp.ShopName || existingLoja.NomeNormalizado != nomeNormalizado)
                        {
                            existingLoja.Nome = resolveResp.ShopName;
                            existingLoja.NomeNormalizado = nomeNormalizado;
                            existingLoja.UpdatedAt = DateTime.UtcNow;
                            await db.SaveChangesAsync(ct);
                        }
                        return Results.Ok(new
                        {
                            id = existingLoja.ShopId.ToString(),
                            nome = existingLoja.Nome,
                            nome_normalizado = existingLoja.NomeNormalizado,
                            marketplace = existingLoja.Marketplace,
                            cron = existingLoja.CronExpression,
                            origem = existingLoja.OrigemPadrao,
                            monitorada = !string.IsNullOrEmpty(existingLoja.CronExpression)
                        });
                    }

                    var novaLoja = new Loja
                    {
                        OwnerUid = "", // Será populado pelo interceptor do EF
                        ShopId = resolveResp.ShopId,
                        Nome = resolveResp.ShopName,
                        NomeNormalizado = nomeNormalizado,
                        Marketplace = mktStr,
                        OrigemPadrao = req.Origem,
                        SourceUrl = req.Input.StartsWith("http", StringComparison.OrdinalIgnoreCase) ? req.Input : null
                    };

                    db.Lojas.Add(novaLoja);
                    await db.SaveChangesAsync(ct);

                    return Results.Ok(new
                    {
                        id = novaLoja.ShopId.ToString(),
                        nome = novaLoja.Nome,
                        nome_normalizado = novaLoja.NomeNormalizado,
                        marketplace = novaLoja.Marketplace,
                        cron = novaLoja.CronExpression,
                        origem = novaLoja.OrigemPadrao,
                        monitorada = !string.IsNullOrEmpty(novaLoja.CronExpression)
                    });
                }

                return Results.BadRequest(new { error = $"Loja não encontrada ou link inválido no marketplace {marketplace.ToString().ToLowerInvariant()}." });
            }
            catch (Grpc.Core.RpcException ex) when (ex.StatusCode == Grpc.Core.StatusCode.NotFound || ex.StatusCode == Grpc.Core.StatusCode.InvalidArgument)
            {
                var mktStr = marketplace.ToString().ToLowerInvariant();
                return Results.BadRequest(new { error = $"Loja não encontrada ou link inválido no marketplace {mktStr}." });
            }
            catch (Grpc.Core.RpcException ex) when (ex.StatusCode == Grpc.Core.StatusCode.Unimplemented)
            {
                var mktStr = marketplace.ToString().ToLowerInvariant();
                return Results.BadRequest(new { error = $"Resolução de loja ainda não suportada para {mktStr}." });
            }
            catch
            {
                return Results.BadRequest(new { error = "Falha ao resolver o ID da loja via Collector." });
            }
        }).RequireAuthorization().WithTags("Lojas");

        // /api/lojas — listar/adicionar/remover lojas monitoradas (buscas com ShopIDs)
        app.MapGet("/api/lojas", async (AppDbContext db, CancellationToken ct) =>
        {
            var buscas = await db.Buscas
                .Where(b => b.Active)
                .OrderByDescending(b => b.CreatedAt)
                .ToListAsync(ct);

            return Results.Ok(new
            {
                lojas = buscas.Select(b => new
                {
                    id = b.Id,
                    shop_ids = b.ShopIds,
                    keywords = b.Keywords ?? Array.Empty<string>(),
                    shop_names = b.ShopNames,
                    cron_expression = b.CronExpression,
                    source_url = b.SourceUrl,
                    ativo = b.Active,
                    criado_em = b.CreatedAt
                }),
                total = buscas.Count
            });
        }).RequireAuthorization().WithTags("Lojas");

        app.MapPost("/api/lojas", async (
            AppDbContext db,
            Collector.V1.CollectorService.CollectorServiceClient collectorClient,
            Scheduler.V1.SchedulerService.SchedulerServiceClient schedulerClient,
            ILogger<AppDbContext> logger,
            AdicionarLojaRequest req,
            CancellationToken ct) =>
        {
            // "input" pode ser URL da loja ou keyword
            var keyword = req.Input ?? "";

            // Resolve marketplace a partir do campo origem_padrao (default: shopee)
            var marketplace = (req.OrigemPadrao?.ToLowerInvariant()) switch
            {
                "amazon" => Collector.V1.Marketplace.Amazon,
                "mercadolivre" or "ml" => Collector.V1.Marketplace.Mercadolivre,
                _ => Collector.V1.Marketplace.Shopee
            };

            var busca = new Busca
            {
                OwnerUid = "",
                SortBy = "relevance",
                Limit = 50,
                SourceUrl = keyword.StartsWith("http", StringComparison.OrdinalIgnoreCase) ? keyword : null,
                Keywords = req.Keywords is { Length: > 0 } ? req.Keywords : null,
                CronExpression = req.Cron
            };

            try
            {
                var resolveResp = await collectorClient.ResolveShopAsync(new Collector.V1.ResolveShopRequest
                {
                    UsernameOrUrl = keyword,
                    Marketplace = marketplace
                }, cancellationToken: ct);

                if (resolveResp.ShopId > 0)
                {
                    busca.ShopIds = [resolveResp.ShopId];
                    if (!string.IsNullOrEmpty(resolveResp.ShopName))
                    {
                        busca.ShopNames = new Dictionary<string, string>
                        {
                            [resolveResp.ShopId.ToString()] = resolveResp.ShopName
                        };
                    }
                }
            }
            catch (Grpc.Core.RpcException ex) when (ex.StatusCode == Grpc.Core.StatusCode.NotFound || ex.StatusCode == Grpc.Core.StatusCode.InvalidArgument)
            {
                return Results.BadRequest(new { error = $"Loja não encontrada ou link inválido no marketplace {marketplace}." });
            }
            catch (Grpc.Core.RpcException ex) when (ex.StatusCode == Grpc.Core.StatusCode.Unimplemented)
            {
                return Results.BadRequest(new { error = $"Resolução de loja ainda não suportada para {req.OrigemPadrao ?? "shopee"}." });
            }
            catch
            {
                return Results.BadRequest(new { error = "Falha ao resolver o ID da loja via Collector." });
            }

            db.Buscas.Add(busca);
            await db.SaveChangesAsync(ct);

            // Registra job no Scheduler (eventual consistency — se falhar, Busca persiste).
            // Loja monitorada sempre coleta periodicamente (default a cada 8h).
            if (busca.ShopIds is { Length: > 0 })
                await SchedulerJobs.RegisterAsync(schedulerClient, busca, logger, ct);

            return Results.Ok(new { id = busca.Id, keywords = busca.Keywords, shop_ids = busca.ShopIds, source_url = busca.SourceUrl, status = "adicionada" });
        }).RequireAuthorization().WithTags("Lojas");

        app.MapDelete("/api/lojas", async (
            AppDbContext db,
            Scheduler.V1.SchedulerService.SchedulerServiceClient schedulerClient,
            ILogger<AppDbContext> logger,
            string id,
            CancellationToken ct) =>
        {
            if (!Guid.TryParse(id, out var guid))
                return Results.BadRequest(new { error = "id inválido" });

            var busca = await db.Buscas.FindAsync([guid], ct);
            if (busca is null) return Results.NotFound();

            busca.Active = false;
            busca.UpdatedAt = DateTime.UtcNow;
            await db.SaveChangesAsync(ct);

            // Pausa o job no Scheduler
            await SchedulerJobs.PauseAsync(schedulerClient, busca, logger, ct);

            return Results.Ok(new { status = "removida", id });
        }).RequireAuthorization().WithTags("Lojas");

        // /api/lojas/novidades — novidades das lojas (produtos novos + variações preço)
        app.MapGet("/api/lojas/novidades", async (
            HttpClient httpClient,
            IConfiguration config,
            string? busca_id,
            int? dias,
            CancellationToken ct) =>
        {
            var analyzerUrl = config["Analyzer:BaseUrl"] ?? "http://localhost:8060";
            var url = $"{analyzerUrl}/novidades?busca_id={busca_id ?? ""}&dias={dias ?? 7}";
            try
            {
                var response = await httpClient.GetFromJsonAsync<object>(url, ct);
                return Results.Ok(response);
            }
            catch
            {
                // Analyzer offline — retorna vazio
                return Results.Ok(new
                {
                    busca_id = busca_id ?? "",
                    dias_janela = dias ?? 7,
                    produtos_novos = Array.Empty<object>(),
                    variacoes = Array.Empty<object>(),
                    total_atual = 0
                });
            }
        }).RequireAuthorization().WithTags("Lojas");

        // /api/lojas/evolucao — evolução de preço das lojas monitoradas
        app.MapGet("/api/lojas/evolucao", async (
            HttpClient httpClient,
            IConfiguration config,
            int? dias,
            CancellationToken ct) =>
        {
            var analyzerUrl = config["Analyzer:BaseUrl"] ?? "http://localhost:8060";
            var url = $"{analyzerUrl}/evolucao?dias={dias ?? 30}";
            try
            {
                var response = await httpClient.GetFromJsonAsync<object>(url, ct);
                return Results.Ok(response);
            }
            catch
            {
                return Results.Ok(new
                {
                    dias_janela = dias ?? 30,
                    lojas = Array.Empty<object>(),
                    resumo = new
                    {
                        total_lojas = 0,
                        total_produtos = 0,
                        preco_medio_global = 0.0,
                        variacao_media_global_pct = 0.0,
                        total_quedas = 0,
                        total_altas = 0
                    }
                });
            }
        }).RequireAuthorization().WithTags("Lojas");

        return app;
    }
}

public sealed record AdicionarLojaRequest
{
    public string? Input { get; init; }
    public string? Cron { get; init; }
    public string? OrigemPadrao { get; init; }
    public string[]? Keywords { get; init; }
}

public sealed record ResolverLojaRequest
{
    public required string Input { get; init; }
    public string? Marketplace { get; init; }
    public string? Origem { get; init; }
}
