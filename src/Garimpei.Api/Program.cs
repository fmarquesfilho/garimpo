using Garimpei.Api.Middleware;
using Garimpei.Application;
using Garimpei.Infrastructure;
using Garimpei.Infrastructure.Tenancy;
using OpenTelemetry.Resources;
using OpenTelemetry.Trace;
using OpenTelemetry.Metrics;
using Serilog;

var builder = WebApplication.CreateBuilder(args);

// Serilog
builder.Host.UseSerilog((context, config) =>
    config.ReadFrom.Configuration(context.Configuration));

// Application & Infrastructure layers
builder.Services.AddApplication();
builder.Services.AddInfrastructure(builder.Configuration);

// Auth (Firebase JWT — validates tokens issued by Firebase Auth)
builder.Services.AddAuthentication("Bearer")
    .AddJwtBearer("Bearer", options =>
    {
        var projectId = builder.Configuration["Auth:ProjectId"]
            ?? throw new InvalidOperationException("Auth:ProjectId is required");

        options.Authority = $"https://securetoken.google.com/{projectId}";
        options.TokenValidationParameters = new()
        {
            ValidateIssuer = true,
            ValidIssuer = $"https://securetoken.google.com/{projectId}",
            ValidateAudience = true,
            ValidAudience = projectId,
            ValidateLifetime = true
        };
    });

builder.Services.AddAuthorization();

// Health checks
builder.Services.AddHealthChecks()
    .AddNpgSql(
        builder.Configuration.GetConnectionString("PostgreSQL") ?? "",
        name: "postgresql",
        tags: ["db", "ready"]);

// HttpClient for analyzer service
builder.Services.AddHttpClient();

// OpenTelemetry
builder.Services.AddOpenTelemetry()
    .ConfigureResource(r => r.AddService("garimpei-api-v2"))
    .WithTracing(tracing => tracing
        .AddAspNetCoreInstrumentation()
        .AddHttpClientInstrumentation()
        .AddOtlpExporter())
    .WithMetrics(metrics => metrics
        .AddAspNetCoreInstrumentation()
        .AddHttpClientInstrumentation()
        .AddOtlpExporter());

// OpenAPI
builder.Services.AddOpenApi();

var app = builder.Build();

// Middleware pipeline
app.UseSerilogRequestLogging();
app.UseAuthentication();
app.UseAuthorization();
app.UseMiddleware<TenantMiddleware>();

// Health checks
app.MapHealthChecks("/health");
app.MapHealthChecks("/health/ready", new()
{
    Predicate = check => check.Tags.Contains("ready")
});

// OpenAPI endpoint
app.MapOpenApi();

// API routes
app.MapGet("/", () => Results.Ok(new { service = "garimpei-api", version = "v2", status = "ok" }));

// Compatibility routes (/api/*) for frontend during migration
app.MapCompatEndpoints();

app.MapGroup("/api/v2")
    .RequireAuthorization()
    .MapBuscasEndpoints()
    .MapCuradoriaEndpoints()
    .MapLojasEndpoints()
    .MapPublicacaoEndpoints();

app.Run();

// Endpoint groups — each in its own file
public static partial class EndpointExtensions { }
