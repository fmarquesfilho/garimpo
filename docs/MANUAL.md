# Manual do Garimpo (para quem usa no dia a dia)

O Garimpo é uma peneira inteligente: de centenas de produtos da Shopee, ele
separa os que **valem o esforço de divulgar** e te ajuda a publicá-los nos seus
canais (Telegram, WhatsApp) com formatação rica, foto e agendamento.

## Primeiro acesso

Ao abrir o app, você vê a **landing page**. Clique em **Entrar com Google** para
acessar. Sem login, nenhuma funcionalidade fica disponível.

## As telas

### 🔍 Buscar (página principal)

Busque produtos para divulgar. Digite um termo (perfume, sérum, batom) e veja
os melhores resultados com foto, preço, comissão e nome da loja.

**Busca:** campo principal no topo — digita, aperta Enter, resultados aparecem.
**Filtros avançados:** botão "⚙️ Filtros" — expande opções de comissão mínima,
vendas, nota e categoria.
**Cards de produto:** cada card mostra:
- Imagem (clicável — abre o produto na Shopee)
- Nome do produto
- Nome da loja (🏪)
- Categoria real da Shopee (Beleza, Moda, Eletrônicos, etc.)
- Preço + comissão + vendas + nota
- Botão "📤 Publicar" para enviar direto

### 🏪 Lojas

Monitore lojas específicas da Shopee. Adicione diretamente pela página (cole a
URL ou ID numérico) — não precisa mais ir à Curadoria.

- **Adicionar loja** — formulário no topo aceita:
  - URL: `https://shopee.com.br/shop/123456`
  - ID numérico: `123456`
- **Remover** — botão ✕ no card da loja.
- **Produtos** — lista completa da loja (sem filtro de elegibilidade), com botão
  de publicar direto.
- **🆕 Novidades** — produtos que apareceram pela primeira vez nos últimos 7 dias.
- **📉 Preços** — variações de preço com badges coloridos (verde ↓ queda,
  vermelho ↑ subida). Botão 📤 para publicar direto como oferta.
- **🔔 Alertas Telegram** — painel colapsável para configurar notificações
  automáticas de preço (ver seção Alertas abaixo).

### 🔔 Alertas de Preço

Notificações automáticas enviadas para um grupo de Telegram quando variações
significativas de preço são detectadas nas lojas monitoradas.

**Configuração (na página /lojas → 🔔 Alertas Telegram):**
- **Chat ID** — ID do grupo Telegram (ex.: `-1001234567890`).
- **Threshold** — variação mínima para disparar alerta (padrão: 15%).
- **Apenas quedas** — se ativo, só notifica quedas de preço (oportunidades).
- **Testar** — envia uma mensagem de confirmação ao grupo.

**Como funciona:** a cada coleta periódica de uma loja, o sistema compara preços
com os snapshots anteriores. Se detectar variação acima do threshold, envia
mensagem formatada ao grupo configurado.

**Formato da mensagem:**
```
🔔 Alerta de Preço
🏪 Loja: loja-123456

📉 Sérum Vitamina C 30ml
   R$ 89.90 → R$ 69.90 (↓22.2%)

⏰ 25/06 08:15
```

Também notifica produtos novos detectados (🆕).

**Env vars necessárias (Cloud Run):**
- `ALERTAS_TELEGRAM_CHAT_ID` — chat_id do grupo
- `ALERTAS_THRESHOLD` — ex.: `0.15` para 15%
- `ALERTAS_APENAS_QUEDAS` — `true` ou `false`
- `TELEGRAM_BOT_TOKEN` — já deve estar configurado

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

Visão geral da sua operação — o que importa para o dia a dia:

- **Cards de resumo:** lojas monitoradas, produtos rastreados, publicações enviadas
- **Lojas monitoradas:** lista com nome e cron configurado
- **Últimas publicações:** as 5 mais recentes com tempo relativo
- **Evolução de preço:** mini gráficos por loja (quando houver 2+ coletas)

Seletor de período: 7, 30 ou 90 dias.

## Os termos

**Teor** — o "grau de ouro" da pepita. Número de 0 a 1 que mede o quanto o
produto rende pelo esforço. Combina:
- **comissão** — quanto da venda volta pra você
- **valor esperado** — comissão × preço × vendas (retorno provável)
- **avaliação** — nota como sinal de confiança

**Nicho vs. Diversificada** — duas estratégias de ranking (descontinuada da interface):
- O sistema usa internamente a estratégia "nicho" (prioriza comissão + avaliação)
- A interface simplificada não expõe essa escolha ao usuário
- O ranking ordena automaticamente pelo melhor potencial de retorno

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
