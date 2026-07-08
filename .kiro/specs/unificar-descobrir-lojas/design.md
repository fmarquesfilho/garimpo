# Design Document

## Overview

Unificar as páginas Descobrir (`/`) e Lojas (`/lojas`) em uma única página "Garimpar" (`/`).
A mudança é 100% frontend — backend permanece inalterado. A estratégia é:

1. Estender `montarResultados` com fonte "lojas" (backward-compatible)
2. Adicionar nova função `carregarProdutosLojas` em `descobrir.js`
3. Mover componentes de `/lojas` para dentro da página `/` em seção colapsável
4. Criar redirect `/lojas` → `/`
5. Atualizar NavDrawer

## Components and Interfaces

### descobrir-logic.js — `montarResultados` (estendido)

```javascript
// Assinatura estendida (backward-compatible: dadosLojas é opcional)
export function montarResultados({
  fontes,           // { curadoria, quedas, novos, favoritos, lojas }
  dadosCuradoria,
  dadosQuedas,
  dadosNovos,
  dadosLojas,       // NOVO — array de produtos das lojas monitoradas
  favoritos,
  busca,
  categorias,
  comissaoMin,
  vendasMin
})
```

Nova lógica adicionada:
```javascript
if (fontes.lojas && dadosLojas?.length) {
  todos.push(...dadosLojas.map(p => ({ ...p, _fonte: 'loja' })));
}
```

O filtro por keyword, categorias, comissão e vendas já se aplica a todos os itens
de `todos[]` — funciona sem mudança adicional.

### descobrir.js — `carregarProdutosLojas` (novo)

```javascript
let cacheLojas = { em: 0, produtos: [] };

export async function carregarProdutosLojas(buscasComLojas) {
  // Cache de 2 minutos (mesmo padrão de oportunidades)
  if (Date.now() - cacheLojas.em < 120000 && cacheLojas.produtos.length > 0) {
    return cacheLojas.produtos;
  }

  const promises = buscasComLojas.map(b =>
    buscarCandidatos({
      estrategia: 'nicho',
      top: 50,
      fonte: 'shopee-shop',
      shopIds: b.shop_ids.join(','),
      semFiltro: true
    })
    .then(r => (r.candidatos ?? []).map(c => ({
      ...c,
      _fonte: 'loja',
      _loja_id: b.id,
      loja: b.nome || b.id
    })))
    .catch(() => [])
  );

  const resultados = await Promise.all(promises);
  const produtos = resultados.flat();
  cacheLojas = { em: Date.now(), produtos };
  return produtos;
}
```

### +page.svelte — Layout unificado

```
┌────────────────────────────────────────────────────────────────┐
│ "O que publicar hoje?"                                         │
│ Busque produtos, monitore lojas e publique com um clique       │
├────────────────────────────────────────────────────────────────┤
│ [input busca]                                                  │
│ [🔍 Busca] [📉 Quedas] [🆕 Novos] [⭐ Favoritos] [🏪 Lojas] │
│ [FilterBar: comissão, vendas, categorias]                      │
├────────────────────────────────────────────────────────────────┤
│ [Seletor de loja: Todas | Loja A | Loja B] (quando 🏪 ativa) │
├────────────────────────────────────────────────────────────────┤
│ [Grid de ProductCards — resultados unificados]                 │
├────────────────────────────────────────────────────────────────┤
│ [⚙️ Configuração ▾] (colapsável, default fechado)             │
│   ├─ FormAdicionarLoja                                         │
│   ├─ GerenciarBuscas (+ PainelNovidades)                       │
│   └─ PainelAlertas                                             │
└────────────────────────────────────────────────────────────────┘
```

### /lojas/+page.server.js — Redirect

```javascript
export function load() {
  return { status: 308, redirect: '/' };
}
```

Ou via `+page.js`:
```javascript
import { redirect } from '@sveltejs/kit';
export function load() { redirect(308, '/'); }
```

### NavDrawer.svelte — Atualização

- Remover link `/lojas`
- Renomear "Descobrir" → "Garimpar"

## Data Models

Nenhuma mudança de modelo. Os dados de "produtos de loja" vêm do mesmo endpoint
`GET /api/candidatos?fonte=shopee-shop&shopIds=X` que a página `/lojas` já usava.
São mapeados com `_fonte: 'loja'` e filtrados pela mesma lógica existente.

## Error Handling

| Cenário | Comportamento |
|---------|--------------|
| Loja falha ao carregar produtos | Produtos das outras lojas aparecem + warning não-bloqueante |
| API candidatos indisponível | Fonte Lojas mostra "Erro ao carregar" mas outras fontes continuam |
| Nenhuma loja monitorada + toggle ativo | Empty state: "Adicione uma loja na seção ⚙️ Configuração abaixo" |
| Cache expirado (>2min) + toggle rápido | Re-fetch silencioso, loading indicator |

## Testing Strategy

| O que | Como | Impacto |
|-------|------|---------|
| montarResultados com dadosLojas | Novos testes em `descobrir.test.js` (adicionar cenários para fonte lojas) |  |
| Testes existentes (40+) | Devem passar sem modificação (backward compat) | Zero regressão |
| E2E descobrir.spec.js | Atualizar para refletir novo toggle Lojas | |
| E2E antigos de /lojas | Atualizar para testar redirect ou remover | |

## Summary of Changes

| Arquivo | Tipo | Descrição |
|---------|------|-----------|
| `web/src/lib/descobrir-logic.js` | Ajuste | Adicionar `dadosLojas` + `fontes.lojas` no `montarResultados` |
| `web/src/lib/descobrir.js` | Ajuste | Nova função `carregarProdutosLojas` com cache 2min |
| `web/src/routes/+page.svelte` | Refactor | Adicionar toggle 🏪, seletor de loja, seção Configuração (FormAdicionarLoja, GerenciarBuscas, PainelAlertas) |
| `web/src/routes/lojas/+page.svelte` | Removido | Substituído por redirect |
| `web/src/routes/lojas/+page.js` | Novo | Redirect 308 → `/` |
| `web/src/lib/components/NavDrawer.svelte` | Ajuste | Remover link /lojas, renomear Descobrir→Garimpar |
| `web/src/lib/components/ListaProdutosLoja.svelte` | Pode ser removido | Funcionalidade absorvida pela fonte 🏪 |
| `web/tests/descobrir.spec.js` | Ajuste | Adicionar cenários para toggle Lojas |
| `web/tests/lojas-*.spec.js` | Ajuste | Atualizar navegação (/ em vez de /lojas) |
| `web/src/tests/descobrir.test.js` | Ajuste | Adicionar testes para `dadosLojas` + `fontes.lojas` |
