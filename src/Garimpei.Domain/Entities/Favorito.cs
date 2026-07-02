using Garimpei.Domain.Interfaces;

namespace Garimpei.Domain.Entities;

/// <summary>
/// Produto salvo pelo usuário para análise posterior.
/// </summary>
public sealed class Favorito : IOwnedEntity
{
    public Guid Id { get; init; } = Guid.NewGuid();
    public required string ProdutoId { get; init; }
    public required string Nome { get; init; }
    public decimal Preco { get; set; }
    public double Comissao { get; set; }
    public string? Link { get; set; }
    public string? Imagem { get; set; }
    public string? Loja { get; set; }
    public string? Categoria { get; set; }
    public string? Origem { get; set; }
    public bool Ativo { get; set; } = true;
    public string OwnerUid { get; set; } = string.Empty;
    public DateTime CreatedAt { get; init; } = DateTime.UtcNow;
    public DateTime UpdatedAt { get; set; } = DateTime.UtcNow;
}
