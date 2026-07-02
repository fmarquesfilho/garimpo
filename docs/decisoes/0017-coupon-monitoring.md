# ADR-0017: Monitoramento de cupons cross-marketplace

## Status

Aceito (2026-07-02)

## Contexto

O Garimpei monitora produtos e preços, mas cupons e vouchers são uma categoria
de oportunidade diferente — têm janela temporal curta, desconto geralmente maior
que variação de preço, e aparecem/desaparecem frequentemente nas APIs dos marketplaces.

Afiliados que publicam cupons rapidamente capturam mais conversões (fator tempo).
Precisamos detectar cupons novos e alertar o afiliado em minutos, não horas.

## Decisão

### Arquitetura

Seguimos o mesmo pipeline existente de produtos, aplicado a cupons:

```
Scheduler → Coupon Collector → BigQuery (coupon_snapshots) → Detector → Alert Matcher → Alerter
```

### Princípios de design

1. **Mesmos padrões** que o product pipeline:
   - Go: `CouponSource` interface + Registry + adapters (Strategy Pattern)
   - C#: `ICouponSource` com Keyed Services
   - Proto: `CouponCollectorService` rPC

2. **Um binary, N deploys**: o `coupon-collector` lê a env var `MARKETPLACE`
   e serve apenas esse marketplace. Deploy separado por marketplace (resiliência).

3. **Append-only BigQuery**: snapshots nunca são editados. Detecção é diff
   entre snapshot atual e anterior (query SQL pura no Python analyzer).

4. **Deduplicação em PostgreSQL**: 24h window por (coupon_id, alert_rule_id).
   Re-alerta só se desconto aumentou. Reset ao editar regra.

5. **Coleta sequencial**: Shopee → Amazon → ML por tenant. Evita picos
   concorrentes e simplifica raciocínio sobre rate limits.

### Componentes novos

| Componente | Stack | Porta |
|-----------|-------|-------|
| `coupon-collector` | Go gRPC | :50061-50063 |
| `POST /detect-coupons` | Python (analyzer) | :8060 |
| `POST /internal/coupon-alerts/evaluate` | C# API | :8080 |
| `CouponAlertRules` CRUD | C# API | :8080 |
| `SendCouponAlert` RPC | Go (alerter) | :50053 |

### Modelo de dados (coupon unificado)

```go
type Coupon struct {
    ID, Marketplace, Code, DiscountType string
    DiscountValue, MinSpend float64
    StartTime, EndTime int64
    ApplicableCategories []string
    Status, OwnerUID string
}
```

Normalizado a partir de:
- Shopee: `productOfferV2` com `priceDiscountRate`, `periodEndTime`
- Amazon: `Offers.Listings.SavingBasis` → calcula % desconto
- Mercado Livre: `Promotions API` (futuro)

## Consequências

### Positivas

- Afiliados recebem alertas de cupons em < 5 min da descoberta
- Reutiliza 100% da infra existente (scheduler, alerter, BigQuery, PostgreSQL)
- Adicionar marketplace = criar adapter + registrar no registry (zero mudança em endpoints)
- Histórico de cupons permite análise de padrões (quando cupons aparecem, categorias recorrentes)

### Negativas

- +3 containers no Cloud Run (um por marketplace)
- BigQuery custo adicional (storage particionado, mas append-only cresce)
- Complexidade de deduplicação (24h window + reset + discount comparison)

### Riscos

- APIs de marketplace podem não expor cupons diretamente (inferimos de `discountRate`)
- Volume de cupons pode ser alto em datas promocionais (Black Friday)
- Dedup window de 24h pode ser curta para cupons que duram 1 semana

## Referências

- [Spec: coupon-monitoring requirements](.kiro/specs/coupon-monitoring/requirements.md)
- [Spec: coupon-monitoring design](.kiro/specs/coupon-monitoring/design.md)
- ADR-0016: Multi-marketplace (base para esta decisão)
