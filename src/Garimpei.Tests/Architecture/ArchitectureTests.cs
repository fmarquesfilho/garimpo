using System.Reflection;
using NetArchTest.Rules;
using Xunit;

namespace Garimpei.Tests.Architecture;

/// <summary>
/// Fitness functions — testes automatizados que validam as regras arquiteturais do projeto.
/// Baseado em Clean Architecture: Domain → Application → Infrastructure → Api.
/// Referência: "Building Evolutionary Architectures" (Ford, Parsons, Kua).
/// </summary>
public class ArchitectureTests
{
    private static readonly Assembly DomainAssembly = typeof(Domain.Entities.Busca).Assembly;
    private static readonly Assembly ApplicationAssembly = typeof(Application.DependencyInjection).Assembly;
    private static readonly Assembly InfrastructureAssembly = typeof(Infrastructure.DependencyInjection).Assembly;
    private static readonly Assembly ApiAssembly = typeof(Program).Assembly;

    // ═══════════════════════════════════════════════════════════════════════
    // REGRA 1: Domain não depende de nada externo (camada mais interna)
    // ═══════════════════════════════════════════════════════════════════════

    [Fact]
    public void Domain_ShouldNot_DependOn_Application()
    {
        var result = Types.InAssembly(DomainAssembly)
            .ShouldNot()
            .HaveDependencyOn("Garimpei.Application")
            .GetResult();

        Assert.True(result.IsSuccessful,
            $"Domain depende de Application: {FormatFailures(result)}");
    }

    [Fact]
    public void Domain_ShouldNot_DependOn_Infrastructure()
    {
        var result = Types.InAssembly(DomainAssembly)
            .ShouldNot()
            .HaveDependencyOn("Garimpei.Infrastructure")
            .GetResult();

        Assert.True(result.IsSuccessful,
            $"Domain depende de Infrastructure: {FormatFailures(result)}");
    }

    [Fact]
    public void Domain_ShouldNot_DependOn_Api()
    {
        var result = Types.InAssembly(DomainAssembly)
            .ShouldNot()
            .HaveDependencyOn("Garimpei.Api")
            .GetResult();

        Assert.True(result.IsSuccessful,
            $"Domain depende de Api: {FormatFailures(result)}");
    }

    [Fact]
    public void Domain_ShouldNot_DependOn_EntityFramework()
    {
        var result = Types.InAssembly(DomainAssembly)
            .ShouldNot()
            .HaveDependencyOn("Microsoft.EntityFrameworkCore")
            .GetResult();

        Assert.True(result.IsSuccessful,
            $"Domain depende de EF Core (framework leak): {FormatFailures(result)}");
    }

    // ═══════════════════════════════════════════════════════════════════════
    // REGRA 2: Application depende apenas de Domain (use cases + interfaces)
    // ═══════════════════════════════════════════════════════════════════════

    [Fact]
    public void Application_ShouldNot_DependOn_Infrastructure()
    {
        var result = Types.InAssembly(ApplicationAssembly)
            .ShouldNot()
            .HaveDependencyOn("Garimpei.Infrastructure")
            .GetResult();

        Assert.True(result.IsSuccessful,
            $"Application depende de Infrastructure: {FormatFailures(result)}");
    }

    [Fact]
    public void Application_ShouldNot_DependOn_Api()
    {
        var result = Types.InAssembly(ApplicationAssembly)
            .ShouldNot()
            .HaveDependencyOn("Garimpei.Api")
            .GetResult();

        Assert.True(result.IsSuccessful,
            $"Application depende de Api: {FormatFailures(result)}");
    }

    // ═══════════════════════════════════════════════════════════════════════
    // REGRA 3: Infrastructure não depende de Api (apenas implementa interfaces)
    // ═══════════════════════════════════════════════════════════════════════

    [Fact]
    public void Infrastructure_ShouldNot_DependOn_Api()
    {
        var result = Types.InAssembly(InfrastructureAssembly)
            .ShouldNot()
            .HaveDependencyOn("Garimpei.Api")
            .GetResult();

        Assert.True(result.IsSuccessful,
            $"Infrastructure depende de Api: {FormatFailures(result)}");
    }

    // ═══════════════════════════════════════════════════════════════════════
    // REGRA 4: Entities devem implementar IOwnedEntity (multi-tenancy)
    // ═══════════════════════════════════════════════════════════════════════

    [Fact]
    public void Entities_InDomain_Should_ImplementIOwnedEntity_OrBeTenant()
    {
        var result = Types.InAssembly(DomainAssembly)
            .That()
            .ResideInNamespace("Garimpei.Domain.Entities")
            .And()
            .AreClasses()
            .Should()
            .ImplementInterface(typeof(Domain.Interfaces.IOwnedEntity))
            .Or()
            .HaveNameStartingWith("Tenant")
            .GetResult();

        // Tenant é a exceção — não é owned, é o próprio owner
        var failures = result.FailingTypes?
            .Where(t => t.Name != "Tenant")
            .ToList();

        Assert.True(failures == null || failures.Count == 0,
            $"Entidades sem IOwnedEntity (sem multi-tenancy): {string.Join(", ", failures?.Select(t => t.Name) ?? [])}");
    }

    // ═══════════════════════════════════════════════════════════════════════
    // REGRA 5: Interfaces do Domain devem residir no namespace correto
    // ═══════════════════════════════════════════════════════════════════════

    [Fact]
    public void Interfaces_ShouldResideIn_InterfacesNamespace()
    {
        var result = Types.InAssembly(DomainAssembly)
            .That()
            .AreInterfaces()
            .Should()
            .ResideInNamespace("Garimpei.Domain.Interfaces")
            .GetResult();

        Assert.True(result.IsSuccessful,
            $"Interfaces fora de Domain.Interfaces: {FormatFailures(result)}");
    }

    // ═══════════════════════════════════════════════════════════════════════
    // REGRA 6: Naming conventions
    // ═══════════════════════════════════════════════════════════════════════

    [Fact]
    public void Interfaces_ShouldStartWith_I()
    {
        var result = Types.InAssembly(DomainAssembly)
            .That()
            .AreInterfaces()
            .Should()
            .HaveNameStartingWith("I")
            .GetResult();

        Assert.True(result.IsSuccessful,
            $"Interfaces sem prefixo 'I': {FormatFailures(result)}");
    }

    [Fact]
    public void Entities_ShouldBe_Sealed()
    {
        var result = Types.InAssembly(DomainAssembly)
            .That()
            .ResideInNamespace("Garimpei.Domain.Entities")
            .And()
            .AreClasses()
            .Should()
            .BeSealed()
            .GetResult();

        Assert.True(result.IsSuccessful,
            $"Entidades não seladas (herança acidental): {FormatFailures(result)}");
    }

    // ═══════════════════════════════════════════════════════════════════════
    // REGRA 7: ValueObjects não devem ter setters públicos mutáveis
    // ═══════════════════════════════════════════════════════════════════════

    [Fact]
    public void ValueObjects_ShouldBe_Records()
    {
        var types = Types.InAssembly(DomainAssembly)
            .That()
            .ResideInNamespace("Garimpei.Domain.ValueObjects")
            .GetTypes();

        foreach (var type in types)
        {
            // Records em C# implementam IEquatable<T> e têm método <Clone>$
            var isRecord = type.GetMethods().Any(m => m.Name == "<Clone>$");
            Assert.True(isRecord,
                $"ValueObject '{type.Name}' deveria ser record (imutável por default)");
        }
    }

    // ═══════════════════════════════════════════════════════════════════════
    // REGRA 8: Services no Domain devem ser estáticos (sem estado)
    // ═══════════════════════════════════════════════════════════════════════

    [Fact]
    public void DomainServices_ShouldBe_Static()
    {
        var types = Types.InAssembly(DomainAssembly)
            .That()
            .ResideInNamespace("Garimpei.Domain.Services")
            .And()
            .AreClasses()
            .GetTypes();

        foreach (var type in types)
        {
            Assert.True(type.IsAbstract && type.IsSealed,
                $"Domain Service '{type.Name}' deveria ser static (stateless)");
        }
    }

    // ═══════════════════════════════════════════════════════════════════════
    // Helpers
    // ═══════════════════════════════════════════════════════════════════════

    private static string FormatFailures(TestResult result)
    {
        if (result.FailingTypes == null || !result.FailingTypes.Any())
            return "(nenhum)";

        return string.Join(", ", result.FailingTypes.Select(t => t.FullName).Take(10));
    }
}
