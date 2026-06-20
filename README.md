# Garimpo

Curadoria automática de produtos de afiliado: pega candidatos (CSV hoje, API da
Shopee amanhã), aplica o piso de comissão, pontua por estratégia e devolve a
lista priorizada do dia. Ataca o gargalo da operação — a escolha manual e diária
do produto a anunciar.

## Rodar

```bash
# lista do dia pela estratégia de nicho (padrão)
go run ./cmd/garimpo

# estratégia diversificada, top 8
go run ./cmd/garimpo -estrategia diversificada -top 8

# subir o piso de comissão para 10%
go run ./cmd/garimpo -comissao-min 0.10

# usar outro CSV
go run ./cmd/garimpo -csv data/candidatos_exemplo.csv

# fonte ao vivo: API de afiliados da Shopee (precisa de credenciais)
export SHOPEE_APP_ID=...
export SHOPEE_SECRET=...
go run ./cmd/garimpo -fonte shopee -cat 100017 -categoria "cosméticos"
```

Veja `docs/APIS.md` para o mapeamento dos campos da Shopee e os limites do
Instagram, e `docs/MODELO.md` para o modelo de negócio e o roadmap incremental.

## Frontend (SvelteKit) — Incremento 2

A interface que ela usa fica em `web/`. Consome a API HTTP do Garimpo.

```bash
# 1. suba a API (a fonte é definida aqui; os filtros vêm do front)
go run ./cmd/garimpo-api -fonte shopee      # ou sem flag, para o CSV de exemplo

# 2. em outro terminal, rode o frontend
cd web
npm install          # OBRIGATÓRIO na primeira vez (baixa as dependências)
npm run dev          # abre em http://localhost:5173
```

A API base do frontend é configurável: `VITE_API_BASE=http://localhost:8080`.

Os controles de **busca (keyword), categoria, comissão, vendas mínimas e nota
mínima** ficam na própria interface — não precisam de flags. As flags do
`garimpo-api` (`-keyword`, `-categoria`, `-vendas-min`, `-nota-min`) servem só
como padrões iniciais; o front sobrescreve a cada requisição.

Telas:
- **Curadoria** — a peneira do dia: busca na Shopee + filtros, short-list
  ranqueada por "teor", alternador de estratégia (nicho / diversificada /
  comparar), e o botão **Garimpar** que manda o produto pro quadro.
- **Quadro** — Kanban da operação (Selecionados → Em produção → Publicado →
  Em análise) com limites de WIP, persistido no navegador.

## Testes

```bash
go test ./...
```

## Estrutura (ports & adapters)

```
cmd/garimpo        CLI (composição: escolhe fonte + estratégia)
cmd/garimpo-api    servidor HTTP que serve a curadoria em JSON
internal/domain    núcleo: Product, Scored (sem dependências)
internal/httpapi   handlers HTTP (CORS) sobre o engine
internal/source    PORTA de entrada: ProductSource
                     - csv.go     adaptador que funciona hoje
                     - shopee.go  adaptador da API de afiliados (GraphQL, ✅ implementado)
                     - flex.go    tipos que aceitam número OU string no JSON da Shopee
internal/scoring   matemática neutra: valor esperado + normalização
internal/strategy  PORTA de decisão: Strategy
                     - niche.go        prioriza nicho + comissão + avaliação
                     - diversified.go  persegue valor esperado/volume
internal/engine    orquestra: fonte -> elegibilidade -> scoring -> ranking
data/              CSV de exemplo
web/               frontend SvelteKit (Incremento 2)
docs/MODELO.md     modelo de negócio + roadmap incremental + Kanban
docs/APIS.md       referência das APIs Shopee (campos) e Instagram (limites)
```

O motor depende só das duas portas. Trocar CSV pela API, ou nicho por
diversificada, não toca em mais nada — é onde a prova de conceito vira produto.

## CSV de entrada

Cabeçalho: `id,name,category,price,commission,sales_30d,rating`
`commission` em fração (0.12 = 12%). `sales_30d` é proxy de demanda — se a API
não expuser vendas, use a posição no ranking de best-sellers como proxy.
