using Garimpei.Domain.Entities;
using Garimpei.Domain.Interfaces;
using Garimpei.Infrastructure.Persistence;
using Microsoft.EntityFrameworkCore;

/// <summary>
/// Onboarding endpoints — fluxo multi-step de cadastro do tenant.
/// /api/onboarding/* (status, termos, shopee, telegram, validar, excluir-conta)
/// </summary>
public static partial class EndpointExtensions
{
    public static WebApplication MapOnboardingEndpoints(this WebApplication app)
    {
        var onboarding = app.MapGroup("/api/onboarding")
            .RequireAuthorization()
            .WithTags("Onboarding");

        // GET /api/onboarding/status
        onboarding.MapGet("/status", async (
            AppDbContext db,
            ITenantContext tenant,
            CancellationToken ct) =>
        {
            var cfg = await db.TenantConfigs
                .FirstOrDefaultAsync(c => c.OwnerUid == tenant.OwnerUid, ct);

            if (cfg is null)
            {
                return Results.Ok(new
                {
                    onboarding_step = 0,
                    configurado = false,
                    aceitou_termos = false
                });
            }

            return Results.Ok(new
            {
                onboarding_step = cfg.OnboardingStep,
                configurado = cfg.Configurado,
                aceitou_termos = cfg.AceitouTermos,
                tem_shopee = !string.IsNullOrEmpty(cfg.ShopeeAppId),
                tem_telegram = !string.IsNullOrEmpty(cfg.TelegramChatId),
                tem_whatsapp = !string.IsNullOrEmpty(cfg.WhatsappPhoneNumberId)
            });
        });

        // POST /api/onboarding/termos
        onboarding.MapPost("/termos", async (
            AppDbContext db,
            ITenantContext tenant,
            CancellationToken ct) =>
        {
            var cfg = await GetOrCreateTenantConfig(db, tenant.OwnerUid, ct);
            cfg.AceitouTermos = true;
            cfg.AceitouTermosEm = DateTime.UtcNow;
            cfg.OnboardingStep = Math.Max(cfg.OnboardingStep, 1);
            cfg.UpdatedAt = DateTime.UtcNow;
            await db.SaveChangesAsync(ct);

            return Results.Ok(new { step = cfg.OnboardingStep, status = "termos_aceitos" });
        });

        // POST /api/onboarding/shopee
        onboarding.MapPost("/shopee", async (
            AppDbContext db,
            ITenantContext tenant,
            OnboardingShopeeRequest req,
            CancellationToken ct) =>
        {
            var cfg = await GetOrCreateTenantConfig(db, tenant.OwnerUid, ct);
            cfg.ShopeeAppId = req.AppId;
            cfg.ShopeeSecretEnc = req.Secret; // TODO: encrypt
            cfg.OnboardingStep = Math.Max(cfg.OnboardingStep, 2);
            cfg.UpdatedAt = DateTime.UtcNow;
            await db.SaveChangesAsync(ct);

            return Results.Ok(new { step = cfg.OnboardingStep, status = "shopee_configurado" });
        });

        // POST /api/onboarding/telegram
        onboarding.MapPost("/telegram", async (
            AppDbContext db,
            ITenantContext tenant,
            OnboardingTelegramRequest req,
            CancellationToken ct) =>
        {
            var cfg = await GetOrCreateTenantConfig(db, tenant.OwnerUid, ct);

            if (req.Pular != true)
            {
                cfg.TelegramTokenEnc = req.Token; // TODO: encrypt
                cfg.TelegramChatId = req.ChatId;
            }

            cfg.OnboardingStep = Math.Max(cfg.OnboardingStep, 3);
            cfg.UpdatedAt = DateTime.UtcNow;
            await db.SaveChangesAsync(ct);

            return Results.Ok(new { step = cfg.OnboardingStep, status = req.Pular == true ? "telegram_pulado" : "telegram_configurado" });
        });

        // POST /api/onboarding/whatsapp
        onboarding.MapPost("/whatsapp", async (
            AppDbContext db,
            ITenantContext tenant,
            OnboardingWhatsappRequest req,
            CancellationToken ct) =>
        {
            var cfg = await GetOrCreateTenantConfig(db, tenant.OwnerUid, ct);

            if (req.Pular != true)
            {
                cfg.WhatsappPhoneNumberId = req.PhoneNumberId;
                cfg.WhatsappTokenEnc = req.AccessToken; // TODO: encrypt
            }

            cfg.OnboardingStep = Math.Max(cfg.OnboardingStep, 3);
            cfg.UpdatedAt = DateTime.UtcNow;
            await db.SaveChangesAsync(ct);

            return Results.Ok(new { step = cfg.OnboardingStep, status = req.Pular == true ? "whatsapp_pulado" : "whatsapp_configurado" });
        });

        // POST /api/onboarding/validar
        onboarding.MapPost("/validar", async (
            AppDbContext db,
            ITenantContext tenant,
            CancellationToken ct) =>
        {
            var cfg = await GetOrCreateTenantConfig(db, tenant.OwnerUid, ct);

            // TODO: validar credenciais Shopee com chamada de teste real
            cfg.OnboardingStep = 4;
            cfg.UpdatedAt = DateTime.UtcNow;
            await db.SaveChangesAsync(ct);

            return Results.Ok(new { step = 4, status = "validado", configurado = true });
        });

        // POST /api/onboarding/excluir-conta (LGPD)
        onboarding.MapPost("/excluir-conta", async (
            AppDbContext db,
            ITenantContext tenant,
            CancellationToken ct) =>
        {
            // Remove todos os dados do tenant
            var cfg = await db.TenantConfigs.FirstOrDefaultAsync(c => c.OwnerUid == tenant.OwnerUid, ct);
            if (cfg is not null) db.TenantConfigs.Remove(cfg);

            // Soft-delete de buscas, favoritos, destinos, templates
            var buscas = await db.Buscas.ToListAsync(ct);
            foreach (var b in buscas) b.Active = false;

            var favoritos = await db.Favoritos.ToListAsync(ct);
            foreach (var f in favoritos) f.Ativo = false;

            var destinos = await db.Destinos.ToListAsync(ct);
            foreach (var d in destinos) d.Ativo = false;

            var templates = await db.Templates.ToListAsync(ct);
            foreach (var t in templates) t.Ativo = false;

            await db.SaveChangesAsync(ct);

            return Results.Ok(new { status = "conta_excluida" });
        });

        return app;
    }

    private static async Task<TenantConfig> GetOrCreateTenantConfig(
        AppDbContext db, string ownerUid, CancellationToken ct)
    {
        var cfg = await db.TenantConfigs
            .FirstOrDefaultAsync(c => c.OwnerUid == ownerUid, ct);

        if (cfg is null)
        {
            cfg = new TenantConfig { OwnerUid = ownerUid };
            db.TenantConfigs.Add(cfg);
        }

        return cfg;
    }
}

public sealed record OnboardingShopeeRequest
{
    public required string AppId { get; init; }
    public required string Secret { get; init; }
}

public sealed record OnboardingTelegramRequest
{
    public string? Token { get; init; }
    public string? ChatId { get; init; }
    public bool? Pular { get; init; }
}

public sealed record OnboardingWhatsappRequest
{
    public string? PhoneNumberId { get; init; }
    public string? AccessToken { get; init; }
    public bool? Pular { get; init; }
}
