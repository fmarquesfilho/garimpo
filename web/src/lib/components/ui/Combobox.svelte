<script>
	/**
	 * Combobox — autocomplete de adição. Filtra `items` conforme o texto e chama
	 * `onselect(item)` ao escolher. Diferente do Select, não mantém um valor
	 * selecionado: cada escolha é uma ação de adicionar (categorias, lojas…).
	 *
	 * Suporta entrada livre (`allowFree`): quando o texto não casa com nenhum item
	 * e parece uma entrada válida, oferece uma opção extra que chama `onfree(texto)`.
	 * Usado na raia de Lojas para "resolver e adicionar" uma loja nova por URL/ID.
	 *
	 * @prop items — lista de objetos a filtrar
	 * @prop getLabel — (item) => string usada no filtro e no fallback de render
	 * @prop onselect — (item) => void ao escolher um item
	 * @prop option — snippet(item) para renderizar cada opção (nome + metadados)
	 * @prop allowFree — habilita a opção de entrada livre
	 * @prop isFree — (texto) => boolean: quando mostrar a opção livre (ex.: parece URL/ID)
	 * @prop freeLabel — (texto) => string do rótulo da opção livre
	 * @prop onfree — (texto) => void ao escolher a opção livre
	 */
	import { cn } from '$lib/utils';

	const SIZES = { sm: 'h-8 text-xs', md: 'h-9 text-sm', lg: 'h-11 text-base' };

	let {
		items = [],
		getLabel = (i) => i?.nome ?? String(i),
		onselect = null,
		option = null,
		placeholder = '',
		allowFree = false,
		isFree = (t) => t.trim().length > 0,
		freeLabel = (t) => `↳ adicionar “${t}”`,
		onfree = null,
		size = 'md',
		class: className = ''
	} = $props();

	let query = $state('');
	let open = $state(false);
	let active = $state(0);
	let root;

	let filtrados = $derived(
		query.trim() ? items.filter((i) => getLabel(i).toLowerCase().includes(query.trim().toLowerCase())) : items
	);
	let temExato = $derived(filtrados.some((i) => getLabel(i).toLowerCase() === query.trim().toLowerCase()));
	let mostrarLivre = $derived(allowFree && query.trim().length > 0 && !temExato && isFree(query));
	// total de linhas navegáveis (itens + opção livre)
	let total = $derived(filtrados.length + (mostrarLivre ? 1 : 0));

	function abrir() {
		open = true;
		active = 0;
	}
	function fechar() {
		open = false;
	}
	function escolher(i) {
		onselect?.(i);
		query = '';
		fechar();
	}
	function escolherLivre() {
		onfree?.(query.trim());
		query = '';
		fechar();
	}
	function ativar(idx) {
		if (idx < filtrados.length) escolher(filtrados[idx]);
		else if (mostrarLivre) escolherLivre();
	}

	function onkeydown(e) {
		if (!open && (e.key === 'ArrowDown' || e.key === 'Enter')) {
			abrir();
			return;
		}
		if (e.key === 'ArrowDown') {
			e.preventDefault();
			active = total ? (active + 1) % total : 0;
		} else if (e.key === 'ArrowUp') {
			e.preventDefault();
			active = total ? (active - 1 + total) % total : 0;
		} else if (e.key === 'Enter') {
			e.preventDefault();
			if (total) ativar(active);
		} else if (e.key === 'Escape') {
			fechar();
			e.currentTarget.blur();
		}
	}

	function onblurRoot(e) {
		// fecha ao sair do componente (não ao mover foco interno)
		if (!root?.contains(e.relatedTarget)) fechar();
	}
</script>

<div class={cn('relative', className)} bind:this={root} onfocusout={onblurRoot}>
	<input
		type="text"
		bind:value={query}
		{placeholder}
		autocomplete="off"
		role="combobox"
		aria-expanded={open}
		aria-controls="combobox-list"
		class={cn(
			'w-full rounded-sm border border-input bg-background px-3 text-foreground placeholder:text-muted-foreground focus-visible:border-ring focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/20',
			SIZES[size] ?? SIZES.md
		)}
		onfocus={abrir}
		oninput={abrir}
		{onkeydown}
	/>

	{#if open && (total > 0 || query.trim())}
		<ul
			id="combobox-list"
			role="listbox"
			class="absolute left-0 right-0 top-[calc(100%+4px)] z-50 max-h-64 overflow-y-auto rounded-md border border-border bg-popover p-1 shadow-md"
		>
			{#each filtrados as item, i (getLabel(item))}
				<li role="option" aria-selected={active === i}>
					<button
						type="button"
						class={cn(
							'flex w-full items-center justify-between gap-3 rounded-sm px-3 py-2 text-left text-sm transition-colors',
							active === i ? 'bg-accent text-accent-foreground' : 'hover:bg-accent'
						)}
						onmouseenter={() => (active = i)}
						onclick={() => escolher(item)}
					>
						{#if option}{@render option(item)}{:else}<span>{getLabel(item)}</span>{/if}
					</button>
				</li>
			{/each}

			{#if mostrarLivre}
				<li role="option" aria-selected={active === filtrados.length}>
					<button
						type="button"
						class={cn(
							'flex w-full items-center gap-2 rounded-sm px-3 py-2 text-left text-sm font-medium text-primary transition-colors',
							active === filtrados.length ? 'bg-accent' : 'hover:bg-accent'
						)}
						onmouseenter={() => (active = filtrados.length)}
						onclick={escolherLivre}
					>
						{freeLabel(query.trim())}
					</button>
				</li>
			{/if}

			{#if total === 0 && !mostrarLivre}
				<li class="px-3 py-2 text-center text-sm text-muted-foreground">Nenhum resultado</li>
			{/if}
		</ul>
	{/if}
</div>
