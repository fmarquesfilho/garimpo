using System.Security.Claims;
using Garimpei.Infrastructure.Tenancy;

namespace Garimpei.Api.Middleware;

/// <summary>
/// Resolves the tenant (OwnerUid) from the authenticated user's JWT claims.
/// Firebase JWT contains "user_id" claim which maps to our OwnerUid.
///
/// For unauthenticated endpoints (health, root), the middleware is a no-op.
/// For authenticated endpoints without a valid user_id claim, returns 401.
/// </summary>
public sealed class TenantMiddleware(RequestDelegate next)
{
    public async Task InvokeAsync(HttpContext context, TenantContext tenantContext)
    {
        if (context.User.Identity?.IsAuthenticated != true)
        {
            await next(context);
            return;
        }

        // Firebase JWT "user_id" claim (standard Firebase Auth)
        var ownerUid = context.User.FindFirstValue("user_id")
            ?? context.User.FindFirstValue(ClaimTypes.NameIdentifier);

        if (string.IsNullOrWhiteSpace(ownerUid))
        {
            context.Response.StatusCode = StatusCodes.Status401Unauthorized;
            await context.Response.WriteAsJsonAsync(new
            {
                error = "tenant_not_resolved",
                detail = "Missing user_id claim in JWT"
            });
            return;
        }

        tenantContext.Set(ownerUid);
        await next(context);
    }
}
