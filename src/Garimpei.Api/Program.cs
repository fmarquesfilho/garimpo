using Garimpei.Application;
using Garimpei.Infrastructure;
using Serilog;

var builder = WebApplication.CreateBuilder(args);

// Serilog
builder.Host.UseSerilog((context, config) =>
    config.ReadFrom.Configuration(context.Configuration));

// Application & Infrastructure layers
builder.Services.AddApplication();
builder.Services.AddInfrastructure(builder.Configuration);

// Auth (Firebase JWT)
builder.Services.AddAuthentication("Bearer")
    .AddJwtBearer("Bearer", options =>
    {
        options.Authority = builder.Configuration["Auth:Authority"];
        options.TokenValidationParameters = new()
        {
            ValidateIssuer = true,
            ValidIssuer = builder.Configuration["Auth:Authority"],
            ValidateAudience = true,
            ValidAudience = builder.Configuration["Auth:Audience"],
            ValidateLifetime = true
        };
    });

builder.Services.AddAuthorization();

// Health checks
builder.Services.AddHealthChecks();

// OpenAPI
builder.Services.AddOpenApi();

var app = builder.Build();

// Middleware pipeline
app.UseSerilogRequestLogging();
app.UseAuthentication();
app.UseAuthorization();

// Health check
app.MapHealthChecks("/health");

// OpenAPI endpoint
app.MapOpenApi();

// API routes
app.MapGet("/", () => Results.Ok(new { service = "garimpei-api", status = "ok" }));

app.MapGroup("/api/v2")
    .RequireAuthorization()
    .MapBuscasEndpoints()
    .MapCuradoriaEndpoints();

app.Run();

// Endpoint groups — each in its own file
public static partial class EndpointExtensions { }
