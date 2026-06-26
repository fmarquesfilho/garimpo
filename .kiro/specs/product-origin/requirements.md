# Requirements Document

## Introduction

A funcionalidade de Origem do Produto permite que a afiliada identifique rapidamente se um produto é genuinamente de Coreia ou Japão, diferenciando produtos autênticos de falsificações. A abordagem primária é extrair dados de origem diretamente da API de Afiliados da Shopee (descobrindo campos não explorados via introspecção GraphQL). Caso a API não exponha essa informação, um mecanismo de fallback permite que lojas monitoradas sejam marcadas com uma origem padrão, herdada automaticamente por todos os seus produtos.

## Glossary

- **Garimpei**: Sistema SaaS de descoberta, ranking e publicação de produtos para afiliados Shopee.
- **API_Shopee**: API GraphQL de Afiliados da Shopee (`POST https://open-api.affiliate.shopee.com.br/graphql`), usada para buscar produtos e gerar links de afiliado.
- **Introspecção_GraphQL**: Consulta padrão do protocolo GraphQL (`__type`, `__schema`) que retorna o schema completo da API, incluindo campos disponíveis em cada tipo de nó.
- **Produto_Candidato**: Entidade `domain.Product` que representa um produto retornado pela API Shopee e avaliado pelo motor de ranking.
- **Busca**: Perfil de coleta (struct `store.Busca`) que define keywords, lojas monitoradas e filtros para cada ciclo de coleta.
- **Badge_Origem**: Indicador visual no card do produto que exibe o país de origem com bandeira (ex.: "🇰🇷 Coreia", "🇯🇵 Japão").
- **Origem_Padrão**: Campo configurável numa Busca com `shop_ids`, indicando que todos os produtos daquela loja são de determinada origem verificada.
- **CandidateCard**: Componente SvelteKit que renderiza um produto candidato no feed da afiliada.
- **Snapshot**: Foto periódica do catálogo de uma categoria/loja, armazenada para análise histórica.

## Requirements

### Requirement 1: Introspecção da API para Descoberta de Campos

**User Story:** Como desenvolvedor do Garimpei, eu quero executar uma introspecção GraphQL na API de Afiliados da Shopee, para descobrir se existem campos relacionados à origem do produto (país de origem, tipo de loja, localização, marca) que não estão sendo utilizados atualmente.

#### Acceptance Criteria

1. WHEN o desenvolvedor executa a tarefa de introspecção, THE Garimpei SHALL enviar uma query de introspecção GraphQL (`__type(name: "ProductOfferV2")`) para a API_Shopee e registrar em log estruturado (stdout em formato JSON) a lista completa de campos disponíveis no tipo de nó `productOfferV2`, incluindo para cada campo: nome, tipo e indicador de obrigatoriedade (nullable ou não).
2. WHEN a resposta de introspecção é recebida com sucesso, THE Garimpei SHALL identificar e listar separadamente os campos cujo nome contenha pelo menos um dos seguintes termos (case-insensitive): `origin`, `country`, `shop_type`, `shopType`, `location`, `brand`, `seller`, `warehouse`, `domestic`, `local`, `imported`, `cross_border`.
3. IF a resposta de introspecção não contiver nenhum campo correspondente aos termos de busca do critério 2, THEN THE Garimpei SHALL registrar em log uma mensagem indicando que nenhum campo relacionado a origem foi encontrado no tipo `productOfferV2`.
4. IF a API_Shopee retornar erro na query de introspecção (códigos 10020, 10035 ou 11001) ou a requisição falhar por timeout (limite: 20 segundos) ou erro de rede, THEN THE Garimpei SHALL registrar em log o tipo de falha, código de erro (quando disponível) e mensagem descritiva para diagnóstico.

---

### Requirement 2: Extração de Origem via API (Abordagem Primária)

**User Story:** Como afiliada, eu quero que o sistema extraia automaticamente a informação de origem do produto da API Shopee, para que eu saiba se o produto é genuinamente coreano ou japonês sem verificação manual.

#### Acceptance Criteria

1. WHEN um campo de origem é identificado na introspecção (Requisito 1), THE Adaptador_Shopee SHALL incluir o nome exato do campo descoberto na lista de campos da query GraphQL `productOfferV2`, tanto na variante por keyword quanto na variante por shopId.
2. WHEN a API_Shopee retorna dados de origem para um produto com valor não-vazio, THE Adaptador_Shopee SHALL mapear o valor retornado, sem transformação, para o campo `Origin` da entidade Produto_Candidato.
3. WHEN a API_Shopee retorna um valor de origem vazio, nulo, ou o campo está ausente na resposta JSON de um produto, THE Adaptador_Shopee SHALL atribuir string vazia ("") ao campo `Origin` do Produto_Candidato.
4. THE Entidade Produto_Candidato SHALL conter um campo `Origin` do tipo string, com comprimento máximo de 100 caracteres, para armazenar o país de origem informado pela API.
5. IF a API_Shopee retornar erro na query que inclui o campo de origem (ex.: campo inválido ou não reconhecido), THEN THE Adaptador_Shopee SHALL registrar o erro com código e mensagem, remover o campo de origem da query, e repetir a requisição sem o campo para que a coleta não seja interrompida.

---

### Requirement 3: Origem Padrão por Loja (Fallback)

**User Story:** Como afiliada, eu quero poder marcar lojas monitoradas com uma origem padrão (ex.: "Coreia"), para que todos os produtos dessas lojas herdem automaticamente o badge de origem quando a API não fornece esse dado.

#### Acceptance Criteria

1. THE Entidade Busca SHALL conter um campo `OrigemPadrao` do tipo string, com comprimento máximo de 60 caracteres, para armazenar o país de origem padrão atribuído a Buscas que possuem `shop_ids` configurados. Um valor é considerado "preenchido" quando a string, após remoção de espaços em branco nas extremidades, possui ao menos 1 caractere.
2. WHEN o Motor_de_Coleta conclui a coleta de produtos de uma Busca que possui `shop_ids` configurados e `OrigemPadrao` preenchido, IF o campo `Origin` de um Produto_Candidato estiver vazio após a resposta da API_Shopee, THEN THE Motor_de_Coleta SHALL atribuir o valor de `OrigemPadrao` da Busca ao campo `Origin` desse Produto_Candidato.
3. WHEN o Motor_de_Coleta conclui a coleta de produtos de uma Busca que possui `OrigemPadrao` preenchido, IF a API_Shopee retorna um valor não-vazio de origem para o produto, THEN THE Motor_de_Coleta SHALL manter o valor retornado pela API_Shopee no campo `Origin` do Produto_Candidato, ignorando o `OrigemPadrao` da Busca.
4. WHEN o Motor_de_Coleta conclui a coleta de produtos de uma Busca que não possui `OrigemPadrao` preenchido, THE Motor_de_Coleta SHALL manter o campo `Origin` do Produto_Candidato exatamente como recebido da API (vazio ou preenchido).
5. IF a afiliada tentar salvar uma Busca com `OrigemPadrao` contendo mais de 60 caracteres ou apenas espaços em branco, THEN THE Sistema SHALL rejeitar a operação e retornar uma mensagem de erro indicando que o valor de origem padrão é inválido.

---

### Requirement 4: Exibição do Badge de Origem na UI

**User Story:** Como afiliada, eu quero ver um badge visual com a bandeira e nome do país no card de cada produto, para identificar rapidamente produtos genuínos de Coreia ou Japão no feed.

#### Acceptance Criteria

1. WHEN um Produto_Candidato possui campo `Origin` cujo valor, após normalização case-insensitive e remoção de acentos, corresponde a "coreia", "korea", "south korea" ou "coréia", THE CandidateCard SHALL exibir o Badge_Origem "🇰🇷 Coreia".
2. WHEN um Produto_Candidato possui campo `Origin` cujo valor, após normalização case-insensitive e remoção de acentos, corresponde a "japao", "japão" ou "japan", THE CandidateCard SHALL exibir o Badge_Origem "🇯🇵 Japão".
3. WHEN um Produto_Candidato possui campo `Origin` preenchido com valor que não corresponde a nenhuma das variações de Coreia ou Japão definidas nos critérios 1 e 2, THE CandidateCard SHALL exibir o Badge_Origem contendo apenas o nome do país como recebido no campo `Origin`, truncado em 20 caracteres seguido de "…" caso exceda esse limite, e sem bandeira emoji.
4. WHEN um Produto_Candidato possui campo `Origin` vazio ou composto apenas de espaços em branco, THE CandidateCard SHALL omitir o Badge_Origem do card.
5. THE Badge_Origem SHALL ser renderizado na seção `.meta` do CandidateCard, na mesma linha dos selos existentes (loja, categoria), utilizando o estilo `.selo` já definido no componente.
6. WHEN o Badge_Origem é exibido, THE CandidateCard SHALL renderizar o badge com `font-size` igual ao dos demais selos da seção `.meta`, garantindo que o badge não ultrapasse uma linha de texto.

---

### Requirement 5: Filtro por Origem no Feed

**User Story:** Como afiliada, eu quero filtrar produtos por país de origem no feed, para focar exclusivamente em produtos coreanos ou japoneses durante minha curadoria.

#### Acceptance Criteria

1. THE Interface_de_Feed SHALL oferecer um filtro de origem com opções que incluam no mínimo: "Todos", "Coreia", "Japão", sendo "Todos" a opção selecionada por padrão ao carregar o feed.
2. WHEN a afiliada seleciona um filtro de origem específico (ex.: "Coreia"), THE Interface_de_Feed SHALL exibir apenas Produtos_Candidatos cujo campo `Origin`, após normalização conforme as regras do Requisito 4 (variações mapeadas para "Coreia" ou "Japão"), corresponda ao filtro selecionado, excluindo produtos com `Origin` vazio.
3. WHEN a afiliada seleciona "Todos", THE Interface_de_Feed SHALL exibir todos os Produtos_Candidatos independente do valor de `Origin`, incluindo produtos com `Origin` vazio.
4. WHEN nenhum produto no feed corresponde ao filtro de origem selecionado, THE Interface_de_Feed SHALL exibir uma mensagem informando que nenhum produto foi encontrado para a origem selecionada, no lugar da lista de produtos.
5. WHEN a afiliada seleciona um filtro de origem, THE Interface_de_Feed SHALL atualizar a lista de produtos exibidos em no máximo 1 segundo após a seleção.

---

### Requirement 6: Persistência da Origem nos Snapshots

**User Story:** Como desenvolvedor do Garimpei, eu quero que a informação de origem seja persistida nos snapshots de mercado, para manter o histórico de quais produtos de cada coleta têm origem verificada.

#### Acceptance Criteria

1. THE Entidade `ItemSnapshot` SHALL incluir o campo `Origin` do tipo string com comprimento máximo de 100 caracteres.
2. WHEN um snapshot é registrado, THE Sistema_de_Persistência SHALL armazenar o valor de `Origin` de cada item junto aos demais campos (`Posicao`, `ProdutoID`, `Nome`, `Preco`, `Comissao`, `Vendas`, `Nota`, `Score`).
3. IF o valor de `Origin` de um produto não estiver disponível no momento da coleta, THEN THE Sistema_de_Persistência SHALL armazenar uma string vazia (`""`) no campo `Origin` desse item.
4. WHEN snapshots históricos são consultados (Novidades, Evolução de Lojas, Estatísticas), THE Sistema_de_Persistência SHALL retornar o campo `Origin` junto com os demais dados do produto em todas as respostas que expõem itens de snapshot.

---

### Requirement 7: Configuração de Origem Padrão na Interface de Lojas

**User Story:** Como afiliada, eu quero configurar a origem padrão de uma loja monitorada pela interface web, para que novos produtos dessa loja recebam automaticamente o badge de origem.

#### Acceptance Criteria

1. WHEN a afiliada edita uma Busca com `shop_ids`, THE Interface_de_Configuração SHALL exibir um campo "Origem Padrão" com as opções pré-definidas (Coreia, Japão, China), uma opção de texto livre limitada a 30 caracteres, e uma opção vazia representando "sem origem".
2. WHEN a afiliada salva uma Busca com `OrigemPadrao` preenchido com um valor válido (uma das opções pré-definidas ou texto livre de 1 a 30 caracteres), THE API_Backend SHALL persistir o valor no campo `origem_padrao` da Busca no banco de dados e retornar confirmação de sucesso.
3. IF a afiliada tenta salvar uma Busca com `OrigemPadrao` contendo texto livre que excede 30 caracteres, THEN THE API_Backend SHALL rejeitar a requisição com uma mensagem de erro indicando o limite de caracteres, sem alterar o valor previamente salvo.
4. WHEN a afiliada remove o valor de `OrigemPadrao` de uma Busca, THE API_Backend SHALL salvar o campo como vazio no banco de dados.
5. WHEN o serviço de coleta executa uma Busca que possui `origem_padrao` preenchido, THE Serviço_de_Coleta SHALL atribuir automaticamente o valor de `origem_padrao` como badge de origem a cada produto novo coletado nessa execução.
6. WHEN o serviço de coleta executa uma Busca que possui `origem_padrao` vazio, THE Serviço_de_Coleta SHALL não atribuir badge de origem aos produtos coletados nessa execução.
