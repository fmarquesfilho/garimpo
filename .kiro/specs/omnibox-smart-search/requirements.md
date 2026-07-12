# Requirements Document

## Introduction

O Omnibox atual (T-0055) funciona bem com prefixos opcionais (`@loja`, `#categoria`, `!marketplace`), mas essa mecânica não é intuitiva para usuários casuais. A evolução "Smart Search" faz o Omnibox inferir a intenção do usuário a partir de texto livre, sem depender de wildcards como fluxo primário.

O sistema apresenta um dropdown contextual com opções de intenção ("Pesquisar em Produtos", "Pesquisar em Lojas", "Resolver Link", "Pesquisar na categoria X"), detecta URLs coladas automaticamente, e permite ativar monitoramento de lojas diretamente a partir dos resultados de busca por loja.

Os prefixos (`@`, `#`, `!`) permanecem funcionais para power users. A BuscaEngine FSM não é alterada estruturalmente — apenas novos eventos e handlers são adicionados conforme necessário.

## Glossary

- **Omnibox**: Input unificado da página Descobrir que aceita texto livre e prefixos opcionais. Componente Svelte headless.
- **BuscaEngine**: Máquina de estados headless (classe Svelte 5) que gerencia toda a lógica da página Descobrir.
- **Smart_Dropdown**: Dropdown contextual do Omnibox que exibe opções de intenção baseadas no texto digitado, substituindo o dropdown atual de autocomplete por tipo.
- **Intenção**: Classificação da ação desejada pelo usuário a partir do texto digitado (buscar produtos, buscar lojas, resolver link, buscar por categoria).
- **Detector_de_Intenção**: Função pura que analisa o texto do Omnibox e gera as opções de intenção para o Smart_Dropdown.
- **Busca_por_Loja**: Nova capacidade de pesquisar lojas cujo nome corresponde ao texto digitado, retornando Store_Cards em vez de Product_Cards.
- **Store_Card**: Card de resultado que exibe informações de uma loja (nome, marketplace, status de monitoramento, ação de ativar monitoramento).
- **Resolução_de_Link**: Processo de identificar uma loja a partir de uma URL colada no Omnibox, usando o endpoint existente `/api/lojas/resolver`.
- **Registro_de_Loja**: Tabela PostgreSQL com lojas conhecidas pelo sistema (shopId, nome, marketplace, cron). Fonte autoritativa de lojas locais.
- **Categoria_de_Primeiro_Nível**: Categoria raiz retornada por `/api/categorias`, usada para matching de intenção por categoria.
- **Marketplace_Ativo**: Marketplace selecionado no filtro de marketplaces do contexto atual da BuscaEngine.
- **Loja_Monitorada**: Loja com cron ativo no Registro_de_Loja — coleta automática habilitada.
- **Collector_API**: Serviço gRPC backend que resolve URLs/IDs de lojas e retorna dados enriquecidos (shopId, nome, seguidores, imagem, avaliação). Não suporta busca por nome — a Shopee não expõe essa capacidade.

## Requirements

### Requisito 1: Detecção de Intenção a partir de Texto Livre

**User Story:** Como usuário, eu quero digitar texto livre no Omnibox e ver opções claras do que posso fazer com esse texto, para que eu não precise memorizar prefixos como `@` ou `#`.

#### Critérios de Aceitação

1. WHEN o usuário digita 2 ou mais caracteres sem prefixo no Omnibox, THE Detector_de_Intenção SHALL gerar uma lista de opções de intenção baseadas no texto digitado.
2. WHEN o texto digitado não começa com URL (http:// ou https://), THE Smart_Dropdown SHALL exibir ao menos as opções "Pesquisar em Produtos" e "Pesquisar em Lojas" com o texto digitado como parâmetro.
3. WHEN o texto digitado corresponde (case-insensitive, parcial) ao nome de uma Categoria_de_Primeiro_Nível de algum marketplace ativo, THE Smart_Dropdown SHALL incluir uma opção "Pesquisar produtos na categoria {nome_categoria}".
4. WHEN o texto digitado começa com prefixo válido (`@`, `#`, `!`), THE Detector_de_Intenção SHALL delegar para o sistema de sugestões existente (comportamento atual preservado).
5. THE Detector_de_Intenção SHALL ser implementado como função pura, sem dependência de DOM ou runes, recebendo o texto e o contexto disponível como parâmetros.
6. WHEN o texto digitado tem menos de 2 caracteres e não possui prefixo, THE Smart_Dropdown SHALL não exibir opções de intenção.

### Requisito 2: Opções de Intenção Sensíveis ao Contexto de Marketplace e Lojas

**User Story:** Como usuário, eu quero que as opções do dropdown reflitam meu contexto atual (marketplaces ativos, lojas selecionadas), para que eu saiba exatamente onde a busca vai acontecer.

#### Critérios de Aceitação

1. WHEN apenas um Marketplace_Ativo está selecionado no filtro e o texto casa com uma categoria desse marketplace, THE Smart_Dropdown SHALL exibir a opção com o nome do marketplace explícito: "Pesquisar por #{categoria} na {marketplace}".
2. WHEN múltiplos Marketplace_Ativo estão selecionados e o texto casa com categorias, THE Smart_Dropdown SHALL exibir uma opção por marketplace que possui a categoria correspondente, mais uma opção "Pesquisar em todos os marketplaces".
3. WHEN lojas estão ativas no escopo da busca atual (ctx.shopIds não vazio), THE Smart_Dropdown SHALL incluir uma opção "Pesquisar por #{categoria} nas lojas selecionadas" se o texto casa com uma categoria.
4. WHEN apenas uma loja está ativa no escopo e o texto casa com uma categoria, THE Smart_Dropdown SHALL usar o nome da loja na opção: "Pesquisar por #{categoria} na {nome_loja}".
5. WHEN nenhum Marketplace_Ativo está explicitamente selecionado (array vazio = todos), THE Smart_Dropdown SHALL tratar como se todos os marketplaces suportados estivessem ativos para fins de matching de categoria.

### Requisito 3: Detecção e Resolução de Links

**User Story:** Como usuário, eu quero colar um link de loja no Omnibox e resolver automaticamente para a loja correspondente, sem precisar navegar para outra tela ou usar prefixos.

#### Critérios de Aceitação

1. WHEN o texto digitado no Omnibox começa com "http://" ou "https://", THE Detector_de_Intenção SHALL classificar o texto como URL e gerar a opção "Resolver Link" como primeira opção do Smart_Dropdown.
2. WHEN o usuário seleciona "Resolver Link" no Smart_Dropdown, THE BuscaEngine SHALL executar a resolução da URL via endpoint existente `POST /api/lojas/resolver` com o texto como input.
3. WHEN a resolução da URL retorna sucesso, THE BuscaEngine SHALL exibir o resultado como um Store_Card único nos resultados de busca.
4. IF a resolução da URL falha (loja não encontrada, timeout, ou marketplace não suportado), THEN THE BuscaEngine SHALL exibir mensagem de erro descritiva no painel de resultados sem travar a interface.
5. WHEN uma URL é detectada, THE Smart_Dropdown SHALL não exibir as opções genéricas de "Pesquisar em Produtos" ou "Pesquisar em Lojas" — apenas "Resolver Link" é relevante para URLs.
6. THE Detector_de_Intenção SHALL reconhecer tanto links públicos de loja quanto links de afiliado convertidos como URLs válidas para resolução.

### Requisito 4: Busca de Lojas por Nome

**User Story:** Como usuário, eu quero pesquisar lojas pelo nome diretamente do Omnibox, para que eu possa encontrar e monitorar novas lojas sem precisar colar links.

**Restrição técnica:** A Shopee não expõe busca de lojas por nome via API (nem Afiliados nem pública). A busca opera exclusivamente sobre o Registro_de_Loja local. Lojas entram no registro quando o usuário as resolve por link — o registro enriquece naturalmente com o uso.

#### Critérios de Aceitação

1. WHEN o usuário seleciona "Pesquisar em Lojas" no Smart_Dropdown, THE BuscaEngine SHALL executar uma busca por lojas cujo nome corresponde ao texto digitado.
2. THE Busca_por_Loja SHALL pesquisar exclusivamente no Registro_de_Loja local por matching de Nome_Normalizado (substring). Não há fonte remota de busca por nome disponível nos marketplaces.
3. WHEN a Busca_por_Loja retorna resultados, THE BuscaEngine SHALL exibir os resultados como Store_Cards no painel de resultados, substituindo temporariamente a visualização de Product_Cards.
4. IF a Busca_por_Loja não retorna nenhum resultado, THEN THE BuscaEngine SHALL exibir mensagem indicando que nenhuma loja foi encontrada com o termo pesquisado, com sugestão de colar o link da loja para adicioná-la ao registro.
5. THE Busca_por_Loja SHALL ordenar resultados por relevância de matching (match exato primeiro, depois substring).
6. WHEN o texto da busca tem menos de 2 caracteres, THE BuscaEngine SHALL não executar a Busca_por_Loja e exibir mensagem solicitando mais caracteres.

### Requisito 5: Endpoint Backend de Busca de Lojas por Nome

**User Story:** Como desenvolvedor, eu quero um endpoint que pesquise lojas por nome no registro local, para que o frontend consiga listar lojas conhecidas que casam com o texto digitado.

**Restrição técnica:** Opera exclusivamente sobre a tabela Lojas (registro local). A Shopee não expõe API de busca de lojas por nome.

#### Critérios de Aceitação

1. THE API SHALL expor um endpoint `GET /api/lojas/buscar` que aceita os parâmetros: `q` (string, termo de busca, obrigatório, mínimo 2 caracteres) e `marketplace` (string, opcional, filtra por marketplace específico).
2. WHEN o endpoint recebe uma requisição válida, THE API SHALL pesquisar no Registro_de_Loja local por lojas cujo Nome_Normalizado contém o termo normalizado, e retornar os resultados encontrados com campos enriquecidos (imagem, seguidores, avaliação).
3. THE API SHALL retornar uma lista de objetos com campos: id (string, shopId), nome, nome_normalizado, marketplace, monitorada (booleano), origem (bandeira), imagem (URL avatar), seguidores, total_produtos, avaliacao.
4. IF o parâmetro `q` tem menos de 2 caracteres ou está ausente, THEN THE API SHALL retornar HTTP 400 com mensagem descritiva.
5. THE API SHALL retornar no máximo 20 resultados ordenados por nome.
6. WHEN o filtro `marketplace` é informado, THE API SHALL retornar apenas lojas daquele marketplace específico.

### Requisito 6: Store Card nos Resultados

**User Story:** Como usuário, eu quero ver cards de loja informativos nos resultados de busca por loja, para que eu possa avaliar rapidamente se quero monitorar aquela loja.

**Decisão de design:** O Store Card deve ter estilo visual similar ao card de produtos para consistência. Deve exibir imagem (avatar da loja), bandeira de origem, e dados Shopee-specific. Campos enriquecidos (imagem, seguidores, avaliação, total de produtos) são persistidos na tabela Lojas para disponibilidade imediata sem re-resolução.

#### Critérios de Aceitação

1. THE Store_Card SHALL exibir: imagem da loja (avatar, arredondado), nome da loja, bandeira de origem (campo `origem_padrao`, ex: "🇰🇷"), marketplace (com ícone), e indicador de status de monitoramento.
2. WHEN a loja exibida no Store_Card não está monitorada, THE Store_Card SHALL exibir um botão de ação "Monitorar" que permite ao usuário iniciar o monitoramento.
3. WHEN a loja exibida no Store_Card já está monitorada (cron ativo), THE Store_Card SHALL exibir o indicador visual de monitoramento ativo e não exibir o botão "Monitorar".
4. THE Store_Card SHALL exibir campos Shopee-specific quando disponíveis: total de produtos, avaliação (estrelas), seguidores. Para marketplaces sem esses dados, SHALL funcionar com campos mínimos (imagem + nome + marketplace + bandeira).
5. THE Store_Card SHALL ocupar layout mobile-first com estilo visual similar ao ProductCard (mesma hierarquia de informações: imagem à esquerda, info no centro, ação à direita).
6. THE Store_Card SHALL exibir a bandeira de origem dos produtos (campo `origem_padrao` da tabela Lojas) quando disponível, posicionada ao lado do nome para visibilidade imediata.
7. WHEN a loja possui imagem (campo `imagem` persistido), THE Store_Card SHALL exibir o avatar circular; WHEN não possui, SHALL exibir um fallback com o ícone do marketplace.

### Requisito 7: Ativação de Monitoramento a partir dos Resultados

**User Story:** Como usuário, eu quero ativar o monitoramento de uma loja diretamente do card de resultado, para que eu não precise navegar ou fazer scroll até os controles de agendamento no topo.

#### Critérios de Aceitação

1. WHEN o usuário clica em "Monitorar" no Store_Card, THE BuscaEngine SHALL registrar a loja no Registro_de_Loja via `POST /api/lojas/resolver` (se ainda não registrada) e adicionar a loja ao escopo da busca atual com o evento ADICIONAR_LOJA.
2. WHEN o monitoramento é ativado a partir do Store_Card, THE BuscaEngine SHALL registrar a loja com cron padrão vazio (monitoramento registrado, mas agendamento pendente de configuração no topo da página).
3. WHEN a loja é adicionada ao escopo via Store_Card, THE Store_Card SHALL atualizar seu indicador visual para refletir que a loja agora está no escopo (ícone de check ou feedback visual equivalente).
4. WHEN o monitoramento é ativado, THE BuscaEngine SHALL não rolar a página para a seção de agendamento — a configuração de cron permanece responsabilidade do controle existente no topo.
5. IF a ativação de monitoramento falha (erro de rede, Collector_API indisponível), THEN THE Store_Card SHALL exibir feedback de erro inline no próprio card, sem modal ou alerta global.
6. WHEN a loja é adicionada ao escopo via ativação de monitoramento, THE BuscaEngine SHALL disparar re-fetch dos resultados para incluir dados dessa loja nos resultados de busca atuais.

### Requisito 8: Seleção de Intenção e Execução via Teclado

**User Story:** Como usuário, eu quero navegar e selecionar opções do Smart_Dropdown usando apenas o teclado, para manter a experiência mobile-first e rápida sem depender de clique.

#### Critérios de Aceitação

1. WHEN o Smart_Dropdown está visível e o usuário pressiona Enter sem item destacado, THE Omnibox SHALL executar a primeira opção de intenção da lista (comportamento padrão: "Pesquisar em Produtos").
2. WHEN o Smart_Dropdown está visível e o usuário pressiona Enter com um item destacado via ArrowDown/ArrowUp, THE Omnibox SHALL executar a opção de intenção destacada.
3. WHEN o usuário seleciona "Pesquisar em Produtos", THE Omnibox SHALL despachar o evento DIGITAR para a BuscaEngine com o texto como keyword e fechar o dropdown.
4. WHEN o usuário seleciona "Pesquisar em Lojas", THE Omnibox SHALL despachar um novo evento (BUSCAR_LOJAS) para a BuscaEngine com o texto como termo de busca.
5. WHEN o usuário seleciona uma opção de categoria, THE Omnibox SHALL despachar o evento ADICIONAR_CATEGORIA com a categoria correspondente e o evento DIGITAR com keyword vazia (a categoria substitui a keyword).
6. THE Smart_Dropdown SHALL suportar navegação cíclica via ArrowDown (último item → primeiro) e ArrowUp (primeiro item → último), mantendo paridade com o comportamento atual do dropdown.

### Requisito 9: Coexistência com Sistema de Prefixos

**User Story:** Como power user, eu quero continuar usando prefixos (`@`, `#`, `!`) no Omnibox exatamente como antes, para que meu fluxo avançado não seja prejudicado pela nova funcionalidade.

#### Critérios de Aceitação

1. WHEN o texto digitado começa com `@`, THE Omnibox SHALL usar exclusivamente o sistema de sugestões existente (matchLojas por Nome_Normalizado) e não exibir opções de intenção genéricas.
2. WHEN o texto digitado começa com `#`, THE Omnibox SHALL usar exclusivamente o sistema de sugestões existente (matchCategorias) e não exibir opções de intenção genéricas.
3. WHEN o texto digitado começa com `!`, THE Omnibox SHALL usar exclusivamente o sistema de sugestões existente (matchMarketplaces) e não exibir opções de intenção genéricas.
4. WHEN o texto contém múltiplos tokens e o token ativo (último, incompleto) possui prefixo, THE Omnibox SHALL exibir sugestões apenas para aquele token, conforme comportamento atual.
5. THE Detector_de_Intenção SHALL ativar somente quando o último token do input não possui prefixo — tokens anteriores completos com prefixo não interferem na detecção de intenção do token ativo.
6. FOR ALL tokens com prefixo, o comportamento de seleção (remover token ativo, despachar evento correspondente) SHALL permanecer idêntico ao implementado em T-0055.

### Requisito 10: Engine como Controlador Único de UI (Headless UI Controller)

**User Story:** Como desenvolvedor, eu quero que a BuscaEngine seja o controlador único de todo o estado da página Descobrir (incluindo UI do Omnibox), para que o comportamento seja determinístico, testável sem DOM, configurável via JSON, e observável via telemetria.

#### Critérios de Aceitação

1. THE BuscaEngine SHALL manter todo o estado de apresentação do Omnibox (inputValue, aberto, highlightIdx, modo, opcoes, placeholder) como sub-estado reativo acessível via getter `engine.omnibox`.
2. THE Omnibox.svelte SHALL ser um renderizador puro: ZERO campos `$state` locais, renderiza exclusivamente `engine.omnibox.*`, e emite apenas eventos brutos (OMNIBOX_INPUT, OMNIBOX_KEYDOWN, OMNIBOX_SELECIONAR, OMNIBOX_BLUR).
3. THE BuscaEngine SHALL processar OMNIBOX_INPUT internamente: parsing de tokens, roteamento entre modo 'intencao' (sem prefixo) e modo 'sugestoes' (com prefixo), geração de opções para o dropdown — tudo como lógica interna da engine.
4. THE BuscaEngine SHALL processar OMNIBOX_KEYDOWN internamente: navegação cíclica (ArrowDown/Up), execução (Enter), fechamento (Escape) — componente não interpreta teclas.
5. FOR ALL comportamentos do Omnibox (minChars, ordem de opções, maxCategorias, navegação cíclica, ação do Enter sem seleção), THE BuscaEngine SHALL ler a configuração de `rules/busca-rules.json` (bloco `omnibox.intencao`) em vez de hardcodar lógica.
6. THE BuscaEngine SHALL gerar um span OpenTelemetry para cada evento processado, incluindo tipo do evento, modo atual, e resultado da transição — permitindo observabilidade completa em tempo de execução.
7. THE `rules/busca-rules.schema.json` SHALL validar em CI a completude de todos os blocos de configuração do Omnibox (prefixos, intencao, placeholders, teclado, storeCard) — verificação compile-time.
8. FOR ALL componentes da página Descobrir (Omnibox, BuscaUnificada, StoreCard, LojaCard, CategoriaCard, MarketplaceFilter, BuscasSalvasPanel), o padrão SHALL ser: componente lê estado derivado da engine e emite eventos via `engine.send()` — sem estado local significativo.

### Requisito 11: Acessibilidade do Smart Dropdown

**User Story:** Como usuário de leitor de tela, eu quero que o Smart_Dropdown seja navegável e anuncie opções corretamente, para que eu tenha acesso igualitário às funcionalidades.

#### Critérios de Aceitação

1. THE Smart_Dropdown SHALL manter os atributos ARIA existentes: `role="combobox"` no input, `role="listbox"` no dropdown, `role="option"` em cada item, e `aria-expanded` refletindo o estado de visibilidade.
2. WHEN o conteúdo do Smart_Dropdown muda (novas opções de intenção geradas), THE Omnibox SHALL anunciar a contagem de opções via elemento `aria-live="polite"`.
3. WHEN uma opção de intenção está destacada via navegação por teclado, THE Omnibox SHALL atualizar `aria-activedescendant` para apontar para o item destacado.
4. THE Smart_Dropdown SHALL agrupar opções semanticamente usando `role="group"` com `aria-label` descritivo para cada grupo (ex: "Ações de busca", "Categorias", "Lojas").
5. WHEN o Smart_Dropdown exibe opções de intenção, cada opção SHALL ter um label acessível completo que inclua a ação e o contexto (ex: "Pesquisar glory em Produtos", não apenas "Produtos").
