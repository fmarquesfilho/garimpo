using Garimpei.Domain.Entities;

namespace Garimpei.Domain.Interfaces;

public interface IProductRepository
{
    Task<Product?> GetByIdAsync(Guid id, CancellationToken ct = default);
    Task<IReadOnlyList<Product>> GetByKeywordAsync(string keyword, int limit, CancellationToken ct = default);
    Task AddAsync(Product product, CancellationToken ct = default);
    Task AddRangeAsync(IEnumerable<Product> products, CancellationToken ct = default);
    Task UpdateAsync(Product product, CancellationToken ct = default);
}
