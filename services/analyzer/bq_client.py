"""BigQuery client singleton."""

from functools import lru_cache

from google.cloud import bigquery

from config import settings


@lru_cache(maxsize=1)
def get_client() -> bigquery.Client:
    """Create a BigQuery client. Uses BIGQUERY_EMULATOR_HOST in dev."""
    return bigquery.Client(project=settings.bq_project)


def query(sql: str, params: list | None = None) -> list[dict]:
    """Execute a query and return rows as dicts."""
    client = get_client()
    job_config = bigquery.QueryJobConfig()
    if params:
        job_config.query_parameters = params

    result = client.query(sql, job_config=job_config)
    return [dict(row) for row in result]


def dataset_ref() -> str:
    """Return fully qualified dataset reference for queries."""
    return f"`{settings.bq_project}.{settings.bq_dataset}`"
