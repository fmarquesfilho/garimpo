<script>
	/**
	 * Preview inline da publicação — mostra como ficará no Telegram/WhatsApp.
	 * Sempre exibe foto (se disponível) + legenda + botão fake de compra.
	 */
	let { imagem = '', legenda = '', link = '' } = $props();
</script>

<div class="preview">
	<span class="preview-label">Preview</span>
	<div class="preview-card">
		{#if imagem}
			<img src={imagem} alt="preview" class="preview-img" />
		{/if}
		<div class="preview-body">{@html legenda.replace(/\n/g, '<br>')}</div>
		{#if link}
			<div class="preview-btn"><span>🛒 Comprar</span></div>
		{/if}
	</div>
</div>

<style>
	.preview { margin-bottom: var(--r5); }
	.preview-label { font-size: 0.72rem; font-weight: 600; text-transform: uppercase; color: var(--tinta-suave); letter-spacing: 0.05em; }
	.preview-card { border: 1px solid var(--linha); border-radius: var(--raio); overflow: hidden; background: var(--branco); margin-top: 6px; box-shadow: 0 2px 8px rgba(0,0,0,0.04); }
	.preview-img { width: 100%; max-height: 280px; object-fit: cover; display: block; }
	.preview-body { padding: var(--r4); font-size: 0.92rem; line-height: 1.5; }
	.preview-btn { padding: 0 var(--r4) var(--r4); }
	.preview-btn span { display: inline-block; padding: 8px 18px; background: var(--ouro); color: white; border-radius: 8px; font-size: 0.85rem; font-weight: 600; }
</style>
