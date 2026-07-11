using Garimpei.Domain.Interfaces;

namespace Garimpei.Domain.Entities;

public sealed class Busca : IOwnedEntity
{
    public Guid Id { get; init; } = Guid.NewGuid();
    public required string OwnerUid { get; set; }
    public string SortBy { get; init; } = "relevance";
    public int Limit { get; init; } = 50;
    public bool Active { get; set; } = true;
    public DateTime CreatedAt { get; init; } = DateTime.UtcNow;
    public DateTime UpdatedAt { get; set; } = DateTime.UtcNow;

    /// <summary>
    /// Arrays de Shop IDs (ex: lojas monitoradas). Se nulo, trata-se de uma busca por palavra-chave genérica.
    /// </summary>
    public long[]? ShopIds { get; set; }

    /// <summary>
    /// Mapeamento de shop_id → nome da loja. Persistido ao criar/salvar.
    /// Ex: { "920292999": "Glory of Seoul", "282170857": "Le Botanic" }
    /// </summary>
    public Dictionary<string, string>? ShopNames { get; set; }

    /// <summary>
    /// URL original fornecida pelo usuário ao adicionar a loja (pode ser link de afiliada com tracking).
    /// Preservada para futura geração de links de conversão via generateShortLink.
    /// </summary>
    public string? SourceUrl { get; set; }

    /// <summary>
    /// Keywords de busca/filtragem. Fonte canônica para identificação (BuscaContract).
    /// Para buscas tipo keyword: são os termos de busca.
    /// Para buscas tipo loja: são filtros opcionais dentro da loja.
    /// </summary>
    public string[]? Keywords { get; set; }

    /// <summary>
    /// Expressão cron para agendamento. Default: "0 */8 * * *" (a cada 8h).
    /// Se null, usa o default.
    /// </summary>
    public string? CronExpression { get; set; }

    /// <summary>
    /// Marketplaces ativos para esta busca. Armazenado como jsonb array no PostgreSQL.
    /// Conforme BuscaContract: mínimo 1 marketplace.
    /// </summary>
    public string[] Marketplaces { get; set; } = [Domain.Marketplaces.Shopee];

    /// <summary>
    /// Comissão mínima para filtragem (ex: 0.07 = 7%). Null = sem filtro.
    /// </summary>
    public decimal? ComissaoMin { get; set; }

    /// <summary>
    /// Vendas mínimas para filtragem. Null = sem filtro.
    /// </summary>
    public int? VendasMin { get; set; }

    /// <summary>
    /// Categorias de filtragem (OR). Null = sem filtro.
    /// </summary>
    public string[]? Categorias { get; set; }

    /// <summary>
    /// Fontes de dados ativas: "curadoria", "quedas", "novos", "lojas", "favoritos". Null = todas.
    /// </summary>
    public string[]? Fontes { get; set; }

    /// <summary>
    /// Returns the list of marketplace identifiers for this search.
    /// </summary>
    public string[] GetMarketplaceList() => Marketplaces;
}
