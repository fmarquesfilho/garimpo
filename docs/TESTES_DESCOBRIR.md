# Cenários de teste — Página Descobrir

Checklist para validação manual da página principal unificada.

---

## Filtros de fonte (toggles)

| # | Fontes ativas | Keyword | Resultado esperado |
|---|---|---|---|
| 1 | 🔍 Busca | "sérum" | Produtos da API de afiliados com imagem, preço, loja |
| 2 | 🔍 Busca | vazio | Nenhum resultado (hint: "Digite um termo") |
| 3 | 📉 Quedas | vazio | Todas as quedas de preço das lojas monitoradas |
| 4 | 📉 Quedas | "Skin1004" | Só quedas cujo nome ou loja contém "Skin1004" |
| 5 | 🆕 Novos | vazio | Todos os produtos novos detectados nas lojas |
| 6 | 🆕 Novos | "retinol" | Só novos cujo nome ou loja contém "retinol" |
| 7 | ⭐ Favoritos | vazio | Lista todos os produtos favoritados |
| 8 | ⭐ Favoritos | "perfume" | Só favoritos cujo nome ou loja contém "perfume" |
| 9 | Nenhum ativo | qualquer | Nenhum resultado (hint: "Ative ao menos uma fonte") |
| 10 | 🔍 + 📉 + 🆕 | vazio | Quedas + Novos (Busca não roda sem keyword) |
| 11 | 🔍 + 📉 + 🆕 | "sérum" | Busca + quedas + novos filtrados por "sérum" |
| 12 | Todos | "SKIN1004" | Resultados de todas as fontes filtrados pelo nome da loja |

---

## Busca por nome de loja

| # | Keyword digitada | O que filtra |
|---|---|---|
| 13 | Nome exato da loja (ex: "SKIN1004 Official") | Mostra só produtos dessa loja em qualquer fonte ativa |
| 14 | Parte do nome da loja (ex: "SKIN") | Filtra qualquer produto cuja loja contenha "SKIN" |

---

## Filtragem por categoria

| # | Categorias ativas | Keyword | Resultado esperado |
|---|---|---|---|
| 35 | ["Perfumaria"] | vazio | Só produtos da categoria Perfumaria (+ produtos sem categoria) |
| 36 | ["Perfumaria", "Maquiagem"] | vazio | Produtos de Perfumaria OU Maquiagem (OR) |
| 37 | ["Cuidados com a Pele"] | "sérum" | Keyword AND categoria — interseção |
| 38 | ["Perfumaria"] | "SKIN1004" | Loja + categoria — mostra só perfumaria da SKIN1004 |
| 39 | Nenhuma | qualquer | Sem filtro de categoria — mostra tudo |
| 40 | Busca salva com categorias | (clicar pill) | Ativa categorias da busca + fontes correspondentes |

---

## Favoritos (⭐)

| # | Ação | Resultado esperado |
|---|---|---|
| 15 | Clicar ☆ num produto | Estrela vira ★ (dourada), produto salvo |
| 16 | Clicar ★ num produto já favoritado | Estrela volta a ☆, produto removido dos favoritos |
| 17 | Ativar fonte ⭐ Favoritos | Mostra todos os produtos salvos |
| 18 | Sair e entrar de outro dispositivo | Favoritos sincronizados (mesma lista) |

---

## Buscas salvas (atalhos)

| # | Ação | Resultado esperado |
|---|---|---|
| 19 | Clicar numa pill de keyword | Seta keyword no input, ativa fonte "Busca" |
| 20 | Busca salva com fontes [quedas, novos] | Clicar ativa Quedas + Novos (desativa Busca se não tem keyword) |
| 21 | Busca agendada (⏱ visível) | Mostra ícone de relógio ao lado das pills |
| 22 | Busca com múltiplas keywords | Mostra todas as keywords como pills clicáveis |

---

## Dados e cache

| # | Cenário | Resultado esperado |
|---|---|---|
| 23 | Sem lojas monitoradas + Quedas/Novos ativo | Empty state: "Você ainda não monitora nenhuma loja" com link para /lojas |
| 24 | Toggle rápido Quedas on/off/on | Não re-busca (cache 2 min), resultados aparecem instantaneamente |
| 25 | API demora > 25s | Timeout com mensagem de erro + botão "Tentar novamente" |
| 26 | Produtos de coletas antigas (sem imagem) | Card aparece sem foto (graceful), nome e preço visíveis |
| 27 | Produtos de coletas novas (com imagem) | Card aparece com foto, link clicável, nome da loja |

---

## Botão X no input

| # | Ação | Resultado esperado |
|---|---|---|
| 28 | Clicar ✕ com texto no campo | Campo limpa, foco volta pro input |
| 29 | Pressionar ESC com texto no campo | Campo limpa |
| 30 | Campo vazio | Botão ✕ não aparece |

---

## ProductCard (visual consistente)

| # | Fonte do produto | O que deve aparecer no card |
|---|---|---|
| 31 | Curadoria (🔍) | Imagem, nome, preço, comissão, vendas, nota, loja, score, posição (#1, #2...) |
| 32 | Queda (📉) | Imagem (se disponível), nome, preço anterior → atual, % variação, loja |
| 33 | Novo (🆕) | Imagem (se disponível), nome, preço, comissão, loja, badge "Novo" |
| 34 | Favorito (⭐) | Mesmo layout dos demais, com ★ dourada indicando que é favoritado |
