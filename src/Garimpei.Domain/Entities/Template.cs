using Garimpei.Domain.Interfaces;

namespace Garimpei.Domain.Entities;

/// <summary>
/// Template de mensagem para publicação. Suporta placeholders: {{nome}}, {{preco}}, {{categoria}}, {{estrategia}}, {{link}}.
/// </summary>
public sealed class Template : IOwnedEntity
{
    public Guid Id { get; init; } = Guid.NewGuid();
    public required string Nome { get; init; }
    public required string Corpo { get; init; } // corpo com placeholders (HTML permitido)
    public bool ComFoto { get; set; }
    public bool Ativo { get; set; } = true;
    public string OwnerUid { get; set; } = string.Empty;
    public DateTime CreatedAt { get; init; } = DateTime.UtcNow;
    public DateTime UpdatedAt { get; set; } = DateTime.UtcNow;

    /// <summary>
    /// Renderiza o template substituindo os placeholders.
    /// </summary>
    public string Renderizar(string nome, decimal preco, string categoria, string estrategia, string link)
    {
        return Corpo
            .Replace("{{nome}}", nome.Trim())
            .Replace("{{preco}}", $"R$ {preco:F2}")
            .Replace("{{categoria}}", categoria)
            .Replace("{{estrategia}}", estrategia)
            .Replace("{{link}}", link);
    }
}
