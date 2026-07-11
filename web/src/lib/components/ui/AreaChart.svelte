<script>
	/**
	 * AreaChart — gráfico de área interativo com gradiente usando LayerChart 2.0.
	 * Substitui MiniChart para visualizações mais ricas no dashboard.
	 *
	 * @prop data — array de { date: string, value: number }
	 * @prop altura — altura em px
	 * @prop formatValue — fn para formatar tooltip values
	 * @prop color — cor CSS (variable or hex)
	 * @prop showAxis — mostrar eixo X com datas
	 */
	import { Chart, Area, Axis, Svg, LinearGradient, Highlight, Spline } from 'layerchart';
	import { scaleTime, scaleLinear } from 'd3-scale';

	let { data = [], altura = 120, color = 'hsl(var(--primary))', showAxis = true, class: className = '' } = $props();

	let chartData = $derived(
		data.map((d) => ({
			date: new Date(d.date ?? d.data),
			value: d.value ?? d.valor ?? 0
		}))
	);

	let gradientId = $derived(`area-grad-${Math.random().toString(36).slice(2, 8)}`);
</script>

{#if chartData.length >= 2}
	<div class={className} style="height: {altura}px">
		<Chart
			data={chartData}
			x="date"
			y="value"
			xScale={scaleTime()}
			yScale={scaleLinear()}
			yNice
			padding={{ left: 0, right: 0, top: 8, bottom: showAxis ? 24 : 4 }}
		>
			<Svg>
				<defs>
					<LinearGradient id={gradientId} stops={[color, 'transparent']} vertical />
				</defs>
				<Area fill="url(#{gradientId})" />
				<Spline stroke={color} width={2} />
				{#if showAxis}
					<Axis
						placement="bottom"
						ticks={3}
						format={(d) => d.toLocaleDateString('pt-BR', { day: '2-digit', month: '2-digit' })}
					/>
				{/if}
				<Highlight area lines />
			</Svg>
		</Chart>
	</div>
{:else}
	<p class="m-0 text-sm italic text-muted-foreground">Aguardando dados…</p>
{/if}
