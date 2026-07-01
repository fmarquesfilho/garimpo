using Garimpei.Domain.Entities;

namespace Garimpei.Domain.Interfaces;

public interface IBuscaRepository
{
    Task<Busca?> GetByIdAsync(Guid id, CancellationToken ct = default);
    Task<IReadOnlyList<Busca>> GetActiveByOwnerAsync(string ownerUid, CancellationToken ct = default);
    Task AddAsync(Busca busca, CancellationToken ct = default);
    Task UpdateAsync(Busca busca, CancellationToken ct = default);
    Task DeleteAsync(Guid id, CancellationToken ct = default);
}
