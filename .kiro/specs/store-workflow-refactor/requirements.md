# Requirements Document

## Introduction

O subsistema de lojas (store) da página Descobrir é o ponto fraco atual da plataforma Garimpei.
Após o sucesso do Omnibox (T-0055), a FSM (BuscaEngine) está sólida, porém o workflow de lojas
sofre de acoplamento implícito, modelo de dados inconsistente, e estados de resolução frágeis.

Este documento define os requisitos para refatorar o subsistema mantendo a BuscaEngine e o Omnibox
intactos, criando um modelo de dados de loja independente, robusto e testável.

## Glossary

- **BuscaEngine**: Máquina de estados headless (classe Svelte 5) que gerencia a página Descobrir. Permanece inalterada estruturalmente.
- **Omnibox**: Input unificado com prefixos (`@loja`, `#categoria`, `!marketplace`). Permanece inalterado.
- **Loja_Monitorada**: Loja que possui agendamento de coleta (cron) — dados são coletados periodicamente pelo Collector.
- **Loja_Escopada**: Loja adicionada ao contexto de busca atual para filtrar resultados, sem necessidade de agendamento.
- **Registro_de_Loja**: Entidade independente que representa uma loja conhecida pelo sistema, com nome canônico, marketplace, e shopId.
- **Collector_API**: Serviço gRPC backend que resolve URLs/IDs de lojas em marketplaces e retorna metadados (shopId, nome, marketplace).
- **Resolução_de_Loja**: Processo assíncrono de identificar uma loja a partir de input textual (URL, ID numérico, ou nome) via Collector_API.
- **shopId**: Identificador numérico único de uma loja dentro de um marketplace.
- **shopMeta**: Metadados associados a um shopId no contexto (marketplace, origem, status de monitoramento, cron).
- **Nome_Canônico**: Nome oficial da loja retornado pelo marketplace, armazenado no Registro_de_Loja como referência primária.
- **Nome_Normalizado**: Versão lowercase, sem espaços e sem caracteres especiais do Nome_Canônico, usada para matching.
- **Token_Loja**: Fragmento textual precedido por `@` no Omnibox que identifica intenção de busca por loja.

## Requirements

### Requisito 1: Registro Independente de Lojas

**User Story:** Como usuário, eu quero que as lojas conhecidas existam independentemente das buscas salvas, para que eu possa adicionar lojas ao escopo sem precisar ter uma busca salva que as contenha.

**Decisão arquitetural:** O Registro_de_Loja é persistido server-side — nova tabela PostgreSQL gerenciada via EF Core na API C#, com endpoints REST (`GET /api/lojas`, `POST /api/lojas/resolver`). O frontend consome via chamadas HTTP na inicialização e após resolução.

#### Critérios de Aceitação

1. THE Registro_de_Loja SHALL armazenar para cada loja: shopId (inteiro positivo), Nome_Canônico (string, máximo 200 caracteres), Nome_Normalizado (string derivada), marketplace (string obrigatória), flag de monitoramento (booleano), e cron (string cron válida quando monitorada, null quando não monitorada).
2. THE Registro_de_Loja SHALL garantir unicidade por par (shopId, marketplace) — uma mesma loja não pode existir duplicada no registro.
3. WHEN a BuscaEngine inicializa, THE BuscaEngine SHALL carregar a lista de lojas disponíveis a partir do Registro_de_Loja como fonte primária de lojas conhecidas.
4. IF o Registro_de_Loja está vazio durante a inicialização, THEN THE BuscaEngine SHALL inicializar com lista de lojas vazia sem erro, permitindo adição posterior via resolução ou migração.
5. WHEN uma loja é resolvida com sucesso via Collector_API, THE Registro_de_Loja SHALL persistir a loja automaticamente; se o par (shopId, marketplace) já existe, SHALL atualizar Nome_Canônico e Nome_Normalizado com os dados mais recentes da API.
6. THE Registro_de_Loja SHALL manter compatibilidade retroativa com lojas derivadas de buscas salvas existentes, incluindo-as no registro durante a migração conforme definido no Requisito 7.

### Requisito 2: Separação entre Loja Monitorada e Loja Escopada

**User Story:** Como usuário, eu quero entender claramente a diferença entre lojas que monitoro (coleta automática) e lojas que apenas filtram meus resultados, para que eu saiba o que cada loja faz no meu contexto.

#### Critérios de Aceitação

1. THE BuscaEngine SHALL classificar cada loja no contexto como Loja_Monitorada quando o campo cron no Registro_de_Loja é não-nulo e ativo, e como Loja_Escopada quando o campo cron é nulo ou inexistente.
2. WHEN o usuário adiciona uma Loja_Monitorada ao escopo, THE BuscaEngine SHALL definir `shopMeta.origem` como "monitorada" e habilitar os toggles das fontes Quedas, Novos e Lojas para essa loja no painel de fontes.
3. WHEN o usuário adiciona uma Loja_Escopada ao escopo, THE BuscaEngine SHALL definir `shopMeta.origem` como "escopada" e habilitar apenas a fonte Curadoria filtrada por shopId, mantendo os toggles Quedas, Novos e Lojas desabilitados para essa loja.
4. THE BuscaEngine SHALL exibir um indicador visual persistente no card de cada loja no contexto que identifique inequivocamente o tipo: um badge indicando coleta ativa para Loja_Monitorada e ausência desse badge para Loja_Escopada.
5. WHEN o campo cron de uma loja no Registro_de_Loja é atualizado de nulo para ativo (promoção), THE BuscaEngine SHALL reclassificar a loja como Loja_Monitorada no contexto atual e habilitar os toggles Quedas, Novos e Lojas sem exigir remoção e re-adição manual.
6. WHEN o campo cron de uma loja no Registro_de_Loja é removido ou desativado (demoção), THE BuscaEngine SHALL reclassificar a loja como Loja_Escopada no contexto atual e desabilitar os toggles Quedas, Novos e Lojas para essa loja.

### Requisito 3: Normalização e Matching de Nomes de Loja

**User Story:** Como usuário, eu quero encontrar lojas no autocomplete mesmo digitando nomes parciais, sem espaços, ou com variações de formatação, para que a busca por loja funcione de forma intuitiva.

#### Critérios de Aceitação

1. THE Registro_de_Loja SHALL gerar um Nome_Normalizado ao persistir uma loja, aplicando na ordem: decomposição Unicode (NFD) com remoção de combining marks (acentos), remoção de todos os caracteres que não sejam letras ASCII (a-z) ou dígitos (0-9), e conversão para lowercase — resultando em uma string contendo apenas `[a-z0-9]`.
2. WHEN o usuário digita um Token_Loja com 2 ou mais caracteres após o prefixo `@` no Omnibox, THE Omnibox SHALL normalizar o texto digitado usando a mesma função de normalização do critério 1, e comparar o resultado contra o Nome_Normalizado e o Nome_Canônico (lowercase) de cada loja no Registro_de_Loja.
3. WHEN o texto normalizado do Token_Loja é substring do Nome_Normalizado de uma loja, THE Omnibox SHALL incluir a loja nas sugestões do autocomplete.
4. WHEN o texto normalizado do Token_Loja é substring do Nome_Canônico (lowercase, preservando espaços) de uma loja, THE Omnibox SHALL incluir a loja nas sugestões do autocomplete.
5. IF o Token_Loja após normalização resulta em string vazia (0 caracteres retidos), THEN THE Omnibox SHALL não exibir sugestões de lojas para aquele token.
6. FOR ALL lojas com Nome_Canônico contendo espaços, THE Registro_de_Loja SHALL garantir que normalize(input_sem_espacos) produz substring match em Nome_Normalizado equivalente ao que normalize(input_com_espacos) produziria — propriedade verificável via testes parametrizados com pares como ("gloryofseoul", "Glory of Seoul") e ("glory", "Glory of Seoul").

### Requisito 4: Fluxo de Resolução de Loja com Estados Explícitos

**User Story:** Como usuário, eu quero feedback claro quando o sistema está resolvendo uma loja nova (por URL ou ID), para que eu saiba se deu certo ou o que deu errado.

#### Critérios de Aceitação

1. WHEN o evento ADICIONAR_LOJA é disparado com input textual cujo valor não corresponde a nenhum shop_id já presente no Registro_de_Loja, THE BuscaEngine SHALL transicionar o estado de resolução para "resolvendo" e armazenar o input no campo `input` do sub-estado.
2. WHEN a Collector_API retorna sucesso, THE BuscaEngine SHALL transicionar o estado de resolução para "sucesso", adicionar a loja ao contexto da máquina de estados, e persistir no Registro_de_Loja.
3. IF a Collector_API retorna erro, THEN THE BuscaEngine SHALL transicionar o estado de resolução para "erro" com a mensagem descritiva retornada pela API no campo `erro` do sub-estado.
4. IF a Collector_API não responde em 10 segundos, THEN THE BuscaEngine SHALL transicionar o estado de resolução para "erro" com mensagem indicando timeout.
5. WHILE o estado de resolução é "resolvendo", THE BuscaEngine SHALL rejeitar eventos ADICIONAR_LOJA subsequentes sem efeito colateral (o evento é descartado e o input do usuário permanece inalterado).
6. IF o estado de resolução é "erro", THEN THE BuscaEngine SHALL permitir retry via novo evento ADICIONAR_LOJA com o mesmo input, transicionando de volta para "resolvendo".
7. THE BuscaEngine SHALL representar o estado de resolução como um sub-estado tipado (`{status: 'idle'|'resolvendo'|'sucesso'|'erro', input?: string, erro?: string}`) em vez de flags booleanas separadas.
8. WHEN o estado de resolução transiciona para "sucesso", THE BuscaEngine SHALL retornar automaticamente ao estado "idle" após a loja ser adicionada ao contexto (transição síncrona, sem delay).
9. IF o evento ADICIONAR_LOJA é disparado com input vazio ou composto apenas por espaços em branco, THEN THE BuscaEngine SHALL permanecer no estado "idle" e não invocar a Collector_API.
10. WHEN o evento ADICIONAR_LOJA é disparado com input que corresponde a uma loja já presente no Registro_de_Loja, THE BuscaEngine SHALL permanecer no estado "idle" e selecionar a loja existente no contexto sem invocar a Collector_API.

### Requisito 5: Scoping Explícito de Marketplace

**User Story:** Como usuário, eu quero que o marketplace de cada loja seja sempre rastreável e consistente, para que resultados de busca nunca misturem lojas de marketplaces diferentes sem meu consentimento.

#### Critérios de Aceitação

1. THE Registro_de_Loja SHALL armazenar o marketplace como campo obrigatório para cada loja, aceitando exclusivamente valores presentes na lista de marketplaces suportados definida em busca-rules.
2. WHEN a Collector_API resolve uma loja, THE BuscaEngine SHALL usar o marketplace retornado pela API como fonte autoritativa e atualizar o Registro_de_Loja caso o valor armazenado difira do retornado.
3. WHEN uma loja é adicionada a partir do Registro_de_Loja (já conhecida), THE BuscaEngine SHALL usar o marketplace armazenado no registro sem chamar a Collector_API.
4. IF o marketplace não puder ser determinado (Collector_API retorna campo vazio ou nulo, e loja não existe no Registro_de_Loja), THEN THE BuscaEngine SHALL rejeitar a adição da loja ao contexto e transicionar o estado de resolução para "erro" com mensagem indicando que o marketplace não pôde ser identificado.
5. WHEN o contexto contém lojas de múltiplos marketplaces, THE BuscaEngine SHALL preservar a associação marketplace-por-loja individualmente no shopMeta.
6. IF a Collector_API retorna um valor de marketplace ausente da lista de suportados, THEN THE BuscaEngine SHALL rejeitar a adição da loja ao contexto e transicionar o estado de resolução para "erro" com mensagem indicando marketplace não suportado.

### Requisito 6: Separação dos Caminhos de Adição de Loja

**User Story:** Como desenvolvedor, eu quero que os caminhos de adição de loja (conhecida vs. nova) sejam explícitos e testáveis independentemente, para que bugs em um caminho não afetem o outro.

#### Critérios de Aceitação

1. WHEN o evento ADICIONAR_LOJA contém o campo `loja` com um objeto que possui `id` presente no Registro_de_Loja, THE BuscaEngine SHALL executar o caminho síncrono de adição direta (sem chamar Collector_API).
2. WHEN o evento ADICIONAR_LOJA contém o campo `value` (string textual) sem correspondência no Registro_de_Loja, THE BuscaEngine SHALL executar o caminho assíncrono de resolução via Collector_API.
3. THE BuscaEngine SHALL validar a estrutura do evento ADICIONAR_LOJA antes de decidir o caminho: o evento deve conter exatamente um dos discriminantes (`loja` com `id`, ou `value` string não-vazia); caso contrário, SHALL transicionar o estado de resolução para "erro" com mensagem descritiva.
4. FOR ALL lojas adicionadas por qualquer caminho, o estado final no contexto (shopIds, shopNomes, shopMeta) SHALL ter a mesma estrutura com campos obrigatórios preenchidos: shopId em shopIds, nome em shopNomes, e marketplace + origem + monitorada + cron em shopMeta.
5. WHEN o caminho síncrono é executado (loja conhecida), THE BuscaEngine SHALL não transicionar o estado de resolução (permanece "idle") e não ativar indicador de loading.


