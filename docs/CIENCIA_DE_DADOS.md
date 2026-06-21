# Guia Estratégico de Ciência de Dados para o Garimpo

Curadoria de produtos para marketing de afiliados (Shopee + Instagram/Telegram).
Nível estratégico: objetivo de negócio → pergunta → técnica → pré-requisito de
dados → estágio de maturidade. Foca no que dá para fazer **já** e no que se
destrava com o **acúmulo de histórico**.

## TL;DR

- **Comece descritivo, não preditivo.** Com baixo volume (seleções + snapshots),
  o maior retorno vem de EDA, detecção de outliers para o produto-fantasma,
  filtros de credibilidade no "teor" e análise de séries temporais dos snapshots.
  Modelos preditivos/causais/bandits só passam a fazer sentido quando o
  `conversionReport` da Shopee acumular conversões reais por `subId`/canal.
- **A técnica certa depende do objetivo E do estágio de maturidade.** "Qual
  estratégia rende mais" tende a bandits/Thompson sampling — mas exige o sinal de
  conversão como recompensa, que ainda não existe; até lá, A/B + comparação
  descritiva. "Qual campanha funcionou" pede CausalImpact/BSTS, com a ressalva de
  que ela controla *quando* publica (confounding).
- **A maior armadilha é o feedback loop:** o sistema só observa conversões de
  produtos que ele mesmo recomendou. Isso enviesa qualquer modelo e métrica de
  ranking. A defesa é injetar exploração deliberada (hold-outs aleatórios) desde
  cedo — já implementado como a flag `exploracao`.

## O gargalo de tudo: o dado de conversão

A Shopee Affiliate Open API expõe o `conversionReport` (GraphQL) com `clickTime`,
`purchaseTime`, `totalCommission`, `netCommission`, `buyerType` e o campo
`utmContent` — onde vão os sub-ids (até 5 por link). É isso que permite atribuir
conversão a canal/campanha. A atribuição é **last-click com janela de 7 dias**, e
o pagamento ocorre **15–40 dias** após o pedido marcado "Completed". Esse atraso
material favorece bandits em *batch* (semanal/mensal), não online. Sem esse fluxo
ligado, preditivo/causal/bandit ficam sem variável-alvo.

## Por objetivo de negócio

### 1. Maximizar rentabilidade por hora de esforço

- **EDA / estatística descritiva** *(JÁ)* — distribuições por categoria, medianas
  (não médias — comissão/vendas são cauda longa), correlação comissão×vendas.
  Sem conversões, descreve o que foi *selecionado*, não o que *rendeu*.
- **KPIs de afiliado (EPC, conversion rate, AOV)** *(com conversões)* — EPC
  (comissão ÷ cliques) normaliza por tráfego e é a métrica-rainha. Reporte
  distribuição, não só média.
- **Ranking multicritério (o "teor")** *(JÁ; calibração depois)* — soma ponderada
  com min-max é adequada ao estágio. Alternativas: **TOPSIS** (lida melhor com
  trade-offs, não premia um critério extremo linearmente) e **Pareto/dominância**
  (filtra o pool às escolhas eficientes). Os pesos são uma *hipótese* até serem
  validados contra conversão.
- **Predição de conversão (classificação)** *(futuro — centenas+ de conversões)*
  — comece com regressão logística. Cuidado com o feedback loop.
- **Métricas de ranking (NDCG, precision@k, MAP)** *(futuro)* — a ferramenta para
  fechar o loop de calibração dos pesos do teor. Precisa de conversão como
  ground-truth e de hold-outs aleatórios para evitar position bias.

### 2. Crescer e engajar audiência

- **Cohort / survival analysis** *(médio prazo)* — Kaplan-Meier para a "vida" de
  um produto como conversor; Cox para quais features prolongam essa vida. Sobre
  posição-no-ranking (snapshots) dá para começar antes das conversões.
- **Atribuição multicanal** *(com conversões; instrumente o subId JÁ)* — qual
  canal (Telegram vs. Instagram) converte. É last-click, não a verdade causal
  completa.
- **Impacto de campanha / inferência causal** *(futuro)* — **CausalImpact/BSTS**
  é o método mais honesto sem grupo de controle (Brodersen et al., Google, 2015).
  Diff-in-diff e controle sintético exigem controles difíceis aqui. *Confounding
  central:* ela publica quando já vê o produto aquecendo → estimativas otimistas.

### 3. Diversificar nicho de forma inteligente (nicho vs. diversificada)

- **Bandits / Thompson sampling** *(futuro; é o destino deste objetivo)* — o
  problema "qual braço puxar com dados se acumulando" é literalmente um
  multi-armed bandit. Braços = estratégias/categorias/canais; recompensa =
  comissão ganha. Thompson sampling é robusto a recompensa atrasada (relevante
  dado o pagamento de 15–40 dias). Comece em batch. No curto prazo, a versão
  pobre é A/B manual nicho vs. diversificada.
- **Otimização de portfólio (Markowitz)** *(futuro; lente conceitual já)* — trata
  cada produto/categoria como ativo (retorno = comissão, risco = variância),
  buscando a fronteira eficiente. Já hoje serve como forma de pensar
  "não depender de poucos produtos".
- **Market basket / regras de associação** *(limitado)* — a API não expõe a cesta
  do comprador de forma rica. Baixa prioridade.

### 4. Reduzir risco

- **Detecção de anomalia (produto-fantasma)** *(JÁ — implementado)* — regra
  z-score/IQR + negócio (comissão no topo do pool E sem tração) embutida na
  curadoria como a flag `suspeito`. Com mais dados, Isolation Forest (anomalia
  multivariada). Cuidado: lançamento legítimo sem vendas ainda pode parecer
  fantasma — combine com idade do produto.
- **Séries temporais dos snapshots** *(médio prazo)* — **STL** (tendência +
  sazonalidade + resíduo, robusto a outliers) para medir aquecimento/esfriamento
  de categorias; detecção de change-point para alertas. STL precisa de ≥2 ciclos
  sazonais; tendência simples (variação %) funciona em semanas.
- **Forecasting** *(futuro)* — para demanda intermitente (muitos zeros, comum em
  cauda longa), método de **Croston** (e a correção SBA de Syntetos-Boylan).
  Métodos padrão falham com excesso de zeros. Não force em séries curtas.

## Trilha de maturidade (ordem de adoção)

**Fase 0 — decisões de engenharia para AGORA (alta alavancagem, custo ~zero):**
1. Instrumentar `subId` por canal/campanha. ✅ *Feito* (`canal_estrategia_data`).
2. Ligar o `conversionReport` ao BigQuery assim que possível — é o pré-requisito
   de tudo na metade preditiva/causal.
3. Reservar uma fração das publicações para exploração aleatória. ✅ *Feito*
   (flag `exploracao`).

**Fase 1 — descritivo (JÁ):** EDA; produto-fantasma com regras ✅; Pareto para
enxugar o pool; sanity-check dos pesos do teor (entropia/CRITIC); tendência
simples dos snapshots.

**Fase 2 — diagnóstico (semanas–meses de snapshots):** STL + change-point;
Isolation Forest; survival sobre tempo-no-ranking.

**Fase 3 — preditivo & medição honesta (precisa de conversões):** EPC e conversão
reais; **calibração do teor via NDCG/precision@k** (usando os hold-outs);
atribuição por canal; regressão logística de propensão.

**Fase 4 — causal & prescritivo (meses–um ano):** CausalImpact/BSTS para
campanhas; contextual bandits/Thompson em batch para nicho vs. diversificada;
otimização de portfólio; uplift *só* com randomização.

**Limiares que mudam a recomendação:**
- Conversões nunca ligadas → congele na Fase 2; preditivo/causal seriam overkill.
- Volume de publicações muito baixo → fique em scoring + EDA + regras de anomalia.
- Sem exploração aleatória → trate toda saída de modelo como sugestão enviesada.

## Ressalvas honestas

- Boa parte do valor está **bloqueada até as conversões acumularem**.
- A janela de 7 dias (last-click) e o delay de 15–40 dias são oficiais e
  materiais; parâmetros variam por mercado.
- O **feedback loop** é estrutural: não se corrige com mais dados, só com design
  de exploração desde cedo.
- **Confounding** na inferência causal: como ela escolhe quando publicar, todo
  "efeito de campanha" é otimista. Reporte intervalos e a limitação.
- Volume baixo penaliza métodos sofisticados (STL curto, covariâncias instáveis).
  Começar descritivo é a leitura correta do trade-off, não conservadorismo.

> Referências-chave: Brodersen et al. (2015) *Inferring causal impact using
> Bayesian structural time-series models*, Annals of Applied Statistics;
> Chapelle & Li (2011) *An Empirical Evaluation of Thompson Sampling*, NeurIPS;
> Croston (1972), Op. Research Quarterly; Cleveland et al. (STL, 1990).
