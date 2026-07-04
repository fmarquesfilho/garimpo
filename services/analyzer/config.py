from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    """Configuration loaded from environment variables."""

    bq_project: str = ""
    bq_dataset: str = "garimpo"
    port: int = 8060

    # BigQuery emulator (dev local)
    bigquery_emulator_host: str = ""

    # Mock mode: retorna dados fictícios sem BigQuery
    mock_data: bool = False

    class Config:
        env_prefix = ""
        case_sensitive = False


settings = Settings()
