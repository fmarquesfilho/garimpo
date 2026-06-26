# Spec: Modelagem de Entidades — Busca, Loja e Coleta

## Problema

A struct `Busca` acumula dois conceitos distintos:

1. **Busca por keyword** — perfil de pesquisa com termos (ex: "perfume", "shiseido"), filtros (comissão mín, vendas mín), e cron para coleta periódica.
2. **Monitoramento de loja** — acompanha uma loja específica (shopId), coleta todos os produtos periodicamente, detecta variações de preço.

Isso causa:
- Campo `keyword` preenchido com "loja-457..." no BigQuery (sem sentido semântico)
- Tag "1 loja" redundante no card (toda loja é 1:1 hoje)
- Scheduler cria jobs com nomes confusos ("coleta-loja-457864097-all")
- Frontend precisa de `if (b.shop_ids?.length > 0)` em vários lugares para diferenciar

## Estado atual

```go
type Busca struct {
    ID          string    // "perfume" ou "loja-457864097"
    Nome        string    // nome amigável (da loja ou perfil)
    Keywords    []string  // ["perfume", "kenzo"] ou []
    ShopIDs     []int64   // [] ou [457864097]
    Categoria   string
    Estrategia  string
    Cron        string
    Ativo       bool
    RotationCursor map[int64]int  // só faz sentido para lojas
    FullScanAt     map[int64]string // só faz sentido para lojas
}
```

## Opções

### Opção A: Separar em duas entidades (Busca + Loja)
```
Busca { ID, Keywords, Categoria, Filtros, Cron }
Loja  { ID, ShopID, Nome, Cron, RotationCursor, FullScanAt }
```
- **Prós:** cada entidade é clara, sem campos órfãos, frontend não precisa de `if`
- **Contras:** 2 tabelas BigQuery, 2 endpoints de CRUD, 2 jobs no scheduler, migração de dados

### Opção B: Interface com implementações (Strategy pattern)
```go
type Monitoravel interface {
    ID() string
    Cron() string
    Coletar(ctx) error
}
type BuscaKeyword struct { ... }
type MonitorLoja struct { ... }
```
- **Prós:** polimorfismo, extensível
- **Contras:** over-engineering para 2 tipos, Go não é Java

### Opção C: Manter unificado mas renomear e limpar (recomendado)
Renomear `Busca` para `Monitor` ou `Perfil`. Aceitar que é uma entidade polimórfica (keyword ou loja). Limpar:
- Remover tag "1 loja" (é sempre 1:1)
- Usar `Nome` como label principal em vez de `ID`
- Se `ShopIDs` preenchido, é monitoramento de loja; senão, busca por keyword
- No BigQuery, keyword do snapshot usa o `Nome` da loja em vez de "loja-457..."

- **Prós:** menos mudanças, sem migração de tabela, já funciona
- **Contras:** struct continua com campos que só servem para um dos tipos

## Recomendação

**Opção C** para agora — o app está estabilizando, a Mileny testando. Renomear é cosmético e não muda funcionalidade. Quando houver multi-tenant (3+ tipos de monitoramento), migrar para Opção A.

## Decisões a tomar
- [ ] Renomear "Busca" para "Monitor" ou "Perfil" no código?
- [ ] Monitorar múltiplas lojas numa mesma entidade faz sentido? (Recomendação: não, manter 1:1)
- [ ] Campo `keyword` no snapshot deveria ser o nome da loja ou manter o ID técnico?

## Impacto no frontend
- Remover tag "1 loja" do card de lojas monitoradas
- Usar `b.nome || b.keywords?.[0] || b.id` como label universal
- Seção "Buscas salvas" mostra apenas buscas por keyword (lojas ficam em /lojas)
