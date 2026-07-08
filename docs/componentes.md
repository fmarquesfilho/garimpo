# Componentes UI — Garimpei

**Atualizado:** 2026-07-03
**Stack:** Svelte 5 (runes) + Bits UI + Tailwind CSS v4 (shadcn-svelte style)
**Referência:** ADR-0022, spec `shadcn-svelte-migration`

---

## Arquitetura

```
app.css
├── @import "tailwindcss"        → engine de utilidades
├── @import tokens.css           → variáveis CSS (:root + dark)
├── @theme { ... }               → mapeia tokens para Tailwind/shadcn
└── base styles + legacy utils   → classes .casca, .subtitulo, etc.

$lib/components/ui/
├── Primitivos (Tailwind)        → Button, Alert, Badge, Card, Input
├── Compostos (Bits UI + TW)     → Select, Checkbox, ToggleGroup, Collapsible, Tabs, Dialog, DropdownMenu, Tooltip, ThemeToggle
└── Application                  → DashPanel, MetricCard, Loading, EmptyState, ...

$lib/utils.ts                    → cn() (tailwind-merge + clsx)
```

---

## Primitivos

### Button

```svelte
<script>
  import { Button } from '$lib/components/ui';
</script>

<Button variant="primary" size="md" onclick={...}>Enviar</Button>
<Button variant="secondary" size="sm">Cancelar</Button>
<Button variant="danger">Remover</Button>
<Button variant="ghost" size="icon">⋮</Button>
```

| Prop | Tipo | Default | Valores |
|------|------|---------|---------|
| variant | string | `'primary'` | `primary`, `secondary`, `danger`, `ghost`, `link` |
| size | string | `'md'` | `sm`, `md`, `lg`, `icon` |
| type | string | `'button'` | `button`, `submit`, `reset` |
| disabled | boolean | `false` | |
| class | string | `''` | Classes Tailwind extras |

### Alert

```svelte
<Alert variant="success">Destino adicionado!</Alert>
<Alert variant="error">Falha ao salvar.</Alert>
<Alert variant="warning" inline>Atenção: limite atingido.</Alert>
```

| Prop | Tipo | Default | Valores |
|------|------|---------|---------|
| variant | string | `'info'` | `info`, `success`, `warning`, `error` |
| inline | boolean | `false` | Sem padding/borda (inline text) |

### Badge

```svelte
<Badge variant="success">Ativo</Badge>
<Badge variant="error">Erro</Badge>
<Badge variant="outline">15%</Badge>
```

| Prop | Tipo | Default | Valores |
|------|------|---------|---------|
| variant | string | `'default'` | `default`, `secondary`, `success`, `warning`, `error`, `outline` |

### Card

```svelte
<Card class="p-6">
  <h3>Título</h3>
  <p>Conteúdo</p>
</Card>
```

| Prop | Tipo | Default |
|------|------|---------|
| class | string | `''` |

### Input

```svelte
<Input label="Email" placeholder="seu@email.com" bind:value={email} />
<Input error="Campo obrigatório" bind:value={nome} />
```

| Prop | Tipo | Default |
|------|------|---------|
| label | string | `''` |
| error | string | `''` |
| value | string (bindable) | `''` |
| + todos os atributos HTML de `<input>` | | |

---

## Compostos (Bits UI)

### Select

```svelte
<Select
  bind:value={destinoId}
  options={destinos.map(d => ({ value: d.id, label: d.nome }))}
  placeholder="Selecione…"
  size="md"
/>
```

| Prop | Tipo | Default |
|------|------|---------|
| value | string (bindable) | `''` |
| label | string | `''` |
| options | `{ value, label }[]` | `[]` |
| placeholder | string | `''` |
| size | string | `'md'` |
| disabled | boolean | `false` |
| onchange | `(v) => void` \| null | `null` |

> Para valores numéricos, use o par `value={String(n)}` + `onchange={(v) => (n = Number(v))}`
> (o `value` do Bits UI Select é sempre string).

### Tabs

```svelte
<Tabs {tabs} bind:active={aba}>
  {#if aba === 'produtos'}...{/if}
</Tabs>
```

| Prop | Tipo | Default |
|------|------|---------|
| tabs | `{ id, label, badge?, badgeVariant? }[]` | `[]` |
| active | string (bindable) | primeiro tab |

### Dialog

```svelte
<Dialog bind:open={mostrar} title="Confirmar" description="Tem certeza?">
  <Button onclick={confirmar}>Sim</Button>
</Dialog>
```

| Prop | Tipo | Default |
|------|------|---------|
| open | boolean (bindable) | `false` |
| title | string | `''` |
| description | string | `''` |

### Checkbox

Caixa de seleção acessível (Bits UI) com label clicável. Adicionado na sessão 07/07 para
substituir `<input type="checkbox">` nativos.

```svelte
<Checkbox bind:checked={ativo} label="Alertar apenas quedas de preço" />
```

| Prop | Tipo | Default |
|------|------|---------|
| checked | boolean (bindable) | `false` |
| label | string | `''` |
| disabled | boolean | `false` |

### ToggleGroup

Seleção entre opções (Bits UI — roving tabindex + ARIA).
Substitui grupos de `<button>` ad hoc. Variante `segment` (pílula segmentada) ou `chips`.

Suporta dois modos:
- **`type="single"`** (padrão) — seleção única mutuamente exclusiva.
- **`type="multiple"`** — múltiplas opções ativas simultaneamente (value é `string[]`).

Opções podem exibir **badges** informativos (ex: contagem de itens) via `badge` nas options.

```svelte
<!-- Seleção única -->
<ToggleGroup bind:value={modo} options={[{ value: 'a', label: 'A' }, { value: 'b', label: 'B' }]} variant="segment" />
<ToggleGroup value={cron} onchange={selecionar} options={presets} variant="chips" nullable={false} />

<!-- Seleção múltipla com badges (fontes na página Garimpar) -->
<ToggleGroup
  type="multiple"
  bind:value={fontesAtivas}
  options={[
    { value: 'descobrir', label: '🔍 Descobrir', badge: totalDescobrir },
    { value: 'lojas', label: '🏪 Lojas', badge: totalLojas }
  ]}
  variant="chips"
/>
```

| Prop | Tipo | Default |
|------|------|---------|
| type | `'single'` \| `'multiple'` | `'single'` |
| value | string \| string[] (bindable) | `''` / `[]` |
| options | `{ value, label, badge? }[]` | `[]` |
| variant | `'segment'` \| `'chips'` | `'chips'` |
| size | `'sm'` \| `'md'` | `'md'` |
| nullable | boolean (permite desmarcar) | `true` |
| onchange | `(v) => void` \| null | `null` |

### Collapsible

Seção colapsável acessível (Bits UI — `aria-expanded` + animação). Útil para agrupar
controles secundários sem poluir a visão principal.

```svelte
<Collapsible title="⚙️ Configuração">
  <FormAdicionarLoja />
  <GerenciarBuscas />
  <PainelAlertas />
</Collapsible>
```

| Prop | Tipo | Default |
|------|------|---------|
| title | string | `''` |
| open | boolean (bindable) | `false` |
| class | string | `''` |

### DropdownMenu

```svelte
<DropdownMenu items={[
  { label: '✎ Editar', onclick: editar },
  { label: '✕ Remover', onclick: remover, destructive: true }
]}>
  <button>⋮</button>
</DropdownMenu>
```

| Prop | Tipo | Default |
|------|------|---------|
| items | `{ label, onclick, destructive? }[]` | `[]` |
| children | snippet | trigger element |

### Tooltip

```svelte
<Tooltip content="Negrito">
  <button>B</button>
</Tooltip>
```

| Prop | Tipo | Default |
|------|------|---------|
| content | string | `''` |
| side | string | `'top'` |

**Requisito:** `<Tooltip.Provider>` deve existir na árvore ancestral (está no `+layout.svelte`).

---

## Componentes de Aplicação (Refatorados)

Além da base em `ui/`, os componentes de domínio e layout (localizados em `$lib/components/`) foram totalmente migrados para consumir as primitivas shadcn-svelte e os design tokens:

### Componentes Primitivos e de UI Básica
- **TagInput**: Campo de tags acessível — usa `Input` + `Button` (pílulas em `Badge`).
- **PeriodSelector**: Seleção interativa de janelas de tempo.
- **ScoreMeter**: Termômetro que exibe o "teor" de oportunidade do produto.
- **ErrorMessage**: Wrapper de feedback usando `--color-erro`.

### Navegação e Layout
- **NavDrawer**: Menu lateral deslizante (mobile/desktop).
- **FilterBar**: Barra de filtragem de produtos e buscas — filtro de comissão usa `Select`.
- **LandingHero** / **HeroProduto**: Headers principais da interface, com suporte a dark mode.
- **PainelAlertas**: Gestão de alertas de preço — usa `Input` e `Checkbox`.

### Cards e Componentes de Domínio
- **ProductCard**: Exibição central de ofertas.
- **CandidateCard**: Visualização de leads de produto.
- **BuscaCard**: Resumo da busca configurada pelo usuário.
- **FormAdicionarLoja**: Cadastro de loja com `<Input>`/`<Select>` + palavras-chave (`TagInput`) e agendamento (`AgendadorBusca`) integrados no mesmo formulário (sessão 07/07).
- **GerenciarBuscas**: Buscas por palavra-chave — fontes em `Checkbox`, janela em `Select`.
- **AgendadorBusca**: Seletor de agendamento — modo e frequência em `ToggleGroup`, cron avançado em `Input` (preset "A cada 8h", prop `permitirNunca`).
- **ResolverLink**: Ferramenta de processamento de links curtos.

---

## Tema e Tokens

### Estrutura

```
tokens.css        → :root { --ouro: #9e7422; ... }  (valores base)
app.css @theme    → --color-primary: var(--ouro);   (mapeia para Tailwind)
```

### Como usar cores nos componentes

```svelte
<!-- Tailwind utility (preferido) -->
<span class="text-primary">Destaque</span>
<div class="bg-card border-border">...</div>

<!-- CSS variable (quando necessário em <style>) -->
<div style="color: var(--ouro)">...</div>
```

### Paleta semântica (shadcn)

Baseada no tema **Garimpo Quente** (metáfora: garimpo, rústico, terracota e ouro; estética quente e humana).

| Token | Light | Dark | Uso |
|-------|-------|------|-----|
| `--primary` | `--ouro` (#ab7815) | #eabb4d | Botões, links, foco |
| `--destructive` | `--rosa` (#c05c48) | #d47a66 | Ações destrutivas (terracota intenso) |
| `--background` | `--porcelana` (#fdfaf6) | #241d19 | Fundo da página (areia clara / sépia denso) |
| `--foreground` | `--tinta` (#3d2b1f) | #f5eee9 | Texto primário (marrom carvão quente) |
| `--muted` | `--porcelana` (#fdfaf6) | #241d19 | Superfícies sutis |
| `--accent` | `--ouro-fundo` (#fbf5e6) | #3a2e18 | Hover states |
| `--border` | `--linha` (#e8dbce) | #4a3b31 | Bordas (terracota suave) |

### Dark mode

Automático via `data-theme="dark"` no `<html>`. Os tokens são sobrescritos em `:root[data-theme="dark"]` dentro de `tokens.css`. Tailwind lê os valores via `@theme`. 
A paleta dark é baseada em tons quentes (sépia/marrom profundo) em vez de cinza/preto frio, e o ouro ganha um destaque mais brilhante e luminoso.

---

## Como adicionar um novo componente

### Via CLI (quando disponível)

```bash
npx shadcn-svelte@latest add <component>
```

### Manualmente

1. Criar em `src/lib/components/ui/<Nome>.svelte`
2. Usar `cn()` de `$lib/utils` para classes condicionais
3. Exportar em `index.js`
4. Usar tokens semânticos Tailwind (`bg-primary`, `text-muted-foreground`, etc.)
5. Verificar: `npm run check && npm run lint:css && npm run lint:js`

### Padrão de props

```svelte
<script>
  import { cn } from '$lib/utils';

  let {
    variant = 'default',
    class: className = '',
    children,
    ...rest
  } = $props();
</script>

<div class={cn('base-classes', variantClasses[variant], className)} {...rest}>
  {@render children()}
</div>
```

---

## CI Guards

| Guard | Protege | Quando roda |
|-------|---------|-------------|
| `npm run check` (svelte-check) | Props incorretas, tipos | CI + pre-push |
| `npm run lint:css` (stylelint) | hex colors, at-rules válidas | CI + pre-push |
| `npm run lint:js` (ESLint) | Unused vars, a11y | CI + pre-push |
| `npm run lint:dead` (knip) | Dead code | CI + pre-push |
| `npm run test:unit` (vitest) | Lógica de negócio | CI + pre-push |
| `npm run test` (Playwright) | E2E com auth | CI |
