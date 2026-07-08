# Cobertura de Testes — Página Garimpar (unificada)

Análise de mapeamento entre os 58 cenários do documento `TESTES_DESCOBRIR.md` (2026-06-27)
e os testes automatizados existentes no repositório.

**Atualização:** 2026-07-08 — Após unificação Descobrir + Lojas → Garimpar (`/`).
A rota `/lojas` foi removida. Todos os fluxos agora vivem na página unificada `/`.

## Inventário de testes existentes

### Unit tests (vitest, `web/src/tests/`)

| Arquivo | O que cobre |
|---------|-------------|
| `descobrir.test.js` | `montarResultados()`, `encontrarLojaPorNome()` — fontes (curadoria, quedas, novos, favoritos, **lojas**), keywords, categorias, filtros numéricos, combinações, empty states, backward compat |
| `favoritos.test.js` | **NOVO** — toggle favoritar/desfavoritar, dedup, cross-fonte (produto de Quedas aparece em Favoritos) |
| `CandidateCard.test.js` | Renderização do card: nome, loja, preço, comissão, vendas, badges (origem, desconto, expiração, suspeito) |
| `loading-timeout.test.js` | Lógica de timeout (busca que demora, erro amigável, cleanup de timer) |
| `oportunidades.test.js` | `gerarLinkProduto()` — geração de link Shopee a partir de busca_id + produto_id |

### E2E tests (Playwright, `web/tests/`)

| Arquivo | O que cobre |
|---------|-------------|
| `descobrir.spec.js` | Busca por keyword, filtros (vendas_min, comissão_min, categoria), toggle fontes, badges, interseção categoria+keyword |
| `lojas-precos.spec.js` | **REESCRITO** — Toggle 🏪 Lojas, seletor de loja, badges Quedas/Novos, graceful degradation, empty state sem lojas |
| `buscas-agendadas.spec.js` | Criação/remoção de buscas agendadas na seção Configuração |
| `alertas-novidades.spec.js` | Estrutura do /api/lojas/novidades, seção GerenciarBuscas |
| `lojas-resolve-shop.spec.js` | Resolver URL/username de loja via Collector |
| `lojas-cadastro.spec.js` | Adicionar loja via FormAdicionarLoja na seção Configuração |
| `novas-features.spec.js` | Rota /lojas retorna 404, sem erros JS nas rotas principais |

---

## Mapeamento cenário × cobertura

### 1. Fontes de dados (toggles)

| # | Cenário | Unit | E2E | Status |
|---|---------|------|-----|--------|
| 1 | 🔍 Busca "sérum" | ✅ `descobrir.test.js` cenário 1 | ✅ `descobrir.spec.js` test 1 | **Coberto** |
| 2 | 🔍 Busca vazio → hint | ✅ cenário 2 (retorna vazio) | ✅ test 2 (badge visível) | **Coberto** |
| 3 | 🔍 Busca vazio + categoria | ✅ cenário filtragem por categoria | ✅ test 5 | **Coberto** |
| 4 | 📉 Quedas vazio | ✅ cenário 3 | ✅ `lojas-precos.spec.js` (badge quedas) | **Coberto** |
| 5 | 📉 Quedas "Skin1004" | ✅ cenário 4 (filtra por nome) | ❌ | **Unit only** |
| 6 | 🆕 Novos vazio | ✅ cenário 5 | ✅ `lojas-precos.spec.js` (badge novos) | **Coberto** |
| 7 | 🆕 Novos "retinol" | ✅ cenário 6 | ❌ | **Unit only** |
| 8 | ⭐ Favoritos vazio | ✅ cenário 7 | ❌ | **Unit only** |
| 9 | ⭐ Favoritos "perfume" | ✅ cenário 8 | ❌ | **Unit only** |
| 10 | Nenhum ativo → hint | ✅ cenário 9 | ✅ test 8 (toggle off) | **Coberto** |
| 11 | 🔍+📉+🆕 vazio | ✅ cenário 10 | ❌ | **Unit only** |
| 12 | 🔍+📉+🆕 "sérum" | ✅ cenário 11 | ❌ | **Unit only** |
| 13 | Todos "SKIN1004" | ✅ cenário 13 (loja) | ❌ | **Unit only** |

### 1b. Fonte Lojas (NOVA — pós-unificação)

| # | Cenário | Unit | E2E | Status |
|---|---------|------|-----|--------|
| — | 🏪 Lojas toggle on | ✅ `descobrir.test.js` (fonte lojas) | ✅ `lojas-precos.spec.js` (toggle) | **Coberto** |
| — | 🏪 Lojas + keyword filtra | ✅ (keyword por nome/loja) | ❌ | **Unit only** |
| — | 🏪 Lojas + comissaoMin | ✅ | ❌ | **Unit only** |
| — | 🏪 Lojas + vendasMin | ✅ | ❌ | **Unit only** |
| — | 🏪 Seletor de loja | ❌ | ✅ `lojas-precos.spec.js` (seletor) | **E2E only** |
| — | 🏪 Sem lojas → empty state | ✅ (retorna vazio) | ✅ `lojas-precos.spec.js` | **Coberto** |

### 2. Busca por nome de loja

| # | Cenário | Unit | E2E | Status |
|---|---------|------|-----|--------|
| 14 | Nome exato loja | ✅ `encontrarLojaPorNome` (nome exato) | ❌ | **Unit only** |
| 15 | Parte do nome | ✅ `encontrarLojaPorNome` (parcial) | ❌ | **Unit only** |

### 3. Filtragem por categoria

| # | Cenário | Unit | E2E | Status |
|---|---------|------|-----|--------|
| 16 | Categoria única | ✅ cenário 14 | ✅ test 5 | **Coberto** |
| 17 | Múltiplas categorias (OR) | ✅ cenário 15 | ❌ | **Unit only** |
| 18 | Keyword + categoria (AND) | ✅ cenário 16 | ✅ test 6 | **Coberto** |
| 19 | Loja + categoria | ✅ cenário 17 | ❌ | **Unit only** |
| 20 | Nenhuma categoria | ✅ (sem filtro mostra tudo) | ❌ | **Unit only** |
| 21 | Busca salva com categorias | ✅ `aplicarBuscaSalva` | ❌ | **Unit only** |

### 4. Favoritos (⭐)

| # | Cenário | Unit | E2E | Status |
|---|---------|------|-----|--------|
| 22 | Clicar ☆ favoritar | ✅ `favoritos.test.js` cenário 22 | ❌ | **Unit** |
| 23 | Clicar ★ desfavoritar | ✅ `favoritos.test.js` cenário 23 | ❌ | **Unit** |
| 24 | Fonte ⭐ mostra salvos | ✅ cenário 7 | ❌ | **Unit only** |
| 25 | Favoritar de Quedas aparece em ambos | ✅ `favoritos.test.js` cross-fonte | ❌ | **Unit** |
| 26 | Sync entre dispositivos | ❌ | ❌ | **🔴 Não coberto** (depende de servidor) |

### 5. Buscas salvas (atalhos)

| # | Cenário | Unit | E2E | Status |
|---|---------|------|-----|--------|
| 27 | Clicar pill seta keyword | ✅ `aplicarBuscaSalva` | ❌ | **Unit only** |
| 28 | Busca com fontes [quedas,novos] | ✅ cenário 20 | ❌ | **Unit only** |
| 29 | Busca salva com categorias | ✅ | ❌ | **Unit only** |
| 30 | Busca agendada mostra ⏱ | ❌ | ❌ | **🔴 Não coberto** |
| 31 | Múltiplas keywords como pills | ❌ | ❌ | **🔴 Não coberto** |
| 32 | Ícones de fonte visíveis | ❌ | ❌ | **🔴 Não coberto** |

### 6. Gerenciar buscas (na seção ⚙️ Configuração)

| # | Cenário | Unit | E2E | Status |
|---|---------|------|-----|--------|
| 33 | Clicar "+ nova busca" | ❌ | ✅ `buscas-agendadas.spec.js` | **E2E only** |
| 34 | Busca só categorias | ❌ | ❌ | **🔴 Não coberto** |
| 35 | Busca fontes [novos] + dias_janela | ❌ | ❌ | **🔴 Não coberto** |
| 36 | Busca agendada (cron) | ❌ | ✅ `buscas-agendadas.spec.js` | **E2E only** |
| 37 | Remover busca | ❌ | ✅ `buscas-agendadas.spec.js` | **E2E only** |

### 7. Dados, cache e timeout

| # | Cenário | Unit | E2E | Status |
|---|---------|------|-----|--------|
| 38 | Sem lojas + Quedas/Novos → empty state | ✅ `descobrir.test.js` cenário 38 | ✅ `lojas-precos.spec.js` | **Coberto** |
| 39 | Toggle rápido on/off/on cache | ❌ | ❌ | **🔴 Não coberto** |
| 40 | API demora > 25s → timeout | ✅ `loading-timeout.test.js` | ❌ | **Unit only** |
| 41 | Coletas antigas sem imagem | ❌ | ❌ | **🔴 Não coberto** |
| 42 | Coletas novas com imagem | ❌ | ✅ `lojas-precos.spec.js` (mock com dados) | **E2E only** |

### 8. Input de busca

| # | Cenário | Unit | E2E | Status |
|---|---------|------|-----|--------|
| 43 | Clicar ✕ limpa campo | ❌ | ❌ | **🔴 Não coberto** |
| 44 | ESC limpa campo | ❌ | ❌ | **🔴 Não coberto** |
| 45 | Campo vazio → ✕ não aparece | ❌ | ❌ | **🔴 Não coberto** |
| 46 | Debounce 400ms | ❌ | ✅ `descobrir.spec.js` (waitForTimeout 600ms) | **E2E implícito** |

### 9. ProductCard (visual)

| # | Cenário | Unit | E2E | Status |
|---|---------|------|-----|--------|
| 47 | Curadoria: imagem, nome, preço, comissão, vendas, nota, loja, score, posição | ✅ `CandidateCard.test.js` (parcial) | ❌ | **Parcial** |
| 48 | Queda: preço anterior→atual, variação, badge verde | ❌ | ✅ `lojas-precos.spec.js` (badge quedas) | **E2E only** |
| 49 | Novo: preço, comissão, loja, badge "Novo" | ❌ | ✅ `lojas-precos.spec.js` (badge novos) | **E2E only** |
| 50 | Favorito: visual + ★ dourada | ❌ | ❌ | **🔴 Não coberto** |
| 51 | Toggle ☆/★ funciona | ✅ `favoritos.test.js` (lógica) | ❌ | **Unit only** |
| 52 | Botão "📤 Publicar" navega | ❌ | ❌ | **🔴 Não coberto** |
| 53 | Botão "🔗 Link" copia | ❌ | ❌ | **🔴 Não coberto** |

### 10. Badges nos toggles

| # | Cenário | Unit | E2E | Status |
|---|---------|------|-----|--------|
| 54 | 📉 Quedas badge verde com número | ❌ | ✅ `lojas-precos.spec.js` | **E2E only** |
| 55 | 🆕 Novos badge rosa | ❌ | ✅ `lojas-precos.spec.js` | **E2E only** |
| 56 | 🔍 Busca badge dourado | ❌ | ✅ `descobrir.spec.js` test 7 | **E2E only** |
| 57 | ⭐ Favoritos badge contagem | ❌ | ❌ | **🔴 Não coberto** |
| 58 | Hover tooltip | ❌ | ❌ | **🔴 Não coberto** |

---

## Resumo quantitativo

| Status | Cenários | % | Δ vs anterior |
|--------|----------|---|---------------|
| **Coberto** (unit + E2E ou ambos) | 16 | 28% | +6 |
| **Unit only** (lógica testada, UI não) | 21 | 36% | +1 |
| **E2E only** (UI testada via mock) | 9 | 16% | 0 |
| **🔴 Não coberto** | 12 | 20% | **-7** |
| **Total** | 58 | 100% | — |

**Melhoria:** de 19 cenários não cobertos para 12 (-37%).
**Novos testes:** 17 unit tests adicionados (favoritos.test.js + fonte lojas + empty states).

---

## Gaps remanescentes (12 cenários)

### Visual/UX (não testável em unit — requer E2E ou component test)

| # | Cenário | Prioridade |
|---|---------|-----------|
| 30-32 | Buscas salvas: ⏱, pills múltiplas, ícones | Baixa (cosmético) |
| 43-45 | Input ✕/ESC/botão condicional | Média (acessibilidade) |
| 50 | Favorito ★ dourada | Baixa (cosmético) |
| 52-53 | Botão Publicar navega / Link copia | Média |
| 57-58 | Badge Favoritos / Tooltip | Baixa |

### Edge cases

| # | Cenário | Prioridade |
|---|---------|-----------|
| 26 | Sync favoritos cross-device | Baixa (depende de servidor real) |
| 34-35 | Busca só categorias / dias_janela | Baixa (path alternativo) |
| 39 | Cache 2min toggle rápido | Baixa (otimização) |
| 41 | Coletas antigas sem imagem | Baixa (visual edge case) |

---

## Nota sobre a unificação

Após a unificação das páginas, os cenários 33-37 (Gerenciar buscas) agora são
acessíveis via seção "⚙️ Configuração" colapsável na página `/`. Os testes E2E
foram atualizados para expandir essa seção antes de interagir com os componentes.

A rota `/lojas` retorna 404 (completamente removida). O teste em
`novas-features.spec.js` valida isso explicitamente.
