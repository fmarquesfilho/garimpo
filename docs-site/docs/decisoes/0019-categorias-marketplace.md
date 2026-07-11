# ADR-0019: Categorias dinâmicas por marketplace

## Status

Aceito (2026-07-02)

## Contexto

A página Descobrir tinha categorias hardcoded no frontend ("cosméticos", "casa",
"moda") que não batiam com os nomes reais retornados pela API da Shopee
("Cuidados com a Pele", "Casa & Decoração", "Beleza"). Isso causava:

1. Filtro por categoria sem keyword retornava zero resultados
2. Filtro client-side descartava resultados válidos (match por `includes` falhava)
3. UX confusa — usuário não sabia quais categorias existiam de fato

Além disso, o sistema está evoluindo para multi-marketplace (Shopee + Amazon +
futuros). Cada marketplace tem sua própria taxonomia de categorias.

## Decisão

1. **Endpoint `/api/categorias`** retorna a lista oficial de categorias por marketplace
2. **Frontend busca categorias da API** e mostra autocomplete dinâmico
3. **Nomes das categorias são os mesmos que a API retorna nos produtos** — garante
   que o filtro client-side funciona sem ambiguidade
4. **Arquitetura extensível**: cada marketplace contribui com suas categorias;
   futuramente, mapeamento cross-marketplace (ex: "Beleza" Shopee ≈ "Beauty" Amazon)

## Formato do endpoint

```json
GET /api/categorias
{
  "marketplaces": [
    {
      "marketplace": "shopee",
      "categorias": [
        { "id": 100636, "nome": "Casa & Decoração", "slug": "casa-decoracao" },
        { "id": 100630, "nome": "Beleza", "slug": "beleza" }
      ]
    }
  ]
}
```

## Taxonomia Shopee Brasil (nível 1)

IDs extraídos empiricamente de `productCatIds` retornados pela API de afiliados.
Mapeamento mantido em `internal/source/categories.go`.

| ID | Nome |
|----|------|
| 100001 | Alimentos |
| 100009 | Celulares |
| 100011 | Roupas Femininas |
| 100012 | Calçados |
| 100017 | Roupas Masculinas |
| 100630 | Beleza |
| 100631 | Saúde & Bem-estar |
| 100632 | Brinquedos & Bebês |
| 100633 | Acessórios & Bolsas |
| 100636 | Casa & Decoração |
| 100637 | Moda |
| 100640 | Perfumaria |
| 100643 | Papelaria & Livros |
| 100644 | Áudio & Eletrônicos |
| 100658 | Manicure & Pedicure |
| 100659 | Cuidados com o Cabelo |
| 100663 | Maquiagem |
| 100664 | Cuidados com a Pele |

## Evolução multi-marketplace

Quando Amazon for integrada:
1. Adicionar `{ marketplace: "amazon", categorias: [...] }` ao endpoint
2. O autocomplete mostra todas unificadas com badge do marketplace
3. Futuramente: tabela de equivalência cross-marketplace para buscas federadas

**Realizado (2026-07-09):** a página Descobrir agrupa as categorias por nome e exibe os
marketplaces a que cada uma pertence — tanto no dropdown do autocomplete quanto no card da
categoria adicionada (`agruparCategoriasPorMarketplace`, componente `Combobox`/
`CategoriaCard`). Ver ADR-0004 e `componentes.md`.

## Consequências

- Frontend não tem mais categorias hardcoded
- Autocomplete mostra nomes reais — sem ambiguidade no filtro
- Adicionar um marketplace = adicionar ao endpoint + collector + mapeamento
- Endpoint é público (sem auth) — categorias não são dados sensíveis
