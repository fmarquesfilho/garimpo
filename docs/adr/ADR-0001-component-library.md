# ADR-0001: Adoção de Bits UI como biblioteca de componentes headless

## Status

Aceito

## Contexto

O frontend Garimpo (SvelteKit, Svelte 5) possui ~36 componentes com estilos duplicados e padrões de interação inconsistentes. Cada componente reimplementa botões, modais, dropdowns e inputs de maneira ad-hoc, sem garantias de acessibilidade (WCAG) ou consistência de teclado.

O projeto usa CSS puro com custom properties (variáveis em `:root`) que definem um sistema visual coerente (paleta ouro/porcelana, tipografia editorial com Fraunces/Archivo, espaçamento e superfícies). Não usa Tailwind CSS.

Precisamos de uma camada de comportamento (keyboard nav, ARIA, focus management) que aceite nosso sistema visual existente sem impor opiniões de estilo.

## Decisão

Adotar **Bits UI** como camada headless de componentes primitivos para o frontend Garimpo.

## Critérios de Avaliação

| Critério | Peso | Descrição |
|----------|------|-----------|
| Acessibilidade (WCAG) | Alto | ARIA semantics, keyboard nav, focus trap built-in |
| Headless/Unstyled | Alto | Zero opinião visual — aceita tokens CSS existentes |
| Svelte 5 Runes | Alto | Compatível com $props, $state, $derived, $bindable |
| Manutenção ativa | Médio | Releases frequentes, comunidade ativa |
| Bundle size | Médio | Tree-shakeable, impacto mínimo por rota |

## Alternativas Avaliadas

### shadcn-svelte

- **Veredito**: Rejeitado
- **Motivo**: Requer Tailwind CSS como dependência fundamental. O projeto Garimpo usa CSS puro com custom properties e não planeja adotar Tailwind. Adicionaria complexidade de tooling desnecessária e forçaria reescrita do sistema de tokens existente.

### Skeleton UI

- **Veredito**: Rejeitado
- **Motivo**: Biblioteca com estilo opinado (design system próprio). Conflitaria diretamente com a identidade visual Garimpo (paleta ouro/porcelana, tipografia editorial). Customizar para manter nosso visual seria mais trabalho que construir sobre uma base headless.

### Melt UI

- **Veredito**: Rejeitado (como camada direta)
- **Motivo**: API de nível mais baixo (Builder pattern) que exige mais boilerplate para cada componente. Bits UI é construído sobre Melt UI internamente mas oferece uma abstração de compound components mais ergonômica. Para nosso caso (wrapping rápido com tokens CSS), a API de mais alto nível do Bits UI é preferível.

## Convenção de Naming dos Tokens

Os tokens preservam os nomes existentes das CSS custom properties de `app.css` sem alteração:

- **Cores**: `--ouro`, `--porcelana`, `--rosa`, `--tinta`, `--erro-texto`, etc.
- **Espaçamento**: `--r1` a `--r12`
- **Tipografia**: `--display`, `--ui`, `--mono`, `--text-xs` a `--text-2xl`, `--font-semi`, `--font-bold`
- **Superfícies**: `--raio`, `--raio-sm`, `--raio-lg`, `--raio-full`, `--sombra`

Organizados no arquivo `tokens.css` por categorias com comentários delimitadores.

## Estrutura de Arquivos dos Componentes

```
web/src/lib/components/ui/
├── tokens.css              # Design tokens (single source of truth)
├── Button.svelte           # Primitivo
├── Input.svelte            # Primitivo
├── Badge.svelte            # Primitivo
├── Alert.svelte            # Primitivo
├── Card.svelte             # Primitivo
├── Select.svelte           # Composto (Bits UI Select)
├── Tabs.svelte             # Composto (Bits UI Tabs)
├── Dialog.svelte           # Composto (Bits UI Dialog)
├── Tooltip.svelte          # Composto (Bits UI Tooltip)
├── DropdownMenu.svelte     # Composto (Bits UI DropdownMenu)
└── index.js                # Barrel export
```

Cada componente:
- Reside em `$lib/components/ui/{Nome}.svelte`
- É exportado do barrel `$lib/components/ui/index.js`
- Usa single-file component pattern com `<style>` scoped
- Declara props via `$props()` (Svelte 5 runes)
- Referencia apenas tokens CSS para valores visuais

## Consequências

### Positivas

- Acessibilidade WCAG garantida pela camada Bits UI (keyboard, ARIA, focus)
- Identidade visual preservada — tokens CSS existentes aplicados sem mudança
- Bundle leve — Bits UI é tree-shakeable, apenas componentes importados entram no bundle
- Migração progressiva — novos componentes coexistem com legacy via CSS scoped
- DX melhorada — compound component API consistente, props tipados

### Negativas

- Dependência nova (`bits-ui`) no projeto
- Curva de aprendizado para a API de compound components do Bits UI
- Estilos via `:global([data-*])` podem ser menos intuitivos que classes diretas
- Melt UI (base do Bits UI) como dependência transitiva adiciona ao node_modules

### Riscos Mitigados

- **Lock-in**: Bits UI é headless — se descontinuado, o comportamento pode ser reimplementado sem mudar visual
- **Breaking changes**: Fixar versão no package.json, atualizar com cuidado
- **Bundle bloat**: Monitoramento via budget de 15KB gzip por rota
