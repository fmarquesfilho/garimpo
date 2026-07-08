# Tasks

## Task 1: Estender montarResultados com fonte "lojas"

- [ ] Em `web/src/lib/descobrir-logic.js`, adicionar parâmetro opcional `dadosLojas` à assinatura de `montarResultados`
- [ ] Quando `fontes.lojas` é true e `dadosLojas` não é vazio, incluir os itens com `_fonte: 'loja'` na lista combinada
- [ ] Garantir que os filtros existentes (keyword, categorias, comissaoMin, vendasMin) se aplicam aos itens de lojas também
- [ ] Verificar que os 40+ testes existentes em `descobrir.test.js` continuam passando sem modificação
- [ ] Adicionar novos testes em `descobrir.test.js` para a fonte "lojas": toggle on/off, filtro por keyword, filtro por loja específica

## Task 2: Criar função carregarProdutosLojas em descobrir.js

- [ ] Em `web/src/lib/descobrir.js`, criar função `carregarProdutosLojas(buscasComLojas)` que chama `buscarCandidatos` com `fonte: 'shopee-shop'` para cada loja
- [ ] Implementar cache de 2 minutos (mesmo padrão de `cacheOportunidades`)
- [ ] Mapear resposta com `_fonte: 'loja'`, `_loja_id`, e `loja` (nome)
- [ ] Tratar erros individualmente (loja que falha não bloqueia as outras)
- [ ] Exportar a função

## Task 3: Refatorar +page.svelte — toggle Lojas + seletor de loja

- [ ] Adicionar `lojas: false` ao estado `fontes` (default desativado para não sobrecarregar na carga inicial)
- [ ] Adicionar estado `dadosLojas` ($state) e `lojaFiltro` ($state, default null = todas)
- [ ] Quando toggle 🏪 é ativado, chamar `carregarProdutosLojas(buscasComLojas)` e popular `dadosLojas`
- [ ] Passar `dadosLojas` (filtrado por `lojaFiltro` se definido) para `montarResultados`
- [ ] Renderizar seletor de loja (chips com nome da loja) visível apenas quando fonte 🏪 ativa
- [ ] Exibir badge com contagem de produtos de lojas nos resultados
- [ ] Exibir empty state quando não há lojas monitoradas e fonte 🏪 está ativa

## Task 4: Refatorar +page.svelte — Área de Configuração colapsável

- [ ] Adicionar estado `mostrarConfig` ($state, default false)
- [ ] Renderizar botão "⚙️ Configuração" que faz toggle de `mostrarConfig`
- [ ] Quando expandido, renderizar FormAdicionarLoja, GerenciarBuscas e PainelAlertas (importar de `$lib/components/`)
- [ ] Quando FormAdicionarLoja adiciona uma loja com sucesso, chamar `buscasSalvas.sincronizarDoServidor()` para atualizar a lista
- [ ] Usar a prop `buscaSelecionada` existente do GerenciarBuscas (com PainelNovidades integrado)
- [ ] Importar `PainelAlertas` e passar `buscaSelecionada` se necessário

## Task 5: Remoção de /lojas e redirect

- [ ] Remover o conteúdo de `web/src/routes/lojas/+page.svelte` (ou deletar o arquivo)
- [ ] Criar `web/src/routes/lojas/+page.js` com `redirect(308, '/')` para redirecionar permanentemente
- [ ] Verificar que `ListaProdutosLoja.svelte` não é mais importado por nenhum arquivo (se sim, pode ser removido)
- [ ] Atualizar referências a `/lojas` em empty states e hints dentro do app (grep por `/lojas`)

## Task 6: Atualizar NavDrawer

- [ ] Em `web/src/lib/components/NavDrawer.svelte`, remover o link `<a href="/lojas">`
- [ ] Renomear o link da página principal de "Descobrir" para "Garimpar"
- [ ] Atualizar active state (classe `!bg-accent`) para refletir que `/` é a rota ativa do link "Garimpar"

## Task 7: Atualizar título e identidade visual

- [ ] Alterar `<title>` de "Descobrir — Garimpei" para "Garimpar — Garimpei" (svelte:head)
- [ ] Manter heading "O que publicar hoje?" (já existe na Descobrir)
- [ ] Adicionar subtítulo: "Busque produtos, monitore lojas e publique com um clique"
- [ ] Verificar `npm run check` e `npm run build`

## Task 8: Atualizar testes E2E

- [ ] Em `web/tests/descobrir.spec.js`, adicionar cenário para toggle 🏪 Lojas (com mock)
- [ ] Em `web/tests/lojas-precos.spec.js` e `lojas-cadastro.spec.js`, atualizar navegação de `/lojas` → `/`
- [ ] Verificar que os mocks de `/api/lojas`, `/api/buscas`, `/api/candidatos` continuam funcionando na nova rota
- [ ] Executar `npx vitest run` para validar testes unitários
- [ ] Verificar `npm run check` e `npm run build` sem erros

## Task 9: Documentar alteração

- [ ] Criar `docs/legado/SESSAO_2026-07-08_UNIFICAR_DESCOBRIR_LOJAS.md` documentando a sessão
- [ ] Atualizar `docs/05-manual-do-usuario.md` se menciona a página /lojas separadamente
- [ ] Atualizar `docs/impacto-migracao-ui.md` se relevante
- [ ] Verificar que `mise run check:api-contract` ainda passa (rotas do frontend devem bater com backend)
