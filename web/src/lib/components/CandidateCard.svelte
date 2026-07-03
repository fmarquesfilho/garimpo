<script>
	import ScoreMeter from './ScoreMeter.svelte';
	import { Button } from '$lib/components/ui';

	let { candidato, posicao = null, destaque = false, onpublicar = null } = $props();

	const brl = (v) => v.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' });
	const pct = (v) => `${(v * 100).toLocaleString('pt-BR', { maximumFractionDigits: 0 })}%`;

	function tempoRestante(iso) {
		if (!iso) return '';
		const diff = new Date(iso).getTime() - Date.now();
		if (diff <= 0) return 'expirado';
		const horas = Math.floor(diff / 3600000);
		if (horas < 24) return `${horas}h`;
		const dias = Math.floor(horas / 24);
		return `${dias}d`;
	}

	let copiado = $state(false);
	async function copiarLink() {
		if (!candidato.link) return;
		try {
			await navigator.clipboard.writeText(candidato.link);
			copiado = true;
			setTimeout(() => (copiado = false), 1600);
		} catch {
			copiado = false;
		}
	}
</script>

<article
	class="overflow-hidden rounded-md border border-border bg-card shadow-sm transition-[transform,box-shadow] duration-150 ease-linear hover:-translate-y-0.5 hover:shadow-[var(--sombra)]"
	class:!border-[var(--ouro-claro)]={destaque}
>
	{#if candidato.imagem}
		<a href={candidato.link || '#'} target="_blank" rel="noopener" class="block no-underline">
			<img
				src={candidato.imagem}
				alt={candidato.nome}
				class="block h-[180px] w-full bg-porcelana object-cover max-sm:h-[140px]"
				loading="lazy"
			/>
		</a>
	{/if}

	<div class="flex flex-col gap-3 p-4">
		{#if posicao != null}
			<span class="dado text-xs font-bold text-tinta-suave opacity-60">#{posicao}</span>
		{/if}

		<header>
			<h3 class="m-0 line-clamp-2 text-base font-semibold leading-tight">{candidato.nome}</h3>
			<div class="mt-1 flex flex-wrap items-center gap-x-1.5 gap-y-1">
				{#if candidato.loja}
					<span class="max-w-[140px] truncate text-xs font-semibold text-tinta-suave">🏪 {candidato.loja}</span>
				{/if}
				{#if candidato.origem}
					<span
						class="rounded-full bg-[var(--sucesso-fundo)] px-1.5 py-px text-[0.65rem] font-bold text-[var(--sucesso-texto)]"
						>{#if candidato.origem === 'Coreia'}🇰🇷{:else if candidato.origem === 'Japão'}🇯🇵{:else if candidato.origem === 'China'}🇨🇳{/if}
						{candidato.origem}</span
					>
				{/if}
				{#if candidato.desconto > 0 && candidato.desconto <= 1}
					<span
						class="rounded-full bg-[var(--erro-fundo)] px-1.5 py-px text-[0.65rem] font-bold text-[var(--erro-texto)]"
						>🔥 {Math.round(candidato.desconto * 100)}% OFF</span
					>
				{:else if candidato.desconto > 1 && candidato.desconto <= 100}
					<span
						class="rounded-full bg-[var(--erro-fundo)] px-1.5 py-px text-[0.65rem] font-bold text-[var(--erro-texto)]"
						>🔥 {Math.round(candidato.desconto)}% OFF</span
					>
				{/if}
				{#if candidato.oferta_expira}
					<span
						class="rounded-full bg-ouro-fundo px-1.5 py-px text-[0.65rem] font-bold text-ouro-escuro"
						title="Oferta de afiliado expira em {new Date(candidato.oferta_expira).toLocaleDateString('pt-BR')}"
						>⏳ {tempoRestante(candidato.oferta_expira)}</span
					>
				{/if}
				{#if candidato.categoria}
					<span class="text-xs font-semibold lowercase text-rosa">{candidato.categoria}</span>
				{/if}
				{#if candidato.suspeito}
					<span
						class="rounded-full bg-[var(--erro-fundo)] px-1.5 py-px text-[0.65rem] font-bold text-[var(--erro-texto)]"
						title="Comissão alta com poucas vendas — pode ser produto sem tração real. Avalie antes de publicar."
						>⚠ suspeito</span
					>
				{/if}
			</div>
		</header>

		<div class="flex flex-col items-baseline justify-between gap-1 sm:flex-row">
			<div class="flex items-baseline gap-2">
				<span class="font-mono text-lg font-bold">{brl(candidato.preco)}</span>
				<span class="text-sm font-bold text-ouro">{pct(candidato.comissao)}</span>
			</div>
			<div class="flex gap-3 text-xs text-tinta-suave">
				<span>{candidato.vendas.toLocaleString('pt-BR')} vendas</span>
				<span>★ {candidato.avaliacao.toLocaleString('pt-BR', { minimumFractionDigits: 1 })}</span>
			</div>
		</div>

		<ScoreMeter score={candidato.score} componentes={candidato.componentes} animar={destaque} />

		<footer class="mt-2 flex gap-2">
			{#if onpublicar}
				<Button size="sm" onclick={() => onpublicar(candidato)}>📤 Publicar</Button>
			{/if}
			<Button variant="ghost" size="sm" onclick={copiarLink} disabled={!candidato.link}>
				{copiado ? '✓ Copiado' : '🔗 Link'}
			</Button>
		</footer>
	</div>
</article>
