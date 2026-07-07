# Documentação do Garimpei

## Estrutura

```
docs/
  01-visao-e-negocio.md      ← O quê e pra quem
  02-arquitetura.md          ← Como roda (Cloud Run, BigQuery, CI)
  03-fluxos-e-modelo.md      ← Entidades, buscas agendadas, pipelines
  04-operacao-shopee.md      ← Integração Shopee (afiliados, coleta, ResolveShop)
  05-manual-do-usuario.md    ← Como usar o produto
  06-qualidade-e-testes.md   ← CI, linters, E2E, fitness functions
  07-dados-e-ia.md           ← Análises + IA
  08-fluxos-sequencia.md     ← 11 diagramas de sequência (mermaid)
  gerado/                    ← Não editar manualmente (mise run docs)
    ENTIDADES.md             ← Mermaid ER do schema BigQuery
    env-vars.md              ← Variáveis de ambiente extraídas do código
  decisoes/                  ← ADRs (24 Architecture Decision Records)
  legado/                    ← Arquivos históricos (não usar como referência)
  meta/                      ← Plano de migração e auditoria
contracts/
  registry.yaml              ← Contratos de serviço (ADR-0020)
  schemas/                   ← JSON Schemas para payloads HTTP
api/
docs-site/                   ← Site Rspress (build com `mise run docs:sync`)
```

## Como gerar

```bash
mise run docs          # Gera tudo (ER, env-vars, site)
mise run docs:check    # Confere se docs geradas estão atualizadas (CI)
```

## Princípios

- **Fonte única de verdade (SSOT)** — cada fato vive em um lugar e é linkado dos outros.
- **Gerar, não transcrever** — tudo que pode ser extraído do código é gerado automaticamente.
- **🔮 = aspiracional** — funcionalidades não implementadas levam selo `> 🔮 Planejado`.
- **Português técnico** — sem jargão inglês desnecessário.

## Site de documentação

O site usa [Rspress](https://rspress.dev/) e vive em `docs-site/`.
Deploy automático via GitHub Actions → Cloudflare Pages.
URL: https://garimpei.app.br/docs

```bash
cd docs-site
npm install
npm run dev
```
