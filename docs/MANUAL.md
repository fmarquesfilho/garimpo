# Manual do Garimpo (para quem usa no dia a dia)

O Garimpo é uma peneira inteligente: de centenas de produtos da Shopee, ele
separa os que **valem o esforço de divulgar** e te ajuda a publicá-los nos seus
canais (Telegram, WhatsApp) com formatação rica, foto e agendamento.

## Primeiro acesso

Ao abrir o app, você vê a **landing page**. Clique em **Entrar com Google** para
acessar. Sem login, nenhuma funcionalidade fica disponível.

## As telas

### 🔍 Curadoria (página principal)

A peneira do dia. Busque por produto (perfume, sérum, batom), ajuste os filtros
e veja os melhores candidatos ordenados pelo **teor**.

**Filtros:** comissão mínima, vendas mínimas, nota mínima, quantidade de resultados,
modo explorar (testa produtos fora do topo).

**Buscas salvas:** salve combinações de filtros + keywords + lojas para reusar.
Configure um **cron** (ex.: todo dia 8h) para coleta automática.

### 🏪 Lojas

Monitore lojas específicas da Shopee. Selecione uma busca com shop_ids e veja:

- **Produtos** — lista completa da loja (sem filtro de elegibilidade), com botão
  de publicar direto.
- **🆕 Novidades** — produtos que apareceram pela primeira vez nos últimos 7 dias.
- **📉 Preços** — variações de preço detectadas (verde = baixou, vermelho = subiu).

### 📤 Publicar (página de publicação)

Ao clicar "Publicar" em qualquer produto, ou ao acessar direto:

1. **Cole um link** da Shopee (opcional) — preenche os dados automaticamente.
2. **Edite o produto** — nome, categoria, preço são editáveis inline.
3. **Escolha o destino** — qual grupo Telegram ou WhatsApp vai receber.
4. **Escolha o template** — modelo de mensagem (com ou sem foto 📷).
5. **Edite a legenda** — editor rico (negrito, itálico, links) com preview WYSIWYG.
6. **Envie ou agende** — imediato ou para um horário futuro.

### 📋 Publicações

Histórico de tudo que foi publicado:
- **Agendadas** — esperando o horário (⏱)
- **Enviadas** — publicadas com sucesso (✓)
- **Erros** — falhas de envio (✕)

### 📡 Destinos & Conversões

- **Destinos** — gerencie onde o Garimpo publica. Cada destino tem tipo
  (Telegram ou WhatsApp) e configuração (chat_id ou grupo(s) WhatsApp).
  - WhatsApp suporta até **5 grupos por destino** (a mensagem é enviada para todos).
  - Ao criar/editar um destino WhatsApp, o app mostra um autocomplete com os
    grupos disponíveis (selecione pelo nome, sem lidar com IDs).
  - Botão ✎ para editar um destino existente (adicionar/remover grupos).
- **Conversões** — relatório de publicações por canal/sub_id, mostrando volume
  e comissão estimada.

### ⏱ Coletas

Histórico das coletas periódicas (snapshots gravados pelo scheduler).

### 📊 Estatísticas

Análise de mercado baseada nos dados coletados por categoria: comissão
média/mediana, preço médio, vendas média, teor médio. Permite comparar
janelas de 7, 30 ou 90 dias.

## Os termos

**Teor** — o "grau de ouro" da pepita. Número de 0 a 1 que mede o quanto o
produto rende pelo esforço. Combina:
- **comissão** — quanto da venda volta pra você
- **valor esperado** — comissão × preço × vendas (retorno provável)
- **avaliação** — nota como sinal de confiança

**Nicho vs. Diversificada** — duas estratégias de ranking:
- *Nicho* bonifica cosméticos/perfumaria/bem-estar (foco editorial)
- *Diversificada* ignora categoria, olha só retorno × demanda
- *Comparar* mostra ambas lado a lado

**⚠ Suspeito** — comissão alta sem vendas/nota. Produto-fantasma. Marcado, não
escondido — você decide.

**✦ Exploração** — ~20% das vagas para testar o que converte fora do topo.

**sub_id** — código de atribuição (ex.: `telegram_nicho_20260622`) que identifica
qual canal/estratégia/data gerou cada venda.

**Pipeline de filtros** — o sistema filtra em cadeia (comissão → vendas → nota).
Na página de Lojas, os filtros são desligados para mostrar tudo.

## Dúvidas frequentes

- *Produto de 80% lá embaixo?* Sem vendas → teor baixo. Marcado ⚠ suspeito.
- *Lista vazia?* Busca vazia = sem resultados. Digite um termo.
- *Não encontro uma loja?* Use o modo "sem filtro" na página Lojas, que mostra
  todos os produtos independente de comissão/vendas.
- *Como agendar?* Na página Publicar, preencha o campo de data/hora antes de enviar.
- *Como editar a mensagem?* Use o editor rico na página Publicar — negrito, itálico,
  links, tudo visual.
