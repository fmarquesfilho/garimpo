using Garimpei.Domain.Interfaces;

namespace Garimpei.Infrastructure.Tenancy;

/// <summary>
/// Scoped service that holds the current tenant for the request.
/// Set by TenantMiddleware from the JWT "user_id" claim.
/// </summary>
public sealed class TenantContext : ITenantContext
{
    public string OwnerUid { get; private set; } = string.Empty;
    public bool IsResolved { get; private set; }

    public void Set(string ownerUid)
    {
        OwnerUid = ownerUid;
        IsResolved = true;
    }
}
