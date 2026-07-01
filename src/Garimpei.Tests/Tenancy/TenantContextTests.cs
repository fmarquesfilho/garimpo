using Garimpei.Infrastructure.Tenancy;
using Xunit;

namespace Garimpei.Tests.Tenancy;

public class TenantContextTests
{
    [Fact]
    public void Default_State_HasEmptyOwnerUid_And_IsResolved_False()
    {
        var ctx = new TenantContext();

        Assert.Equal(string.Empty, ctx.OwnerUid);
        Assert.False(ctx.IsResolved);
    }

    [Fact]
    public void Set_SetsOwnerUid_And_IsResolved_True()
    {
        var ctx = new TenantContext();

        ctx.Set("user-abc-123");

        Assert.Equal("user-abc-123", ctx.OwnerUid);
        Assert.True(ctx.IsResolved);
    }

    [Fact]
    public void Set_WithEmptyString_StillMarksResolved()
    {
        var ctx = new TenantContext();

        ctx.Set(string.Empty);

        Assert.Equal(string.Empty, ctx.OwnerUid);
        Assert.True(ctx.IsResolved);
    }
}
