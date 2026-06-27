# Buscas Agendadas — Modelo Conceitual

Documento de entendimento comum. Registra todos os cenários de uso antes de implementar.

---

## Conceito central

Uma **busca agendada** é uma configuração salva que o sistema executa periodicamente. Ela combina:

1. **Fontes** — de onde vêm os dados (curadoria por keyword, lojas monitoradas, ou ambos)
2. **Filtros** — o que refinar (keywords, lojas específicas, categoria, comissão mínima, etc.)
3. **Frequência** — quando executar (cron)
4. **Ação** — o que fazer com os resultados (coletar snapshot, gerar alertas, ou ambos)

A busca agendada NÃO é obrigatoriamente atrelada a uma loja ou a uma keyword. Ela é uma combinação flexível de critérios.

---

## Cenários de uso

### 1. Busca por keyword única

> "Quero monitorar o mercado de sérum vitamina C a cada 4 horas"

```
keywords: ["sérum vitamina c"]
lojas: []
fontes: [curadoria]
cron: "0 */4 * * *"
```

O sistema busca na API de afiliados e salva snapshot dos top produtos para essa keyword.

---

### 2. Busca com múltiplas keywords

> "Quero monitorar 3 marcas coreanas juntas: Skin1004, COSRX e Innisfree"

```
keywords: ["skin1004", "cosrx", "innisfree"]
lojas: []
fontes: [curadoria]
cron: "0 8 * * *"
```

O sistema busca cada keyword separadamente e salva snapshots para cada uma. Na UI, a busca aparece como um grupo com 3 pills clicáveis.

---

### 3. Monitoramento de loja(s) sem keyword

> "Quero monitorar tudo que aparece na loja SKIN1004 Official — produtos novos e quedas de preço"

```
keywords: []
lojas: [258316442]
fontes: [quedas, novos]
cron: "0 */4 * * *"
```

O sistema coleta o catálogo da loja via shopOfferV2, salva snapshot, e detecta novidades/variações.

---

### 4. Monitoramento de múltiplas lojas

> "Quero monitorar 3 lojas coreanas de skincare juntas"

```
keywords: []
lojas: [258316442, 123456789, 987654321]
fontes: [quedas, novos]
cron: "0 8,18 * * *"
```

O sistema coleta as 3 lojas e agrega novidades/variações de todas.

---

### 5. Keyword + loja (busca filtrada)

> "Quero monitorar apenas os sérums da loja SKIN1004 Official"

```
keywords: ["sérum"]
lojas: [258316442]
fontes: [curadoria, quedas, novos]
cron: "0 12 * * *"
```

O sistema busca dentro da loja específica usando a keyword como filtro, e também monitora novidades daquela loja (filtrando pelo termo nos resultados).

---

### 6. Múltiplas keywords + múltiplas lojas

> "Quero monitorar perfumaria e skincare nas minhas 3 lojas favoritas"

```
keywords: ["perfume", "skincare"]
lojas: [258316442, 123456789, 987654321]
fontes: [curadoria, quedas, novos]
cron: "0 8 * * *"
```

Combinação completa: busca cada keyword em cada loja, e monitora novidades de todas as lojas filtrando pelo termo.

---

### 7. Sem keyword, sem loja — todas as lojas monitoradas

> "Quero receber alertas de quedas de preço de TODAS as minhas lojas"

```
keywords: []
lojas: []  (interpreta como "todas as lojas do usuário")
fontes: [quedas]
cron: "0 */6 * * *"
```

O sistema usa todas as lojas monitoradas do usuário e verifica variações.

---

### 8. Busca sem agendamento (salva mas manual)

> "Quero salvar essa combinação de filtros para usar rápido, mas não precisa rodar automático"

```
keywords: ["maquiagem coreana"]
lojas: []
fontes: [curadoria]
cron: ""  (vazio = só manual)
```

Aparece como atalho na página de busca, mas não dispara coleta automática.

---

### 9. Favoritos + alerta (futuro)

> "Me avise quando um produto que eu favoritei cair de preço"

```
keywords: []
lojas: []
fontes: [favoritos]
tipo_alerta: "variacao_favorito"
cron: "0 */4 * * *"
```

O sistema monitora os preços dos produtos favoritados e alerta se cairem.

---

## Modelo de dados proposto

```
Busca {
  id: string
  nome: string                    // nome amigável (ex: "Skincare coreana")
  keywords: string[]              // 0 ou mais keywords
  shop_ids: int64[]               // 0 ou mais lojas (vazio = todas)
  fontes: string[]                // ["curadoria", "quedas", "novos", "favoritos"]
  
  // Filtros opcionais
  categoria: string               // filtro de categoria (opcional)
  comissao_min: float             // comissão mínima (opcional)
  vendas_min: int                 // vendas mínimas (opcional)
  nota_min: float                 // nota mínima (opcional)
  
  // Execução
  cron: string                    // "" = sem agendamento (só manual)
  ativo: bool                     // false = desativada (tombstone)
  
  // Metadados
  owner_uid: string
  salvo_em: timestamp
}
```

---

## Relação com a UI atual

| Onde aparece | O que mostra | Ações |
|---|---|---|
| Página Descobrir (/) | Pills de keywords como atalhos + ⏱ se agendada | Clicar aplica a busca |
| Página Lojas (/lojas) | Lista completa com cron, keywords, lojas, fontes | Criar, editar, remover, testar |

---

## Relação com o backend atual

Hoje a struct `store.Busca` já tem `Keywords`, `ShopIDs`, `Cron`, `Estrategia`. O que falta:

1. **Campo `fontes`** — para saber quais tipos de resultado a busca monitora (hoje é inferido: se tem shop_ids → loja, se tem keywords → curadoria)
2. **Execução unificada** — o scheduler hoje chama apenas `coleta.Service.Executar()` que busca via API de afiliados. Precisa ser estendido para:
   - Se `fontes` inclui "curadoria" → busca por keyword na API
   - Se `fontes` inclui "quedas"/"novos" → monitora lojas e gera novidades
3. **Filtros no agendamento** — hoje os filtros (comissão mín, vendas mín) são aplicados na hora da busca mas não são salvos na Busca de forma que o scheduler use

---

## O que NÃO muda

- O mecanismo de coleta (shopOfferV2, productOfferV2) permanece o mesmo
- O formato de snapshot no BigQuery não muda
- A detecção de novidades/variações continua comparando snapshots
- Os alertas de preço (Telegram) continuam usando o mesmo Alerter
- O scheduler (Cloud Scheduler) continua disparando via HTTP

---

## Próximos passos (implementação)

1. Adicionar campo `fontes` à struct `Busca` (backend + schema evolution)
2. Atualizar `GerenciarBuscas.svelte` para permitir selecionar fontes
3. Atualizar o scheduler para executar com base nas fontes configuradas
4. Na página Descobrir, ao clicar num atalho, ativar as fontes correspondentes além da keyword

---

## Perguntas resolvidas

| Pergunta | Resposta |
|---|---|
| Busca agendada precisa ter keyword? | Não — pode ser só monitoramento de lojas |
| Busca agendada precisa ter loja? | Não — pode ser só keyword |
| Pode ter múltiplas keywords? | Sim — cada uma gera resultados separados |
| Pode ter múltiplas lojas? | Sim — agrega novidades de todas |
| `lojas: []` significa o quê? | Todas as lojas do usuário (quando fontes inclui quedas/novos) |
| Busca sem cron é válida? | Sim — é um atalho salvo, sem execução automática |
