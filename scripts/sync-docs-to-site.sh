#!/usr/bin/env bash
# Sincroniza docs/ para docs-site/docs/ (Rspress)
# Mapeia nomes de arquivo para slugs sem prefixo numérico.

set -euo pipefail

DEST="docs-site/docs"
rm -rf "$DEST"
mkdir -p "$DEST/decisoes" "$DEST/public"

# Copia favicon
cp web/static/favicon.svg "$DEST/public/favicon.svg" 2>/dev/null || true

# Copia docs principais (remove prefixo numérico do nome)
cp docs/01-visao-e-negocio.md "$DEST/visao-e-negocio.md"
cp docs/02-arquitetura.md "$DEST/arquitetura.md"
cp docs/03-fluxos-e-modelo.md "$DEST/fluxos-e-modelo.md"
cp docs/04-operacao-shopee.md "$DEST/operacao-shopee.md"
cp docs/05-manual-do-usuario.md "$DEST/manual-do-usuario.md"
cp docs/06-qualidade-e-testes.md "$DEST/qualidade-e-testes.md"
cp docs/07-dados-e-ia.md "$DEST/dados-e-ia.md"

# Copia ADRs relevantes (se existirem)
for adr in docs/decisoes/0012-*.md docs/decisoes/0013-*.md docs/decisoes/0014-*.md docs/decisoes/0015-*.md; do
  [ -f "$adr" ] && cp "$adr" "$DEST/decisoes/"
done

# Index page
cat > "$DEST/index.md" << 'EOF'
# Garimpei

Plataforma de curadoria e publicação automatizada para afiliados Shopee.

## O que é

O Garimpei busca produtos na API de afiliados da Shopee, ranqueia por potencial
de retorno (comissão × demanda × avaliação), monitora variações de preço, e
publica ofertas em canais (Telegram, WhatsApp) — tudo multi-tenant com
rastreamento de conversão.

## Navegação

- [Visão e Negócio](/visao-e-negocio) — O quê e pra quem
- [Arquitetura](/arquitetura) — Como roda (Cloud Run, gRPC, BigQuery)
- [Fluxos e Modelo](/fluxos-e-modelo) — Entidades e regras de negócio
- [Qualidade e Testes](/qualidade-e-testes) — CI, fitness functions, drift checks
- [Decisões (ADRs)](/decisoes/0015-multi-tenant-credenciais) — Registro de decisões arquiteturais

## Links

- **App:** https://garimpei.app.br
- **Repo:** https://github.com/fmarquesfilho/garimpo
- **API Spec:** [OpenAPI 3.1](/api)
EOF

# API page (OpenAPI reference)
cat > "$DEST/api.md" << 'EOF'
# API Reference

A especificação OpenAPI 3.1 completa da API do Garimpei está em
[`api/openapi.yaml`](https://github.com/fmarquesfilho/garimpo/blob/main/api/openapi.yaml).

## Endpoints principais

| Grupo | Base | Descrição |
|-------|------|-----------|
| Curadoria | `/api/candidatos`, `/api/v2/curadoria/*` | Busca + ranking |
| Buscas/Lojas | `/api/buscas`, `/api/lojas` | CRUD monitoramento |
| Publicação | `/api/publicar`, `/api/publicacoes` | Envio para canais |
| Favoritos | `/api/favoritos` | Produtos salvos |
| Destinos | `/api/destinos` | Canais (Telegram/WhatsApp) |
| Templates | `/api/templates` | Modelos de mensagem |
| Alertas | `/api/alertas` | Configuração + teste |
| Onboarding | `/api/onboarding/*` | Configuração multi-tenant |
| Analytics | `/api/conversoes`, `/api/estatisticas`, `/api/coletas` | Dados |
| Admin | `/api/admin/me`, `/api/health` | Status |

## Autenticação

Todos os endpoints (exceto `/api/health` e `/api/candidatos`) requerem
Firebase Auth JWT no header `Authorization: Bearer <token>`.
EOF

# Remove referências a páginas /docs/* que não existem no site (dead links)
for f in "$DEST"/*.md "$DEST"/decisoes/*.md; do
  [ -f "$f" ] || continue
  # Remove links markdown com /docs/ path (mantém o texto visível)
  sed -i '' 's|\[\([^]]*\)\](/docs/[^)]*)|[\1](#)|g' "$f"
done

echo "✓ Docs sincronizados para $DEST"
