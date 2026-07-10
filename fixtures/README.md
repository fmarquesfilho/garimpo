# Fixtures compartilhados

Dados de teste determinísticos que percorrem toda a stack:
**Collector (Go) → API (C#) → Frontend (Svelte)**

## Estrutura

```
fixtures/
├── lojas.json             ← 3 lojas de teste (Glory of Seoul, Le Botanic, COSRX)
├── produtos.json          ← 10 produtos determinísticos
├── buscas.json            ← 5 buscas (keyword, loja, multi-loja, categoria, mista)
├── respostas/
│   ├── collector-fetch.json       ← Golden: Collector.Fetch("serum")
│   ├── collector-fetchshop.json   ← Golden: Collector.FetchShop(920292999)
│   ├── api-buscas.json            ← Golden: GET /api/buscas (com shop_names!)
│   ├── api-candidatos.json        ← Golden: GET /api/candidatos?keyword=serum
│   ├── api-novidades.json         ← Golden: GET /api/novidades
│   └── frontend-ctx.json          ← Golden: engine.ctx após payloadToConfig
└── README.md
```

## Lojas de teste

| Loja | Shop ID | URL |
|------|---------|-----|
| Glory of Seoul | 920292999 | https://s.shopee.com.br/70IKp57jnV |
| Le Botanic | 282170857 | https://s.shopee.com.br/8fQYnxWQqu |
| COSRX Official | 592884015 | https://s.shopee.com.br/1gGoSgfopD |

## Como usar

### Frontend (contract test)

```javascript
import apiBuscas from '../../fixtures/respostas/api-buscas.json';
import frontendCtx from '../../fixtures/respostas/frontend-ctx.json';
import { payloadToConfig } from '$lib/busca-unificada-logic.js';

// Valida que a transformação produz o resultado esperado
const configs = apiBuscas.buscas.map(payloadToConfig);
expect(configs).toEqual(frontendCtx.configs);
```

### Frontend (mock API em E2E)

```javascript
import apiCandidatos from '../../../fixtures/respostas/api-candidatos.json';
await page.route('**/api/candidatos**', route =>
    route.fulfill({ body: JSON.stringify(apiCandidatos) }));
```

### C# (integration test)

```csharp
var expected = File.ReadAllText("../../../fixtures/respostas/api-buscas.json");
// Valida schema (campos existem com tipos corretos), não valores exatos
```

### Go (golden test)

```go
golden := loadFixture("../../fixtures/respostas/collector-fetch.json")
// Valida que a resposta do Collector.Fetch tem o mesmo schema
```

## Drift check

```bash
mise run check:fixtures-contract
```

Valida:
1. Todos os arquivos de fixture existem
2. JSONs são válidos
3. Shop IDs em buscas.json referenciam lojas existentes
4. shop_names em api-buscas.json são consistentes
5. Frontend contract test passa (payloadToConfig vs golden)

## Como atualizar

1. Editar o fixture relevante
2. Atualizar os golden files dependentes (se formato mudou)
3. Rodar `mise run check:fixtures-contract` para validar
4. Atualizar contract tests se necessário

## Contrato principal: shop_names

O campo `shop_names` é o contrato formal entre backend e frontend para
nomes de loja:

```json
{
  "shop_names": {
    "920292999": "Glory of Seoul",
    "282170857": "Le Botanic"
  }
}
```

- **Backend** persiste no campo `Busca.ShopNames` (jsonb)
- **API** retorna como `shop_names` no GET /api/buscas
- **Frontend** usa diretamente em `payloadToConfig` → `shopNomes`
- **Nunca** usar o campo `nome` (legacy/deprecated) para nomes de loja
