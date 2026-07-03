using System.Text.Json;
using Xunit;

namespace Garimpei.Tests.Integration;

/// <summary>
/// Garante que os DTOs de request deserializam corretamente com snake_case.
/// Previne regressão: o bug de destino_id=null em produção aconteceu porque
/// JsonNamingPolicy.SnakeCaseLower não era aplicado ao desserializar requests.
/// </summary>
public class JsonBindingTests
{
    private static readonly JsonSerializerOptions SnakeCaseOptions = new()
    {
        PropertyNamingPolicy = JsonNamingPolicy.SnakeCaseLower
    };

    [Fact]
    public void AgendarPublicacaoRequest_DeserializesSnakeCase()
    {
        var json = """
        {
            "nome": "Sérum Vitamina C",
            "preco": 49.90,
            "destino_id": "f97e10e4-9b13-4397-84c6-9024237b062d",
            "template_id": "padrao",
            "legenda_custom": "<b>Teste</b>",
            "agendada_em": "2026-07-03T12:00:00Z",
            "produto_id": "P1"
        }
        """;

        var req = JsonSerializer.Deserialize<AgendarPublicacaoRequest>(json, SnakeCaseOptions);

        Assert.NotNull(req);
        Assert.Equal("Sérum Vitamina C", req!.Nome);
        Assert.Equal(49.90m, req.Preco);
        Assert.Equal("f97e10e4-9b13-4397-84c6-9024237b062d", req.DestinoId);
        Assert.Equal("padrao", req.TemplateId);
        Assert.Equal("<b>Teste</b>", req.LegendaCustom);
        Assert.NotNull(req.AgendadaEm);
        Assert.Equal("P1", req.ProdutoId);
    }

    [Fact]
    public void PublicarCompatRequest_DeserializesSnakeCase()
    {
        var json = """
        {
            "nome": "Produto X",
            "preco": 29.90,
            "destino_id": "@mileseleciona",
            "template_id": "foto",
            "link": "https://shope.ee/abc"
        }
        """;

        var req = JsonSerializer.Deserialize<PublicarCompatRequest>(json, SnakeCaseOptions);

        Assert.NotNull(req);
        Assert.Equal("Produto X", req!.Nome);
        Assert.Equal("@mileseleciona", req.DestinoId);
        Assert.Equal("foto", req.TemplateId);
        Assert.Equal("https://shope.ee/abc", req.Link);
    }

    [Fact]
    public void AgendarPublicacaoRequest_DeserializesWithoutAgendadaEm()
    {
        // Frontend envia sem agendada_em quando é envio imediato
        var json = """{"nome":"Test","preco":10,"destino_id":"uuid-here"}""";

        var req = JsonSerializer.Deserialize<AgendarPublicacaoRequest>(json, SnakeCaseOptions);

        Assert.NotNull(req);
        Assert.Equal("uuid-here", req!.DestinoId);
        Assert.Null(req.AgendadaEm);
    }

    [Fact]
    public void AgendarPublicacaoRequest_AlsoWorksWithJsonPropertyName()
    {
        // Testa que o [JsonPropertyName] funciona sem o SnakeCaseLower policy
        // (garante que o binding funciona independente da config global)
        var json = """{"destino_id":"abc","legenda_custom":"<b>X</b>","produto_id":"P1"}""";

        var req = JsonSerializer.Deserialize<AgendarPublicacaoRequest>(json);

        Assert.NotNull(req);
        Assert.Equal("abc", req!.DestinoId);
        Assert.Equal("<b>X</b>", req.LegendaCustom);
        Assert.Equal("P1", req.ProdutoId);
    }
}
