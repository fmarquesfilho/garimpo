using System;
using Garimpei.Domain.Interfaces;

namespace Garimpei.Domain.Entities;

public sealed class Loja : IOwnedEntity
{
    public Guid Id { get; init; } = Guid.NewGuid();
    public required string OwnerUid { get; set; }

    /// <summary>ID numérico da loja no marketplace (ex: 920292999).</summary>
    public required long ShopId { get; init; }

    /// <summary>Nome oficial retornado pelo marketplace.</summary>
    public required string Nome { get; set; }

    /// <summary>Nome normalizado para matching (lowercase, sem espaços/acentos, [a-z0-9]).</summary>
    public required string NomeNormalizado { get; set; }

    /// <summary>Marketplace obrigatório (shopee, mercado_livre, amazon).</summary>
    public required string Marketplace { get; init; }

    /// <summary>Cron de coleta. Null = loja escopada (sem monitoramento).</summary>
    public string? CronExpression { get; set; }

    /// <summary>URL original usada na resolução (preserva tracking de afiliado).</summary>
    public string? SourceUrl { get; set; }

    /// <summary>Origem geográfica padrão (ex: 🇰🇷, 🇧🇷).</summary>
    public string? OrigemPadrao { get; set; }

    // ── Campos enriquecidos (Shopee get_shop_detail) ──────────────────────
    /// <summary>URL do avatar da loja.</summary>
    public string? ImageUrl { get; set; }

    /// <summary>URL da imagem de capa da loja.</summary>
    public string? CoverUrl { get; set; }

    /// <summary>Número de seguidores.</summary>
    public int? FollowerCount { get; set; }

    /// <summary>Quantidade total de produtos na loja.</summary>
    public int? ItemCount { get; set; }

    /// <summary>Avaliação média (0-5 estrelas).</summary>
    public double? RatingStar { get; set; }

    /// <summary>Localização reportada pela loja (ex: "São Paulo").</summary>
    public string? ShopLocation { get; set; }

    /// <summary>Descrição/bio da loja.</summary>
    public string? Description { get; set; }

    public DateTime CreatedAt { get; init; } = DateTime.UtcNow;
    public DateTime UpdatedAt { get; set; } = DateTime.UtcNow;

    /// <summary>Gera NomeNormalizado a partir do Nome canônico.</summary>
    public static string Normalizar(string nome)
    {
        if (string.IsNullOrWhiteSpace(nome)) return string.Empty;
        var normalized = nome.Normalize(System.Text.NormalizationForm.FormD);
        var sb = new System.Text.StringBuilder(normalized.Length);
        foreach (var c in normalized)
        {
            var cat = System.Globalization.CharUnicodeInfo.GetUnicodeCategory(c);
            if (cat == System.Globalization.UnicodeCategory.NonSpacingMark) continue;
            if (char.IsAsciiLetterOrDigit(c)) sb.Append(char.ToLowerInvariant(c));
        }
        return sb.ToString();
    }
}
