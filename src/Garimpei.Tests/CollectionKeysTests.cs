using System.Text.Json;
using Garimpei.Domain;
using Xunit;

namespace Garimpei.Tests;

public class CollectionKeysTests
{
    private record BuscaFixture(
        string Id,
        string Tipo,
        string[] Keywords,
        long[] ShopIds,
        string[] Categorias,
        string[] CollectionKeys
    );

    private static List<BuscaFixture> LoadFixtures()
    {
        var path = Path.Combine(
            AppContext.BaseDirectory, "..", "..", "..", "..", "..", "fixtures", "buscas.json");
        var json = File.ReadAllText(Path.GetFullPath(path));
        var items = JsonSerializer.Deserialize<JsonElement[]>(json)!;

        return items.Select(item => new BuscaFixture(
            Id: item.GetProperty("id").GetString()!,
            Tipo: item.GetProperty("tipo").GetString()!,
            Keywords: item.GetProperty("keywords").EnumerateArray()
                .Select(e => e.GetString()!).ToArray(),
            ShopIds: item.GetProperty("shop_ids").EnumerateArray()
                .Select(e => e.GetInt64()).ToArray(),
            Categorias: item.GetProperty("categorias").EnumerateArray()
                .Select(e => e.GetString()!).ToArray(),
            CollectionKeys: item.GetProperty("collection_keys").EnumerateArray()
                .Select(e => e.GetString()!).ToArray()
        )).ToList();
    }

    public static IEnumerable<object[]> FixtureData()
    {
        foreach (var fx in LoadFixtures())
            yield return [fx.Id, fx.ShopIds, fx.Keywords, fx.Categorias, fx.CollectionKeys];
    }

    [Theory]
    [MemberData(nameof(FixtureData))]
    public void Derive_MatchesFixture(string id, long[] shopIds, string[] keywords, string[] categorias, string[] expected)
    {
        _ = id; // Used for test display name
        var result = CollectionKeys.Derive(shopIds, keywords, categorias);
        Assert.Equal(expected, result);
    }

    [Fact]
    public void Derive_Sorted()
    {
        var result = CollectionKeys.Derive([999, 111, 555], null);
        var sorted = result.OrderBy(x => x, StringComparer.Ordinal).ToArray();
        Assert.Equal(sorted, result);
    }

    [Fact]
    public void Derive_NoDuplicates()
    {
        // "42" appears as both shop_id and keyword
        var result = CollectionKeys.Derive([42], ["42"]);
        Assert.Equal(["42"], result);
    }

    [Fact]
    public void Derive_EmptyKeywordsIgnored()
    {
        var result = CollectionKeys.Derive([], ["  ", "", "valid"]);
        Assert.Equal(["valid"], result);
    }

    [Fact]
    public void Derive_LowercaseTrim()
    {
        var result = CollectionKeys.Derive([], ["  HELLO  ", "World"]);
        Assert.Equal(["hello", "world"], result);
    }
}
