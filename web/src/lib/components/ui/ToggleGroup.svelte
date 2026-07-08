<script>
	/**
	 * ToggleGroup — seleção única ou múltipla (Bits UI + Tailwind).
	 * Substitui grupos de <button> ad hoc por um controle acessível (roving tabindex, ARIA).
	 *
	 * @prop value — bind:value (string para single, string[] para multiple)
	 * @prop options — array de { value, label, badge?, badgeColor? }
	 * @prop type — 'single' | 'multiple' (default: 'single')
	 * @prop variant — 'segment' (pílula segmentada) | 'chips' (botões soltos)
	 * @prop size — 'sm' | 'md'
	 * @prop nullable — se false, não permite desmarcar (clicar o ativo mantém) [single only]
	 * @prop onchange — callback opcional; recebe value atualizado
	 */
	import { ToggleGroup } from 'bits-ui';
	import { cn } from '$lib/utils';

	let {
		value = $bindable(),
		options = [],
		type = 'single',
		variant = 'chips',
		size = 'md',
		nullable = true,
		onchange = null,
		class: className = '',
		...rest
	} = $props();

	const SIZES = { sm: 'h-7 px-2.5 text-xs', md: 'h-9 px-3 text-sm' };

	const BADGE_COLORS = {
		default: 'bg-ouro',
		sucesso: 'bg-sucesso',
		rosa: 'bg-rosa',
		ouro: 'bg-ouro'
	};

	function handleSingle(v) {
		const nv = v ?? '';
		if (!nullable && !nv) return;
		value = nv;
		onchange?.(nv);
	}

	function handleMultiple(v) {
		value = v ?? [];
		onchange?.(value);
	}
</script>

{#if type === 'multiple'}
	<ToggleGroup.Root
		type="multiple"
		value={Array.isArray(value) ? value : []}
		onValueChange={handleMultiple}
		class={cn('flex flex-wrap gap-1.5', className)}
		{...rest}
	>
		{#each options as opt (opt.value)}
			<ToggleGroup.Item
				value={opt.value}
				class={cn(
					'fonte-btn inline-flex cursor-pointer items-center justify-center gap-1 rounded-full border border-border bg-porcelana font-semibold text-tinta-suave transition-[border-color,background] duration-150 hover:border-ouro hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring data-[state=on]:border-ouro-claro data-[state=on]:bg-ouro-fundo data-[state=on]:font-bold data-[state=on]:text-ouro-escuro',
					SIZES[size] ?? SIZES.md,
					size === 'md' ? 'py-[7px] px-3.5 text-[0.82rem]' : ''
				)}
				title={opt.title ?? ''}
			>
				{opt.label}
				{#if opt.badge && opt.badge > 0}
					<span
						class={cn(
							'fonte-badge flex h-4 w-4 items-center justify-center rounded-full text-[0.65rem] font-bold text-white',
							BADGE_COLORS[opt.badgeColor ?? 'default']
						)}>{opt.badge}</span
					>
				{/if}
			</ToggleGroup.Item>
		{/each}
	</ToggleGroup.Root>
{:else}
	<ToggleGroup.Root
		type="single"
		{value}
		onValueChange={handleSingle}
		class={cn(
			variant === 'segment' ? 'inline-flex gap-0.5 rounded-full bg-muted p-0.5' : 'flex flex-wrap gap-2',
			className
		)}
		{...rest}
	>
		{#each options as opt (opt.value)}
			<ToggleGroup.Item
				value={opt.value}
				class={cn(
					'inline-flex cursor-pointer items-center justify-center font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring',
					SIZES[size] ?? SIZES.md,
					variant === 'segment'
						? 'rounded-full text-muted-foreground data-[state=on]:bg-background data-[state=on]:text-foreground data-[state=on]:shadow-sm'
						: 'rounded-full border border-border bg-muted text-foreground hover:border-primary hover:text-primary data-[state=on]:border-primary data-[state=on]:bg-accent data-[state=on]:font-bold data-[state=on]:text-accent-foreground'
				)}
			>
				{opt.label}
			</ToggleGroup.Item>
		{/each}
	</ToggleGroup.Root>
{/if}
