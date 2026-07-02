namespace Garimpei.Domain;

/// <summary>
/// Well-known marketplace identifiers used throughout the system.
/// </summary>
public static class Marketplaces
{
    public const string Shopee = "shopee";
    public const string Amazon = "amazon";
    public const string MercadoLivre = "mercadolivre";

    /// <summary>
    /// All currently supported marketplaces.
    /// </summary>
    public static readonly string[] All = [Shopee, Amazon, MercadoLivre];

    /// <summary>
    /// Returns true if the given value is a recognized marketplace identifier.
    /// </summary>
    public static bool IsValid(string? marketplace) =>
        marketplace is Shopee or Amazon or MercadoLivre;

    /// <summary>
    /// Returns the marketplace string, defaulting to Shopee when null/empty.
    /// </summary>
    public static string ResolveOrDefault(string? marketplace) =>
        string.IsNullOrWhiteSpace(marketplace) ? Shopee : marketplace;
}
