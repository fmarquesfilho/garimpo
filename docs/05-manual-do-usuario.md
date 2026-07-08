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

### BuscaUnificada

O componente de busca unificada (no topo da página) agrupa todas as configurações de pesquisa:

- **Keywords** — palavras-chave para descoberta de produtos
- **Lojas** — seletor plural (multi-marketplace); adicione via URL ou ID numérico, com país de origem
- **Filtros colapsáveis** — comissão mínima, vendas mínimas, categorias
- **Fontes de dados** — toggle entre Descobrir e Lojas (ToggleGroup)
- **Salvar** — grava a configuração como busca nomeada (exibida como chip)
- **Agendar** — define frequência de execução automática (cron)

Buscas salvas aparecem como chips clicáveis abaixo do campo de keywords, permitindo
alternar rapidamente entre configurações.

## Lojas monitoradas

Para adicionar uma loja, use o seletor de lojas dentro da BuscaUnificada:

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
