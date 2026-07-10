<script>
	/**
	 * JsonTree — JSON viewer recursivo minimalista.
	 * Zero dependências externas. Svelte 5.
	 *
	 * @prop data — qualquer valor JSON
	 * @prop depth — profundidade atual (internal)
	 * @prop expanded — se inicia expandido (default: true para depth < 2)
	 */
	let { data, depth = 0, expanded = undefined } = $props();
	let open = $state(expanded ?? depth < 2);

	let isObject = $derived(data !== null && typeof data === 'object' && !Array.isArray(data));
	let isArray = $derived(Array.isArray(data));
	let entries = $derived(isObject ? Object.entries(data) : isArray ? data.map((v, i) => [i, v]) : []);
</script>

{#if isObject || isArray}
	<span class="cursor-pointer select-none text-muted-foreground hover:text-foreground" onclick={() => (open = !open)}>
		{open ? '▾' : '▸'}
		<span class="text-xs opacity-60">{isArray ? `[${entries.length}]` : `{${entries.length}}`}</span>
	</span>
	{#if open}
		<div class="ml-4 border-l border-border pl-2">
			{#each entries as [key, value] (key)}
				<div class="my-0.5">
					<span class="font-semibold text-primary/80">{key}</span><span class="text-muted-foreground">: </span>
					{#if value !== null && typeof value === 'object'}
						<svelte:self data={value} depth={depth + 1} />
					{:else if typeof value === 'string'}
						<span class="text-green-700 dark:text-green-400">"{value}"</span>
					{:else if typeof value === 'number'}
						<span class="text-blue-600 dark:text-blue-400">{value}</span>
					{:else if typeof value === 'boolean'}
						<span class="text-orange-600 dark:text-orange-400">{String(value)}</span>
					{:else}
						<span class="text-muted-foreground">null</span>
					{/if}
				</div>
			{/each}
		</div>
	{/if}
{:else if typeof data === 'string'}
	<span class="text-green-700 dark:text-green-400">"{data}"</span>
{:else if typeof data === 'number'}
	<span class="text-blue-600 dark:text-blue-400">{data}</span>
{:else if typeof data === 'boolean'}
	<span class="text-orange-600 dark:text-orange-400">{String(data)}</span>
{:else}
	<span class="text-muted-foreground">null</span>
{/if}
