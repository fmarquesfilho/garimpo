# Diagrama de Entidades — Garimpei

Atualizado em: 2026-06-27

## Diagrama ER (Mermaid)

```mermaid
erDiagram
    BUSCA {
        string id PK "slug (ex: 'perfume' ou 'loja-457864097')"
        string nome "nome amigável (da loja ou perfil)"
        string[] keywords "termos de busca (pode ser vazio para lojas)"
        int[] shop_ids "IDs de lojas Shopee (vazio para buscas por keyword)"
        string categoria "rótulo opcional"
        string estrategia "sempre 'nicho' (diversificada descontinuada)"
        float comissao_min
        int vendas_min
        float nota_min
        int top
        string cron "expressão cron (ex: '0 */4 * * *')"
        bool ativo
        string owner_uid "FK para Firebase Auth"
        json rotation_cursor "map shopID→próxima página"
        json full_scan_at "map shopID→timestamp última varredura"
        timestamp salvo_em
    }

    SNAPSHOT {
        timestamp coletado_em PK
        string keyword "identifica a busca (= busca.id para lojas)"
        string categoria
        string estrategia
        int posicao
        string produto_id
        string nome
        float preco
        float comissao
        int vendas
        float nota
        float score
    }

    PUBLICACAO {
        string id PK
        string produto_id
        string nome
        string categoria
        float preco
        float comissao
        string link
        string imagem
        string estrategia
        string destino_id "FK para DESTINO"
        string template_id "FK para TEMPLATE"
        string agendada_em
        string status "agendada | enviada | erro"
        string detalhe "sub_id ou mensagem de erro"
        timestamp criada_em
        string enviada_em
        string owner_uid
    }

    DESTINO {
        string id PK
        string nome
        string tipo "telegram | whatsapp"
        string config "chat_id ou group_ids"
        bool ativo
    }

    TEMPLATE {
        string id PK
        string nome
        string corpo "HTML com placeholders"
        bool com_foto
        bool ativo
    }

    EVENTO {
        string tipo "selecao | publicacao"
        string produto_id
        string nome
        string categoria
        string estrategia
        string canal
        string sub_id
        float comissao
        float preco
        int vendas
        float score
        timestamp em
    }

    CONVERSAO_REAL {
        string conversion_id PK
        string sub_id "utmContent da Shopee = canal_estrategia_data"
        string produto_id
        string nome_produto
        float comissao_total
        string status "PENDING | COMPLETED | CANCELLED"
        timestamp clique_em
        timestamp compra_em
        timestamp sincronizado_em
    }

    BUSCA ||--o{ SNAPSHOT : "gera via coleta periódica"
    BUSCA }o--|| PUBLICACAO : "produtos publicados"
    PUBLICACAO }o--|| DESTINO : "enviada para"
    PUBLICACAO }o--|| TEMPLATE : "formatada com"
    PUBLICACAO ||--o| CONVERSAO_REAL : "rastreada via sub_id"
    EVENTO }o--|| PUBLICACAO : "registra ação"
```

## Regras de negócio

| Entidade | Regra |
|----------|-------|
| BUSCA com `shop_ids` | É monitoramento de loja. Gera coleta com `productOfferV2(shopId)`. |
| BUSCA com `keywords` | É busca por palavra-chave. Gera coleta com `productOfferV2(keyword)`. |
| BUSCA sem `keywords` nem `shop_ids` | Inválida (rejeitada pela API). |
| SNAPSHOT.keyword | Para lojas = `busca.id` (ex: "loja-457864097"). Para keywords = o termo buscado. |
| PUBLICACAO.detalhe | Quando status=enviada, contém o `sub_id`. Quando status=erro, contém a mensagem. |
| CONVERSAO_REAL.sub_id | Cruza com `PUBLICACAO.detalhe` para fechar o ciclo. |

## Decisões tomadas

- **1:1 entre Busca e Loja** — cada loja monitorada é uma Busca separada.
- **Estratégia sempre "nicho"** — diversificada foi descontinuada da UI e do service.
- **Categorias são rótulos opcionais** — não vêm da API Shopee, são input do usuário.
- **CONVERSAO_REAL** — tabela futura, endpoint `/api/conversoes/sync` já implementado.
