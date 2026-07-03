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

<div class="agendador">
	<div class="topo">
		<span class="rotulo-secao">coleta automática</span>
		<div class="modo-toggle">
			<button class:ativo={modo === 'atalho'} onclick={() => (modo = 'atalho')} type="button">Atalhos</button>
			<button class:ativo={modo === 'avancado'} onclick={() => (modo = 'avancado')} type="button">Avançado</button>
		</div>
	</div>

	{#if modo === 'atalho'}
		<div class="grade-atalhos">
			{#each atalhos as a}
				<button
					type="button"
					class="atalho"
					class:selecionado={value === a.cron}
					onclick={() => selecionarAtalho(a.cron)}
				>
					{a.label}
				</button>
			{/each}
		</div>
	{:else}
		<div class="campo-avancado">
			<input
				type="text"
				class="entrada dado"
				bind:value
				placeholder="ex.: 0 8 * * * (min hora dia mês semana)"
				spellcheck="false"
			/>
			<p class="dica">
				Formato: <code>minuto hora dia-do-mês mês dia-da-semana</code>. Exemplos: <code>0 8 * * *</code> = todo dia às
				8h;
				<code>0 9 * * 1,4</code> = segunda e quinta às 9h.
			</p>
		</div>
	{/if}

	{#if value}
		<p class="resumo dado">
			⏱ {descricao(value)}
			<button type="button" class="limpar" onclick={() => (value = '')}>remover</button>
		</p>
	{:else}
		<p class="resumo dado mudo">Sem agendamento — a busca só roda quando você clicar.</p>
	{/if}
</div>

<style>
	.agendador {
		background: var(--nevoa);
		border: 1px solid var(--linha);
		border-radius: var(--raio);
		padding: var(--r4);
		display: flex;
		flex-direction: column;
		gap: var(--r3);
	}
	.topo {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: var(--r4);
	}
	.rotulo-secao {
		font-size: 0.8rem;
		font-weight: 700;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		color: var(--tinta-suave);
	}
	.modo-toggle {
		display: inline-flex;
		gap: 2px;
		background: var(--linha);
		border-radius: 999px;
		padding: 2px;
	}
	.modo-toggle button {
		border: none;
		background: transparent;
		font-size: 0.78rem;
		font-weight: 600;
		color: var(--tinta-suave);
		padding: 4px 12px;
		border-radius: 999px;
		cursor: pointer;
	}
	.modo-toggle button.ativo {
		background: var(--porcelana);
		color: var(--tinta);
	}
	.grade-atalhos {
		display: flex;
		flex-wrap: wrap;
		gap: var(--r2);
	}
	.atalho {
		border: 1px solid var(--linha);
		background: var(--porcelana);
		color: var(--tinta);
		font-size: 0.82rem;
		font-weight: 500;
		padding: 6px 12px;
		border-radius: 999px;
		cursor: pointer;
		transition: all 0.12s ease;
	}
	.atalho:hover {
		border-color: var(--ouro);
		color: var(--ouro);
	}
	.atalho.selecionado {
		background: var(--ouro-fundo);
		border-color: var(--ouro);
		color: var(--ouro-escuro);
		font-weight: var(--font-bold);
	}
	.campo-avancado {
		display: flex;
		flex-direction: column;
		gap: var(--r2);
	}
	.entrada {
		font-family: var(--mono);
		font-size: 0.9rem;
		padding: 8px 12px;
		border-radius: 10px;
		border: 1px solid var(--linha);
		background: var(--porcelana);
		color: var(--tinta);
		width: 100%;
	}
	.dica {
		font-size: 0.78rem;
		color: var(--tinta-suave);
		margin: 0;
		line-height: 1.5;
	}
	code {
		background: var(--ouro-fundo);
		padding: 1px 5px;
		border-radius: 4px;
		font-family: var(--mono);
		font-size: 0.85em;
	}
	.resumo {
		font-size: 0.82rem;
		margin: 0;
		display: flex;
		align-items: center;
		gap: var(--r3);
	}
	.mudo {
		color: var(--tinta-suave);
		font-style: italic;
	}
	.limpar {
		border: none;
		background: transparent;
		color: var(--tinta-suave);
		font-size: 0.75rem;
		cursor: pointer;
		padding: 2px 6px;
		border-radius: 6px;
	}
	.limpar:hover {
		color: var(--erro-texto);
	}
</style>
