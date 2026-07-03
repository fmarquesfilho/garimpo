<script>
	// Seletor visual de agendamento. Gera uma expressão cron sem expor a sintaxe
	// para quem não conhece. O campo `cron` (bind:value) é sempre válido ou vazio.
	let { value = $bindable('') } = $props();

	// Atalhos de frequência mais usados
	const atalhos = [
		{ label: 'Nunca', cron: '' },
		{ label: 'Todo dia às 8h', cron: '0 8 * * *' },
		{ label: 'Todo dia às 12h', cron: '0 12 * * *' },
		{ label: 'Todo dia às 18h', cron: '0 18 * * *' },
		{ label: '2× por dia (8h e 18h)', cron: '0 8,18 * * *' },
		{ label: 'Seg e Qui às 9h', cron: '0 9 * * 1,4' },
		{ label: 'Segunda-feira às 8h', cron: '0 8 * * 1' },
		{ label: 'Todo sábado às 9h', cron: '0 9 * * 6' }
	];

	// Modo: 'atalho' (padrão) ou 'avancado' (campo livre)
	let modo = $state('atalho');

	function selecionarAtalho(cron) {
		value = cron;
		modo = 'atalho';
	}

	function descricao(cron) {
		if (!cron) return 'Sem coleta automática — só manual';
		const a = atalhos.find((x) => x.cron === cron);
		return a ? a.label : `Cron personalizado: ${cron}`;
	}
</script>

<div class="flex flex-col gap-3 rounded-md border border-border bg-card p-4">
	<div class="flex items-center justify-between gap-4">
		<span class="text-xs font-bold uppercase tracking-wide text-tinta-suave">coleta automática</span>
		<div class="inline-flex gap-0.5 rounded-full bg-[var(--linha)] p-0.5">
			<button
				class="cursor-pointer rounded-full border-none bg-transparent px-3 py-1 text-xs font-semibold text-tinta-suave"
				class:!bg-porcelana={modo === 'atalho'}
				class:!text-foreground={modo === 'atalho'}
				onclick={() => (modo = 'atalho')}
				type="button">Atalhos</button
			>
			<button
				class="cursor-pointer rounded-full border-none bg-transparent px-3 py-1 text-xs font-semibold text-tinta-suave"
				class:!bg-porcelana={modo === 'avancado'}
				class:!text-foreground={modo === 'avancado'}
				onclick={() => (modo = 'avancado')}
				type="button">Avançado</button
			>
		</div>
	</div>

	{#if modo === 'atalho'}
		<div class="flex flex-wrap gap-2">
			{#each atalhos as a}
				<button
					type="button"
					class="cursor-pointer rounded-full border border-border bg-porcelana px-3 py-1.5 text-sm font-medium text-foreground transition-all duration-100 ease-linear hover:border-ouro hover:text-ouro"
					class:!bg-ouro-fundo={value === a.cron}
					class:!border-ouro={value === a.cron}
					class:!text-ouro-escuro={value === a.cron}
					class:!font-bold={value === a.cron}
					onclick={() => selecionarAtalho(a.cron)}
				>
					{a.label}
				</button>
			{/each}
		</div>
	{:else}
		<div class="flex flex-col gap-2">
			<input
				type="text"
				class="dado w-full rounded-[10px] border border-border bg-porcelana px-3 py-2 font-mono text-[0.9rem] text-foreground"
				bind:value
				placeholder="ex.: 0 8 * * * (min hora dia mês semana)"
				spellcheck="false"
			/>
			<p class="m-0 text-xs leading-relaxed text-tinta-suave">
				Formato: <code class="rounded bg-ouro-fundo px-1.5 py-px font-mono text-[0.85em]"
					>minuto hora dia-do-mês mês dia-da-semana</code
				>. Exemplos: <code class="rounded bg-ouro-fundo px-1.5 py-px font-mono text-[0.85em]">0 8 * * *</code> = todo
				dia às 8h;
				<code class="rounded bg-ouro-fundo px-1.5 py-px font-mono text-[0.85em]">0 9 * * 1,4</code> = segunda e quinta às
				9h.
			</p>
		</div>
	{/if}

	{#if value}
		<p class="dado m-0 flex items-center gap-3 text-sm">
			⏱ {descricao(value)}
			<button
				type="button"
				class="cursor-pointer rounded-md border-none bg-transparent px-1.5 py-0.5 text-xs text-tinta-suave hover:text-[var(--erro-texto)]"
				onclick={() => (value = '')}>remover</button
			>
		</p>
	{:else}
		<p class="dado m-0 text-sm italic text-tinta-suave">Sem agendamento — a busca só roda quando você clicar.</p>
	{/if}
</div>
