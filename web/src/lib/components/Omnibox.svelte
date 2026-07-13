<script>
	/**
	 * Omnibox — renderizador puro. Zero lógica de decisão.
	 * Lê estado derivado de engine.omnibox.
	 * Emite eventos brutos para a engine.
	 */
	let { engine } = $props();

	let om = $derived(engine.omnibox);
	let inputEl;
</script>

<div class="relative" onfocusout={(e) => {
	if (!e.currentTarget.contains(/** @type {Node} */ (e.relatedTarget)))
		engine.send({type: 'OMNIBOX_BLUR'});
}}>
	<span class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 opacity-50">🔍</span>
	<input
		bind:this={inputEl}
		type="text"
		value={om.inputValue}
		placeholder={om.placeholder}
		autocomplete="off"
		spellcheck="false"
		role="combobox"
		aria-expanded={om.aberto && om.opcoes.length > 0}
		aria-controls="omnibox-listbox"
		aria-autocomplete="list"
		aria-activedescendant={om.highlightIdx >= 0 ? `omnibox-opt-${om.highlightIdx}` : undefined}
		class="w-full rounded-sm border border-input bg-background py-2.5 pl-9 pr-4 text-base text-foreground placeholder:text-muted-foreground focus-visible:border-ring focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/20"
		oninput={(e) => engine.send({type: 'OMNIBOX_INPUT', value: e.currentTarget.value})}
		onfocus={() => engine.send({type: 'OMNIBOX_INPUT', value: om.inputValue})}
		onkeydown={(e) => {
			if (['Enter','ArrowDown','ArrowUp','Escape'].includes(e.key)) {
				e.preventDefault();
				engine.send({type: 'OMNIBOX_KEYDOWN', key: e.key});
			}
		}}
	/>

	<span class="sr-only" aria-live="polite">
		{om.aberto && om.opcoes.length > 0
			? `${om.opcoes.length} ${om.opcoes.length === 1 ? 'opção' : 'opções'}`
			: ''}
	</span>

	{#if om.aberto && om.opcoes.length > 0}
		<ul id="omnibox-listbox" role="listbox" aria-label="Opções de busca"
			class="absolute left-0 right-0 top-[calc(100%+4px)] z-50 max-h-80
				   overflow-y-auto rounded-md border border-border bg-popover p-1 shadow-md">
			{#each om.opcoes as opcao, i (opcao.tipo + ':' + i)}
				<li id={`omnibox-opt-${i}`} role="option"
					aria-selected={om.highlightIdx === i}
					aria-label={opcao.labelAcessivel ?? opcao.label}>
					<button type="button" tabindex="-1"
						class="flex w-full items-center gap-2 rounded-sm px-3 py-2 text-left text-sm
							   transition-colors {om.highlightIdx === i
								   ? 'bg-accent text-accent-foreground' : 'hover:bg-accent'}"
						onmouseenter={() => engine.send({type:'OMNIBOX_KEYDOWN', key:'highlight', idx: i})}
						onclick={() => engine.send({type:'OMNIBOX_SELECIONAR', indice: i})}>
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
