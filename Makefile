# Makefile — Garimpei
# Alvos de documentação, build e deploy

.PHONY: docs docs-api docs-er docs-env docs-site docs-check test lint

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

# ─── Desenvolvimento ────────────────────────────────────────────

test: ## Roda todos os testes
	go test ./...
	cd web && npx vitest --run

lint: ## Roda linters
	golangci-lint run ./...
	cd web && npx eslint .

build: ## Build da imagem Docker
	docker build -t garimpo-api .

# ─── Ajuda ──────────────────────────────────────────────────────

help: ## Mostra esta ajuda
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'
