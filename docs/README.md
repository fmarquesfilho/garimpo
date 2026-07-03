# Documentação do Garimpei

## Estrutura

```
docs/
  01-visao-e-negocio.md      ← O quê e pra quem
  02-arquitetura.md          ← Como roda (Cloud Run, BigQuery, CI)
  03-fluxos-e-modelo.md      ← Entidades, buscas, regras de negócio
  04-operacao-shopee.md      ← Integração Shopee (afiliados, coleta, origem)
  05-manual-do-usuario.md    ← Como usar o produto
  06-qualidade-e-testes.md   ← CI, linters, BDD, cenários
  07-dados-e-ia.md           ← Análises + IA
  gerado/                    ← Não editar manualmente (mise run docs)
    ENTIDADES.md             ← Mermaid ER do schema BigQuery
    env-vars.md              ← Variáveis de ambiente extraídas do código
  decisoes/                  ← ADRs (Architecture Decision Records)
  legado/                    ← Arquivos históricos (não usar como referência)
  meta/                      ← Plano de migração e auditoria
api/
  openapi.yaml               ← Fonte única do contrato da API (OpenAPI 3.1)
docs-site/                   ← Site Astro Starlight (build com `mise run docs:sync`)
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

O site usa [Astro Starlight](https://starlight.astro.build/) e vive em `docs-site/`.
Para rodar localmente:

```bash
cd docs-site
npm install
npm run dev
```
