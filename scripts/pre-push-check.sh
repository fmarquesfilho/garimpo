#!/usr/bin/env bash
# =============================================================================
# pre-push-check.sh — Roda verificações pré-push via mise
# =============================================================================
# Instalar como git hook:
#   ln -sf ../../scripts/pre-push-check.sh .git/hooks/pre-push
# =============================================================================

set -euo pipefail

if ! command -v mise &> /dev/null; then
    echo "⚠️  mise não instalado. Instale: brew install mise"
    echo "   Rodando checks diretamente..."
    # Fallback: roda os file tasks diretamente
    .mise/tasks/check/service-contracts
    .mise/tasks/check/api-contract
    .mise/tasks/check/config-consistency
    .mise/tasks/check/schema-sync
    .mise/tasks/check/data-ownership
    .mise/tasks/check/stale-refs
    exit $?
fi

exec mise run prepush
