# Estratégia BDD — Rastreabilidade Ponta a Ponta

## Visão

Cada **fluxo de valor** do Garimpei (user story) é definido em linguagem natural
(Gherkin), vinculado a testes automatizados em 3 camadas, e validado a cada deploy.

```
Story (Gherkin)
  → Testes E2E (Playwright) — valida o fluxo completo no browser
  → Testes de integração (Go) — valida a API isoladamente
  → Testes unitários (Go/Vitest) — valida a lógica de negócio
```

Se qualquer camada falhar, o deploy é bloqueado.

---

## Estrutura de arquivos proposta

```
specs/
  features/
    curadoria.feature          ← Gherkin: stories do fluxo de curadoria
    monitoramento.feature      ← Gherkin: stories de lojas e alertas
    publicacao.feature         ← Gherkin: stories de publicação
web/
  tests/
    features/                  ← Playwright steps (E2E)
      curadoria.spec.js
      monitoramento.spec.js
      publicacao.spec.js
internal/
  coleta/
    service_test.go            ← Testes unitários (lógica pura)
  httpapi/
    regression_test.go         ← Testes de integração (API)
    coleta_test.go
    lojas_test.go
```

---

## Fluxos de valor identificados (com exemplo Gherkin)

### Fluxo 1: Curadoria → Publicação (o que a Mileny faz todo dia)

```gherkin
Feature: Curadoria e publicação de produtos

  Como Mileny (operadora de afiliados)
  Eu quero encontrar produtos com bom teor e publicá-los com um clique
  Para maximizar minha receita por hora de trabalho

  Scenario: Buscar e ver detalhes sem sair da página
    Given estou logada na página de Curadoria
    When busco por "sérum vitamina c"
    Then vejo uma lista de produtos ranqueados por teor
    And cada produto mostra: nome, preço, comissão, nota, teor
    When clico em um produto
    Then vejo um modal com: imagem grande, descrição, botão publicar
    And não saio da página de Curadoria

  Scenario: Publicar direto do modal de detalhes
    Given estou vendo os detalhes de um produto no modal
    When clico em "Publicar"
    Then sou redirecionada para /publicar com os dados preenchidos
    And o destino padrão está selecionado

  Scenario: Publicar produto de oportunidade (queda de preço)
    Given estou na página de Oportunidades
    And existe uma queda de preço > 15%
    When clico em "📤 Publicar" no card da oportunidade
    Then sou redirecionada para /publicar com os dados do produto
```

### Fluxo 2: Monitoramento de lojas

```gherkin
Feature: Monitoramento automático de lojas

  Como Mileny
  Eu quero adicionar uma loja e receber alertas automáticos
  Para não perder oportunidades de preço

  Scenario: Adicionar loja por link curto
    Given estou logada na página de Lojas
    When colo "https://s.shopee.com.br/5L9p1PQbR2" no campo de adicionar
    And clico em "Adicionar"
    Then a loja aparece na lista com seu nome real
    And um job de coleta é criado (a cada 4h)

  Scenario: Coleta periódica grava snapshot
    Given a loja "koksara" está cadastrada com cron "0 */4 * * *"
    When o scheduler dispara a coleta
    Then um snapshot com 100 produtos é gravado no BigQuery
    And o keyword do snapshot é "loja-457864097"

  Scenario: Variação de preço gera alerta
    Given existem 2+ snapshots da loja "koksara"
    And um produto caiu de R$100 para R$80 (queda de 20%)
    And alertas estão ativos com threshold 15%
    When o sistema verifica variações após a coleta
    Then uma mensagem é enviada ao grupo "Alerta | Garimpei 🚨"
    And a mensagem contém: nome do produto, preço anterior, preço atual, variação %
```

### Fluxo 3: Publicação

```gherkin
Feature: Publicação de ofertas

  Como Mileny
  Eu quero publicar uma oferta com foto e legenda personalizada
  Para engajar minha audiência no Telegram

  Scenario: Publicar com template e foto
    Given estou na página de Publicar com produto preenchido
    And selecionei o destino "Ofertas Beleza"
    And selecionei o template "foto"
    When clico em "Enviar"
    Then a publicação é enviada ao canal Telegram
    And aparece no histórico com status "enviada"
    And o título do produto é preservado no histórico

  Scenario: Agendar publicação para o futuro
    Given preenchi os dados do produto
    And defini agendada_em para amanhã 10:00
    When clico em "Agendar"
    Then a publicação fica com status "agendada"
    And o publicar-pendentes a envia quando chegar a hora
    And todos os campos (nome, preço, link, imagem) são preservados
```

---

## Mapeamento: Story → Testes existentes

| Story | Teste E2E (Playwright) | Teste integração (Go) | Teste unitário |
|-------|------------------------|----------------------|----------------|
| Buscar produtos | `smoke.spec.js` (parcial) | `TestCandidatosRankeiaEFiltra` | `engine_test.go` |
| Adicionar loja por link | ❌ **a criar** | `TestAdicionarLojaComURL` | `TestParseShopInput*` |
| Coleta grava keyword | — | `TestColetaLojaGravaKeywordComBuscaID` | `TestExecutarColetaComBuscaID` |
| Alerta de preço | — | `TestAlertasConfigRefletaEnvVars` | — |
| Publicar preserva título | — | `TestPublicarPendentesPreservaTodosOsCampos` | — |
| Ver detalhes do produto | ❌ **a criar** | — | — |

---

## Implementação incremental

### Fase 1: Feature files + E2E para o fluxo crítico (agora)
Criar `.feature` files e implementar os cenários Playwright para o fluxo
"Curadoria → ver detalhes → publicar" (o que a Mileny mais usa).

### Fase 2: Integração no CI
Adicionar step no workflow que roda os cenários e reporta qual feature passou/falhou.

### Fase 3: Coverage por feature
Dashboard que mostra: "Feature X está coberta por N testes E2E + M testes de integração".

---

## Ferramentas

| Camada | Ferramenta | Motivo |
|--------|-----------|--------|
| Feature files | Gherkin (`.feature`) | Linguagem universal, legível por não-dev |
| E2E | Playwright | Já configurado, roda no CI |
| Integração API | Go test | Já existente, rápido |
| Unitários | Go test + Vitest | Lógica pura, sem IO |
| Binding Gherkin→Playwright | `playwright-bdd` ou manual | Avaliando |

---

## Decisão: usar `playwright-bdd` ou manter manual?

**Manual (recomendado agora):** os `.feature` files servem como spec legível,
os testes Playwright implementam os cenários diretamente. Menos tooling, menos
manutenção. O mapeamento é por convenção de nomenclatura:

```
curadoria.feature → Scenario: "Buscar e ver detalhes"
  ↕
web/tests/curadoria.spec.js → test('Buscar e ver detalhes', ...)
```

**playwright-bdd (futuro):** quando tiver 50+ cenários e precisar de step reuse.
Hoje com ~15 cenários, o overhead não compensa.
