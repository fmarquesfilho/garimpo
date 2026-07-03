using Garimpei.Api.Endpoints.Coupons;
using Garimpei.Api.Middleware;
using Garimpei.Application;
using Garimpei.Infrastructure;
using Garimpei.Infrastructure.Tenancy;
using OpenTelemetry.Resources;
using OpenTelemetry.Trace;
using OpenTelemetry.Metrics;
using Serilog;
using System.Text.Json;

var builder = WebApplication.CreateBuilder(args);

// JSON: snake_case para compatibilidade com frontend
builder.Services.ConfigureHttpJsonOptions(options =>
{
    options.SerializerOptions.PropertyNamingPolicy = JsonNamingPolicy.SnakeCaseLower;
});

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

// CORS (dev only — production uses same-origin via Cloudflare Worker)
builder.Services.AddCors();

var app = builder.Build();

// Middleware pipeline
app.UseSerilogRequestLogging();

// In Development, allow CORS from any origin (frontend dev server)
if (app.Environment.IsDevelopment())
{
    app.UseCors(policy => policy
        .AllowAnyOrigin()
        .AllowAnyMethod()
        .AllowAnyHeader());
}

// In Development, bypass Firebase JWT auth with a fake user for local testing
if (app.Environment.IsDevelopment())
{
    app.Use(async (context, next) =>
    {
        if (context.User.Identity?.IsAuthenticated != true)
        {
            // Allow overriding tenant via X-Dev-User header for multi-tenant testing
            var devUser = context.Request.Headers["X-Dev-User"].FirstOrDefault() ?? "dev-user-001";
            var devEmail = context.Request.Headers["X-Dev-Email"].FirstOrDefault() ?? "dev@garimpei.local";

            var claims = new[]
            {
                new System.Security.Claims.Claim("user_id", devUser),
                new System.Security.Claims.Claim(System.Security.Claims.ClaimTypes.Email, devEmail),
            };
            var identity = new System.Security.Claims.ClaimsIdentity(claims, "DevBypass");
            context.User = new System.Security.Claims.ClaimsPrincipal(identity);
        }
        await next();
    });
}

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
app.MapBuscasCompatEndpoints();
app.MapLojasCompatEndpoints();
app.MapFavoritosEndpoints();
app.MapDestinosEndpoints();
app.MapTemplatesEndpoints();
app.MapPublicacoesEndpoints();
app.MapAlertasEndpoints();
app.MapOnboardingEndpoints();
app.MapAnalyticsEndpoints();
app.MapResolverLinkEndpoints();

// V2 API routes (native C# format)
app.MapGroup("/api/v2")
    .RequireAuthorization()
    .MapBuscasEndpoints()
    .MapCuradoriaEndpoints()
    .MapLojasEndpoints()
    .MapPublicacaoEndpoints()
    .MapCouponRulesEndpoints()
    .MapCouponListingEndpoints();

// Internal endpoints (no auth — internal network only)
app.MapCouponAlertEvaluationEndpoints();

app.Run();

// Endpoint groups — each in its own file
public static partial class EndpointExtensions { }
