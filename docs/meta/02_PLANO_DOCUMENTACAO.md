# Plano de Documentação v2 — Estrutura, Ferramenta e Geração Automática

Objetivo: sair de **18 arquivos soltos** em `docs/` para **7 documentos canônicos
editados à mão + 3 artefatos gerados automaticamente**, publicados como um site de
documentação que se atualiza a cada deploy.

---

## 1. Ferramenta escolhida

### Site de documentação: **Astro Starlight**

Por que Starlight e não as alternativas:

| Opção | Veredito |
|---|---|
| **Astro Starlight** ✅ | Fica no ecossistema **Node/JS** que você já roda (SvelteKit, Vite, npm na CI). Markdown/MDX, busca local, versionamento, dark mode, zero runtime pesado. Build estático que o Cloud Run já serve como qualquer asset. |
| Docusaurus | Ótimo, mas traz React e é mais "pesado" do que você precisa para ~7 docs. |
| MkDocs Material | Excelente, porém **adiciona Python** a uma stack Go+JS — mais uma toolchain na CI. |
| VitePress | Bom e leve (Vue); Starlight só ganha por integração e plugins de docs (ex.: Scalar). |

> Decisão: **Starlight**. Mantém uma toolchain só (Node), build estático,
> e integra direto com o renderizador de OpenAPI.

### Referência de API: **OpenAPI 3.1 (já existe) + Scalar**

Você já tem `openapi.yaml` (≈40 endpoints) e Swagger UI em `/api/docs`. Eleve isso:

- **O `openapi.yaml` é a fonte única do contrato de API.** Nenhum endpoint é
  descrito em prosa em outro lugar — os docs **linkam** para a referência.
- **Renderização:** trocar/duplicar o Swagger UI por **Scalar** (`@scalar/api-reference`),
  que tem UX melhor e plugin para Starlight (`starlight-openapi` ou embed do Scalar).
  Mantém também o Swagger UI em `/api/docs` se quiser (custo zero).
- **Cliente tipado de graça:** gerar um client TypeScript do `openapi.yaml` para o
  SvelteKit com **`openapi-typescript`** (tipos) + `openapi-fetch` (client). Isso
  elimina a divergência entre o que o backend devolve e o que o front espera —
  uma fonte recorrente dos seus bugs (ex.: `erro` vs `detail` no Problem Details).
- **Validação na CI:** `redocly lint openapi.yaml` (ou `spectral lint`) bloqueia
  spec inválido. Opcional: teste de contrato que valida respostas reais do servidor
  contra o schema.

### Modelo de dados: **Mermaid ER gerado do schema**

Hoje o `ENTIDADES.md` é escrito à mão e já divergiu (I6, I7). Inverta:

- A **fonte** vira o schema real. Você tem duas candidatas a SSOT:
  1. `deploy/bigquery_schema.sql` (DDL), ou
  2. As structs Go + `EnsureSchema` (que evolui o schema em runtime).
- Escreva um pequeno **gerador** (`cmd/gen-er`) que lê a SSOT e emite
  `docs/gerado/ENTIDADES.md` com o bloco Mermaid `erDiagram`. As **regras de
  negócio** (que hoje estão no ENTIDADES) ficam num arquivo conceitual separado e
  estável; só o diagrama é gerado.
- A CI roda o gerador e **falha se o arquivo gerado estiver desatualizado**
  (`git diff --exit-code`). Assim, mudou o schema → tem que regenerar o ER no mesmo
  PR. Acaba o drift por construção.

> Resumo da estratégia: **tudo que pode ser gerado, é gerado** (API, ER, lista de
> env vars, números de teste). O que sobra para escrever à mão é só o conceitual.

---

## 2. Nova estrutura de arquivos

```
docs/
  README.md                 ← índice + "comece por aqui" (aponta para o site)
  01-visao-e-negocio.md      ← O QUÊ e PRA QUEM (de: MODELO, JORNADA, parte de APIS §2)
  02-arquitetura.md          ← COMO roda (de: DEPLOY_GCP, COLETA, ANALISE_ESTATICA, REFATORACAO)
  03-fluxos-e-modelo.md       ← entidades + buscas agendadas + regras (de: ENTIDADES, BUSCAS_AGENDADAS)
  04-operacao-shopee.md       ← integração Shopee + origem (de: APIS §1, SHOPEE_INTROSPECT_RESULT)
  05-manual-do-usuario.md     ← como a Mileny usa (de: MANUAL)
  06-qualidade-e-testes.md    ← CI, linters, BDD, cenários (de: ANALISE_ESTATICA, BDD_STRATEGY, TESTES_DESCOBRIR)
  07-dados-e-ia.md           ← análises + IA integrada (de: CIENCIA_DE_DADOS + 05_CIENCIA_DE_DADOS_E_IA.md)
  gerado/
    api.html                 ← Scalar/Redoc a partir de openapi.yaml (não editar)
    ENTIDADES.md             ← Mermaid ER gerado do schema (não editar)
    env-vars.md              ← lista de variáveis extraída do código (não editar)
  legado/
    DEPLOY.md                ← OCI, arquivado com aviso no topo
    MELHORIAS_28_06.md       ← vira backlog (ver 03); manter como registro histórico
  decisoes/
    0001-nome-garimpei.md    ← ADRs curtos (Architecture Decision Records)
    0002-so-nicho.md
    0003-deploy-gcp.md
    ...
api/
  openapi.yaml               ← SSOT do contrato (movido de docs/ para a raiz da API)
backlog/                     ← ver 03_BACKLOG_COMO_CODIGO.md
```

### Mapa de migração (de → para)

| Doc atual | Vai para | Observação |
|---|---|---|
| `MODELO.md` | `01-visao-e-negocio.md` | rebaixar diversificada/Instagram/Kanban a 🔮 |
| `JORNADA.md` | `01-visao-e-negocio.md` | mesclar como "jornada do usuário" |
| `APIS.md` §2 (Instagram) | `01-visao-e-negocio.md` (visão) | marcar 🔮 |
| `APIS.md` §1 (Shopee) | `04-operacao-shopee.md` | parte técnica |
| `SHOPEE_INTROSPECT_RESULT.md` | `04-operacao-shopee.md` | apêndice "origem do produto" |
| `DEPLOY_GCP.md` | `02-arquitetura.md` | vira o runbook único |
| `DEPLOY.md` | `legado/DEPLOY.md` | arquivar (I5) |
| `COLETA.md` | `02-arquitetura.md` | seção "coleta e scheduler" |
| `REFATORACAO.md` | `02-arquitetura.md` + `06-qualidade` | princípios + limites |
| `ANALISE_ESTATICA.md` | `06-qualidade-e-testes.md` | + remover números à mão (I11) |
| `ENTIDADES.md` | `03-fluxos-e-modelo.md` (regras) + `gerado/ENTIDADES.md` (ER) | dividir conceito vs diagrama |
| `BUSCAS_AGENDADAS.md` | `03-fluxos-e-modelo.md` | já é excelente; vira a referência de buscas |
| `BDD_STRATEGY.md` | `06-qualidade-e-testes.md` | estratégia de testes |
| `TESTES_DESCOBRIR.md` | `06-qualidade-e-testes.md` | checklist de cenários (ou mover para `specs/`) |
| `MANUAL.md` | `05-manual-do-usuario.md` | corrigir Quadro/novos (I3, I9, I10) |
| `CIENCIA_DE_DADOS.md` | `07-dados-e-ia.md` | + conteúdo de `05_CIENCIA_DE_DADOS_E_IA.md` |
| `MELHORIAS_28_06.md` | `backlog/` + `legado/` | itens viram tarefas |
| `openapi.yaml` | `api/openapi.yaml` | promovido a SSOT, renderizado em `gerado/api.html` |

### ADRs (Architecture Decision Records)

As "decisões canônicas" do `00_LEIA-ME.md` viram ADRs curtos em `docs/decisoes/`.
Cada ADR: **contexto → decisão → consequências**, ~15 linhas. Isso dá rastreável o
*porquê* (ex.: por que só nicho, por que GCP e não OCI) sem poluir os docs vivos.

---

## 3. Integração com a CI/build (geração automática)

Adicione um alvo `mise run docs` e um job na pipeline (que hoje já roda
go test → lint → vitest → playwright):

```makefile
# Makefile (trecho)
docs: docs-api docs-er docs-env docs-board docs-site

docs-api:    ## renderiza openapi.yaml -> docs/gerado/api.html
	npx @scalar/cli document bundle api/openapi.yaml -o docs/gerado/api.html

docs-er:     ## gera o Mermaid ER do schema
	go run ./cmd/gen-er > docs/gerado/ENTIDADES.md

docs-env:    ## extrai variáveis de ambiente referenciadas no código
	./.mise/tasks/docs/env > docs/gerado/env-vars.md

docs-board:  ## gera o quadro Kanban do backlog (ver 03)
	go run ./cmd/gen-board

docs-site:   ## build do site Starlight
	cd docs-site && npm run build

docs-check:  ## CI: falha se algo gerado estiver desatualizado
	$(MAKE) docs-api docs-er docs-env
	git diff --exit-code docs/gerado || (echo "Docs geradas desatualizadas: rode 'mise run docs'"; exit 1)
```

Na CI (`deploy-gcp.yml`), depois dos testes:

```yaml
  docs:
    needs: [test-go, test-web]
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Validar contrato OpenAPI
        run: npx @redocly/cli lint api/openapi.yaml
      - name: Conferir docs geradas (sem drift)
        run: mise run docs:check
      - name: Verificar links quebrados
        run: npx lychee --no-progress 'docs/**/*.md'
      - name: Build do site
        run: mise run docs-site
      - name: Publicar (GitHub Pages ou /docs no Cloud Run)
        run: ./scripts/publish-docs.sh
```

**Onde publicar o site:** duas opções, ambas baratas:
- **GitHub Pages** (separado do app) — mais simples, isola docs do runtime.
- **Rota `/docs` no Cloud Run** — serve o `docs-site/dist` como asset estático,
  mesma origem do app. Bom se quiser docs autenticadas no futuro.

---

## 4. Esqueleto mínimo do Starlight

`docs-site/astro.config.mjs`:

```js
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

export default defineConfig({
  integrations: [
    starlight({
      title: 'Garimpei — Documentação',
      locales: { root: { label: 'Português', lang: 'pt-BR' } },
      sidebar: [
        { label: 'Comece aqui', link: '/' },
        { label: 'Visão e negócio', link: '/01-visao-e-negocio/' },
        { label: 'Arquitetura', link: '/02-arquitetura/' },
        { label: 'Fluxos e modelo', link: '/03-fluxos-e-modelo/' },
        { label: 'Integração Shopee', link: '/04-operacao-shopee/' },
        { label: 'Manual do usuário', link: '/05-manual-do-usuario/' },
        { label: 'Qualidade e testes', link: '/06-qualidade-e-testes/' },
        { label: 'Dados e IA', link: '/07-dados-e-ia/' },
        { label: 'Referência da API', link: '/gerado/api/' },
        { label: 'Modelo de dados', link: '/gerado/entidades/' },
        { label: 'Decisões (ADRs)', autogenerate: { directory: 'decisoes' } },
      ],
    }),
  ],
});
```

> O Mermaid renderiza nativamente no Starlight (via `rehype-mermaid` ou
> `expressive-code`); seus blocos ` ```mermaid ` continuam funcionando.

---

## 5. Ordem de execução da migração

Trate como o epic `docs-migration` no backlog. Sugestão de fatiamento:

1. **Fundação (Now):** criar `docs-site/` Starlight + mover `openapi.yaml` para
   `api/` + renderizar com Scalar. Entregável testável: site sobe com a referência
   de API funcionando.
2. **Geração de ER + env (Now):** escrever `cmd/gen-er` e `gen-env-doc.sh`; ligar o
   `docs-check` na CI. Resolve I6, I7, I11 por construção.
3. **Consolidação conceitual (Next):** mesclar os 18 docs nos 7 canônicos seguindo o
   mapa, resolvendo as inconsistências 🔴 de cada arquivo ao tocá-lo.
4. **ADRs + legado (Next):** extrair as decisões canônicas como ADRs; arquivar
   `DEPLOY.md`.
5. **Backlog-como-código (Next):** ver `03`.
6. **Polimento (Later):** busca, versionamento de docs (quando lançar v1 público),
   docs em `/docs` autenticadas se necessário.

Cada passo é um PR pequeno e verificável — o mesmo princípio incremental que você
já usa no produto.
