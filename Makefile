# Makefile — Garimpei (mono-repo: Go + C# + protos)

.PHONY: docs docs-api docs-er docs-env docs-site docs-check test lint build \
        proto up down test-go test-csharp build-go build-csharp

# ─── Documentação ───────────────────────────────────────────────

docs: docs-api docs-er docs-env docs-board docs-sync docs-site ## Gera toda a documentação

docs-api: ## Renderiza openapi.yaml → HTML com Scalar
	npx @scalar/cli bundle api/openapi.yaml -o docs/gerado/api.html

docs-er: ## Gera o Mermaid ER do schema BigQuery
	go run ./cmd/gen-er > docs/gerado/ENTIDADES.md

docs-env: ## Extrai variáveis de ambiente referenciadas no código
	./scripts/gen-env-doc.sh > docs/gerado/env-vars.md

docs-board: ## Gera quadro Kanban e roadmap do backlog
	go run ./cmd/gen-board

docs-sync: ## Sincroniza docs canônicos para docs-site
	./scripts/sync-docs-to-site.sh

docs-site: docs-sync ## Build do site Starlight (sync + build)
	cd docs-site && npm run build

docs-publish: docs docs-site ## Gera, sincroniza e builda — pronto para commit

docs-check: ## CI: falha se docs geradas estiverem desatualizados
	$(MAKE) docs-er docs-env docs-board
	git diff --exit-code docs/gerado || (echo "❌ Docs geradas desatualizadas: rode 'make docs'"; exit 1)

# ─── Proto (gRPC) ──────────────────────────────────────────────

proto: ## Gera código Go e C# a partir dos .proto
	cd protos && buf generate
	@echo "✅ Proto gerados em gen/go/ e src/Garimpei.Protos/Generated/"

proto-lint: ## Lint nos .proto files
	cd protos && buf lint

proto-breaking: ## Verifica breaking changes nos .proto
	cd protos && buf breaking --against '.git#subdir=protos'

# ─── Build ──────────────────────────────────────────────────────

build: build-go build-csharp ## Build de tudo

build-go: ## Build do monólito Go + microserviços
	go build ./...

build-csharp: ## Build da solution C#
	cd src && dotnet build --no-restore

restore-csharp: ## Restore NuGet packages
	cd src && dotnet restore

# ─── Testes ─────────────────────────────────────────────────────

test: test-go test-csharp ## Roda todos os testes

test-go: ## Testes Go
	go test ./...

test-csharp: ## Testes C#
	cd src && dotnet test --no-build

# ─── Lint ───────────────────────────────────────────────────────

lint: lint-go lint-web ## Roda linters

lint-go: ## Lint Go
	golangci-lint run ./...

lint-web: ## Lint frontend
	cd web && npx eslint .

# ─── Docker Compose (dev local) ────────────────────────────────

up: ## Sobe todos os serviços (dev local)
	docker compose up -d

down: ## Para todos os serviços
	docker compose down

logs: ## Logs de todos os containers
	docker compose logs -f

ps: ## Status dos containers
	docker compose ps

up-deps: ## Sobe apenas dependências (PG + BQ emulator)
	docker compose up -d postgres bigquery-emulator

# ─── Ajuda ──────────────────────────────────────────────────────

help: ## Mostra esta ajuda
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
