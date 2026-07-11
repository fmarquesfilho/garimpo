<script>
	import { onMount, onDestroy } from 'svelte';
	import {
		buscarSaudeColetas,
		buscarOportunidadesAgora,
		buscarResumoConversoes,
		buscarEficaciaAlertas,
		buscarEvolucaoLojas,
		buscarDashboardChanges
	} from '$lib/api.js';
	import { createPollingTimer } from '$lib/polling.svelte.js';
	import { brl, num, pctSinal } from '$lib/formatters.js';
	import {
		MetricCard,
		Select,
		Badge,
		Collapsible,
		StatSection,
		OpportunityCard,
		FreshnessBar,
		AnimatedMetric
	} from '$lib/components/ui/index.js';
	import AreaChart from '$lib/components/ui/AreaChart.svelte';

	let dias = $state(7);
	const diasOpcoes = [7, 30, 90].map((d) => ({ value: String(d), label: `${d} dias` }));

	// Section states
	let saude = $state({ loading: true, error: null, data: null });
	let oportunidades = $state({ loading: true, error: null, data: null });
	let performance = $state({ loading: true, error: null, data: null });
	let eficacia = $state({ loading: true, error: null, data: null });
	let evolucao = $state({ loading: true, error: null, data: null });

	// Polling state: local timestamps for change detection
	let lastFetched = $state({ saude: null, oportunidades: null, performance: null });
	let highlightSection = $state(null); // section name to highlight on update

	// ── Initial load ──────────────────────────────────────────────────────────
	async function carregar() {
		saude = { loading: true, error: null, data: null };
		oportunidades = { loading: true, error: null, data: null };
		performance = { loading: true, error: null, data: null };
		eficacia = { loading: true, error: null, data: null };
		evolucao = { loading: true, error: null, data: null };
		lastFetched = { saude: null, oportunidades: null, performance: null };

		buscarSaudeColetas()
			.then((d) => { saude = { loading: false, error: null, data: d }; })
			.catch((e) => { saude = { loading: false, error: e.message, data: null }; });

		buscarOportunidadesAgora({ dias })
			.then((d) => { oportunidades = { loading: false, error: null, data: d }; })
			.catch((e) => { oportunidades = { loading: false, error: e.message, data: null }; });

		buscarResumoConversoes({ dias })
			.then((d) => { performance = { loading: false, error: null, data: d }; })
			.catch((e) => { performance = { loading: false, error: e.message, data: null }; });

		buscarEficaciaAlertas({ dias })
			.then((d) => { eficacia = { loading: false, error: null, data: d }; })
			.catch((e) => { eficacia = { loading: false, error: e.message, data: null }; });

		buscarEvolucaoLojas({ dias })
			.then((d) => { evolucao = { loading: false, error: null, data: d }; })
			.catch((e) => { evolucao = { loading: false, error: e.message, data: null }; });
	}

	// ── Smart polling ─────────────────────────────────────────────────────────
	async function onPollTick() {
		const changes = await buscarDashboardChanges();
		const promises = [];

		if (changes.saude_updated_at && changes.saude_updated_at !== lastFetched.saude) {
			promises.push(
				buscarSaudeColetas().then((d) => {
					saude = { loading: false, error: null, data: d };
					lastFetched.saude = changes.saude_updated_at;
					flashSection('saude');
				})
			);
		}

		if (changes.oportunidades_updated_at && changes.oportunidades_updated_at !== lastFetched.oportunidades) {
			promises.push(
				buscarOportunidadesAgora({ dias }).then((d) => {
					oportunidades = { loading: false, error: null, data: d };
					lastFetched.oportunidades = changes.oportunidades_updated_at;
					flashSection('oportunidades');
				})
			);
		}

		if (changes.performance_updated_at && changes.performance_updated_at !== lastFetched.performance) {
			promises.push(
				buscarResumoConversoes({ dias }).then((d) => {
					performance = { loading: false, error: null, data: d };
					lastFetched.performance = changes.performance_updated_at;
					flashSection('performance');
				}),
				buscarEficaciaAlertas({ dias }).then((d) => {
					eficacia = { loading: false, error: null, data: d };
				})
			);
		}

		if (promises.length > 0) await Promise.all(promises);
	}

	function flashSection(name) {
		highlightSection = name;
		setTimeout(() => { highlightSection = null; }, 1200);
	}

	const poll = createPollingTimer({
		onTick: onPollTick,
		onFullRefresh: carregar
	});

	onMount(async () => {
		await carregar();
		poll.start();
	});

	onDestroy(() => poll.destroy());

	// ── Derived ───────────────────────────────────────────────────────────────
	let tempoRelativo = $derived(() => {
		const min = saude.data?.minutos_desde_ultima;
		if (min == null) return '—';
		if (min < 60) return `${min}min atrás`;
		const h = Math.floor(min / 60);
		if (h < 24) return `${h}h atrás`;
		return `${Math.floor(h / 24)}d atrás`;
	});

	let evolucaoLojas = $derived(evolucao.data?.lojas ?? []);
	let evolucaoKeywords = $derived(evolucao.data?.keywords ?? []);
	let topEvolucao = $derived(
		[...evolucaoLojas, ...evolucaoKeywords]
			.filter((e) => e.pontos?.length >= 2)
			.sort((a, b) => Math.abs(b.variacao_media_pct) - Math.abs(a.variacao_media_pct))
			.slice(0, 3)
	);
</script>

<svelte:head>
	<title>Dashboard — Garimpei</title>
</svelte:head>

<div class="mx-auto max-w-[960px] space-y-5 pb-12">
	<!-- Header with FreshnessBar -->
	<header class="flex flex-wrap items-center justify-between gap-3">
		<h1 class="m-0 text-2xl font-bold">📊 Dashboard</h1>
		<FreshnessBar lastUpdate={poll.lastTickAt} countdown={poll.countdown} status={poll.status} />
		<Select
			value={String(dias)}
			onchange={(v) => { dias = Number(v); carregar(); }}
			options={diasOpcoes}
			size="sm"
			class="w-28"
		/>
	</header>

	<!-- ═══ SEÇÃO 1: SAÚDE ═══ -->
	<StatSection
		icon="🔄"
		title="Saúde das coletas"
		subtitle="Suas buscas agendadas estão rodando?"
		loading={saude.loading}
		error={saude.error}
		empty={saude.data?.status === 'sem_dados'}
		emptyMessage="Nenhuma coleta registrada. Configure uma busca com agendamento para começar."
		class={highlightSection === 'saude' ? 'ring-2 ring-primary/30 transition-all duration-700' : ''}
	>
		<div class="flex flex-wrap items-center gap-4">
			<span class="inline-flex items-center gap-2 rounded-full border border-border bg-card px-3 py-1 text-sm font-medium"
				class:text-emerald-600={saude.data?.status === 'ok'}
				class:text-amber-600={saude.data?.status === 'atrasado'}
			>
				<span class="relative flex h-2.5 w-2.5">
					<span class="absolute inline-flex h-full w-full animate-ping rounded-full opacity-75"
						class:bg-emerald-500={saude.data?.status === 'ok'}
						class:bg-amber-500={saude.data?.status === 'atrasado'}
					></span>
					<span class="relative inline-flex h-2.5 w-2.5 rounded-full"
						class:bg-emerald-500={saude.data?.status === 'ok'}
						class:bg-amber-500={saude.data?.status === 'atrasado'}
						class:bg-muted-foreground={!saude.data?.status || saude.data?.status === 'sem_dados'}
					></span>
				</span>
				{saude.data?.status === 'ok' ? 'Operando' : saude.data?.status === 'atrasado' ? 'Atrasado' : 'Sem dados'}
			</span>
			<span class="text-sm text-muted-foreground">
				Última coleta: <strong>{tempoRelativo()}</strong>
			</span>
			<span class="text-sm text-muted-foreground">
				Coletas 24h: <AnimatedMetric value={saude.data?.coletas_24h ?? 0} class="font-bold" /> / {saude.data?.coletas_esperadas_24h ?? '?'}
			</span>
		</div>
		{#if saude.data?.keywords_atrasadas?.length > 0}
			<div class="mt-3 flex flex-wrap gap-1.5">
				<span class="text-xs text-muted-foreground">Sem coleta recente:</span>
				{#each saude.data.keywords_atrasadas.slice(0, 5) as kw}
					<Badge variant="warning">{kw}</Badge>
				{/each}
				{#if saude.data.keywords_atrasadas.length > 5}
					<Badge variant="outline">+{saude.data.keywords_atrasadas.length - 5}</Badge>
				{/if}
			</div>
		{/if}
	</StatSection>

	<!-- ═══ SEÇÃO 2: OPORTUNIDADES ═══ -->
	<StatSection
		icon="💰"
		title="Oportunidades agora"
		subtitle="Produtos para publicar — quedas, novos, alto valor"
		loading={oportunidades.loading}
		error={oportunidades.error}
		empty={!oportunidades.data?.total_quedas && !oportunidades.data?.total_novos && !oportunidades.data?.total_alto_valor}
		emptyMessage="Nenhuma oportunidade detectada no período. Os dados aparecerão após as próximas coletas."
		class={highlightSection === 'oportunidades' ? 'ring-2 ring-primary/30 transition-all duration-700' : ''}
	>
		<div class="mb-4 grid grid-cols-3 gap-3">
			<div class="rounded-lg border border-border bg-muted/30 p-3 text-center">
				<p class="m-0 text-2xl font-bold text-emerald-600"><AnimatedMetric value={oportunidades.data?.total_quedas ?? 0} /></p>
				<p class="m-0 text-xs text-muted-foreground">📉 Quedas</p>
			</div>
			<div class="rounded-lg border border-border bg-muted/30 p-3 text-center">
				<p class="m-0 text-2xl font-bold"><AnimatedMetric value={oportunidades.data?.total_novos ?? 0} /></p>
				<p class="m-0 text-xs text-muted-foreground">🆕 Novos</p>
			</div>
			<div class="rounded-lg border border-border bg-muted/30 p-3 text-center">
				<p class="m-0 text-2xl font-bold text-primary"><AnimatedMetric value={oportunidades.data?.total_alto_valor ?? 0} /></p>
				<p class="m-0 text-xs text-muted-foreground">💎 Alto valor</p>
			</div>
		</div>

		{#if oportunidades.data?.quedas?.length > 0}
			<h3 class="mb-2 mt-0 text-sm font-semibold text-muted-foreground">Maiores quedas</h3>
			<div class="space-y-2">
				{#each oportunidades.data.quedas.slice(0, 5) as produto (produto.produto_id)}
					<OpportunityCard {produto} tipo="queda" />
				{/each}
			</div>
		{/if}

		{#if oportunidades.data?.novos?.length > 0}
			<h3 class="mb-2 mt-4 text-sm font-semibold text-muted-foreground">Produtos novos</h3>
			<div class="space-y-2">
				{#each oportunidades.data.novos.slice(0, 3) as produto (produto.produto_id)}
					<OpportunityCard {produto} tipo="novo" />
				{/each}
			</div>
		{/if}
	</StatSection>

	<!-- ═══ SEÇÃO 3: PERFORMANCE ═══ -->
	<StatSection
		icon="📈"
		title="Performance"
		subtitle="Receita, conversões e eficácia dos alertas"
		loading={performance.loading && eficacia.loading}
		error={performance.error && eficacia.error ? 'Indisponível' : null}
		empty={performance.data?.status === 'sem_dados' && !eficacia.data?.quedas_detectadas}
		emptyMessage="Sem dados de conversão ainda. Publique produtos e as métricas aparecerão aqui."
		class={highlightSection === 'performance' ? 'ring-2 ring-primary/30 transition-all duration-700' : ''}
	>
		<div class="grid grid-cols-2 gap-3 sm:grid-cols-4">
			<div class="rounded-lg border border-border bg-muted/30 p-3 text-center">
				<p class="m-0 text-lg font-bold text-primary">
					<AnimatedMetric value={performance.data?.comissao_total ?? 0} format={brl} />
				</p>
				<p class="m-0 text-xs text-muted-foreground">Comissão</p>
			</div>
			<div class="rounded-lg border border-border bg-muted/30 p-3 text-center">
				<p class="m-0 text-lg font-bold"><AnimatedMetric value={performance.data?.conversoes ?? 0} /></p>
				<p class="m-0 text-xs text-muted-foreground">Conversões</p>
			</div>
			<div class="rounded-lg border border-border bg-muted/30 p-3 text-center">
				<p class="m-0 text-lg font-bold">
					<AnimatedMetric value={eficacia.data?.taxa_deteccao ?? 0} format={(v) => `${Math.round(v)}%`} />
				</p>
				<p class="m-0 text-xs text-muted-foreground">Taxa detecção</p>
			</div>
			<div class="rounded-lg border border-border bg-muted/30 p-3 text-center">
				<p class="m-0 text-lg font-bold">{performance.data?.melhor_canal ?? '—'}</p>
				<p class="m-0 text-xs text-muted-foreground">Melhor canal</p>
			</div>
		</div>

		{#if eficacia.data?.quedas_detectadas > 0}
			<div class="mt-4 grid grid-cols-3 gap-3 rounded-lg border border-border bg-card p-3">
				<div class="text-center">
					<p class="m-0 text-lg font-bold"><AnimatedMetric value={eficacia.data.quedas_detectadas} /></p>
					<p class="m-0 text-xs text-muted-foreground">Quedas detectadas</p>
				</div>
				<div class="text-center">
					<p class="m-0 text-lg font-bold"><AnimatedMetric value={eficacia.data.alertas_enviados} /></p>
					<p class="m-0 text-xs text-muted-foreground">Alertas enviados</p>
				</div>
				<div class="text-center">
					<p class="m-0 text-lg font-bold"><AnimatedMetric value={eficacia.data.conversoes_atribuidas} /></p>
					<p class="m-0 text-xs text-muted-foreground">Converteram</p>
				</div>
			</div>
		{/if}
	</StatSection>

	<!-- ═══ SEÇÃO 4: EVOLUÇÃO (colapsável) ═══ -->
	<Collapsible title="📈 Evolução de preços" class="rounded-xl border border-border">
		{#if evolucao.loading}
			<div class="flex items-center justify-center py-6">
				<div class="h-4 w-4 animate-spin rounded-full border-2 border-primary border-t-transparent"></div>
				<span class="ml-2 text-sm text-muted-foreground">Carregando…</span>
			</div>
		{:else if topEvolucao.length > 0}
			<div class="space-y-5 p-4">
				{#each topEvolucao as item (item.busca_id)}
					<div>
						<div class="mb-2 flex items-center justify-between">
							<span class="text-sm font-semibold">{item.busca_id}</span>
							<span
								class="text-sm font-bold"
								class:text-emerald-600={item.variacao_media_pct < 0}
								class:text-destructive={item.variacao_media_pct > 0}
							>
								{pctSinal(item.variacao_media_pct)}
							</span>
						</div>
						<AreaChart
							data={item.pontos?.map((p) => ({ date: p.data, value: p.preco_medio })) ?? []}
							formatValue={brl}
							altura={100}
							color={item.variacao_media_pct < 0 ? 'hsl(142 71% 45%)' : 'hsl(var(--destructive))'}
						/>
					</div>
				{/each}
			</div>
		{:else}
			<p class="m-0 p-4 text-sm italic text-muted-foreground">Sem dados de evolução no período.</p>
		{/if}
	</Collapsible>
</div>
