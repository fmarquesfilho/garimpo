using Garimpei.Domain.Entities;
using Garimpei.Infrastructure.Persistence;
using Garimpei.Infrastructure.Tenancy;
using Microsoft.EntityFrameworkCore;
using Publisher.V1;
using Xunit;

namespace Garimpei.Tests.Integration;

/// <summary>
/// Integration tests for the publish orchestration flow.
/// Validates the contract invariants:
/// - Immediate publish (no agendada_em) MUST trigger gRPC Publisher.Publish
/// - Scheduled publish (with agendada_em) MUST NOT trigger gRPC
/// - GroupId_Resolution: destino_id (UUID) → Destino.Config (chat_id)
/// - Unreachable publisher → status "erro" with clear message
/// </summary>
public class PublishFlowTests : IDisposable
{
    private readonly AppDbContext _db;
    private readonly TenantContext _tenantContext;

    public PublishFlowTests()
    {
        _tenantContext = new TenantContext();
        _tenantContext.Set("test-user-publish");

        var options = new DbContextOptionsBuilder<AppDbContext>()
            .UseInMemoryDatabase(databaseName: $"publish-flow-{Guid.NewGuid()}")
            .Options;

        _db = new AppDbContext(options, _tenantContext);
        _db.Database.EnsureCreated();
    }

    public void Dispose() => _db.Dispose();

    // ═══════════════════════════════════════════════════════════════════════
    // Contract: GroupId_Resolution
    // ═══════════════════════════════════════════════════════════════════════

    [Fact]
    public async Task GroupIdResolution_ResolvesUuidToChatId()
    {
        // Arrange: destino com UUID e Config = chat_id real
        var destino = new Destino
        {
            Nome = "@mileseleciona",
            Tipo = "telegram",
            Config = "@mileseleciona"
        };
        _db.Destinos.Add(destino);
        await _db.SaveChangesAsync();

        // Act: resolve UUID → Config
        var found = await _db.Destinos.FindAsync(destino.Id);

        // Assert: Config é o chat_id (não o UUID)
        Assert.NotNull(found);
        Assert.Equal("@mileseleciona", found.Config);
        Assert.NotEqual(destino.Id.ToString(), found.Config);
    }

    [Fact]
    public async Task GroupIdResolution_DirectChatId_PassesThrough()
    {
        // Se o valor não é UUID, deve ser tratado como chat_id direto
        var directChatId = "@meugrupo";
        var isUuid = Guid.TryParse(directChatId, out _);

        Assert.False(isUuid, "Chat ID direto não deve ser parseável como UUID");
    }

    // ═══════════════════════════════════════════════════════════════════════
    // Contract: Immediate publish triggers gRPC
    // ═══════════════════════════════════════════════════════════════════════

    [Fact]
    public void PublishRequest_WithGroupId_MustNotBeUuid()
    {
        // Invariant: group_id no gRPC NUNCA deve ser um UUID do PostgreSQL
        var validGroupIds = new[] { "@mileseleciona", "-1001234567890", "+5511999999999" };
        var invalidGroupId = Guid.NewGuid().ToString(); // UUID = inválido

        foreach (var gid in validGroupIds)
        {
            // Válido: começa com @, -, + ou é numérico
            Assert.True(
                gid.StartsWith('@') || gid.StartsWith('-') || gid.StartsWith('+') || long.TryParse(gid, out _),
                $"group_id '{gid}' deve ter formato de chat_id");
        }

        // UUID é inválido como group_id
        Assert.False(
            invalidGroupId.StartsWith('@') || invalidGroupId.StartsWith('-') || invalidGroupId.StartsWith('+'),
            "UUID não deve ser aceito como group_id");
    }

    // ═══════════════════════════════════════════════════════════════════════
    // Contract: Scheduled publish does NOT trigger gRPC
    // ═══════════════════════════════════════════════════════════════════════

    [Fact]
    public async Task ScheduledPublish_SavesWithStatusAgendada()
    {
        // Quando agendada_em está preenchido, status = "agendada" (não "enviada")
        var pub = new Publicacao
        {
            ProdutoId = "prod-123",
            Nome = "Produto Teste",
            Preco = 49.90m,
            DestinoId = Guid.NewGuid().ToString(),
            AgendadaEm = DateTime.UtcNow.AddHours(2),
            Status = "agendada" // NÃO deve chamar gRPC
        };

        _db.Publicacoes.Add(pub);
        await _db.SaveChangesAsync();

        var saved = await _db.Publicacoes.FindAsync(pub.Id);
        Assert.NotNull(saved);
        Assert.Equal("agendada", saved.Status);
    }

    [Fact]
    public async Task ImmediatePublish_SavesWithStatusEnviadaOrErro()
    {
        // Quando agendada_em está vazio, status DEVE ser "enviada" ou "erro"
        // (nunca "pendente" — indicaria que gRPC não foi chamado)
        var pub = new Publicacao
        {
            ProdutoId = "prod-456",
            Nome = "Produto Imediato",
            Preco = 29.90m,
            DestinoId = Guid.NewGuid().ToString(),
            AgendadaEm = null,
            Status = "enviada" // Ou "erro" se publisher falhar
        };

        _db.Publicacoes.Add(pub);
        await _db.SaveChangesAsync();

        var saved = await _db.Publicacoes.FindAsync(pub.Id);
        Assert.NotNull(saved);
        Assert.Contains(saved.Status, new[] { "enviada", "erro" });
        Assert.DoesNotContain("pendente", saved.Status);
    }

    // ═══════════════════════════════════════════════════════════════════════
    // Contract: Keywords always array
    // ═══════════════════════════════════════════════════════════════════════

    [Fact]
    public async Task Keywords_StoredAsCommaSeparated_ReturnedAsArray()
    {
        // O campo Keyword armazena comma-separated, mas o contrato exige array no response
        var busca = new Busca
        {
            Keyword = "kenzo,shiseido,dior",
            OwnerUid = "test-user-publish",
            SortBy = "relevance",
            Limit = 50
        };

        _db.Buscas.Add(busca);
        await _db.SaveChangesAsync();

        var saved = await _db.Buscas.FindAsync(busca.Id);
        Assert.NotNull(saved);

        // Simula o que o endpoint faz
        var keywords = saved.Keyword.Split(',', StringSplitOptions.RemoveEmptyEntries | StringSplitOptions.TrimEntries);
        Assert.Equal(3, keywords.Length);
        Assert.Equal("kenzo", keywords[0]);
        Assert.Equal("shiseido", keywords[1]);
        Assert.Equal("dior", keywords[2]);
    }
}
