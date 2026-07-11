# Visão e negócio

## O que é o Garimpei

Garimpei é uma plataforma de curadoria inteligente de produtos de afiliados Shopee.
Seleciona automaticamente os melhores produtos por "teor" (score composto de comissão,
vendas e avaliação) e publica em canais de venda (Telegram, WhatsApp).

## Proposta de valor

Transformar a seleção diária de produtos — hoje manual, demorada e sem dado — numa
decisão assistida, rastreável e comparável.

O diferencial defensável não é o ranking em si: é **cruzar a curadoria com o resultado
real** ao longo do tempo. O Garimpei aprende o que converte para a audiência específica,
no nicho específico.

## Público-alvo

Afiliados Shopee que precisam selecionar produtos de nicho com alta chance de conversão
sem gastar horas navegando manualmente.

## Modelo de receita

Comissão de afiliado Shopee sobre vendas originadas pelos links publicados.
Atribuição via `sub_id` (utm_content) no conversionReport.

## Jornada do usuário

O fluxo de valor ponta a ponta:

```
mercado Shopee → COLETA → CURADORIA → DECISÃO → PUBLICAÇÃO → ATRIBUIÇÃO → APRENDIZADO
  (oferta)     (snapshot)   (teor)    (escolha)   (canal)     (sub_id)     (análise)
```

### Etapas

1. **Descobrir** (página `/`) — O motor de ranking entrega os melhores candidatos.
   A ordenação por teor poupa o garimpo manual diário (a maior dor original).

2. **Avaliar** — Entender *por que* um produto é boa aposta: barra de componentes
   (explicabilidade), selos ⚠ suspeito / ✦ exploração. Decisão de risco: descartar
   produto-fantasma de comissão alta.

3. **Selecionar** (Garimpar) — Comprometer-se com um candidato. Gera evento de
   seleção (dado de comportamento). É aqui que a estratégia vira ação.

4. **Publicar** — A oferta chega à audiência com formatação rica (foto, negrito,
   link inline). Escolha de destino, template, agendamento. Gera evento de
   publicação com `sub_id`.

5. **Monitorar** — Lojas monitoradas alertam sobre novos produtos e quedas de preço.

6. **Analisar** — Dashboard de conversões reais, evolução de mercado, estatísticas.
   Fechar o laço — saber o que realmente converteu, por canal e estratégia.

### Momentos que decidem a rentabilidade

- **Filtrar o fantasma** — evitar comissão alta sem tração protege a audiência.
- **Explorar vs. explotar** — a flag `exploracao` é investimento em aprender.
- **Escolha de canal** — só mensurável com `sub_id` + conversões.

## Estratégia de ranking

Apenas a estratégia **nicho** está ativa. A "diversificada" foi descontinuada da UI.
O código do Strategy pattern permanece como dívida técnica documentada.

Ver [ADR 0002](/decisoes/0002-so-nicho).

## Canais de publicação

Ativos: **Telegram** e **WhatsApp** (Meta Cloud API).

> 🔮 Planejado: Instagram como canal de publicação (não implementado).
> A Graph API permite publicar conteúdo mas não coloca link clicável em feed
> nem expõe clique de afiliado. Atribuição continuaria vindo da Shopee.

## Visão futura

> 🔮 Planejado: Multi-tenant com onboarding self-service.

> 🔮 Planejado: Comparação de estratégias com dados reais de conversão.

> 🔮 Planejado: IA integrada para recomendação e geração de copy.
