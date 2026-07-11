using Garimpei.Infrastructure.Persistence;
using Microsoft.EntityFrameworkCore;

/// <summary>
/// GET /api/buscas/{id}/job — retorna status do job no Scheduler para feedback ao usuário.
/// Permite saber: última execução, próxima execução, status atual.
/// </summary>
public static partial class EndpointExtensions
{
    public static WebApplication MapJobStatusEndpoints(this WebApplication app)
    {
        app.MapGet("/api/buscas/{id}/job", async (
            Guid id,
            AppDbContext db,
            Scheduler.V1.SchedulerService.SchedulerServiceClient scheduler,
            ILogger<AppDbContext> logger,
            CancellationToken ct) =>
        {
            var busca = await db.Buscas.FirstOrDefaultAsync(b => b.Id == id, ct);
            if (busca is null)
                return Results.NotFound(new { error = "busca não encontrada" });

            var jobId = $"busca-{busca.Id}";

            try
            {
                var response = await scheduler.ListJobsAsync(
                    new Scheduler.V1.ListJobsRequest { StatusFilter = "all" },
                    cancellationToken: ct);

                var job = response.Jobs.FirstOrDefault(j => j.Id == jobId);
                if (job is null)
                {
                    return Results.Ok(new
                    {
                        job_id = jobId,
                        status = "não_registrado",
                        cron = busca.CronExpression ?? "",
                        detail = "Job não encontrado no Scheduler (pode ter sido reiniciado)"
                    });
                }

                return Results.Ok(new
                {
                    job_id = job.Id,
                    name = job.Name,
                    cron = job.CronExpression,
                    status = job.Status,
                    last_run_at = job.LastRunAt,
                    next_run_at = job.NextRunAt
                });
            }
            catch (Exception ex)
            {
                logger.LogWarning(ex, "Falha ao consultar Scheduler para job {JobId}", jobId);
                return Results.Ok(new
                {
                    job_id = jobId,
                    status = "indisponível",
                    cron = busca.CronExpression ?? "",
                    detail = "Scheduler temporariamente indisponível"
                });
            }
        }).RequireAuthorization().WithTags("Buscas");

        return app;
    }
}
