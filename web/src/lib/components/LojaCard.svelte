<script>
	/**
	 * LojaCard — loja no escopo da busca. Mostra o nome, o marketplace (esq.),
	 * a bandeira de origem cadastrada (China/Japão/Coreia — ver operação Shopee)
	 * e um indicador de monitoramento (temporizador se monitorada).
	 * @prop nome, marketplace, origem, monitorada, cron
	 * @prop onremover — () => void
	 */
	import { cronLabel } from '$lib/busca-engine.svelte.js';

	let {
		nome = '',
		marketplace = 'shopee',
		origem = null,
		monitorada = false,
		cron = '',
		tipo = '',
		onremover = null
	} = $props();

	// origem pode vir como emoji ("🇰🇷"), código ("coreia") ou rótulo ("🇰🇷 Coreia")
	const BANDEIRAS = {
		coreia: '🇰🇷',
		'coreia do sul': '🇰🇷',
		kr: '🇰🇷',
		japao: '🇯🇵',
		japão: '🇯🇵',
		jp: '🇯🇵',
		china: '🇨🇳',
		cn: '🇨🇳'
	};
	let bandeira = $derived.by(() => {
		if (!origem) return null;
		if (/\p{Regional_Indicator}/u.test(origem)) return origem.match(/\p{Regional_Indicator}{2}/u)?.[0] ?? origem;
		return BANDEIRAS[String(origem).toLowerCase().trim()] ?? '🏳️';
	});
</script>

<div class="relative min-w-[210px] rounded-sm border border-border bg-muted px-3 py-2.5">
	<button
		type="button"
		class="absolute right-1.5 top-1.5 rounded px-1 text-sm leading-none text-muted-foreground hover:text-destructive"
		onclick={() => onremover?.()}
		aria-label="Remover loja {nome}">✕</button
	>
	<div class="pr-5 text-[0.95rem] font-bold text-foreground">{nome}</div>
	<div class="mt-2 flex items-center justify-between gap-2.5">
		<span class="font-[var(--mono)] text-xs text-muted-foreground">{marketplace}</span>
		<span class="flex items-center gap-2">
			{#if bandeira}<span class="text-base" title="Origem cadastrada">{bandeira}</span>{/if}
			{#if tipo === 'monitorada' || monitorada}
				<span
					title={cron ? `Monitorada: ${cron}` : 'Loja monitorada'}
					class="inline-flex items-center gap-1 rounded-full border border-[var(--sucesso-borda)] bg-[var(--sucesso-fundo)] px-1.5 py-0.5 font-[var(--mono)] text-[0.68rem] text-[var(--sucesso-texto)]"
					>⏱ {cronLabel(cron) || 'monitorada'}</span
				>
			{:else}
				<span
					title="Loja apenas escopada (sem monitoramento ativo)"
					class="inline-flex items-center gap-1 rounded-full border border-border bg-card px-1.5 py-0.5 font-[var(--mono)] text-[0.68rem] text-muted-foreground"
					>○ escopada</span
				>
			{/if}
		</span>
	</div>
</div>
