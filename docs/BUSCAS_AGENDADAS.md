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

### 10. Produtos novos sem keyword ou loja

> "Quero ver todos os produtos novos que aparecerem em qualquer loja monitorada"

```
keywords: []
lojas: []  (todas as lojas do usuário)
fontes: [novos]
config_novos: { dias_janela: 2 }
cron: "0 */6 * * *"
```

O sistema verifica todas as lojas e retorna produtos que apareceram pela primeira vez nos últimos N dias (configurável).

---

### 11. Produtos novos em loja(s) específica(s), sem keyword

> "Quero ver novidades apenas da loja SKIN1004 Official"

```
keywords: []
lojas: [258316442]
fontes: [novos]
config_novos: { dias_janela: 3 }
cron: "0 8 * * *"
```

---

### 12. Produtos novos com keyword (filtra por nome)

> "Quero saber quando qualquer loja monitorada lançar algo com 'retinol' no nome"

```
keywords: ["retinol"]
lojas: []
fontes: [novos]
config_novos: { dias_janela: 7 }
cron: "0 12 * * *"
```

O sistema detecta produtos novos de todas as lojas e filtra pelo termo "retinol" no nome.

---

### 13. Produtos novos em loja(s) com keyword

> "Quero saber quando a SKIN1004 Official lançar algo com 'sérum' no nome"

```
keywords: ["sérum"]
lojas: [258316442]
fontes: [novos]
config_novos: { dias_janela: 2 }
cron: "0 */4 * * *"
```

---

### 14. Busca por categoria(s) sem keyword ou loja

> "Quero monitorar tudo na categoria 'cosméticos' + 'skincare' de qualquer loja"

```
keywords: []
lojas: []
categorias: ["cosméticos", "skincare"]
fontes: [curadoria]
cron: "0 8 * * *"
```

O sistema busca por categoria na API de afiliados (sem keyword), retornando os top produtos nessas categorias.

---

### 15. Categoria(s) + loja(s), sem keyword

> "Quero ver produtos de perfumaria apenas nas minhas 2 lojas favoritas"

```
keywords: []
lojas: [258316442, 123456789]
categorias: ["perfumaria"]
fontes: [curadoria, novos]
cron: "0 12 * * *"
```

---

### 16. Categoria(s) + keyword, sem loja

> "Quero buscar sérums dentro da categoria skincare no mercado todo"

```
keywords: ["sérum"]
lojas: []
categorias: ["skincare"]
fontes: [curadoria]
cron: "0 8 * * *"
```

---

### 17. Categoria(s) + loja(s) + keyword (combinação máxima)

> "Quero monitorar produtos de 'retinol' na categoria skincare das lojas SKIN1004 e COSRX"

```
keywords: ["retinol"]
lojas: [258316442, 123456789]
categorias: ["skincare"]
fontes: [curadoria, novos, quedas]
cron: "0 8,18 * * *"
```

Combinação completa: busca por keyword, filtra por categoria, restringe a lojas específicas, e monitora novidades/quedas.

---

## Modelo de dados proposto

```
Busca {
  id: string
  nome: string                    // nome amigável (ex: "Skincare coreana")
  keywords: string[]              // 0 ou mais keywords
  shop_ids: int64[]               // 0 ou mais lojas (vazio = todas do usuário)
  categorias: string[]            // 0 ou mais categorias Shopee (filtro adicional)
  fontes: string[]                // ["curadoria", "quedas", "novos", "favoritos"]
  
  // Configuração de "produtos novos"
  config_novos: {
    dias_janela: int              // janela para considerar "novo" (default: 7)
  }
  
  // Filtros opcionais
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

### Critérios combinatórios

Os filtros se compõem por **interseção** (AND):

| Campo | Vazio significa | Preenchido significa |
|-------|-----------------|---------------------|
| `keywords` | Não filtra por termo | Filtra nome/título por esses termos |
| `shop_ids` | Todas as lojas do usuário | Apenas essas lojas |
| `categorias` | Todas as categorias | Apenas essas categorias |
| `fontes` | Nenhum resultado | Quais tipos de resultado incluir |
| `config_novos.dias_janela` | 7 (default) | Janela personalizada (1, 2, 3, 7, 14, 30 dias) |

Exemplos de como os filtros se combinam:
- `keywords=["retinol"] + lojas=[] + categorias=["skincare"]` → Busca "retinol" em todas as lojas, só na categoria skincare
- `keywords=[] + lojas=[258316442] + categorias=[]` → Todos os produtos da loja, qualquer categoria
- `keywords=[] + lojas=[] + categorias=["cosméticos", "perfumaria"]` → Top produtos nessas 2 categorias, de qualquer loja

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
| Busca agendada precisa ter keyword? | Não — pode ser só monitoramento de lojas ou categorias |
| Busca agendada precisa ter loja? | Não — pode ser só keyword ou categoria |
| Pode ter múltiplas keywords? | Sim — cada uma gera resultados separados |
| Pode ter múltiplas lojas? | Sim — agrega novidades de todas |
| Pode ter múltiplas categorias? | Sim — filtra por qualquer uma delas (OR entre categorias) |
| `lojas: []` significa o quê? | Todas as lojas do usuário (quando fontes inclui quedas/novos) |
| `categorias: []` significa o quê? | Sem filtro de categoria (todas) |
| Busca sem cron é válida? | Sim — é um atalho salvo, sem execução automática |
| O que define "produto novo"? | Configurável pelo usuário (dias_janela: 1-30 dias) |
| Categoria é obrigatória? | Não — é um filtro opcional que se combina com keywords e lojas |

## Cenários totais mapeados

| # | Keywords | Lojas | Categorias | Fontes | Descrição |
|---|:---:|:---:|:---:|---|---|
| 1 | ✓ | — | — | curadoria | Busca simples por termo |
| 2 | ✓✓ | — | — | curadoria | Múltiplas keywords |
| 3 | — | ✓ | — | quedas, novos | Monitorar loja(s) |
| 4 | — | ✓✓ | — | quedas, novos | Múltiplas lojas |
| 5 | ✓ | ✓ | — | curadoria, quedas, novos | Keyword filtrada em loja |
| 6 | ✓✓ | ✓✓ | — | curadoria, quedas, novos | Combinação completa |
| 7 | — | — | — | quedas | Todas as lojas, só quedas |
| 8 | ✓ | — | — | curadoria | Atalho manual (sem cron) |
| 9 | — | — | — | favoritos | Monitorar preço de favoritos |
| 10 | — | — | — | novos | Produtos novos em todas as lojas |
| 11 | — | ✓ | — | novos | Novos em loja(s) específica(s) |
| 12 | ✓ | — | — | novos | Novos filtrados por keyword |
| 13 | ✓ | ✓ | — | novos | Novos em loja(s) + keyword |
| 14 | — | — | ✓ | curadoria | Busca por categoria(s) |
| 15 | — | ✓ | ✓ | curadoria, novos | Loja(s) + categoria(s) |
| 16 | ✓ | — | ✓ | curadoria | Keyword + categoria(s) |
| 17 | ✓ | ✓ | ✓ | curadoria, novos, quedas | Combinação máxima |
