<script>
	// Seletor visual de agendamento. Gera uma expressão cron sem expor a sintaxe
	// para quem não conhece. O campo `cron` (bind:value) é sempre válido ou vazio.
	// `permitirNunca=false` esconde a opção "Nunca" (ex.: loja monitorada, que sempre
	// coleta periodicamente).
	import { Input, ToggleGroup, Button } from '$lib/components/ui';

	let { value = $bindable(''), permitirNunca = true } = $props();

	// Sentinel para representar "Nunca" (cron vazio) no ToggleGroup sem colidir com
	// o estado "nada selecionado".
	const NUNCA = '__nunca__';

	// Atalhos de frequência mais usados
	const atalhos = $derived([
		...(permitirNunca ? [{ label: 'Nunca', cron: '' }] : []),
		{ label: 'A cada 8h', cron: '0 */8 * * *' },
		{ label: 'Todo dia às 8h', cron: '0 8 * * *' },
		{ label: 'Todo dia às 12h', cron: '0 12 * * *' },
		{ label: 'Todo dia às 18h', cron: '0 18 * * *' },
		{ label: '2× por dia (8h e 18h)', cron: '0 8,18 * * *' },
		{ label: 'Seg e Qui às 9h', cron: '0 9 * * 1,4' },
		{ label: 'Segunda-feira às 8h', cron: '0 8 * * 1' },
		{ label: 'Todo sábado às 9h', cron: '0 9 * * 6' }
	]);

	const modoOpcoes = [
		{ value: 'atalho', label: 'Atalhos' },
		{ value: 'avancado', label: 'Avançado' }
	];

	let presetOpcoes = $derived(atalhos.map((a) => ({ value: a.cron === '' ? NUNCA : a.cron, label: a.label })));
	let presetValor = $derived(value === '' ? (permitirNunca ? NUNCA : '') : value);

	// Modo: 'atalho' (padrão) ou 'avancado' (campo livre)
	let modo = $state('atalho');

	function selecionarPreset(v) {
		value = v === NUNCA ? '' : v;
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
		<span class="text-xs font-bold uppercase tracking-wide text-muted-foreground">coleta automática</span>
		<ToggleGroup
			value={modo}
			onchange={(v) => (modo = v)}
			options={modoOpcoes}
			variant="segment"
			size="sm"
			nullable={false}
		/>
	</div>

	{#if modo === 'atalho'}
		<ToggleGroup
			value={presetValor}
			onchange={selecionarPreset}
			options={presetOpcoes}
			variant="chips"
			nullable={false}
		/>
	{:else}
		<div class="flex flex-col gap-2">
			<Input bind:value placeholder="ex.: 0 8 * * * (min hora dia mês semana)" spellcheck="false" class="font-mono" />
			<p class="m-0 text-xs leading-relaxed text-muted-foreground">
				Formato: <code class="rounded bg-accent px-1.5 py-px font-mono text-[0.85em]"
					>minuto hora dia-do-mês mês dia-da-semana</code
				>. Exemplos: <code class="rounded bg-accent px-1.5 py-px font-mono text-[0.85em]">0 8 * * *</code> = todo dia às
				8h;
				<code class="rounded bg-accent px-1.5 py-px font-mono text-[0.85em]">0 9 * * 1,4</code> = segunda e quinta às 9h.
			</p>
		</div>
	{/if}

	{#if value}
		<p class="dado m-0 flex items-center gap-2 text-sm">
			⏱ {descricao(value)}
			{#if permitirNunca}
				<Button variant="ghost" size="sm" onclick={() => (value = '')}>remover</Button>
			{/if}
		</p>
	{:else if permitirNunca}
		<p class="dado m-0 text-sm italic text-muted-foreground">Sem agendamento — a busca só roda quando você clicar.</p>
	{/if}
</div>
