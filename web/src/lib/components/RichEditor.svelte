<script>
	/**
	 * RichEditor: editor WYSIWYG baseado no Tiptap (ProseMirror).
	 * Suporta negrito, itálico, links. Preview em tempo real.
	 * @prop content — HTML inicial (bind:content para two-way)
	 * @prop placeholder — texto placeholder
	 * @prop onchange — callback quando o conteúdo muda
	 */
	import { onMount, onDestroy } from 'svelte';
	import { Editor } from '@tiptap/core';
	import StarterKit from '@tiptap/starter-kit';
	import Link from '@tiptap/extension-link';
	import Placeholder from '@tiptap/extension-placeholder';

	let { content = $bindable(''), placeholder = 'Escreva a legenda…', onchange = null } = $props();

	let element = $state(null);
	let editor = $state(null);

	onMount(() => {
		editor = new Editor({
			element: element,
			extensions: [
				StarterKit.configure({ heading: false, codeBlock: false, blockquote: false }),
				Link.configure({ openOnClick: false, HTMLAttributes: { class: 'editor-link' } }),
				Placeholder.configure({ placeholder })
			],
			content: content || '',
			onUpdate: ({ editor: e }) => {
				content = e.getHTML();
				onchange?.(content);
			},
			onTransaction: () => {
				// Force reactivity
				editor = editor;
			}
		});
	});

	onDestroy(() => {
		editor?.destroy();
	});

	// Sync external content changes (e.g. template switch)
	$effect(() => {
		if (editor && content !== editor.getHTML()) {
			editor.commands.setContent(content, false);
		}
	});

	function toggleBold() { editor?.chain().focus().toggleBold().run(); }
	function toggleItalic() { editor?.chain().focus().toggleItalic().run(); }
	function setLink() {
		const url = prompt('URL do link:');
		if (url) {
			editor?.chain().focus().setLink({ href: url }).run();
		}
	}
	function removeLink() { editor?.chain().focus().unsetLink().run(); }
</script>

<div class="rich-editor">
	<div class="toolbar">
		<button type="button" class="tb-btn" class:ativo={editor?.isActive('bold')} onclick={toggleBold} title="Negrito">
			<strong>B</strong>
		</button>
		<button type="button" class="tb-btn" class:ativo={editor?.isActive('italic')} onclick={toggleItalic} title="Itálico">
			<em>I</em>
		</button>
		<button type="button" class="tb-btn" onclick={setLink} title="Inserir link">🔗</button>
		{#if editor?.isActive('link')}
			<button type="button" class="tb-btn" onclick={removeLink} title="Remover link">✕🔗</button>
		{/if}
	</div>
	<div class="editor-content" bind:this={element}></div>
</div>

<style>
	.rich-editor {
		border: 1px solid var(--linha);
		border-radius: 10px;
		overflow: hidden;
		background: white;
	}
	.toolbar {
		display: flex;
		gap: 2px;
		padding: 6px 8px;
		border-bottom: 1px solid var(--linha);
		background: var(--porcelana);
	}
	.tb-btn {
		border: 1px solid transparent;
		background: transparent;
		padding: 4px 10px;
		border-radius: 6px;
		cursor: pointer;
		font-size: 0.85rem;
		color: var(--tinta-suave);
	}
	.tb-btn:hover { background: white; color: var(--tinta); }
	.tb-btn.ativo { background: var(--ouro-fundo); color: var(--ouro-escuro); border-color: var(--ouro); }

	.editor-content {
		min-height: 120px;
		padding: 12px;
		font-size: 0.92rem;
		line-height: 1.6;
	}
	.editor-content :global(.tiptap) {
		outline: none;
		min-height: 100px;
	}
	.editor-content :global(.tiptap p) {
		margin: 0 0 0.5em;
	}
	.editor-content :global(.tiptap p.is-editor-empty:first-child::before) {
		content: attr(data-placeholder);
		color: var(--tinta-suave);
		opacity: 0.5;
		pointer-events: none;
		float: left;
		height: 0;
	}
	.editor-content :global(.editor-link) {
		color: var(--ouro);
		text-decoration: underline;
	}
</style>
