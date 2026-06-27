<script>
	/**
	 * Card de oportunidade: queda de preço, alta de preço ou produto novo.
	 * Usado no feed de oportunidades.
	 */
	import { brl, pct, tempoAtras } from '$lib/formatters.js';

	let { item, tipo = 'queda', nomeLoja = '', onpublicar = null } = $props();
</script>

<div class="card-oportunidade {tipo}">
	<div class="card-header">
		{#if tipo === 'queda'}
			<span class="badge-variacao badge-queda">↓ {Math.abs(item.variacao_pct * 100).toFixed(0)}%</span>
		{:else if tipo === 'alta'}
			<span class="badge-variacao badge-alta">↑ {Math.abs(item.variacao_pct * 100).toFixed(0)}%</span>
		{:else}
			<span class="badge-novo">Novo</span>
		{/if}
		<span class="loja-tag">{nomeLoja || item.loja}</span>
		<span class="tempo">{tempoAtras(item.detectado_em)}</span>
	</div>

	<h3 class="card-nome">{item.nome}</h3>

	<div class="card-precos">
		{#if tipo === 'queda'}
			<span class="preco-antes">{brl(item.preco_anterior)}</span>
			<span class="seta">→</span>
			<span class="preco-atual destaque-queda">{brl(item.preco_atual)}</span>
			<span class="economia">(-{brl(item.preco_anterior - item.preco_atual)})</span>
		{:else if tipo === 'alta'}
			<span class="preco-antes">{brl(item.preco_anterior)}</span>
			<span class="seta">→</span>
			<span class="preco-atual destaque-alta">{brl(item.preco_atual)}</span>
		{:else}
			<span class="preco-atual">{brl(item.preco)}</span>
			{#if item.comissao > 0}
				<span class="comissao">{pct(item.comissao)} comissão</span>
			{/if}
			{#if item.vendas > 0}
				<span class="vendas">{item.vendas} vendas</span>
			{/if}
		{/if}
	</div>

	{#if onpublicar}
		<div class="card-acoes">
			<button class="btn-publicar" onclick={() => onpublicar(item)}>
				📤 Publicar
			</button>
		</div>
	{/if}
</div>

<style>
	.card-oportunidade {
		border: 1px solid var(--linha);
		border-radius: var(--raio);
		padding: var(--r4);
		background: var(--branco);
		transition: border-color 0.15s;
	}
	.card-oportunidade:hover { border-color: var(--ouro-claro); }
	.card-oportunidade.queda { border-left: 3px solid var(--sucesso-texto); }
	.card-oportunidade.alta { border-left: 3px solid var(--erro-texto); }
	.card-oportunidade.novo { border-left: 3px solid var(--ouro); }

	.card-header { display: flex; align-items: center; gap: var(--r2); margin-bottom: 6px; }
	.badge-variacao { padding: 2px 8px; border-radius: var(--raio-full); font-size: 0.72rem; font-weight: 700; }
	.badge-queda { background: var(--sucesso-fundo); color: var(--sucesso-texto); }
	.badge-alta { background: var(--erro-fundo); color: var(--erro-texto); }
	.badge-novo {
		padding: 2px 8px; border-radius: var(--raio-full);
		font-size: 0.72rem; font-weight: 700;
		background: var(--ouro-fundo); color: var(--ouro-escuro);
	}
	.loja-tag {
		font-size: 0.72rem; color: var(--tinta-suave);
		background: var(--porcelana); padding: 1px 6px; border-radius: 4px;
	}
	.tempo { font-size: 0.72rem; color: var(--tinta-suave); margin-left: auto; }

	.card-nome {
		font-size: 0.95rem; font-weight: 600; margin: 0 0 8px;
		line-height: 1.3; display: -webkit-box;
		-webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden;
	}

	.card-precos { display: flex; align-items: center; gap: var(--r2); font-size: 0.88rem; flex-wrap: wrap; }
	.preco-antes { text-decoration: line-through; color: var(--tinta-suave); }
	.seta { color: var(--tinta-suave); font-size: 0.8rem; }
	.preco-atual { font-weight: 700; }
	.destaque-queda { color: var(--sucesso-texto); }
	.destaque-alta { color: var(--erro-texto); }
	.economia { font-size: 0.78rem; color: var(--sucesso-texto); font-weight: 600; }
	.comissao { font-size: 0.78rem; color: var(--ouro); font-weight: 600; }
	.vendas { font-size: 0.78rem; color: var(--tinta-suave); }

	.card-acoes { margin-top: 10px; }
	.btn-publicar {
		padding: 6px 14px; border: 1px solid var(--ouro-claro);
		background: var(--ouro-fundo); color: var(--ouro-escuro);
		border-radius: var(--raio-sm); font-size: 0.82rem;
		font-weight: 600; cursor: pointer;
	}
	.btn-publicar:hover { background: var(--ouro-claro); }
</style>
