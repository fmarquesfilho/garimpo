using Garimpei.Domain;
using Garimpei.Domain.Entities;
using Microsoft.Extensions.Logging;

/// <summary>
/// Helper compartilhado para registrar/pausar jobs de coleta no Scheduler a partir de
/// uma <see cref="Busca"/>. Centraliza a montagem do <c>SetScheduleRequest</c> usada
/// tanto por /api/lojas quanto por /api/buscas, garantindo que TODO agendamento passe
/// pelo Scheduler (que por sua vez enfileira alertas via Cloud Tasks — ver ADR-0023).
/// </summary>
public static class SchedulerJobs
{
    /// <summary>Cron padrão quando a busca não define um (a cada 8h).</summary>
    public static string DefaultCron => "0 */8 * * *";

    /// <summary>Registra (ou atualiza) o job periódico da busca no Scheduler.</summary>
    public static Task RegisterAsync(
        Scheduler.V1.SchedulerService.SchedulerServiceClient schedulerClient,
        Busca busca,
        ILogger logger,
        CancellationToken ct)
        => SetAsync(schedulerClient, busca, enabled: true, logger, ct);

    /// <summary>Pausa o job periódico da busca no Scheduler (soft delete).</summary>
    public static Task PauseAsync(
        Scheduler.V1.SchedulerService.SchedulerServiceClient schedulerClient,
        Busca busca,
        ILogger logger,
        CancellationToken ct)
        => SetAsync(schedulerClient, busca, enabled: false, logger, ct);

    private static async Task SetAsync(
        Scheduler.V1.SchedulerService.SchedulerServiceClient schedulerClient,
        Busca busca,
        bool enabled,
        ILogger logger,
        CancellationToken ct)
    {
        try
        {
            await schedulerClient.SetScheduleAsync(BuildRequest(busca, enabled), cancellationToken: ct);
        }
        catch (Exception ex)
        {
            // Eventual consistency: a Busca já persistiu no PostgreSQL. Se o Scheduler
            // estiver indisponível, o job pode ser reconciliado depois.
            logger.LogWarning(ex, "Falha ao {Acao} job no Scheduler para busca {BuscaId}",
                enabled ? "registrar" : "pausar", busca.Id);
        }
    }

    /// <summary>
    /// Monta o <c>SetScheduleRequest</c> a partir da busca. Uma busca com <c>ShopIds</c>
    /// vira um job <c>shop_collection</c>; sem loja, vira <c>keyword_search</c> (Fetch por
    /// keyword). As keywords vêm de <c>Busca.Keywords</c> ou, na ausência, do campo
    /// <c>Busca.Keyword</c> (formato legado separado por vírgula das buscas por termo).
    /// </summary>
    public static Scheduler.V1.SetScheduleRequest BuildRequest(Busca busca, bool enabled)
    {
        var hasShop = busca.ShopIds is { Length: > 0 };
        var hasKeywords = busca.Keywords is { Length: > 0 }
            || (!hasShop && !string.IsNullOrWhiteSpace(busca.Keyword));

        var req = new Scheduler.V1.SetScheduleRequest
        {
            JobId = $"busca-{busca.Id}",
            CronExpression = string.IsNullOrWhiteSpace(busca.CronExpression) ? DefaultCron : busca.CronExpression,
            Enabled = enabled
        };

        // busca_id: ALWAYS present (BuscaContract requirement)
        req.Params.Add("busca_id", busca.Id.ToString());

        // Determine type based on actual fields
        var type = (hasShop, hasKeywords) switch
        {
            (true, true) => "mixed",
            (true, false) => "shop_collection",
            _ => "keyword_search"
        };

        if (hasShop)
            req.Params.Add("shop_id", busca.ShopIds![0].ToString());
        req.Params.Add("owner_uid", busca.OwnerUid);
        req.Params.Add("type", type);

        var keywords = busca.Keywords is { Length: > 0 }
            ? busca.Keywords
            : (!hasShop && !string.IsNullOrWhiteSpace(busca.Keyword)
                ? busca.Keyword.Split(',', StringSplitOptions.RemoveEmptyEntries | StringSplitOptions.TrimEntries)
                : null);

        if (keywords is { Length: > 0 })
            req.Params.Add("keywords", string.Join(",", keywords));

        // collection_keys: ALWAYS present (BuscaContract requirement)
        var collectionKeys = CollectionKeys.Derive(busca.ShopIds, keywords, busca.Categorias);
        req.Params.Add("collection_keys", string.Join(",", collectionKeys));

        return req;
    }
}
