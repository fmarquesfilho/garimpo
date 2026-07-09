# Manual do usuário

## Página Garimpar (/)

A página principal unifica descoberta e lojas monitoradas em uma única interface.
Produtos são ranqueados por **teor** (score composto de comissão × vendas × avaliação).

### Fontes de produtos

Use o toggle de fontes para alternar a origem dos produtos exibidos no grid:

- **🔍 Descobrir** — produtos da keyword/categoria configurada (ranking por teor)
- **🏪 Lojas** — produtos das lojas monitoradas (novidades e variações de preço)

Ambas as fontes podem estar ativas simultaneamente. Quando 🏪 Lojas está ativo,
chips de lojas aparecem para filtrar por loja específica.

### Ações disponíveis

- **Garimpar** — seleciona o produto e registra evento de curadoria
- **Publicar** — envia para um canal (Telegram ou WhatsApp)
- **Favoritar** — salva para consulta futura
- **Ver origem** — badge com país de origem (se configurado na loja)

### Selos informativos

- ⚠ **Suspeito** — comissão alta com zero vendas/nota (produto-fantasma)
- ✦ **Exploração** — produto fora do ranking usual (epsilon-greedy para diversificar)

### Controles em raias

As configurações de pesquisa ficam no topo da página, organizadas em **raias**
horizontais. No console superior há o campo de palavras-chave e três botões — **Filtros**,
**Lojas** e **Buscas** — cada um abrindo sua raia e mostrando um contador de quantas
configurações estão aplicadas. Ainda no topo: **colapsar tudo** e **limpar tudo**.

- **Raia Filtros** — toggles de fontes (🆕 Novos, 📉 Quedas, ⭐ Favoritos), filtros
  quantitativos (comissão mínima, vendas mínimas) e **categorias**. Ao digitar uma
  categoria, o autocomplete mostra o nome e os marketplaces a que ela pertence; cada
  categoria adicionada vira um card.
- **Raia Lojas** — escopa a busca em lojas específicas. Digite para buscar entre as **lojas
  monitoradas** (nome + marketplace) ou **cole um link/ID para adicionar uma loja nova**
  ("↳ resolver e adicionar"). Cada loja no escopo aparece como card com nome, marketplace,
  **bandeira de origem** e um indicador de monitoramento. Com lojas no escopo, a busca só
  roda nelas.
- **Raia Buscas** — buscas salvas e agendadas, cada uma num card com as seções que tiver
  (palavras-chave, categorias, lojas, marketplaces) e a info de agendamento. **Rodar**
  reexecuta a busca; **✎ editar** entra no edit mode para alterar e re-salvar a mesma busca
  (inclusive reagendar); **✕** remove.

Cada raia tem seu próprio **limpar raia**.

## Lojas monitoradas

Para adicionar uma loja nova ao monitoramento, na raia Lojas cole o link ou ID e escolha
"↳ resolver e adicionar":

1. Cole a URL (`shopee.com.br/shop/123456`) ou o ID numérico
2. Selecione o país de origem (ex: "🇰🇷 Coreia") — herdado por todos os produtos
3. Opcionalmente ajuste o cron (padrão: a cada 4h)

O sistema coleta automaticamente e detecta:
- **Novos produtos** — itens que apareceram desde a última coleta
- **Variações de preço** — quedas e altas com porcentagem

### Evolução de preço

A tela "Evolução" mostra série temporal de preço médio por loja monitorada,
com resumo global e top variações.

## Alertas de preço

Configure em ⚙️ Configuração → Alertas (seção colapsável) para receber notificações no Telegram quando um produto
monitorado cai de preço acima do threshold.

Configuração:
- **Chat ID** — grupo Telegram para notificações
- **Threshold** — variação mínima para alerta (ex: 15%)
- **Apenas quedas** — alertar só quando preço cai (oportunidades)

## Publicação

Ao publicar um produto:

1. Escolha o **destino** (canal Telegram ou grupo WhatsApp)
2. Escolha o **template** de mensagem
3. Opcionalmente edite a legenda ou agende para horário específico
4. O link de afiliado é gerado com `sub_id` para rastreamento

### Templates

Modelos de mensagem com placeholders: `{{nome}}`, `{{preco}}`, `{{categoria}}`.
Preview ao vivo antes de enviar.

### Agendamento

Publicações podem ser agendadas. O Cloud Scheduler dispara `POST /api/publicar-pendentes`
a cada hora para enviar as que passaram do horário.

## Conversões

Em "Conversões", veja o relatório real de vendas originadas pelos seus links.
Dados do `conversionReport` da Shopee, sincronizados 1×/dia.

Mostra: produto, loja, comissão, status (PENDING/COMPLETED/CANCELLED), canal.

## Estatísticas

Resumo de mercado por categoria nos últimos N dias:
- Comissão média e mediana
- Preço médio
- Vendas média
- Teor médio

## Onboarding

Configuração inicial (multi-tenant):
1. Aceitar termos (LGPD)
2. Configurar credenciais Shopee (App ID + Secret)
3. Configurar Telegram (opcional)
4. Validar credenciais com chamada de teste
