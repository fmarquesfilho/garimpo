# Próxima Sessão — Itens Prioritários

## Contexto

A migração arquitetural (ADR-0012) está completa. O C# serve toda a API, os
microserviços Go estão rodando como sidecars, o frontend está no Cloudflare Pages.
O monólito Go foi decomissionado.

Porém, muitas áreas do site ainda precisam dos endpoints portados (T-0026), e há
um **bug grave** na detecção de variações de preço que precisa ser investigado.

---

## 1. Bug grave: detecção de variações de preço não funciona

**Sintoma:** O monólito Go executou milhares de coletas mas nunca detectou nenhuma
variação de preço (quedas/altas). Os alertas nunca dispararam.

**Impacto:** A funcionalidade core de "Quedas" na página publicar sempre retorna
vazio. Alertas de preço (T-0005) nunca funcionaram.

**Hipóteses a investigar:**
1. Os snapshots estão gravando o mesmo produto com IDs diferentes a cada coleta (sem match para comparar)
2. O campo `preco` não é populado corretamente (sempre 0 ou sempre o mesmo valor)
3. A query de novidades/variações tem um bug no JOIN (nunca encontra o mesmo produto_id em dias diferentes)
4. O throttling/rotação faz com que nunca colete a mesma loja duas vezes no período da janela

**Ação proposta:**
- Investigar os dados no BigQuery (`snapshots` table) — existem produto_ids repetidos em datas diferentes?
- Se os dados estão corrompidos/inutilizáveis → **reset do BigQuery** (truncar tabelas de snapshots)
- Reimplementar a coleta no novo analyzer Python com validação de que o mesmo produto aparece em múltiplos snapshots

---

## 2. Reset do BigQuery

**Decisão:** Truncar as tabelas de snapshots no BigQuery e recomeçar coletas do zero com a nova arquitetura.

**Justificativa:**
- Dados coletados pelo monólito Go nunca produziram variações úteis
- O bug pode estar nos dados em si (IDs inconsistentes, preços faltando)
- Começar limpo permite validar que o pipeline novo (scheduler → collector → snapshots) funciona end-to-end

**Script (a executar no início da sessão):**
```sql
-- BigQuery: truncar snapshots para recomeçar
TRUNCATE TABLE `garimpo-500114.garimpei.snapshots`;
-- Manter conversoes (dados reais de vendas da Shopee)
-- Manter eventos (histórico, baixo volume)
```

---

## 3. Port dos endpoints restantes (T-0026)

Endpoints que o frontend usa mas não existem no C#. Portar por prioridade:

### Prioridade Alta (site não funciona sem eles)
- `/api/buscas` — CRUD de buscas salvas
- `/api/lojas` + `/novidades` + `/evolucao` — monitoramento de lojas
- `/api/favoritos` — salvar/remover favoritos
- `/api/publicar` + `/publicacoes` — publicação (core)
- `/api/destinos` — configurar canais
- `/api/onboarding/*` — fluxo de onboarding (multi-tenant)

### Prioridade Média (funcionalidades secundárias)
- `/api/templates` + `/preview` — templates de mensagem
- `/api/alertas` + `/testar` + `/configurar` — T-0005
- `/api/coletas` — histórico de coletas
- `/api/estatisticas` — dashboard
- `/api/conversoes` + `/reais` — analytics de atribuição

### Prioridade Baixa (pode esperar)
- `/api/resolver-link` — resolve link curto Shopee

---

## 4. Testar WhatsApp Meta (T-0024)

Configurar o app Meta Business e testar envio real. Guia em `docs/guias/configurar-whatsapp-meta.md`.

---

## 5. Tarefas pendentes no backlog

| Task | Título | Prioridade |
|------|--------|-----------|
| T-0024 | Testar WhatsApp Meta Cloud API | Alta |
| T-0026 | Portar endpoints restantes | Alta |
| T-0005 | Alertas configuráveis por usuário | Média |
| T-0002 | Persistir conversões no BigQuery | Média |
| T-0007 | Recomendação IA personalizada | Backlog |

---

## Ordem sugerida para a sessão

1. **Investigar bug de variações** (query no BigQuery, verificar dados)
2. **Reset BigQuery** (se dados estiverem inutilizáveis)
3. **Portar endpoints alta prioridade** (buscas, lojas, favoritos, publicar, onboarding)
4. **Configurar coleta no scheduler** (recriar crons para popular snapshots limpos)
5. **Validar pipeline de variações** (coleta → snapshot → analyzer → quedas funciona)
6. **T-0024** (testar WhatsApp se sobrar tempo)
