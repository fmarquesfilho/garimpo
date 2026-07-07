<script>
	import { onMount } from 'svelte';
	import { buscarColetas, listarBuscasServidor } from '$lib/api.js';
	import { Select } from '$lib/components/ui';

	let dias = $state(30);
	const diasOpcoes = [7, 30, 90].map((d) => ({ value: String(d), label: `${d} dias` }));
	let coletas = $state([]);
	let buscas = $state([]);
	let carregando = $state(true);
	let erro = $state(null);

	const dataHora = (v) => {
		if (!v) return '—';
		const d = new Date(v);
		return d.toLocaleString('pt-BR', {
			day: '2-digit',
			month: '2-digit',
			year: '2-digit',
			hour: '2-digit',
			minute: '2-digit'
		});
	};

	const tempoAtras = (v) => {
		if (!v) return '';
		const diff = Date.now() - new Date(v).getTime();
		const min = Math.floor(diff / 60000);
		if (min < 60) return `${min}min atrás`;
		const h = Math.floor(min / 60);
		if (h < 24) return `${h}h atrás`;
		const d = Math.floor(h / 24);
		return `${d}d atrás`;
	};

	async function carregar() {
		carregando = true;
		erro = null;
		try {
			const [c, b] = await Promise.all([buscarColetas({ dias }), listarBuscasServidor()]);
			coletas = c?.coletas ?? [];
			buscas = b?.buscas ?? [];
		} catch (e) {
			erro = e.message;
		} finally {
			carregando = false;
		}
	}

	onMount(carregar);

	// Agrupa coletas por keyword para mostrar resumo
	let resumoPorKeyword = $derived(() => {
		const mapa = new Map();
		for (const c of coletas) {
			const key = c.keyword || '(sem keyword)';
			if (!mapa.has(key)) {
				mapa.set(key, {
					keyword: key,
					categoria: c.categoria,
					total: 0,
					ultimaColeta: c.coletado_em,
					produtosTotal: 0
				});
			}
			const r = mapa.get(key);
			r.total++;
			r.produtosTotal += c.produtos;
			if (new Date(c.coletado_em) > new Date(r.ultimaColeta)) {
				r.ultimaColeta = c.coletado_em;
			}
		}
		return [...mapa.values()].sort((a, b) => new Date(b.ultimaColeta).getTime() - new Date(a.ultimaColeta).getTime());
	});
</script>

<section class="max-w-[42rem] mb-8">
	<p class="rotulo">monitoramento</p>
	<h1 class="text-[clamp(2rem,6vw,3rem)] my-2 mb-4">Coletas</h1>
	<p class="text-tinta-suave mb-4">
		Histórico de coletas executadas e status das buscas agendadas. Cada coleta grava um snapshot dos top produtos do
		momento.
	</p>
	<span class="flex items-center gap-1.5 text-sm text-tinta-suave">
		janela:
		<Select
			value={String(dias)}
			onchange={(v) => {
				dias = Number(v);
				carregar();
			}}
			options={diasOpcoes}
			size="sm"
			class="w-28"
		/>
	</span>
</section>

{#if carregando}
	<p class="text-tinta-suave italic">Carregando histórico…</p>
{:else if erro}
	<div class="bg-nevoa border border-border rounded-md p-8 text-center">
		<p><strong>Não consegui carregar as coletas.</strong></p>
		<p>{erro}</p>
	</div>
{:else}
	<!-- Buscas agendadas -->
	{#if buscas.length > 0}
		<section class="mb-8">
			<h2 class="text-xl mb-2">Buscas agendadas</h2>
			<div class="grid grid-cols-[repeat(auto-fill,minmax(240px,1fr))] gap-4 mt-3">
				{#each buscas as b (b.id)}
					<div class="bg-nevoa border border-border rounded-md py-3 px-4 flex flex-col gap-2">
						<div class="flex justify-between items-center">
							<span class="font-bold text-sm">{b.id}</span>
							{#if b.cron}
								<span class="text-[0.7rem] font-mono py-0.5 px-2 rounded-full bg-ouro-fundo text-ouro-escuro"
									>⏱ {b.cron}</span
								>
							{:else}
								<span class="text-[0.7rem] py-0.5 px-2 rounded-full bg-porcelana text-tinta-suave border border-border"
									>manual</span
								>
							{/if}
						</div>
						<div class="flex flex-wrap gap-1">
							{#each b.keywords ?? [] as kw}
								<span
									class="text-xs font-semibold py-0.5 px-2 rounded-full bg-porcelana border border-border text-tinta"
									>{kw}</span
								>
							{/each}
						</div>
						<div class="text-xs text-tinta-suave">
							<span>{b.estrategia}</span> · <span>{b.categoria}</span> · <span>top {b.top}</span>
						</div>
					</div>
				{/each}
			</div>
		</section>
	{/if}

	<!-- Resumo por keyword -->
	{#if resumoPorKeyword().length > 0}
		<section class="mb-8">
			<h2 class="text-xl mb-2">Resumo por keyword</h2>
			<p class="text-tinta-suave text-sm mb-4">Quantas coletas cada keyword teve na janela de {dias} dias.</p>
			<div class="border border-border rounded-md overflow-hidden bg-nevoa">
				<div
					class="grid grid-cols-[1.4fr_1.2fr_1fr_1fr_0.8fr] gap-2 py-3 px-4 items-center bg-[color-mix(in_srgb,var(--porcelana)_70%,white)] text-[0.7rem] font-semibold tracking-wide uppercase text-tinta-suave border-b border-border max-md:hidden"
				>
					<span>keyword</span>
					<span>categoria</span>
					<span>coletas</span>
					<span>produtos</span>
					<span>última coleta</span>
				</div>
				{#each resumoPorKeyword() as r (r.keyword)}
					<div
						class="grid grid-cols-[1.4fr_1.2fr_1fr_1fr_0.8fr] max-md:grid-cols-[1fr_1fr] gap-2 max-md:gap-1 py-3 px-4 items-center border-t border-border first:border-t-0 text-sm"
					>
						<span class="font-semibold text-rosa">{r.keyword}</span>
						<span class="dado">{r.categoria || '—'}</span>
						<span class="dado">{r.total}</span>
						<span class="dado">{r.produtosTotal}</span>
						<span class="dado font-mono text-xs">{tempoAtras(r.ultimaColeta)}</span>
					</div>
				{/each}
			</div>
		</section>
	{/if}

	<!-- Histórico detalhado -->
	<section class="mb-8">
		<h2 class="text-xl mb-2">Histórico de execuções</h2>
		{#if coletas.length === 0}
			<div class="bg-nevoa border border-border rounded-md p-8 text-center">
				<p>Nenhuma coleta encontrada nesta janela.</p>
				<p class="text-tinta-suave text-sm">
					As coletas acontecem nos horários agendados (cron) ou quando disparadas manualmente.
				</p>
			</div>
		{:else}
			<p class="text-tinta-suave text-sm mb-4">{coletas.length} execuções nos últimos {dias} dias.</p>
			<div class="border border-border rounded-md overflow-hidden bg-nevoa">
				<div
					class="grid grid-cols-[1.4fr_1.2fr_1fr_1fr_0.8fr] gap-2 py-3 px-4 items-center bg-[color-mix(in_srgb,var(--porcelana)_70%,white)] text-[0.7rem] font-semibold tracking-wide uppercase text-tinta-suave border-b border-border max-md:hidden"
				>
					<span>quando</span>
					<span>keyword</span>
					<span>categoria</span>
					<span>estratégia</span>
					<span>produtos</span>
				</div>
				{#each coletas as c, i (i)}
					<div
						class="grid grid-cols-[1.4fr_1.2fr_1fr_1fr_0.8fr] max-md:grid-cols-[1fr_1fr] gap-2 max-md:gap-1 py-3 px-4 items-center border-t border-border first:border-t-0 text-sm"
					>
						<span class="dado font-mono text-xs">{dataHora(c.coletado_em)}</span>
						<span class="font-semibold text-rosa">{c.keyword || '—'}</span>
						<span class="dado">{c.categoria || '—'}</span>
						<span class="dado">{c.estrategia}</span>
						<span class="dado font-bold text-ouro">{c.produtos}</span>
					</div>
				{/each}
			</div>
		{/if}
	</section>
{/if}
