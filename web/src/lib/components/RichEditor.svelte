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
	import { Tooltip } from '$lib/components/ui';

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

	function toggleBold() {
		editor?.chain().focus().toggleBold().run();
	}
	function toggleItalic() {
		editor?.chain().focus().toggleItalic().run();
	}
	function setLink() {
		const url = prompt('URL do link:');
		if (url) {
			editor?.chain().focus().setLink({ href: url }).run();
		}
	}
	function removeLink() {
		editor?.chain().focus().unsetLink().run();
	}
</script>

<div class="overflow-hidden rounded-[10px] border border-border bg-white">
	<div class="flex gap-0.5 border-b border-border bg-porcelana px-2 py-1.5">
		<Tooltip content="Negrito">
			<button
				type="button"
				class="rounded-md border border-transparent bg-transparent px-2.5 py-1 text-sm text-tinta-suave hover:bg-white hover:text-foreground"
				class:!bg-ouro-fundo={editor?.isActive('bold')}
				class:!text-ouro-escuro={editor?.isActive('bold')}
				class:!border-ouro={editor?.isActive('bold')}
				onclick={toggleBold}
			>
				<strong>B</strong>
			</button>
		</Tooltip>
		<Tooltip content="Itálico">
			<button
				type="button"
				class="rounded-md border border-transparent bg-transparent px-2.5 py-1 text-sm text-tinta-suave hover:bg-white hover:text-foreground"
				class:!bg-ouro-fundo={editor?.isActive('italic')}
				class:!text-ouro-escuro={editor?.isActive('italic')}
				class:!border-ouro={editor?.isActive('italic')}
				onclick={toggleItalic}
			>
				<em>I</em>
			</button>
		</Tooltip>
		<Tooltip content="Inserir link">
			<button
				type="button"
				class="rounded-md border border-transparent bg-transparent px-2.5 py-1 text-sm text-tinta-suave hover:bg-white hover:text-foreground"
				onclick={setLink}>🔗</button
			>
		</Tooltip>
		{#if editor?.isActive('link')}
			<Tooltip content="Remover link">
				<button
					type="button"
					class="rounded-md border border-transparent bg-transparent px-2.5 py-1 text-sm text-tinta-suave hover:bg-white hover:text-foreground"
					onclick={removeLink}>✕🔗</button
				>
			</Tooltip>
		{/if}
	</div>
	<div class="min-h-[120px] p-3 text-[0.92rem] leading-relaxed" bind:this={element}></div>
</div>

<style>
	:global(.tiptap) {
		outline: none;
		min-height: 100px;
	}
	:global(.tiptap p) {
		margin: 0 0 0.5em;
	}
	:global(.tiptap p.is-editor-empty:first-child::before) {
		content: attr(data-placeholder);
		color: var(--tinta-suave);
		opacity: 0.5;
		pointer-events: none;
		float: left;
		height: 0;
	}
	:global(.editor-link) {
		color: var(--ouro);
		text-decoration: underline;
	}
</style>
