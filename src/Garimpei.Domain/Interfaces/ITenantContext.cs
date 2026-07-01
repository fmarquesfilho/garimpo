namespace Garimpei.Domain.Interfaces;

/// <summary>
/// Provides the current tenant (owner_uid) resolved from the request context.
/// Injected as Scoped — one instance per HTTP request.
/// </summary>
public interface ITenantContext
{
    /// <summary>
    /// The owner_uid (Firebase user_id) for the current request.
    /// Empty string if not resolved (anonymous/health endpoints).
    /// </summary>
    string OwnerUid { get; }

    /// <summary>
    /// Whether the tenant has been resolved for this request.
    /// </summary>
    bool IsResolved { get; }
}
