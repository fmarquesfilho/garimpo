using System.Text.Json;

/// <summary>
/// Admin logs endpoint — proxy para Cloud Logging API.
/// GET /api/admin/logs?severity=ERROR&service=scheduler&keyword=falhou&limit=50
/// Requer autenticação admin.
/// </summary>
public static partial class EndpointExtensions
{
    public static WebApplication MapLogsEndpoints(this WebApplication app)
    {
        app.MapGet("/api/admin/logs", async (
            HttpClient httpClient,
            IConfiguration config,
            HttpContext context,
            string? severity,
            string? service,
            string? keyword,
            string? traceId,
            int? limit,
            int? minutes,
            CancellationToken ct) =>
        {
            // Verify admin
            var email = context.User.FindFirst("email")?.Value ?? "";
            var adminEmails = config["AdminEmails"] ?? "";
            var isAdmin = !string.IsNullOrEmpty(email)
                && adminEmails.Split(',', StringSplitOptions.RemoveEmptyEntries)
                    .Any(e => e.Trim().Equals(email, StringComparison.OrdinalIgnoreCase));

            if (!isAdmin)
                return Results.Forbid();

            var projectId = config["GCP_PROJECT_ID"] ?? "garimpo-500114";
            var maxEntries = Math.Min(limit ?? 50, 200);
            var timeWindow = Math.Min(minutes ?? 60, 1440); // max 24h

            // Build Cloud Logging filter query
            var filters = new List<string>
            {
                $"resource.type=\"cloud_run_revision\"",
                $"resource.labels.project_id=\"{projectId}\"",
                $"timestamp >= \"{DateTime.UtcNow.AddMinutes(-timeWindow):yyyy-MM-ddTHH:mm:ssZ}\""
            };

            if (!string.IsNullOrWhiteSpace(severity))
                filters.Add($"severity >= \"{severity.ToUpper()}\"");

            if (!string.IsNullOrWhiteSpace(service))
                filters.Add($"resource.labels.container_name=\"{service}\"");

            if (!string.IsNullOrWhiteSpace(keyword))
                filters.Add($"textPayload:\"{keyword}\" OR jsonPayload.message:\"{keyword}\"");

            if (!string.IsNullOrWhiteSpace(traceId))
                filters.Add($"trace=\"projects/{projectId}/traces/{traceId}\"");

            var filter = string.Join("\n", filters);

            // Call Cloud Logging API v2 (entries:list)
            var requestBody = new
            {
                resourceNames = new[] { $"projects/{projectId}" },
                filter,
                orderBy = "timestamp desc",
                pageSize = maxEntries
            };

            try
            {
                // Use Application Default Credentials (ADC) — available in Cloud Run
                var loggingUrl = "https://logging.googleapis.com/v2/entries:list";
                var request = new HttpRequestMessage(HttpMethod.Post, loggingUrl)
                {
                    Content = JsonContent.Create(requestBody)
                };

                // In Cloud Run, use metadata server for auth token
                var tokenUrl = "http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/token";
                var tokenReq = new HttpRequestMessage(HttpMethod.Get, tokenUrl);
                tokenReq.Headers.Add("Metadata-Flavor", "Google");

                string accessToken;
                try
                {
                    var tokenResp = await httpClient.SendAsync(tokenReq, ct);
                    var tokenJson = await tokenResp.Content.ReadFromJsonAsync<JsonElement>(ct);
                    accessToken = tokenJson.GetProperty("access_token").GetString() ?? "";
                }
                catch
                {
                    // Not in Cloud Run (local dev) — return mock data
                    return Results.Ok(new
                    {
                        entries = new[]
                        {
                            new { timestamp = DateTime.UtcNow.ToString("o"), severity = "INFO", service = "garimpei-api", message = "Mock log entry (Cloud Logging não disponível em dev local)", jsonPayload = (object?)null },
                            new { timestamp = DateTime.UtcNow.AddSeconds(-5).ToString("o"), severity = "WARNING", service = "scheduler", message = "Use mise run test:e2e-traces para testar em produção", jsonPayload = (object?)null }
                        },
                        totalEntries = 2,
                        filter,
                        source = "mock"
                    });
                }

                request.Headers.Authorization = new System.Net.Http.Headers.AuthenticationHeaderValue("Bearer", accessToken);

                var response = await httpClient.SendAsync(request, ct);
                var content = await response.Content.ReadAsStringAsync(ct);

                if (!response.IsSuccessStatusCode)
                {
                    return Results.Ok(new
                    {
                        entries = Array.Empty<object>(),
                        totalEntries = 0,
                        error = $"Cloud Logging API: {response.StatusCode}",
                        filter
                    });
                }

                // Parse Cloud Logging response
                var loggingResponse = JsonSerializer.Deserialize<JsonElement>(content);
                var entries = new List<object>();

                if (loggingResponse.TryGetProperty("entries", out var entriesJson))
                {
                    foreach (var entry in entriesJson.EnumerateArray())
                    {
                        var ts = entry.TryGetProperty("timestamp", out var tsVal) ? tsVal.GetString() : "";
                        var sev = entry.TryGetProperty("severity", out var sevVal) ? sevVal.GetString() : "DEFAULT";
                        var containerName = "";
                        if (entry.TryGetProperty("resource", out var res) && res.TryGetProperty("labels", out var labels))
                            containerName = labels.TryGetProperty("container_name", out var cn) ? cn.GetString() ?? "" : "";

                        string? message = null;
                        object? payload = null;

                        if (entry.TryGetProperty("jsonPayload", out var jp))
                        {
                            payload = jp;
                            if (jp.TryGetProperty("message", out var msg))
                                message = msg.GetString();
                            else if (jp.TryGetProperty("msg", out var msg2))
                                message = msg2.GetString();
                        }
                        else if (entry.TryGetProperty("textPayload", out var tp))
                        {
                            message = tp.GetString();
                        }

                        var traceField = entry.TryGetProperty("trace", out var tr) ? tr.GetString() : null;
                        var spanField = entry.TryGetProperty("spanId", out var sp) ? sp.GetString() : null;

                        entries.Add(new
                        {
                            timestamp = ts,
                            severity = sev,
                            service = containerName,
                            message,
                            trace = traceField,
                            spanId = spanField,
                            jsonPayload = payload
                        });
                    }
                }

                return Results.Ok(new
                {
                    entries,
                    totalEntries = entries.Count,
                    filter,
                    source = "cloud_logging"
                });
            }
            catch (Exception ex)
            {
                return Results.Ok(new
                {
                    entries = Array.Empty<object>(),
                    totalEntries = 0,
                    error = ex.Message,
                    filter
                });
            }
        }).RequireAuthorization().WithTags("Admin");

        // POST /api/telemetry — frontend error/event reporting
        // Aceita eventos do browser (erros JS, Web Vitals, custom events).
        // Loga no stdout em JSON estruturado → Cloud Logging captura automaticamente.
        // Inclui o service "garimpei-web" para ser filtrável na página de logs.
        app.MapPost("/api/telemetry", (
            HttpContext context,
            ILogger<Program> logger,
            JsonElement body) =>
        {
            var type = body.TryGetProperty("type", out var t) ? t.GetString() : "unknown";
            var message = body.TryGetProperty("message", out var m) ? m.GetString() : "";
            var url = body.TryGetProperty("url", out var u) ? u.GetString() : "";
            var stack = body.TryGetProperty("stack", out var s) ? s.GetString() : null;
            var uid = context.User.FindFirst("user_id")?.Value ?? context.User.FindFirst("sub")?.Value ?? "anonymous";

            switch (type)
            {
                case "error":
                    logger.LogError("Frontend error: {Message} | url={Url} | uid={Uid} | stack={Stack}",
                        message, url, uid, stack ?? "(no stack)");
                    break;
                case "web-vital":
                    var name = body.TryGetProperty("name", out var n) ? n.GetString() : "";
                    var value = body.TryGetProperty("value", out var v) ? v.GetDouble() : 0;
                    var rating = body.TryGetProperty("rating", out var r) ? r.GetString() : "";
                    logger.LogInformation("Web Vital: {Name}={Value} ({Rating}) | url={Url} | uid={Uid}",
                        name, value, rating, url, uid);
                    break;
                default:
                    logger.LogInformation("Frontend event: type={Type} | message={Message} | url={Url} | uid={Uid}",
                        type, message, url, uid);
                    break;
            }

            return Results.Ok(new { received = true });
        }).RequireAuthorization().WithTags("Telemetry");

        return app;
    }
}
