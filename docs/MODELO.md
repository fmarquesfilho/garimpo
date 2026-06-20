# Garimpo — Modelo de Negócio e Processo

Documento-arcabouço da prova de conceito. Trata a operação de afiliados da sua
esposa como um **fluxo de entrega de valor** e aplica nela os conceitos de
Processos de Software: VSM para achar o gargalo, Kanban para operar, entrega
incremental para construir, e métricas para decidir.

---

## 1. O fluxo atual (Value Stream Mapping do estado atual)

A operação dela, hoje, ponta a ponta:

| # | Etapa | Tipo | Onde dói |
|---|-------|------|----------|
| 1 | Abrir aba de afiliados / best-sellers da Shopee | manual, **diária** | repetitivo |
| 2 | Varrer produtos e filtrar comissão ≥ 7% | manual, a olho | **cansativo, propenso a erro** |
| 3 | Escolher o produto do dia | julgamento | **gargalo — sem dado de apoio** |
| 4 | Selecionar um vídeo pronto | manual | ok |
| 5 | Montar post + link de afiliado | manual | repetitivo |
| 6 | Publicar no Instagram | manual | ok |
| 7 | Acompanhar resultado | quase inexistente | **sem feedback** |

**O gargalo é o bloco 1–3: a descoberta e seleção do produto.** É o passo que
ela refaz todo dia, na mão, sem dado — o equivalente a um *build step* manual num
pipeline. Em linguagem Lean, é a *muda* (desperdício) mais cara: trabalho
repetitivo de baixo valor agregado que consome justamente o tempo que deveria ir
para criação de conteúdo. E o passo 7 (medição) praticamente não existe, então
hoje não há como saber se uma escolha foi boa.

**Estado futuro (alvo):** o Garimpo assume os passos 1–3 entregando, toda manhã,
uma lista priorizada de candidatos elegíveis. Ela passa a *escolher de uma
short-list curada* em vez de garimpar do zero. E o passo 7 deixa de ser opcional:
o encurtador próprio captura cliques e fecha o laço de feedback.

---

## 2. Proposta de valor

> Transformar a seleção diária de produtos — hoje manual, demorada e sem dado —
> numa decisão assistida, rastreável e comparável entre estratégias.

O diferencial defensável não é o ranking em si: é **cruzar a curadoria com o
resultado real** ao longo do tempo. Nenhuma ferramenta paga sabe o que converte
*para a audiência dela*, no *nicho dela*. O Garimpo aprende isso.

---

## 3. As duas estratégias (e por que não são excludentes)

| | **Nicho** | **Diversificada** |
|---|---|---|
| Tese | construir audiência fiel em cosméticos/perfumaria/bem-estar | pegar o pico de valor esperado do momento |
| Valoriza | comissão + avaliação + aderência ao nicho | volume × comissão × preço (valor esperado) |
| Retorno | composto, cauda longa (confiança da marca) | imediato, volátil |
| Risco | crescimento mais lento | audiência dispersa, menos fidelização |

Elas não competem — **coexistem como um experimento**. A ideia é alocar parte dos
slots da semana a cada estratégia (ex.: 70% nicho / 30% diversificada),
**etiquetar cada post com a estratégia que o gerou**, e comparar os resultados.
A sobreposição entre os dois rankings (produtos que ambas escolhem) é a aposta de
menor risco; a divergência é o que o experimento mede.

No CSV de exemplo, com piso de 7%, o nicho enche o topo de cosméticos/bem-estar
e a diversificada sobe eletrônicos de alto volume — com 3 de 5 produtos em comum.
Esse é exatamente o tipo de sinal que você quer observar com dados reais.

---

## 4. Roadmap incremental (cada incremento é testável e entrega valor)

Espelha a lógica de sprints da disciplina: incrementos sucessivos, cada um com um
critério de viabilidade verificável.

**Incremento 0 — Curadoria offline (ESTE arcabouço).**
Fonte CSV (export manual da Shopee) → elegibilidade → scoring → ranking.
*Teste de viabilidade:* a lista priorizada bate com, ou supera, a escolha que ela
faria no olho? Rode os dois rankings sobre um dia real e compare com o palpite
dela. Se a curadoria for ao menos tão boa e mais rápida, o conceito se provou.

**Incremento 1 — Fonte ao vivo (API da Shopee).** ✅ adaptador implementado.
`ShopeeAPISource` chama o `productOfferV2` (GraphQL, assinado com SHA256) atrás da
mesma porta — nada mais muda. A API já entrega `sales` (demanda) e `ratingStar`
(avaliação), então o scoring roda sem proxy inventado. Filtre o nicho por
`productCatId` e ordene por comissão (`sortType: 5`).
*Teste:* rodar `-fonte shopee -cat <id> -categoria cosméticos` e comparar o
ranking com o do CSV. Ver `docs/APIS.md` para os campos.

**Incremento 2 — Quadro Kanban operacional (SvelteKit).**
A short-list do dia alimenta a coluna "Candidatos", já com o `offerLink`. Ela
opera o dia a partir do quadro. *Teste:* ela toca a rotina sem abrir planilha?

**Incremento 3 — Atribuição (Shopee nativa + encurtador para o resto).**
Para links Shopee, a atribuição é **nativa**: gere o link com `generateShortLink`
embutindo `subIds` (ex.: `["instagram", "<estrategia>", "<data>", "<itemId>"]`) e
puxe o resultado do `conversionReport` (os subIds voltam em `utmContent`). Não
precisa de encurtador próprio só para saber o que a Shopee converteu. O
encurtador caseiro (Go + Postgres) entra para o que a Shopee não cobre: outras
redes e volume bruto de cliques (o relatório só mostra pedidos, não todo clique).
*Teste:* dá para atribuir cada venda à estratégia e ao post que a gerou?

**Incremento 4 — Comparação de estratégias com dados reais.**
Com resultado por post + etiqueta de estratégia, comparar receita-por-post e
**receita-por-hora** entre nicho e diversificada; ajustar pesos.
*Teste:* qual estratégia paga melhor o esforço, e em que proporção combiná-las?

**Incremento 5 — Painel (Metabase sobre o Postgres).**
Visualização sem reinventar gráfico. Energia fica na coleta e na modelagem.

> Construa o que toca o **esforço** e a **coleta crua** (curadoria, encurtador,
> etiquetagem). Use pronto o que é commodity (Metabase, quadro). O dado de
> esforço — quanto tempo cada post levou — é o mais frágil e o mais valioso:
> proteja-o desde o Incremento 2.

---

## 5. Quadro Kanban da operação

Colunas para o dia a dia dela, com limites de WIP para evitar sobrecarga
(prática direta de Kanban — puxar, não empurrar):

| Coluna | WIP | O que é |
|--------|-----|---------|
| **Candidatos** | — | preenchida pelo Garimpo toda manhã, já ranqueada |
| **Selecionados do dia** | 3–5 | ela puxa os melhores da curadoria |
| **Em produção** | 1–2 | montando post + vídeo + link |
| **Publicado** | — | no ar |
| **Em análise** | — | acompanhando 7–14 dias (cauda longa) |
| **Arquivado** | — | com aprendizado registrado (virou dado) |

Os limites de WIP em "Selecionados" e "Em produção" são o coração do método:
forçam foco e tornam o gargalo visível. Se "Em produção" vive cheio, o gargalo
saiu da seleção e foi para a montagem — e o próximo incremento deve atacar ali.

---

## 6. Métricas (DORA adaptado à operação)

Traduzindo as quatro chaves para o fluxo dela — útil para o vídeo final da
disciplina, inclusive, como estudo de caso de processo fora de software:

| DORA | Equivalente aqui |
|------|------------------|
| Frequência de implantação | posts publicados por dia/semana |
| Lead time para mudanças | tempo de "decidi anunciar" até "publicado" |
| Taxa de falha | posts sem engajamento/clique relevante |
| Tempo de restauração | quão rápido ela troca de produto quando algo não engaja |

Mais duas métricas de negócio, que são o objetivo real:
**receita por post** e **receita por hora de esforço** (a métrica-rainha — só
existe se o tempo for registrado, daí o cuidado com o dado de esforço).

---

## 7. Riscos e premissas

- **API da Shopee é incógnita.** ~~Mitigado pelo padrão de adaptador~~ →
  **resolvido**: a API GraphQL de afiliados entrega produtos (com vendas e
  avaliação), gera short links com subIds e tem relatório de conversão. Detalhes
  e mapeamento em `docs/APIS.md`. O padrão de adaptador segue valendo para trocar
  de fonte sem tocar no motor.
- **Instagram não fecha o loop de clique.** A Graph API dá performance de
  conteúdo (alcance/engajamento), mas não coloca link clicável em feed, não
  adiciona sticker de link em Stories via API e não expõe o clique de afiliado.
  Logo: a **colocação do link permanece manual** e a **atribuição vem da Shopee**
  (subIds → conversionReport). Ver `docs/APIS.md`.
- **Termos de uso.** Coleta via API oficial autorizada é segura; raspar o site
  direto pode violar os termos e arriscar a conta — não compensa.
- **Comissão alta ≠ bom produto (achado dos dados reais).** Ao plugar a API,
  vimos muitos itens com comissão de 60–83% e **zero venda / zero nota** —
  produto-fantasma típico de dropship. Por isso a elegibilidade agora aceita
  pisos opcionais de **vendas** e **nota**, além da comissão. Promover o que já
  tem tração é menos arriscado; o piso de credibilidade protege a curadoria.
- **Volume de dados baixo no começo.** Por isso o foco inicial é
  curadoria + instrumentação, não predição. Modelo preditivo só quando houver
  histórico que o justifique.
