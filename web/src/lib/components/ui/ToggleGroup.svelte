<script>
	/**
	 * ToggleGroup — seleção única entre opções mutuamente exclusivas (Bits UI + Tailwind).
	 * Substitui grupos de <button> ad hoc por um controle acessível (roving tabindex, ARIA).
	 * @prop value — bind:value (string da opção ativa)
	 * @prop options — array de { value, label }
	 * @prop variant — 'segment' (pílula segmentada) | 'chips' (botões soltos)
	 * @prop size — 'sm' | 'md'
	 * @prop nullable — se false, não permite desmarcar (clicar o ativo mantém)
	 * @prop onchange — callback opcional (v) => void; se ausente, usa bind:value
	 */
	import { ToggleGroup } from 'bits-ui';
	import { cn } from '$lib/utils';

	let {
		value = $bindable(''),
		options = [],
		variant = 'chips',
		size = 'md',
		nullable = true,
		onchange = null,
		class: className = '',
		...rest
	} = $props();

	const SIZES = { sm: 'h-7 px-2.5 text-xs', md: 'h-9 px-3 text-sm' };

	function handle(v) {
		const nv = v ?? '';
		if (!nullable && !nv) return;
		// Atualiza o valor (preserva bind:value) antes de notificar via callback.
		value = nv;
		onchange?.(nv);
	}
</script>

<ToggleGroup.Root
	type="single"
	{value}
	onValueChange={handle}
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
