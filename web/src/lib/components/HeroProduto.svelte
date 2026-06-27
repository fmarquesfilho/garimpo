<script>
	/**
	 * Card hero do produto — mostra imagem + campos editáveis (nome, categoria, preço, link).
	 * Usado na página de publicação.
	 */
	let { produto = $bindable({}) } = $props();
</script>

<div class="hero-produto">
	{#if produto.imagem}
		<img src={produto.imagem} alt={produto.nome} class="hero-img" />
	{:else}
		<div class="hero-placeholder">📦</div>
	{/if}
	<div class="hero-info">
		<input class="nome-edit" bind:value={produto.nome} placeholder="Nome do produto" />
		<div class="hero-meta">
			<input class="campo-mini" bind:value={produto.categoria} placeholder="Categoria" />
			<input class="campo-mini preco" type="number" step="0.01" bind:value={produto.preco} placeholder="Preço" />
		</div>
		{#if produto.link}
			<a class="link-prod" href={produto.link} target="_blank" rel="noopener">{produto.link.substring(0, 60)}…</a>
		{/if}
	</div>
</div>

<style>
	.hero-produto { display: flex; gap: var(--r4); padding: var(--r4); border: 1px solid var(--linha); border-radius: var(--raio); background: var(--nevoa); margin-bottom: var(--r5); }
	.hero-img { width: 100px; height: 100px; object-fit: cover; border-radius: 10px; flex-shrink: 0; }
	.hero-placeholder { width: 100px; height: 100px; border-radius: 10px; background: var(--porcelana); display: flex; align-items: center; justify-content: center; font-size: 2.5rem; flex-shrink: 0; }
	.hero-info { flex: 1; display: flex; flex-direction: column; gap: var(--r2); min-width: 0; }
	.nome-edit { font-size: 1rem; font-weight: 700; border: 1px solid var(--linha); background: var(--branco); border-radius: 8px; width: 100%; padding: 8px 12px; }
	.nome-edit:focus { outline: 2px solid var(--ouro); outline-offset: 1px; }
	.hero-meta { display: flex; gap: var(--r3); flex-wrap: wrap; }
	.campo-mini { font-size: 0.85rem; padding: 6px 10px; border: 1px solid var(--linha); border-radius: 8px; background: var(--porcelana); width: 120px; }
	.campo-mini.preco { width: 90px; font-weight: 600; }
	.link-prod { font-size: 0.72rem; color: var(--tinta-suave); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; text-decoration: none; }
	.link-prod:hover { color: var(--ouro); }

	@media (max-width: 420px) {
		.hero-produto { flex-direction: column; align-items: center; text-align: center; }
		.hero-img { width: 100%; height: 180px; border-radius: var(--raio); }
	}
</style>
