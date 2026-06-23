# Jornada do Usuário — Fluxo de Valor e Pontos de Decisão

Mapeia a jornada da operadora (sua esposa) pelo Garimpo, organizada pelo **fluxo
de valor** (do dado bruto da Shopee à comissão ganha) e pelas **decisões** que
determinam a rentabilidade. Complementa o VSM do `MODELO.md`.

## O fluxo de valor, ponta a ponta

```
mercado Shopee ─► COLETA ─► CURADORIA ─► DECISÃO ─► PUBLICAÇÃO ─► ATRIBUIÇÃO ─► APRENDIZADO
   (oferta)      (snapshot)   (teor)     (escolha)   (canal)      (sub_id)      (análise)
                     │            │           │           │            │             │
                  BigQuery     filtros     Garimpar    Telegram    conversion-    Looker/
                  snapshots   + selos     /Publicar    WhatsApp     Report        Python
                     │                                    │
                     └── LOJAS ──► novidades             editor
                        (shopOfferV2)  variações preço   WYSIWYG
                                                         templates
                                                         agendamento
```

## Etapas, valor e decisão

**1. Descoberta (buscar)** — *Valor:* transformar um catálogo gigante num punhado
de candidatos relevantes. *Decisão:* que termo/categoria buscar hoje? *Apoio do
sistema:* busca + filtros de credibilidade. *Atrito a evitar:* paralisia diante de
opções — o teor já ordena.

**2. Avaliação (ler o teor)** — *Valor:* entender *por que* um produto é boa
aposta, não só *que* é. *Decisão:* confio no topo ou investigo? *Apoio:* a barra
de componentes (explicabilidade) e os selos ⚠ suspeito / ✦ exploração. *Decisão de
risco:* descartar um produto-fantasma de comissão alta.

**3. Seleção (Garimpar)** — *Valor:* comprometer-se com um candidato; gera o
**evento de seleção** (dado de comportamento). *Decisão:* nicho seguro ou aposta
diversificada? *Ponto-chave de rentabilidade:* é aqui que a estratégia vira ação.

**4. Comparação (modo Comparar)** — *Valor:* ver onde nicho e diversificada
discordam. *Decisão:* quando fugir do nicho compensa? *Maturidade:* hoje é
julgamento; no futuro, um bandit recomenda a fração ideal.

**5. Publicação (Publicar)** — *Valor:* a oferta chega à audiência com
formatação rica (foto, negrito, botão inline). *Decisão:* qual destino? qual
template? editar a legenda? agendar ou enviar agora? *Apoio:* editor WYSIWYG
com preview, templates com placeholders, múltiplos destinos, agendamento.
Gera o **evento de publicação** com o `sub_id`.

**6. Produção (Quadro)** — *Valor:* visibilidade do trabalho em andamento, sem
gargalo. *Decisão:* o que priorizar? *Apoio:* limites de WIP no Kanban.

**7. Atribuição & aprendizado** *(destrava com conversões)* — *Valor:* fechar o
laço — saber o que realmente converteu, por canal e estratégia. *Decisão:*
realocar esforço para o que paga melhor por hora. *Apoio futuro:* `conversionReport`
→ BigQuery → Looker/Python (ver `CIENCIA_DE_DADOS.md`).

## Os momentos que decidem a rentabilidade

1. **Filtrar o fantasma** (etapa 2) — evitar comissão alta sem tração protege a
   audiência e o tempo.
2. **Nicho vs. diversificada** (etapa 3) — o trade-off central de retorno/risco.
3. **Explorar vs. explotar** (etapa 1–3) — a flag `exploracao` é o investimento em
   aprender; sem ela, a rentabilidade estagna no que já se conhece.
4. **Escolha de canal** (etapa 5) — só fica mensurável com o `sub_id` + conversões.

## Onde o sistema reduz esforço (o norte: rentabilidade por hora)

- Ordenar por teor poupa o garimpo manual diário (a maior dor original).
- Os selos transformam decisão de risco em algo visual e rápido.
- O registro automático de seleção/publicação cria, sem trabalho extra, o
  histórico que sustenta toda a análise futura.
