using Garimpei.Domain.Entities;
using Garimpei.Domain.Interfaces;
using Garimpei.Infrastructure.Persistence;
using Microsoft.EntityFrameworkCore;

/// <summary>
/// Alertas endpoints — configuração e teste de alertas de preço.
/// /api/alertas, /api/alertas/testar, /api/alertas/configurar
/// </summary>
public static partial class EndpointExtensions
{
    public static WebApplication MapAlertasEndpoints(this WebApplication app)
    {
        var alertas = app.MapGroup("/api/alertas")
            .RequireAuthorization()
            .WithTags("Alertas");

        // GET /api/alertas — configuração atual
        alertas.MapGet("/", async (
            AppDbContext db,
            ITenantContext tenant,
            CancellationToken ct) =>
        {
            var cfg = await db.TenantConfigs
                .FirstOrDefaultAsync(c => c.OwnerUid == tenant.OwnerUid, ct);

            return Results.Ok(new
            {
                habilitado = cfg?.TelegramChatId is not null,
                chat_id = cfg?.TelegramChatId ?? "",
                threshold = cfg?.AlertaThreshold ?? 0.15,
                apenas_quedas = cfg?.AlertaApenasQuedas ?? true
            });
        });

        // POST /api/alertas/testar — envia alerta de teste
        alertas.MapPost("/testar", async (
            AppDbContext db,
            ITenantContext tenant,
            HttpClient httpClient,
            IConfiguration config,
            CancellationToken ct) =>
        {
            var cfg = await db.TenantConfigs
                .FirstOrDefaultAsync(c => c.OwnerUid == tenant.OwnerUid, ct);

            if (cfg?.TelegramChatId is null || cfg.TelegramTokenEnc is null)
            {
                return Results.BadRequest(new
                {
                    error = "Telegram não configurado",
                    detail = "Configure o Telegram no onboarding primeiro"
                });
            }

            // TODO(T-0045): enviar mensagem de teste via Telegram Bot API
            // Por ora retorna sucesso simulado
            return Results.Ok(new
            {
                status = "enviado",
                chat_id = cfg.TelegramChatId,
                mensagem = "🔔 Alerta de teste — se você recebeu esta mensagem, os alertas estão funcionando!"
            });
        });

        // POST /api/alertas/configurar — atualizar threshold e filtros
        alertas.MapPost("/configurar", async (
            AppDbContext db,
            ITenantContext tenant,
            ConfigurarAlertasRequest req,
            CancellationToken ct) =>
        {
            var cfg = await db.TenantConfigs
                .FirstOrDefaultAsync(c => c.OwnerUid == tenant.OwnerUid, ct);

            if (cfg is null)
            {
                cfg = new TenantConfig { OwnerUid = tenant.OwnerUid };
                db.TenantConfigs.Add(cfg);
            }

            if (req.ChatId is not null) cfg.TelegramChatId = req.ChatId;
            if (req.Threshold is not null) cfg.AlertaThreshold = req.Threshold.Value;
            if (req.ApenasQuedas is not null) cfg.AlertaApenasQuedas = req.ApenasQuedas.Value;
            cfg.UpdatedAt = DateTime.UtcNow;

            await db.SaveChangesAsync(ct);

            return Results.Ok(new
            {
                status = "configurado",
                threshold = cfg.AlertaThreshold,
                apenas_quedas = cfg.AlertaApenasQuedas
            });
        });

        return app;
    }
}

public sealed record ConfigurarAlertasRequest
{
    public string? ChatId { get; init; }
    public double? Threshold { get; init; }
    public bool? ApenasQuedas { get; init; }
}
