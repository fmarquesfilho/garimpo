# Cenários de teste — Página Descobrir

Checklist completo para validação manual. Organizado por fluxo.

---

## 1. Fontes de dados (toggles)

| # | Fontes ativas | Keyword | Categorias | Resultado esperado |
|---|---|---|---|---|
| 1 | 🔍 Busca | "sérum" | — | Produtos da API com imagem, preço, loja, comissão |
| 2 | 🔍 Busca | vazio | — | Hint: "Digite um termo" |
| 3 | 🔍 Busca | vazio | ["Perfumaria"] | Produtos da categoria Perfumaria (sem precisar keyword) |
| 4 | 📉 Quedas | vazio | — | Todas as quedas de preço das lojas monitoradas |
| 5 | 📉 Quedas | "Skin1004" | — | Só quedas cujo nome ou loja contém "Skin1004" |
| 6 | 🆕 Novos | vazio | — | Todos os produtos novos das lojas monitoradas |
| 7 | 🆕 Novos | "retinol" | — | Só novos cujo nome ou loja contém "retinol" |
| 8 | ⭐ Favoritos | vazio | — | Todos os produtos favoritados |
| 9 | ⭐ Favoritos | "perfume" | — | Só favoritos cujo nome ou loja contém "perfume" |
| 10 | Nenhum ativo | qualquer | — | Hint: "Ative ao menos uma fonte" |
| 11 | 🔍 + 📉 + 🆕 | vazio | — | Quedas + Novos (curadoria não roda sem keyword nem categoria) |
| 12 | 🔍 + 📉 + 🆕 | "sérum" | — | Todos os resultados filtrados por "sérum" |
| 13 | Todos | "SKIN1004" | — | Resultados de todas as fontes filtrados pelo nome da loja |

---

## 2. Busca por nome de loja

| # | Keyword digitada | Resultado esperado |
|---|---|---|
| 14 | Nome exato (ex: "SKIN1004 Official") | Mostra só produtos dessa loja em qualquer fonte ativa |
| 15 | Parte do nome (ex: "SKIN") | Filtra qualquer produto cuja loja contenha "SKIN" |

---

## 3. Filtragem por categoria

| # | Categorias ativas | Keyword | Resultado esperado |
|---|---|---|---|
| 16 | ["Perfumaria"] | vazio | Só produtos da categoria Perfumaria |
| 17 | ["Perfumaria", "Maquiagem"] | vazio | Produtos de Perfumaria OU Maquiagem |
| 18 | ["Cuidados com a Pele"] | "sérum" | Keyword AND categoria (interseção) |
| 19 | ["Perfumaria"] | "SKIN1004" | Loja AND categoria |
| 20 | Nenhuma | qualquer | Sem filtro de categoria — mostra tudo |
| 21 | Busca salva com categorias | (clicar pill) | Ativa categorias + fontes da busca |

---

## 4. Favoritos (⭐)

| # | Ação | Resultado esperado |
|---|---|---|
| 22 | Clicar ☆ num produto | Estrela vira ★ (dourada), produto salvo |
| 23 | Clicar ★ num produto já favoritado | Volta a ☆, produto removido |
| 24 | Ativar fonte ⭐ Favoritos | Mostra todos os produtos salvos |
| 25 | Favoritar produto de "Quedas" | Aparece tanto em Quedas quanto em Favoritos |
| 26 | Sair e entrar de outro dispositivo | Mesmos favoritos aparecem (sync servidor) |

---

## 5. Buscas salvas (atalhos)

| # | Ação | Resultado esperado |
|---|---|---|
| 27 | Clicar pill de keyword | Seta keyword no input + ativa fontes da busca |
| 28 | Busca salva com fontes [quedas, novos] | Clicar ativa Quedas + Novos |
| 29 | Busca salva com categorias | Clicar ativa categorias + fontes |
| 30 | Busca agendada | Mostra ícone ⏱ ao lado das pills |
| 31 | Busca com múltiplas keywords | Mostra todas as keywords como pills clicáveis |
| 32 | Busca com 📉 e 🆕 mostra ícones | Ícones de fonte visíveis ao lado das pills |

---

## 6. Gerenciar buscas (em /lojas)

| # | Ação | Resultado esperado |
|---|---|---|
| 33 | Clicar "+ nova busca" | Abre formulário com keywords, categorias, fontes, agendamento |
| 34 | Criar busca só com categorias (sem keyword) | Salva e aparece como atalho na Descobrir |
| 35 | Criar busca com fontes [novos] + dias_janela=3 | Salva corretamente, badge mostra "janela: 3d" |
| 36 | Criar busca agendada (cron) | Mostra ⏱ no card e aparece na Descobrir |
| 37 | Remover busca | Desaparece da lista e dos atalhos |

---

## 7. Dados, cache e timeout

| # | Cenário | Resultado esperado |
|---|---|---|
| 38 | Sem lojas monitoradas + Quedas/Novos ativo | Empty state com link para /lojas |
| 39 | Toggle rápido on/off/on | Não re-busca (cache 2 min) |
| 40 | API demora > 25s | Timeout com erro + botão "Tentar novamente" |
| 41 | Coletas antigas (sem imagem no snapshot) | Card sem foto, nome e preço visíveis |
| 42 | Coletas novas (com imagem no snapshot) | Card com foto, link clicável, nome da loja |

---

## 8. Input de busca

| # | Ação | Resultado esperado |
|---|---|---|
| 43 | Clicar ✕ com texto | Campo limpa |
| 44 | Pressionar ESC | Campo limpa |
| 45 | Campo vazio | Botão ✕ não aparece |
| 46 | Digitar e esperar 400ms | Busca dispara automaticamente (debounce) |

---

## 9. ProductCard (visual)

| # | Fonte do produto | O que aparece |
|---|---|---|
| 47 | Curadoria (🔍) | Imagem, nome, preço, comissão %, vendas, nota ★, loja, score, posição #N |
| 48 | Queda (📉) | Imagem, nome, preço anterior → atual, variação %, loja, badge verde |
| 49 | Novo (🆕) | Imagem, nome, preço, comissão, loja, badge "Novo" |
| 50 | Favorito (⭐) | Mesmo visual + ★ dourada |
| 51 | Qualquer produto | Botão ☆/★ funciona (toggle) |
| 52 | Qualquer produto | Botão "📤 Publicar" navega para /publicar com dados |
| 53 | Qualquer produto | Botão "🔗 Link" copia para clipboard |

---

## 10. Badges nos toggles

| # | Cenário | Resultado esperado |
|---|---|---|
| 54 | Dados carregados | 📉 Quedas mostra badge verde com número |
| 55 | Dados carregados | 🆕 Novos mostra badge rosa com número |
| 56 | Dados carregados | 🔍 Busca mostra badge dourado com número |
| 57 | Favoritos não-vazio | ⭐ Favoritos mostra badge com contagem |
| 58 | Hover nos toggles | Tooltip explica o que cada fonte faz |
