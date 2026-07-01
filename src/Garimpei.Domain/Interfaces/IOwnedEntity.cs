namespace Garimpei.Domain.Interfaces;

/// <summary>
/// Marker interface for entities that belong to a tenant.
/// Entities implementing this will have automatic query filtering by OwnerUid.
/// </summary>
public interface IOwnedEntity
{
    string OwnerUid { get; set; }
}
