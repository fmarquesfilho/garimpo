<script>
	/**
	 * Card hero do produto — mostra imagem + campos editáveis (nome, categoria, preço, link).
	 * Usado na página de publicação.
	 */
	let { produto = $bindable({}) } = $props();
</script>

<div
	class="mb-5 flex flex-col items-center gap-4 rounded-md border border-border bg-card p-4 text-center sm:flex-row sm:items-start sm:text-left"
>
	{#if produto.imagem}
		<img
			src={produto.imagem}
			alt={produto.nome}
			class="h-[100px] w-[100px] shrink-0 rounded-[10px] object-cover sm:h-[100px] sm:w-[100px] max-sm:h-[180px] max-sm:w-full max-sm:rounded-md"
		/>
	{:else}
		<div
			class="flex h-[100px] w-[100px] shrink-0 items-center justify-center rounded-[10px] bg-porcelana text-4xl max-sm:h-[180px] max-sm:w-full max-sm:rounded-md"
		>
			📦
		</div>
	{/if}
	<div class="flex min-w-0 flex-1 flex-col gap-2">
		<input
			class="w-full rounded-lg border border-border bg-[var(--branco)] px-3 py-2 text-base font-bold focus:outline-2 focus:outline-ouro focus:outline-offset-1"
			bind:value={produto.nome}
			placeholder="Nome do produto"
		/>
		<div class="flex flex-wrap gap-3">
			<input
				class="w-[120px] rounded-lg border border-border bg-porcelana px-2.5 py-1.5 text-sm"
				bind:value={produto.categoria}
				placeholder="Categoria"
			/>
			<input
				class="w-[90px] rounded-lg border border-border bg-porcelana px-2.5 py-1.5 text-sm font-semibold"
				type="number"
				step="0.01"
				bind:value={produto.preco}
				placeholder="Preço"
			/>
		</div>
		{#if produto.link}
			<a
				class="truncate text-xs text-tinta-suave no-underline hover:text-ouro"
				href={produto.link}
				target="_blank"
				rel="noopener">{produto.link.substring(0, 60)}…</a
			>
		{/if}
	</div>
</div>
