<script>
	/**
	 * Omnibox — renderizador puro com chips inline.
	 * Loja chips (ouro) e categoria chips (rosa) renderizam DENTRO do input container.
	 * Zero $state local. Lê engine.omnibox + engine.lojaCards + engine.categoriaCards.
	 * Emite eventos brutos para a engine.
	 */
	import Badge from '$lib/components/ui/Badge.svelte';

	let { engine } = $props();

	let om = $derived(engine.omnibox);
	let lojas = $derived(engine.lojaCards);
	let categorias = $derived(engine.categoriaCards);
	let inputEl;
</script>

<div
	class="relative"
	onfocusout={(e) => {
		if (!e.currentTarget.contains(/** @type {Node} */ (e.relatedTarget))) engine.send({ type: 'OMNIBOX_BLUR' });
	}}
>
	<!-- Input container: chips + texto (flex-wrap) -->
	<div
		class="flex flex-wrap items-center gap-1.5 rounded-sm border border-input
			   bg-background px-2 py-1.5 focus-within:border-ring
			   focus-within:ring-2 focus-within:ring-ring/20"
		role="group"
		aria-label="Filtros ativos"
	>
		<span class="pointer-events-none shrink-0 opacity-50">🔍</span>

		{#each lojas as l (l.id)}
			<Badge variant="loja" aria-label="Loja: {l.nome} — ativa">
				🏪 {l.nome}
				<button
					type="button"
					class="ml-0.5 cursor-pointer border-none bg-transparent px-0.5 text-inherit opacity-70 hover:opacity-100"
					aria-label="Remover loja {l.nome}"
					onclick={(e) => {
						e.stopPropagation();
						engine.send({ type: 'REMOVER_LOJA', shopId: l.id });
					}}>✕</button
				>
			</Badge>
		{/each}

		{#each categorias as c (c.nome)}
			<Badge variant="categoria" aria-label="Categoria: {c.nome} — ativa">
				🏷️ {c.nome}
				<button
					type="button"
					class="ml-0.5 cursor-pointer border-none bg-transparent px-0.5 text-inherit opacity-70 hover:opacity-100"
					aria-label="Remover categoria {c.nome}"
					onclick={(e) => {
						e.stopPropagation();
						engine.send({ type: 'REMOVER_CATEGORIA', nome: c.nome });
					}}>✕</button
				>
			</Badge>
		{/each}

		<input
			bind:this={inputEl}
			type="text"
			value={om.inputValue}
			placeholder={lojas.length || categorias.length ? 'Adicionar...' : om.placeholder}
			autocomplete="off"
			spellcheck="false"
			role="combobox"
			aria-expanded={om.aberto && om.opcoes.length > 0}
			aria-controls="omnibox-listbox"
			aria-autocomplete="list"
			aria-activedescendant={om.highlightIdx >= 0 ? `omnibox-opt-${om.highlightIdx}` : undefined}
			class="min-w-[120px] flex-1 border-none bg-transparent text-base text-foreground
				   placeholder:text-muted-foreground outline-none"
			oninput={(e) => engine.send({ type: 'OMNIBOX_INPUT', value: e.currentTarget.value })}
			onfocus={() => engine.send({ type: 'OMNIBOX_INPUT', value: om.inputValue })}
			onkeydown={(e) => {
				if (['Enter', 'ArrowDown', 'ArrowUp', 'Escape'].includes(e.key)) {
					e.preventDefault();
					engine.send({ type: 'OMNIBOX_KEYDOWN', key: e.key });
				}
			}}
		/>
	</div>

	<!-- Aria-live: opcoes count -->
	<span class="sr-only" aria-live="polite">
		{om.aberto && om.opcoes.length > 0 ? `${om.opcoes.length} ${om.opcoes.length === 1 ? 'opção' : 'opções'}` : ''}
	</span>

	<!-- Dropdown listbox -->
	{#if om.aberto && om.opcoes.length > 0}
		<ul
			id="omnibox-listbox"
			role="listbox"
			aria-label="Opções de busca"
			class="absolute left-0 right-0 top-[calc(100%+4px)] z-50 max-h-80
				   overflow-y-auto rounded-md border border-border bg-popover p-1 shadow-md"
		>
			{#each om.opcoes as opcao, i (opcao.tipo + ':' + i)}
				<li
					id={`omnibox-opt-${i}`}
					role="option"
					aria-selected={om.highlightIdx === i}
					aria-label={opcao.labelAcessivel ?? opcao.label}
				>
					<button
						type="button"
						tabindex="-1"
						class="flex w-full items-center gap-2 rounded-sm px-3 py-2 text-left text-sm
							   transition-colors {om.highlightIdx === i ? 'bg-accent text-accent-foreground' : 'hover:bg-accent'}"
						onmouseenter={() => engine.send({ type: 'OMNIBOX_KEYDOWN', key: 'highlight', idx: i })}
						onclick={() => engine.send({ type: 'OMNIBOX_SELECIONAR', indice: i })}
					>
						<span aria-hidden="true">{opcao.icone}</span>
						<span class="flex-1 truncate">{opcao.label}</span>
						{#if opcao.tipo === 'produtos' || opcao.tipo === 'lojas' || opcao.tipo === 'resolver_link'}
							<kbd class="text-xs text-muted-foreground">↵</kbd>
						{/if}
					</button>
				</li>
			{/each}
		</ul>
	{/if}
</div>
