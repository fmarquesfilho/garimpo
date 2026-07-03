# Biblioteca de Componentes UI — Garimpo

## Visão Geral

A biblioteca de componentes `$lib/components/ui/` fornece primitivos e compostos reutilizáveis para o frontend Garimpo. Construída sobre [Bits UI](https://bits-ui.com) para acessibilidade e Svelte 5 runes para reatividade.

**Princípios:**

- Headless + tokens CSS — comportamento via Bits UI, visual via design tokens
- WCAG AA — contraste ≥4.5:1, keyboard nav, ARIA semântico
- Tree-shakeable — apenas componentes importados entram no bundle
- Migração progressiva — coexiste com componentes legados

## Design Tokens

Arquivo: `src/lib/components/ui/tokens.css`

Todas as decisões visuais estão centralizadas em CSS custom properties organizadas por categoria:

| Categoria | Exemplos | Uso |
|-----------|----------|-----|
| Color: Neutrals | `--porcelana`, `--tinta`, `--linha` | Fundos, texto, bordas |
| Color: Ouro | `--ouro`, `--ouro-hover`, `--ouro-fundo` | Ações primárias, destaques |
| Color: Rosa | `--rosa`, `--rosa-hover`, `--rosa-fundo` | Tags, categorias (nicho beleza) |
| Color: Feedback | `--erro-texto`, `--sucesso-fundo`, `--aviso-borda` | Alertas, validação |
| Spacing | `--r1` a `--r12` | Padding, gap, margin |
| Typography | `--display`, `--ui`, `--mono`, `--text-xs` a `--text-2xl` | Fontes e escalas |
| Surfaces | `--raio`, `--raio-sm`, `--raio-full`, `--sombra` | Border-radius, shadows |

## Componentes Primitivos

### Button

```svelte
<script>
  import { Button } from '$lib/components/ui';
</script>

<Button variant="primary" size="md" onclick={handleClick}>
  Salvar
</Button>
```

| Prop | Tipo | Default | Descrição |
|------|------|---------|-----------|
| `variant` | `'primary' \| 'secondary' \| 'danger' \| 'ghost'` | `'primary'` | Estilo visual |
| `size` | `'sm' \| 'md' \| 'lg'` | `'md'` | Tamanho |
| `disabled` | `boolean` | `false` | Desabilita interação |
| `type` | `'button' \| 'submit' \| 'reset'` | `'button'` | Tipo HTML |
| `onclick` | `function` | `null` | Handler de clique |

### Input

```svelte
<script>
  import { Input } from '$lib/components/ui';
  let nome = $state('');
</script>

<Input bind:value={nome} label="Nome" placeholder="Digite..." />
```

| Prop | Tipo | Default | Descrição |
|------|------|---------|-----------|
| `value` | `string` (bindable) | `''` | Valor do campo |
| `label` | `string` | `''` | Rótulo acima do input |
| `variant` | `'default' \| 'mono'` | `'default'` | Família tipográfica |
| `size` | `'sm' \| 'md' \| 'lg'` | `'md'` | Tamanho |
| `type` | `string` | `'text'` | Tipo do input HTML |
| `placeholder` | `string` | `''` | Placeholder |
| `disabled` | `boolean` | `false` | Desabilita |

### Badge

```svelte
<script>
  import { Badge } from '$lib/components/ui';
</script>

<Badge variant="gold">Destaque</Badge>
<Badge variant="pink">Beleza</Badge>
```

| Prop | Tipo | Default | Descrição |
|------|------|---------|-----------|
| `variant` | `'default' \| 'gold' \| 'pink' \| 'green' \| 'red'` | `'default'` | Cor da badge |

### Alert

```svelte
<script>
  import { Alert } from '$lib/components/ui';
</script>

<Alert variant="success">Operação concluída!</Alert>
<Alert variant="error" inline>Campo obrigatório</Alert>
```

| Prop | Tipo | Default | Descrição |
|------|------|---------|-----------|
| `variant` | `'error' \| 'success' \| 'warning'` | `'error'` | Tipo de feedback |
| `inline` | `boolean` | `false` | Estilo compacto (sem fundo) |

### Card

```svelte
<script>
  import { Card } from '$lib/components/ui';
</script>

<Card variant="highlight" padding="lg">
  <h3>Produto premium</h3>
  <p>Conteúdo do card</p>
</Card>
```

| Prop | Tipo | Default | Descrição |
|------|------|---------|-----------|
| `variant` | `'default' \| 'highlight' \| 'success' \| 'error'` | `'default'` | Estilo visual |
| `padding` | `'sm' \| 'md' \| 'lg'` | `'md'` | Espaçamento interno |

## Componentes Compostos (Bits UI)

### Select

```svelte
<script>
  import { Select } from '$lib/components/ui';
  let status = $state('');
  const options = [
    { value: 'ativo', label: 'Ativo' },
    { value: 'inativo', label: 'Inativo' },
  ];
</script>

<Select bind:value={status} label="Status" {options} placeholder="Escolha..." />
```

| Prop | Tipo | Default | Descrição |
|------|------|---------|-----------|
| `value` | `string` (bindable) | `''` | Valor selecionado |
| `label` | `string` | `''` | Rótulo |
| `options` | `{ value, label }[]` | `[]` | Opções do select |
| `placeholder` | `string` | `''` | Texto quando vazio |
| `size` | `'sm' \| 'md' \| 'lg'` | `'md'` | Tamanho |
| `disabled` | `boolean` | `false` | Desabilita |

### Tabs

```svelte
<script>
  import { Tabs } from '$lib/components/ui';
  import { Tabs as BitsTab } from 'bits-ui';
  let active = $state('geral');
  const tabs = [
    { id: 'geral', label: 'Geral' },
    { id: 'config', label: 'Config', badge: '3' },
  ];
</script>

<Tabs {tabs} bind:active>
  <BitsTab.Content value="geral">Conteúdo geral</BitsTab.Content>
  <BitsTab.Content value="config">Configurações</BitsTab.Content>
</Tabs>
```

| Prop | Tipo | Default | Descrição |
|------|------|---------|-----------|
| `tabs` | `{ id, label, badge?, badgeVariant? }[]` | `[]` | Definição das abas |
| `active` | `string` (bindable) | `''` | Aba ativa |

### Dialog

```svelte
<script>
  import { Dialog, Button } from '$lib/components/ui';
  let aberto = $state(false);
</script>

<Button onclick={() => aberto = true}>Abrir</Button>

<Dialog bind:open={aberto} title="Confirmar" description="Tem certeza?">
  <p>Conteúdo do modal</p>
  <Button onclick={() => aberto = false}>Fechar</Button>
</Dialog>
```

| Prop | Tipo | Default | Descrição |
|------|------|---------|-----------|
| `open` | `boolean` (bindable) | `false` | Estado aberto/fechado |
| `title` | `string` | `''` | Título do modal |
| `description` | `string` | `''` | Descrição abaixo do título |

### Tooltip

```svelte
<script>
  import { Tooltip, Button } from '$lib/components/ui';
</script>

<Tooltip content="Copiar link" side="bottom">
  <Button variant="ghost">📋</Button>
</Tooltip>
```

| Prop | Tipo | Default | Descrição |
|------|------|---------|-----------|
| `content` | `string` | `''` | Texto do tooltip |
| `side` | `'top' \| 'bottom' \| 'left' \| 'right'` | `'top'` | Posição |

### DropdownMenu

```svelte
<script>
  import { DropdownMenu, Button } from '$lib/components/ui';
  const items = [
    { label: 'Editar', onclick: () => edit() },
    { label: 'Excluir', onclick: () => remove(), destructive: true },
  ];
</script>

<DropdownMenu {items}>
  <Button variant="ghost">⋮</Button>
</DropdownMenu>
```

| Prop | Tipo | Default | Descrição |
|------|------|---------|-----------|
| `items` | `{ label, onclick, disabled?, destructive? }[]` | `[]` | Itens do menu |

## Props Universais

Todos os componentes aceitam:
- `...rest` — atributos HTML extras são repassados ao elemento raiz (`data-testid`, `id`, `aria-label`, `class`, etc.)
- Valores inválidos de `variant` ou `size` fazem fallback silencioso para o default

## Acessibilidade

- **Keyboard**: Arrow keys em Select/Tabs/DropdownMenu, Escape para fechar Dialog/Tooltip
- **Focus**: `focus-visible` com outline dourado (2px solid `--ouro`, offset 2px)
- **ARIA**: Roles, states e properties gerenciados automaticamente pelo Bits UI
- **Reduced motion**: Animações desativadas quando `prefers-reduced-motion: reduce`
- **Contraste**: Todas as combinações de cor passam WCAG AA (≥4.5:1 texto normal)

## Migração

A migração é progressiva. Componentes legados (ex: `TabBar`) continuam funcionando.
Para migrar, substitua imports gradualmente:

```diff
- import { TabBar } from '$lib/components/ui';
+ import { Tabs } from '$lib/components/ui';
```

O `TabBar` permanece exportado durante a transição.

## Componentes de Negócio Refatorados

Os seguintes componentes de `$lib/components/` foram atualizados para usar os primitivos UI:

| Componente | Primitivos usados |
|---|---|
| `ErrorMessage` | Card, Button |
| `EmptyState` | Card |
| `FormAdicionarLoja` | Card, Button, Input, Alert |
| `TagInput` | Badge |
| `PeriodSelector` | Tokens (+ ARIA radiogroup) |
| `NavDrawer` | Button |
| `PainelAlertas` | Button, Badge, Alert, Input |
| `ListaProdutosLoja` | Alert |
| `BuscaCard` | Badge, Button |
| `ScoreMeter` | Tokens (sem hex) |

### Páginas migradas

| Rota | Primitivos usados |
|---|---|
| `/configurar` | Button, Alert, Input, Card |

### Progresso da migração

| Padrão legado | Antes | Agora | Meta |
|---|---|---|---|
| `<button>` inline | 75 | 62 | 0 (todos via `<Button>`) |
| `<input>` inline | 30 | 28 | 0 (todos via `<Input>`) |
| Badge utility class | 25 | 16 | 0 (todos via `<Badge>`) |
| msg-erro/sucesso class | 6 | 2 | 0 (todos via `<Alert>`) |
| Hex colors hardcoded | 50 | 46 | 0 (todos via tokens) |

Os restantes estão em componentes complexos como `ProductCard` (multi-layout), `FilterBar` (autocomplete), e pages com lógica de form. A migração continua nas próximas sessões.
