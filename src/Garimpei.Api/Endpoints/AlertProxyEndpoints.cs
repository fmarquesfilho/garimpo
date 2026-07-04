/// <summary>
/// Thin proxy: forwards /process-alert to the scheduler HTTP handler.
/// The scheduler owns alert orchestration (ADR-0023). The C# API only
/// routes the Cloud Tasks callback to the scheduler sidecar.
/// No business logic here — just a pass-through.
/// </summary>
public static partial class EndpointExtensions
{
    public static WebApplication MapAlertProxyEndpoints(this WebApplication app)
    {
        app.MapPost("/process-alert", async (
            HttpClient httpClient,
            HttpContext context,
            ILogger<Program> logger,
            CancellationToken ct) =>
        {
            // Forward the request body to scheduler HTTP handler
            var schedulerUrl = "http://localhost:8054/process-alert";

            using var body = new StreamContent(context.Request.Body);
            body.Headers.ContentType = new System.Net.Http.Headers.MediaTypeHeaderValue("application/json");

            try
            {
                var response = await httpClient.PostAsync(schedulerUrl, body, ct);
                var content = await response.Content.ReadAsStringAsync(ct);

                context.Response.StatusCode = (int)response.StatusCode;
                context.Response.ContentType = "application/json";
                await context.Response.WriteAsync(content, ct);
            }
            catch (HttpRequestException ex)
            {
                logger.LogError(ex, "Scheduler HTTP unreachable for alert processing");
                context.Response.StatusCode = 502;
                await context.Response.WriteAsync("{\"error\":\"scheduler_unavailable\"}", ct);
            }
        }).WithTags("Internal");

        return app;
    }
}
