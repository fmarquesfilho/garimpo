using Garimpei.Domain.Interfaces;

namespace Garimpei.Domain.Entities;

/// <summary>
/// Canal de publicação (grupo Telegram, número WhatsApp, etc.).
/// </summary>
public sealed class Destino : IOwnedEntity
{
    public Guid Id { get; init; } = Guid.NewGuid();
    public required string Nome { get; init; }
    public required string Tipo { get; init; } // "telegram" | "whatsapp"
    public string Config { get; set; } = string.Empty; // chat_id, telefone, etc.
    public bool Ativo { get; set; } = true;
    public string OwnerUid { get; set; } = string.Empty;
    public DateTime CreatedAt { get; init; } = DateTime.UtcNow;
    public DateTime UpdatedAt { get; set; } = DateTime.UtcNow;
}
