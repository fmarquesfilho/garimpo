# Pacote de Documentação v2 — Garimpei

> **Para:** Fernando (mantenedor) e a IA assistente que co-mantém o projeto.
> **Gerado em:** 2026-06-27 · **Versão do pacote:** docs-v2.0.0

Este pacote é a base para reorganizar e modernizar a documentação do Garimpei.
Ele **não** substitui o código nem o `docs/` atual de imediato: é um plano de
migração + artefatos novos para entrar no controle de versão e ser executado
incrementalmente.

## O que tem aqui

| Arquivo | Responde ao pedido | O que é |
|---|---|---|
| `00_LEIA-ME.md` | (3) | Este guia + instruções para a IA + changelog |
| `01_INCONSISTENCIAS.md` | (1) | Auditoria de inconsistências entre os 18 documentos, com severidade e correção sugerida |
| `02_PLANO_DOCUMENTACAO.md` | (2) e (3) | Nova estrutura de docs (de 18 → 7 arquivos + site), ferramenta escolhida (Starlight + OpenAPI/Scalar + geração automática de ER e API), e como plugar na CI |
| `03_BACKLOG_COMO_CODIGO.md` | (5) | Backlog versionado em YAML + gerador de quadro Kanban a cada build |
| `04_NEGOCIO_E_RELEASES.md` | (4) | Técnicas de business plan SaaS, plano de releases (Now/Next/Later), e roteiro de multi-tenant + LGPD |
| `05_CIENCIA_DE_DADOS_E_IA.md` | (6) | Roadmap de análises sobre os dados coletados + desenho de uma IA integrada ao produto |

Leia na ordem. O `01` e o `02` são pré-requisitos dos demais (definem a verdade
canônica e onde cada coisa passa a morar).

---

## Instruções para a IA que co-mantém o projeto

Estas regras valem para qualquer sessão de trabalho na documentação a partir de agora.

### Princípios

1. **Fonte única de verdade (SSOT) por tipo de informação.** Cada fato vive em
   exatamente um lugar e é referenciado (link) dos outros. Em particular:
   - **Contrato de API** → `api/openapi.yaml` (já existe, 3.1, ~40 endpoints). Nunca
     descreva endpoints em prosa em outro doc; **linke** para a referência gerada.
   - **Modelo de dados** → gerado a partir do schema Go/BigQuery (ver `02`). O
     `ENTIDADES.md` atual vira **saída gerada**, não fonte editada à mão.
   - **Estado do trabalho** (backlog/roadmap) → `backlog/*.yaml` (ver `03`). Não
     duplicar status de tarefa em mais nenhum `.md`.
2. **Documento não repete código.** Se algo pode ser gerado (ER, API, lista de
   variáveis de ambiente, lista de tabelas), gere — não transcreva.
3. **Nada de "data-bomba".** Evite datas absolutas em docs conceituais; use
   horizontes (Now/Next/Later) e deixe datas só no changelog e no backlog.
4. **Português técnico, sem jargão inglês desnecessário** (preferência do Fernando:
   "Segurança" em vez de "Hardening", "token original" em vez de "token cru").
5. **Marque o que é aspiracional.** Tudo que descreve algo ainda não implementado
   leva um selo `> 🔮 Planejado` no topo da seção. Hoje há confusão entre o que
   existe e o que é visão (ver `01`, itens de Instagram e Kanban).
6. **Toda mudança estrutural de docs entra no changelog deste arquivo** (seção
   final), com data e motivo.

### Fluxo de trabalho ao mexer na documentação

```
1. Ler 01_INCONSISTENCIAS.md → conferir se a mudança resolve/cria inconsistência
2. Editar a FONTE (código anotado, openapi.yaml, ou .md conceitual) — nunca a saída gerada
3. Rodar `make docs` localmente (ver 02) → conferir que ER/API/board regeneram
4. Atualizar o changelog (este arquivo) se a estrutura mudou
5. Abrir PR; a CI valida (openapi lint, links quebrados, drift de schema, file-size)
```

### Regras que não mudam (herdadas do projeto)

- Máx. **400 linhas** por arquivo de produção (CI bloqueia). Docs não têm limite,
  mas prefira arquivos coesos.
- Go template em GitHub Classroom usa o module path da turma — **não aplicável aqui**,
  é regra de outro contexto do Fernando; ignore no Garimpei.
- Não comitar segredos. `SHOPEE_*`, `TELEGRAM_*`, `ENCRYPTION_KEY` vivem no Secret
  Manager / `garimpo.env`, nunca no repo nem na doc.

---

## Decisões canônicas (resolvem ambiguidades hoje espalhadas)

Estas decisões **encerram** discussões que aparecem inconsistentes nos 18 docs.
Trate-as como verdade; o `01` lista onde cada doc precisa ser ajustado.

| Tema | Decisão canônica |
|---|---|
| **Nome do produto** | **Garimpei** (produto/marca, domínio `garimpei.app.br`). O binário, o serviço Cloud Run e o dataset BigQuery permanecem `garimpo`/`garimpo-api` por compatibilidade. Docs antigos que dizem "Garimpo" como nome do produto devem ser atualizados. |
| **Página principal** | **Descobrir** (`/`). Unifica a antiga Busca + Oportunidades. "Curadoria"/"Buscar"/"Oportunidades" são nomes legados. |
| **Estratégias de ranking** | Só **nicho** está ativa. "Diversificada" foi descontinuada da UI e do service; o código do Strategy pattern permanece como dívida técnica documentada, não como feature. |
| **Quadro Kanban (do produto)** | **Removido** (não usado pela Mileny). Não é feature. |
| **Canal de publicação** | **Telegram + WhatsApp**. Instagram é **visão futura** (não implementado) — toda menção a Instagram como canal ativo é aspiracional. |
| **Deploy** | **GCP (Cloud Run + BigQuery)** é a arquitetura real. O `DEPLOY.md` (OCI + nginx + Postgres) está **obsoleto** e descreve um caminho não seguido. |
| **Persistência de favoritos/buscas** | localStorage (imediato) **+** sync BigQuery (servidor). |
| **Categorias** | Campo é **plural** (`categorias[]`), filtro opcional por OR. O `categoria` singular no ENTIDADES está defasado. |
| **Alertas de "produtos novos"** | Implementados no backend, mas **desabilitados** por ora (aguardando config por usuário). |

---

## Changelog da documentação

Mantenha no topo o mais recente. Formato: `versão — data — autor — resumo`.

### docs-v2.0.0 — 2026-06-27 — Fernando + IA
- **Reestruturação proposta**: de 18 arquivos soltos em `docs/` para 7 documentos
  canônicos + site gerado (ver `02_PLANO_DOCUMENTACAO.md`).
- **Auditoria de inconsistências** registrada em `01_INCONSISTENCIAS.md` (12 itens).
- **Decisões canônicas** acima encerram ambiguidades de nome, página principal,
  estratégias, canal, deploy e modelo de categorias.
- **Backlog vira código**: introdução de `backlog/*.yaml` + gerador de Kanban (`03`).
- **Plano de negócio e releases** (`04`) e **roadmap de dados + IA** (`05`)
  adicionados como documentos vivos.
- **API e modelo de dados** passam a ser documentação gerada, não escrita à mão.

> A partir daqui, cada PR que alterar a estrutura da documentação adiciona uma
> linha de changelog. Mudanças de conteúdo dentro de um doc não precisam de entrada.
