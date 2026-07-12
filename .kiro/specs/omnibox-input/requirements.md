# Requirements Document

## Introduction

Substituir o sistema multi-lane de inputs na página Descobrir (keywords, lojas, categorias separados) por um campo unificado estilo omnibox. O novo componente `Omnibox` aceita texto livre com inferência automática de tipo e suporta prefixos opcionais (`@`, `#`, `!`) como atalhos para filtrar por loja, categoria ou marketplace respectivamente. O dropdown de sugestões agrupa resultados por tipo. Filtros numéricos (comissão, vendas, nota) permanecem como controles separados fora do input.

## Glossary

- **Omnibox**: Componente unificado de input que substitui os inputs separados de keywords, lojas e categorias na página Descobrir.
- **BuscaEngine**: Máquina de estados finitos (FSM) que gerencia o estado completo da busca na página Descobrir, recebendo eventos e aplicando transições conforme `busca-rules.json`.
- **Token**: Unidade semântica extraída do texto digitado no Omnibox — pode ser keyword, loja, categoria ou marketplace.
- **Prefixo**: Caractere especial (`@`, `#`, `!`) que antecede um token para forçar o tipo sem depender de inferência.
- **Inferência**: Processo de classificar automaticamente o texto digitado em tipos (keyword, loja, categoria) quando nenhum prefixo é usado.
- **Sugestão**: Item exibido no dropdown agrupado por tipo, gerado a partir do texto parcial digitado.
- **Loja_Monitorada**: Loja derivada das buscas salvas do usuário, disponível para autocompletar via prefixo `@` ou inferência.
- **Categoria**: Categoria de produtos obtida do endpoint `/api/categorias`, filtrável por marketplace.
- **Marketplace**: Plataforma de e-commerce suportada (Shopee, Mercado Livre, Amazon), filtrável via prefixo `!`.
- **Busca_Salva**: Configuração de busca previamente persistida pelo usuário, que pode aparecer como sugestão no dropdown.

## Requirements

### Requisito 1: Input Unificado

**User Story:** Como usuário da página Descobrir, quero um único campo de busca que aceite keywords, lojas, categorias e marketplaces, para não precisar navegar entre múltiplos inputs e economizar espaço vertical.

#### Critérios de Aceitação

1. THE Omnibox SHALL renderizar um único campo de texto que substitui os inputs separados de keywords, lojas e categorias na página Descobrir.
2. THE Omnibox SHALL aceitar texto livre sem exigir prefixos para iniciar uma busca.
3. THE Omnibox SHALL manter o texto digitado pelo usuário como texto literal, sem transformar tokens em chips visuais dentro do campo.
4. WHEN o usuário pressiona Enter, THE Omnibox SHALL emitir um evento para a BuscaEngine com todo o contexto resolvido (keywords, shopIds, categorias, marketplacesFiltro).

### Requisito 2: Inferência Automática de Tipo

**User Story:** Como usuário, quero que o sistema reconheça automaticamente se estou buscando uma keyword, loja ou categoria sem precisar memorizar prefixos, para que a experiência seja intuitiva.

#### Critérios de Aceitação

1. WHEN o usuário digita texto sem prefixo, THE Omnibox SHALL executar inferência comparando o texto contra keywords, lojas monitoradas e categorias disponíveis simultaneamente.
2. WHEN o texto sem prefixo coincide parcialmente com o nome de uma Loja_Monitorada, THE Omnibox SHALL incluir a loja como sugestão no grupo "Lojas" do dropdown.
3. WHEN o texto sem prefixo coincide parcialmente com o nome de uma Categoria, THE Omnibox SHALL incluir a categoria como sugestão no grupo "Categorias" do dropdown.
4. WHEN o texto sem prefixo não coincide com nenhuma loja ou categoria, THE Omnibox SHALL tratar o texto como keyword global.

### Requisito 3: Prefixos Opcionais como Atalhos

**User Story:** Como usuário avançado, quero usar prefixos (@, #, !) para filtrar rapidamente o tipo de busca, para compor buscas complexas com precisão.

#### Critérios de Aceitação

1. WHEN o usuário digita `@` seguido de texto, THE Omnibox SHALL filtrar sugestões exclusivamente para Lojas_Monitoradas cujo nome coincide parcialmente com o texto.
2. WHEN o usuário digita `#` seguido de texto, THE Omnibox SHALL filtrar sugestões exclusivamente para Categorias cujo nome coincide parcialmente com o texto.
3. WHEN o usuário digita `!` seguido de texto, THE Omnibox SHALL filtrar sugestões exclusivamente para Marketplaces cujo nome coincide parcialmente com o texto.
4. THE Omnibox SHALL tratar texto sem prefixo como keyword quando combinado com tokens prefixados (ex: `serum @lebotanic` → keyword "serum" + loja "lebotanic").

### Requisito 4: Dropdown de Sugestões Agrupadas

**User Story:** Como usuário, quero ver sugestões organizadas por tipo em um dropdown, para identificar rapidamente o que estou selecionando.

#### Critérios de Aceitação

1. WHEN o usuário digita 2 ou mais caracteres de um token, THE Omnibox SHALL exibir um dropdown com sugestões agrupadas por tipo (Keywords, Lojas, Categorias, Marketplaces).
2. THE Omnibox SHALL exibir no máximo 7 sugestões por grupo no dropdown.
3. WHEN existe uma Busca_Salva cujo contexto coincide com o texto digitado, THE Omnibox SHALL exibir a busca salva como primeira sugestão no dropdown.
4. THE Omnibox SHALL realizar match parcial case-insensitive ao filtrar sugestões.
5. WHEN o dropdown está vazio (nenhuma sugestão encontra match), THE Omnibox SHALL ocultar o dropdown.

### Requisito 5: Seleção de Sugestões e Navegação por Teclado

**User Story:** Como usuário, quero navegar e selecionar sugestões com o teclado, para manter fluidez sem usar o mouse.

#### Critérios de Aceitação

1. WHEN o usuário pressiona Tab ou clica em uma sugestão, THE Omnibox SHALL completar o token com o valor selecionado no campo de texto.
2. WHEN o usuário pressiona Enter com o dropdown aberto e uma sugestão destacada, THE Omnibox SHALL selecionar a sugestão destacada e completar o token.
3. WHEN o usuário pressiona Enter sem sugestão destacada, THE Omnibox SHALL executar a busca com o contexto atualmente resolvido.
4. WHEN o usuário pressiona Escape, THE Omnibox SHALL fechar o dropdown sem alterar o texto.
5. WHEN o usuário pressiona as teclas ArrowUp ou ArrowDown, THE Omnibox SHALL navegar entre as sugestões no dropdown.

### Requisito 6: Integração com BuscaEngine

**User Story:** Como desenvolvedor, quero que o Omnibox comunique mudanças de estado à BuscaEngine via eventos, para manter o padrão arquitetural existente baseado em FSM.

#### Critérios de Aceitação

1. WHEN o usuário seleciona uma loja no dropdown, THE Omnibox SHALL emitir o evento `ADICIONAR_LOJA` para a BuscaEngine com o shopId correspondente.
2. WHEN o usuário seleciona uma categoria no dropdown, THE Omnibox SHALL emitir o evento `ADICIONAR_CATEGORIA` para a BuscaEngine com o nome da categoria.
3. WHEN o usuário altera o texto de keyword e confirma com Enter, THE Omnibox SHALL emitir o evento `DIGITAR` para a BuscaEngine com a keyword resolvida.
4. WHEN o usuário seleciona um marketplace no dropdown, THE Omnibox SHALL emitir o evento `MUDAR_MARKETPLACES` para a BuscaEngine com a lista atualizada de marketplaces.
5. THE Omnibox SHALL respeitar as transições definidas em `busca-rules.json`, incluindo debounce de 400ms para o evento `DIGITAR`.

### Requisito 7: Parser de Tokens

**User Story:** Como desenvolvedor, quero uma função pura que tokenize o texto do input em tokens tipados, para facilitar testes e reutilização.

#### Critérios de Aceitação

1. THE Parser SHALL tokenizar o texto do input em um array de objetos com propriedades `tipo` ("keyword" | "loja" | "categoria" | "marketplace"), `valor` (texto sem prefixo) e `completo` (boolean indicando se o token foi finalizado).
2. WHEN o texto contém múltiplos tokens separados por espaço, THE Parser SHALL identificar cada token individualmente respeitando a gramática definida em T-0055.
3. THE Parser SHALL tratar texto sem prefixo como tipo "keyword".
4. THE Parser SHALL tratar `@texto` como tipo "loja", `#texto` como tipo "categoria", e `!texto` como tipo "marketplace".
5. FOR ALL textos válidos de input, parsear e depois imprimir (serializar de volta para string) e parsear novamente SHALL produzir tokens equivalentes (propriedade round-trip).

### Requisito 8: Geração de Sugestões

**User Story:** Como desenvolvedor, quero uma função pura que gere sugestões a partir do token incompleto atual e do contexto disponível, para manter a lógica testável separada do componente visual.

#### Critérios de Aceitação

1. WHEN o último token tem 2 ou mais caracteres, THE Gerador_Sugestoes SHALL produzir sugestões filtradas por match parcial case-insensitive.
2. WHEN o último token tem prefixo, THE Gerador_Sugestoes SHALL restringir sugestões ao tipo correspondente ao prefixo.
3. WHEN o último token não tem prefixo, THE Gerador_Sugestoes SHALL produzir sugestões de todos os tipos (keywords, lojas, categorias).
4. WHEN não existem lojas monitoradas nem categorias disponíveis, THE Gerador_Sugestoes SHALL retornar sugestões apenas do tipo keyword global.
5. THE Gerador_Sugestoes SHALL respeitar o limite máximo de 7 sugestões conforme configurado em `busca-rules.json`.

### Requisito 9: Estado Zero e Degradação Graceful

**User Story:** Como usuário novo sem buscas salvas, quero que o Omnibox funcione como um campo de busca por keyword simples, para não ficar bloqueado pela ausência de lojas ou categorias.

#### Critérios de Aceitação

1. WHEN o usuário não possui lojas monitoradas, THE Omnibox SHALL funcionar como input de keyword global, sem exibir sugestões de lojas.
2. WHEN o endpoint `/api/categorias` retorna lista vazia ou erro, THE Omnibox SHALL funcionar sem sugestões de categorias, mantendo funcionalidade de keyword e lojas.
3. IF o carregamento de sugestões exceder 2 segundos, THEN THE Omnibox SHALL exibir o dropdown sem as sugestões pendentes, mantendo as já carregadas.
4. WHEN o usuário digita um prefixo inválido (diferente de `@`, `#`, `!`), THE Omnibox SHALL tratar o texto inteiro como keyword.

### Requisito 10: Acessibilidade

**User Story:** Como usuário que utiliza tecnologia assistiva, quero que o Omnibox seja navegável por teclado e anunciado corretamente por leitores de tela, para ter uma experiência equivalente.

#### Critérios de Aceitação

1. THE Omnibox SHALL implementar o padrão ARIA combobox (`role="combobox"`, `aria-expanded`, `aria-controls`, `aria-activedescendant`).
2. THE Omnibox SHALL anunciar o número de sugestões disponíveis via `aria-live` region quando o dropdown abre ou atualiza.
3. THE Omnibox SHALL associar cada grupo de sugestões com um `role="group"` e `aria-label` descritivo do tipo.
4. WHEN o foco sai do Omnibox, THE Omnibox SHALL fechar o dropdown.
