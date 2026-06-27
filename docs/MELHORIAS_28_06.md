# Melhorias planejadas — 28/06/2026

Baseado em testes de uso real com Mileny e observações do Fernando. Documentado antes de implementar.

---

## 1. Input de busca: botão de limpar (X)

**Problema:** Para apagar a keyword digitada, o usuário precisa clicar no campo e apagar manualmente. Em mobile isso é especialmente chato.

**Solução:** Adicionar um ícone ✕ à direita do input principal da busca que, ao ser clicado, limpa o campo e foca o input.

**Onde:** `web/src/lib/components/FilterBar.svelte` (campo de busca principal)

**Comportamento:**
- O ✕ só aparece quando há texto digitado
- Clicar no ✕ limpa o campo e foca o input
- Atalho ESC no input também limpa (se não tiver, adicionar)

**Complexidade:** Baixa (~15 min)

---

## 2. Componente unificado de card de produto

**Problema:** Hoje existem representações de produto em múltiplas páginas com diferentes componentes:
- `CandidateCard.svelte` — página de busca (mostra loja, imagem, score completo)
- `CardOportunidade.svelte` — página de oportunidades (sem imagem, sem nome da loja como vem do card de busca)
- `ListaProdutosLoja.svelte` — página de lojas (lista compacta)
- Cards inline na aba "novidades" e "preços"

Isso gera:
- Inconsistência visual (foto aparece num lugar e não em outro)
- Informações úteis (nome da loja) não aparecem em todos os contextos
- Manutenção multiplicada (corrigir bug em 4 lugares)

**Solução:** Criar um componente `ProductCard.svelte` unificado e configurável que substitua todos os acima. O componente aceita props de configuração para adaptar ao contexto:

```svelte
<ProductCard
  produto={item}
  layout="full|compact|feed"     <!-- full=busca, compact=lista, feed=oportunidades -->
  mostrarImagem={true}
  mostrarLoja={true}
  mostrarScore={true}
  mostrarVariacao={variacao_pct}  <!-- para oportunidades -->
  mostrarBadges={true}
  onpublicar={handler}
/>
```

**Layouts:**
- `full` — card grande com imagem (página de busca). É o `CandidateCard` atual.
- `compact` — linha horizontal com thumb pequena (página de lojas, lista de produtos)
- `feed` — card de feed com borda lateral colorida (oportunidades: queda/alta/novo)

**Informações que devem estar disponíveis em todos os layouts:**
- Nome do produto
- Preço
- Comissão (%)
- Nome da loja (quando disponível)
- Imagem (quando disponível)
- Badges (origem, desconto, expiração)

**Migração:**
1. Criar `ProductCard.svelte` com API de props
2. Migrar `CandidateCard` → `ProductCard layout="full"` (manter CandidateCard como alias por compatibilidade de testes)
3. Migrar `CardOportunidade` → `ProductCard layout="feed"`
4. Migrar `ListaProdutosLoja` → usar `ProductCard layout="compact"` no loop
5. Remover componentes antigos após validação

**Complexidade:** Alta (~2-3h). Precisa de cuidado para não regredir UX existente.

---

## 3. Unificação de Busca + Oportunidades

**Problema:** As páginas de busca (`/`) e oportunidades (`/oportunidades`) resolvem o mesmo problema — encontrar produtos para divulgar — mas por caminhos diferentes:
- Busca: por palavra-chave, resultado instantâneo
- Oportunidades: por variação de preço das lojas monitoradas, análise periódica

Na prática, Mileny navega entre as duas com o mesmo objetivo. A separação cria fricção.

**Visão proposta:** Unificar em uma única página de "Descoberta" com abas ou seções:
1. **Busca** — input de keyword + resultados (como hoje)
2. **Oportunidades** — feed de quedas/altas/novos das lojas monitoradas (como hoje)
3. **Favoritos** — produtos salvos para análise posterior (ver item 4)

**Alternativa (mais conservadora):** Manter as páginas separadas mas:
- Usar o mesmo `ProductCard` em ambas (item 2)
- Adicionar link "Ver oportunidades desta loja" na busca
- Adicionar link "Buscar mais como este" nas oportunidades

**Decisão necessária (Fernando):** Unificar em uma página só, ou manter separadas e apenas alinhar visualmente?

**Complexidade:** Média-Alta se unificar (reescrever rota + estado). Baixa se apenas alinhar visual.

---

## 4. Favoritos (salvar produtos para análise posterior)

**Problema:** Mileny publica produtos no grupo de WhatsApp (só ela e Fernando) apenas para "salvar para depois". O antigo quadro Kanban resolvia isso mas era complexo demais. Ela precisa de algo simples: favoritar e listar.

**Solução:** Botão ⭐ no `ProductCard` que salva o produto localmente (localStorage) e opcionalmente sincroniza com o servidor. Uma seção "Favoritos" acessível da página principal (ou como aba na busca unificada).

**Comportamento:**
- Clicar ⭐ salva o produto (id, nome, preço, link, imagem, data de favoritação)
- Lista de favoritos mostra produtos salvos em ordem cronológica reversa
- Ação "Publicar" disponível direto da lista de favoritos
- Ação "Remover" (desfavoritar)
- Persistência: localStorage (imediato) + sync best-effort com servidor (BigQuery tabela `favoritos`)

**Onde aparece o ⭐:**
- Em todo `ProductCard` (busca, oportunidades, lojas)
- Feedback visual: ⭐ dourada se favoritado, ☆ outline se não

**Complexidade:** Média (~1.5h frontend + ~30min backend para sync)

---

## 5. Mover "Publicações" para submenu de Monitoramento

**Problema:** A página de publicações é um relatório (histórico + desempenho), não uma ação primária. Hoje está no menu como item de mesmo peso que "Publicar" (ação). Na hierarquia mental de Mileny, publicações/desempenho são "ver como está indo" — mesma categoria de Estatísticas.

**Solução:** No menu lateral (NavDrawer), mover "Publicações" de:
```
Publicar
  🔗 Link
  📤 Publicações    ← aqui hoje
```
Para:
```
Monitoramento
  📊 Estatísticas
  📤 Publicações    ← mover para cá
```

**Impacto:** Apenas alterar `NavDrawer.svelte`. Rota `/publicacoes` permanece a mesma.

**Complexidade:** Mínima (~5 min)

---

## 6. Bug: aba Desempenho trava no loading

**Problema:** A aba "Desempenho" em Publicações fica presa em "Consultando relatório de conversões da Shopee... Isso pode levar até 15 segundos..." indefinidamente.

**Diagnóstico provável:**
- O timeout de 20s foi adicionado, mas o `$effect` que dispara `carregarReais()` pode estar re-disparando (loop): toda vez que `conversoesReais` muda para `null` (ex: por outro effect ou re-render), o effect re-executa.
- Outra hipótese: a API retorna erro mas o `erroReais` não é exibido corretamente (o erro é um objeto mas o template espera `.message`).

**Investigar:**
1. Verificar se o `$effect` que chama `carregarReais()` pode entrar em loop
2. Verificar se após timeout, `carregandoReais` é de fato setado para `false`
3. Testar localmente com API mockada que demora 25s (deve mostrar erro após 20s)

**Ação:** Debug + fix + teste automatizado

**Complexidade:** Baixa-Média (~30 min debug + fix)

---

## 7. Área de alertas configuráveis

**Problema:** Hoje os alertas de variação de preço estão enterrados na página de Lojas (painel colapsável). Mileny está recebendo alertas de "produtos novos" no Telegram — útil, mas atualmente misturado com alertas de preço no mesmo grupo. Ela não tem controle sobre quais alertas quer receber nem para onde.

**Visão:** Criar uma página `/alertas` (ou seção em `/configurar`) onde o usuário configura:

### Tipos de alerta disponíveis:
1. **Variação de preço** (já implementado) — produto caiu/subiu X%
2. **Produto novo** (já enviado, mas sem config) — produto apareceu pela primeira vez na loja
3. **Oportunidade** (futuro) — produto com score alto apareceu
4. **Conversão** (futuro) — alguém comprou pelo seu link

### Para cada tipo, o usuário configura:
- **Habilitado:** sim/não
- **Destino:** Telegram (grupo A, grupo B), WhatsApp, E-mail (futuro)
- **Filtros:** threshold (ex: só variação > 15%), lojas específicas, categorias
- **Frequência:** tempo real, resumo diário, resumo semanal

### UX proposta:
```
/configurar (ou /alertas)

┌─────────────────────────────────────────────────┐
│ 🔔 Alertas                                      │
├─────────────────────────────────────────────────┤
│ ☑ Variação de preço                             │
│   Destino: Telegram (Grupo Alertas)             │
│   Threshold: > 15%                              │
│   Apenas quedas: ☑                              │
│                                                 │
│ ☑ Produtos novos                                │
│   Destino: Telegram (Grupo Novidades)           │
│   Lojas: todas                                  │
│                                                 │
│ ☐ Conversões                                    │
│   (habilite para receber aviso quando vender)   │
│                                                 │
│ [+ Adicionar alerta]                            │
└─────────────────────────────────────────────────┘
```

### Impacto no backend:
- Tabela `alertas_config` no BigQuery (tipo, destino, filtros, ativo, owner_uid)
- Refatorar `internal/alerts` para ler config do banco em vez de env vars
- Suportar múltiplos destinos por tipo de alerta

### Migração:
- O painel de alertas atual em `/lojas` (`PainelAlertas.svelte`) vira a configuração de "Variação de preço"
- Adicionar UI para "Produtos novos" (que já funciona no backend mas sem config)
- A página `/lojas` perde o painel de alertas (fica mais limpa)

**Complexidade:** Alta (~4-6h total: backend + frontend + migração)

**Prioridade:** Média — funciona hoje de forma hardcoded, mas precisa ser configurável antes de ter mais usuários.

---

## Ordem de implementação sugerida

1. **Item 5** — Mover Publicações no menu (5 min, impacto imediato de organização)
2. **Item 6** — Fix aba Desempenho (30 min, bug que atrapalha uso diário)
3. **Item 1** — Botão X no input (15 min, QoL)
4. **Item 2** — ProductCard unificado (2-3h, base para tudo depois)
5. **Item 4** — Favoritos (1.5h, resolve a dor imediata da Mileny)
6. **Item 3** — Unificação busca+oportunidades (decisão de design necessária antes)
7. **Item 7** — Alertas configuráveis (próxima sessão, precisa de spec)

---

## Perguntas para o Fernando antes de implementar

1. **Item 3:** ✅ Unificar busca+oportunidades em uma página só (com abas). Após unificação, fazer refatoração geral e busca por código morto.
2. **Item 4 (Favoritos):** ✅ Sync com servidor desde o início (BigQuery). Mileny troca de dispositivo e precisa ver os mesmos favoritos.
3. **Item 7 (Alertas):** ✅ Desabilitar alertas de produtos novos por enquanto (mantendo a funcionalidade no código, pronta para ser habilitada via configuração futura).
