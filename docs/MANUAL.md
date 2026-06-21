# Manual do Garimpo (para quem usa no dia a dia)

O Garimpo é uma peneira: de centenas de produtos da Shopee, ele separa os que
**valem o esforço de divulgar** e te ajuda a publicá-los. Boa parte deste manual
também está **dentro da tela** — no card de cada produto há um botão **?** ao lado
de "teor" que abre essa explicação.

## Os termos

**Teor** — o "grau de ouro" da pepita. Um número de 0 a 1 que mede o quanto o
produto tende a render *pelo esforço de divulgar*. Quanto maior, melhor a aposta.
Não é a comissão sozinha — é a combinação de três sinais, cada um comparado aos
outros produtos daquele momento:

- **comissão** — quanto da venda volta pra você;
- **valor esperado** — comissão × preço × vendas: o retorno *provável*, não só a
  porcentagem (uma comissão de 80% num produto que ninguém compra vale pouco);
- **avaliação** — a nota, como sinal de confiança de quem comprou.

A barrinha colorida embaixo do teor mostra o peso de cada sinal naquele produto.

**Nicho vs. Diversificada** — duas formas de ranquear:
- *Nicho* dá um bônus para cosméticos, perfumaria e bem-estar (o seu foco). Esses
  produtos aparecem com a marca `×nicho`.
- *Diversificada* ignora categoria e olha só retorno × demanda.
- *Comparar* mostra as duas listas lado a lado, para você ver onde elas discordam.

**⚠ suspeito** — o produto-fantasma: comissão alta, mas sem vendas ou sem nota.
É o padrão de item de baixa qualidade que ninguém compra. O Garimpo **não esconde**
— ele marca, para você decidir com a informação à vista. Os filtros de *vendas
mínimas* e *nota mínima* tiram esses da peneira quando você quer.

**✦ exploração** — quando você liga o "explorar", o Garimpo reserva ~20% das vagas
para produtos **fora do topo**, sorteados. Parece contraintuitivo, mas é o que
permite descobrir o que converte de verdade — sem isso, você só publica o que o
sistema já recomenda e nunca testa o resto. Esses dados valem ouro mais pra frente.

**atribuição (sub_id)** — quando você publica, aparece um código como
`telegram_nicho_20260321`. É a etiqueta que, lá na frente, vai dizer **qual canal**
trouxe a venda (Telegram vs. Instagram). Por enquanto fica registrado; quando as
conversões da Shopee estiverem ligadas, ele fecha o ciclo.

## O fluxo

1. **Buscar** — digite o que quer divulgar (perfume, sérum, batom). Ajuste os
   filtros (comissão, vendas e nota mínimas) até a peneira ficar limpa.
2. **Ler o teor** — o topo da lista é a melhor aposta do momento. Olhe os selos.
3. **Garimpar** — manda o produto para o **Quadro** (seu Kanban de produção).
4. **Publicar** — dispara a oferta no canal. A mensagem é montada sem mostrar a
   comissão (que é sua margem). Por ora o envio é simulado (mock).
5. **Quadro** — acompanhe cada produto de "Selecionados" até "Publicado".

## Dúvidas frequentes

- *Por que esse produto de 80% de comissão está lá embaixo?* Porque tem zero
  vendas — o teor pesa demanda, não só comissão. Provavelmente está com ⚠ suspeito.
- *Por que a lista veio vazia?* Com a fonte Shopee, busca vazia volta vazia —
  comece sempre por um termo.
- *Devo ligar o "explorar" sempre?* Não precisa. Ligue de vez em quando: é um
  investimento em aprender o que funciona, ao custo de algumas publicações menos
  "certeiras".
