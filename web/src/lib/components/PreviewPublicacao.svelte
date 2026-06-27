<script>
	/**
	 * Preview visual da publicação como ficará no Telegram.
	 * Mostra imagem (se com_foto), legenda formatada e botão fake de compra.
	 */
	let { legenda = '', imagem = '', link = '', previewFoto = false } = $props();
</script>

<div class="preview-col">
	<h2>Preview</h2>
	<div class="preview-card">
		{#if previewFoto && imagem}
			<img src={imagem} alt="preview" class="preview-img" />
		{/if}
		<div class="preview-corpo">{@html legenda.replace(/\n/g, '<br>')}</div>
		{#if link}
			<div class="preview-botao">
				<span class="btn-fake">🛒 Comprar</span>
			</div>
		{/if}
	</div>
	<p class="preview-nota">Como ficará no Telegram</p>
</div>

<style>
	.preview-col h2 { font-size: 1rem; margin: 0 0 var(--r3); color: var(--tinta-suave); }
	.preview-card {
		border: 1px solid var(--linha); border-radius: var(--raio); overflow: hidden;
		background: var(--branco); box-shadow: 0 2px 8px rgba(0,0,0,0.06);
	}
	.preview-img { width: 100%; max-height: 240px; object-fit: cover; }
	.preview-corpo { padding: var(--r4); font-size: 0.92rem; line-height: 1.5; }
	.preview-botao { padding: 0 var(--r4) var(--r4); }
	.btn-fake {
		display: inline-block; padding: 8px 18px; background: var(--ouro);
		color: white; border-radius: 8px; font-size: 0.85rem; font-weight: 600;
	}
	.preview-nota { font-size: 0.75rem; color: var(--tinta-suave); margin-top: var(--r2); font-style: italic; }
</style>
