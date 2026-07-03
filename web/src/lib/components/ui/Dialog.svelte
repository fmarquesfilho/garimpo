<script>
	/**
	 * Dialog — modal acessível com Bits UI + Tailwind.
	 * @prop open — bind:open para controle externo
	 * @prop title — título do dialog
	 * @prop description — descrição opcional abaixo do título
	 * @prop children — conteúdo do body
	 */
	import { Dialog } from 'bits-ui';

	let {
		open = $bindable(false),
		title = '',
		description = '',
		children,
		...rest
	} = $props();
</script>

<Dialog.Root bind:open {...rest}>
	<Dialog.Portal>
		<Dialog.Overlay class="fixed inset-0 z-50 bg-black/50 backdrop-blur-sm animate-in fade-in-0" />
		<Dialog.Content class="fixed left-1/2 top-1/2 z-50 w-full max-w-md -translate-x-1/2 -translate-y-1/2 rounded-lg border border-border bg-popover p-6 shadow-lg animate-in fade-in-0 zoom-in-95">
			{#if title}
				<Dialog.Title class="text-lg font-semibold text-foreground">{title}</Dialog.Title>
			{/if}
			{#if description}
				<Dialog.Description class="mt-1 text-sm text-muted-foreground">{description}</Dialog.Description>
			{/if}
			<div class="mt-4">
				{@render children()}
			</div>
			<Dialog.Close class="absolute right-4 top-4 rounded-sm opacity-70 transition-opacity hover:opacity-100 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring">
				<span class="text-lg">✕</span>
				<span class="sr-only">Fechar</span>
			</Dialog.Close>
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>
