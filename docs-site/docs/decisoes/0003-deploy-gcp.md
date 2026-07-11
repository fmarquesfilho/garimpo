# ADR 0003 — Deploy em GCP (Cloud Run + BigQuery)

## Contexto

O projeto nasceu com deploy em OCI (Oracle Cloud) com nginx + Postgres (descrito
no antigo `DEPLOY.md`). Posteriormente migrou para GCP.

## Decisão

A arquitetura real e única é **GCP**:
- **Cloud Run** para o backend (`garimpo-api`)
- **BigQuery** para persistência analítica
- **Cloud Scheduler** para coletas agendadas
- **Artifact Registry** para imagens Docker
- **Secret Manager** para credenciais

## Consequências

- O `DEPLOY.md` (OCI) está arquivado em `docs/legado/` com aviso de obsolescência.
- Todo runbook de operação refere-se exclusivamente a GCP.
- Não há Postgres no projeto — queries são BigQuery SQL.
