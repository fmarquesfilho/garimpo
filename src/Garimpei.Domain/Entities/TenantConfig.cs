using Garimpei.Domain.Interfaces;

namespace Garimpei.Domain.Entities;

/// <summary>
/// Configurações e credenciais do tenant (multi-tenancy). Armazena estado de onboarding,
/// credenciais Shopee e Telegram (encriptadas fora da entidade).
/// </summary>
public sealed class TenantConfig : IOwnedEntity
{
    public Guid Id { get; init; } = Guid.NewGuid();
    public string OwnerUid { get; set; } = string.Empty;
    public string? Email { get; set; }
    public string? ShopeeAppId { get; set; }
    public string? ShopeeSecretEnc { get; set; } // criptografado
    public string? TelegramTokenEnc { get; set; } // criptografado
    public string? TelegramChatId { get; set; }

    // WhatsApp (Meta Cloud API)
    public string? WhatsappPhoneNumberId { get; set; }
    public string? WhatsappTokenEnc { get; set; } // criptografado

    // Amazon (Creators API — OAuth 2.0)
    public string? AmazonAccessKeyEnc { get; set; } // criptografado
    public string? AmazonSecretKeyEnc { get; set; } // criptografado
    public string? AmazonPartnerTag { get; set; }

    public int OnboardingStep { get; set; } // 0=início, 4=completo
    public bool AceitouTermos { get; set; }
    public DateTime? AceitouTermosEm { get; set; }

    // Alertas — price drop notification settings
    public double AlertaThreshold { get; set; } = 0.15; // 15% de queda
    public bool AlertaApenasQuedas { get; set; } = true;

    public DateTime CreatedAt { get; init; } = DateTime.UtcNow;
    public DateTime UpdatedAt { get; set; } = DateTime.UtcNow;

    public bool Configurado => OnboardingStep >= 4 && !string.IsNullOrEmpty(ShopeeAppId);
}
