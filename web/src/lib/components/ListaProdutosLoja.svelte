<script>
	/**
	 * Lista de produtos de uma loja monitorada.
	 * Exibe grid com thumb, preço, comissão, vendas e botão de publicar.
	 */
	import { brl, pct } from '$lib/formatters.js';
	import { Loading } from '$lib/components/ui/index.js';

	let { produtos = [], carregando = false, erro = null, onpublicar = null } = $props();
</script>

{#if carregando}
	<Loading mensagem="Buscando produtos da loja…" />
{:else if erro}
	<div class="msg-erro">{erro}</div>
{:else if produtos.length === 0}
	<p class="vazio-tab">Nenhum produto encontrado. A coleta periódica pode ainda não ter rodado.</p>
{:else}
	<div class="grade-produtos">
		{#each produtos as p (p.id)}
			<div class="card-produto-loja">
				{#if p.imagem}
					<img src={p.imagem} alt={p.nome} class="prod-thumb" />
				{/if}
				<div class="prod-info">
					<h4>{p.nome}</h4>
					<div class="prod-dados">
						<span class="prod-preco">{brl(p.preco)}</span>
						<span class="prod-comissao">{pct(p.comissao)}</span>
						<span class="prod-vendas">{p.vendas} vendas</span>
						<span class="prod-nota">★ {p.avaliacao?.toFixed(1)}</span>
					</div>
					<div class="prod-score">teor: {p.score?.toFixed(3)}</div>
				</div>
				{#if onpublicar}
					<button class="btn-pub-mini" onclick={() => onpublicar(p)} title="Publicar este produto">
						📤
					</button>
				{/if}
			</div>
		{/each}
	</div>
{/if}

<style>
	.msg-erro {
		background: var(--erro-fundo); color: var(--erro-texto);
		padding: var(--r3) var(--r4); border-radius: 8px; margin-bottom: var(--r4);
	}
	.vazio-tab { color: var(--tinta-suave); font-size: 0.9rem; font-style: italic; }
	.grade-produtos { display: flex; flex-direction: column; gap: var(--r3); }
	.card-produto-loja {
		display: flex; gap: var(--r3); padding: var(--r3) var(--r4);
		border: 1px solid var(--linha); border-radius: var(--raio-sm);
		background: var(--branco); align-items: center;
	}
	.prod-thumb { width: 56px; height: 56px; border-radius: 8px; object-fit: cover; flex-shrink: 0; }
	.prod-info { flex: 1; min-width: 0; }
	.prod-info h4 {
		font-size: 0.9rem; margin: 0 0 4px;
		white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
	}
	.prod-dados { display: flex; flex-wrap: wrap; gap: var(--r2); font-size: 0.78rem; color: var(--tinta-suave); }
	.prod-preco { font-weight: 700; color: var(--ouro); }
	.prod-comissao { font-weight: 600; }
	.prod-score { font-size: 0.72rem; color: var(--tinta-suave); margin-top: 2px; }
	.btn-pub-mini {
		border: 1px solid var(--linha); background: var(--porcelana);
		border-radius: 8px; width: 36px; height: 36px;
		display: flex; align-items: center; justify-content: center;
		cursor: pointer; font-size: 1rem; flex-shrink: 0;
	}
	.btn-pub-mini:hover { border-color: var(--rosa); background: color-mix(in srgb, var(--rosa) 8%, white); }
</style>
