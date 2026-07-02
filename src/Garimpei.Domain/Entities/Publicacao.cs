using Garimpei.Domain.Interfaces;

namespace Garimpei.Domain.Entities;

/// <summary>
/// Publicação agendada ou enviada em um canal.
/// </summary>
public sealed class Publicacao : IOwnedEntity
{
    public Guid Id { get; init; } = Guid.NewGuid();
    public required string ProdutoId { get; init; }
    public required string Nome { get; init; }
    public string? Categoria { get; set; }
    public decimal Preco { get; set; }
    public double Comissao { get; set; }
    public string? Link { get; set; }
    public string? Imagem { get; set; }
    public string? Estrategia { get; set; }
    public string? DestinoId { get; set; }
    public string? TemplateId { get; set; }
    public DateTime? AgendadaEm { get; set; }
    public string Status { get; set; } = "pendente"; // pendente | enviada | erro
    public string? Detalhe { get; set; }
    public string OwnerUid { get; set; } = string.Empty;
    public DateTime CreatedAt { get; init; } = DateTime.UtcNow;
    public DateTime? EnviadaEm { get; set; }
}
