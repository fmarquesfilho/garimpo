<script>
	/**
	 * Dialog — modal acessível usando Bits UI.
	 * @prop open — bind:open para controlar abertura/fechamento
	 * @prop title — título do modal
	 * @prop description — descrição opcional
	 * @prop children — conteúdo do modal
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
		<Dialog.Overlay class="dialog-overlay" />
		<Dialog.Content class="dialog-content">
			{#if title}
				<Dialog.Title class="dialog-title">{title}</Dialog.Title>
			{/if}
			{#if description}
				<Dialog.Description class="dialog-description">{description}</Dialog.Description>
			{/if}
			<div class="dialog-body">
				{@render children()}
			</div>
			<Dialog.Close class="dialog-close" aria-label="Fechar">✕</Dialog.Close>
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>

<style>
	:global(.dialog-overlay) {
		position: fixed;
		inset: 0;
		background: rgba(46, 34, 38, 0.4);
		z-index: 99;
		animation: fadeIn 0.15s ease;
	}
	:global(.dialog-content) {
		position: fixed;
		top: 50%;
		left: 50%;
		transform: translate(-50%, -50%);
		background: var(--nevoa);
		border: 1px solid var(--linha);
		border-radius: var(--raio-lg);
		box-shadow: var(--sombra);
		padding: var(--r6);
		max-width: 520px;
		width: calc(100% - var(--r8));
		max-height: 85vh;
		overflow-y: auto;
		z-index: 100;
		animation: slideUp 0.2s ease;
	}
	:global(.dialog-content:focus-visible) {
		outline: 2px solid var(--ouro);
		outline-offset: 2px;
	}
	:global(.dialog-title) {
		font-family: var(--display);
		font-size: var(--text-xl);
		font-weight: var(--font-semi);
		color: var(--tinta);
		margin-bottom: var(--r2);
	}
	:global(.dialog-description) {
		font-size: var(--text-sm);
		color: var(--tinta-suave);
		margin-bottom: var(--r4);
	}
	.dialog-body {
		margin-top: var(--r4);
	}
	:global(.dialog-close) {
		position: absolute;
		top: var(--r4);
		right: var(--r4);
		width: 32px;
		height: 32px;
		display: flex;
		align-items: center;
		justify-content: center;
		border: none;
		background: transparent;
		color: var(--tinta-suave);
		font-size: var(--text-lg);
		cursor: pointer;
		border-radius: var(--raio-sm);
		transition: background 0.15s ease, color 0.15s ease;
	}
	:global(.dialog-close:hover) {
		background: var(--porcelana);
		color: var(--tinta);
	}
	:global(.dialog-close:focus-visible) {
		outline: 2px solid var(--ouro);
		outline-offset: 2px;
	}

	@keyframes fadeIn {
		from { opacity: 0; }
		to { opacity: 1; }
	}
	@keyframes slideUp {
		from { opacity: 0; transform: translate(-50%, -48%); }
		to { opacity: 1; transform: translate(-50%, -50%); }
	}

	@media (prefers-reduced-motion: reduce) {
		:global(.dialog-overlay),
		:global(.dialog-content) {
			animation-duration: 0ms;
		}
		:global(.dialog-close) { transition-duration: 0ms; }
	}
</style>
