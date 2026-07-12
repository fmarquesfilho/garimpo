using System;
using System.IO;
using System.Text.Json;
using Garimpei.Domain.Entities;
using Xunit;

namespace Garimpei.Domain.Tests.Entities;

public class LojaTests
{
    private record NormalizacaoPar(string input, string expected);

    [Fact]
    public void Normalizar_DeveProcessarParesParametrizadosCorretamente()
    {
        var basePath = AppDomain.CurrentDomain.BaseDirectory;
        var fixturesPath = Path.GetFullPath(Path.Combine(basePath, "..", "..", "..", "..", "..", "fixtures", "normalizacao-pares.json"));
        
        var json = File.ReadAllText(fixturesPath);
        var pares = JsonSerializer.Deserialize<NormalizacaoPar[]>(json, new JsonSerializerOptions { PropertyNameCaseInsensitive = true });
        
        Assert.NotNull(pares);
        Assert.NotEmpty(pares);

        foreach (var par in pares)
        {
            var result = Loja.Normalizar(par.input);
            Assert.Equal(par.expected, result);
        }
    }
}
