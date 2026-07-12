<script>
	/**
	 * Omnibox — input unificado da página Descobrir. Substitui os inputs separados
	 * de keyword/loja/categoria por um único campo que infere o tipo pelo conteúdo
	 * e aceita prefixos opcionais (`@loja`, `#categoria`, `!marketplace`).
	 *
	 * O texto é literal (sem chips no campo). Ao selecionar loja/categoria/marketplace,
	 * o token ativo é removido do input e a seleção vira um card abaixo (via engine) —
	 * mesmo padrão do ui/Combobox.svelte. Keyword dispara busca via evento DIGITAR
	 * (debounce 400ms na própria engine).
	 *
	 * Headless: toda lógica de negócio mora na BuscaEngine; aqui só parser + sugestões
	 * (funções puras) e a interação de teclado/ARIA.
	 */
	import { parsearInput, serializarTokens, tokensParaContexto } from '$lib/omnibox-parser.js';
	import { gerarSugestoes } from '$lib/omnibox-sugestoes.js';
	import { OMNIBOX, MARKETPLACES } from '$lib/busca-config.js';

	let {
		engine,
		lojasMonitoradas = [],
		placeholder = 'Buscar produtos, lojas ou categorias… (ex: sérum @loja #beleza)'
	} = $props();

	const CFG = OMNIBOX ?? { minChars: 2, maxSugestoes: 7, matchBuscaSalva: true };
	const GRUPO_LABEL = {
		busca_salva: 'Buscas salvas',
		loja: 'Lojas',
		categoria: 'Categorias',
		marketplace: 'Marketplaces'
	};

	let inputValue = $state('');
	let aberto = $state(false);
	let highlightIdx = $state(-1);
	let inputEl;

	// ── Derivados (parser + sugestões, tudo puro) ────────────────────────────
	/** @type {import('$lib/omnibox-parser.js').Token} */
	const TOKEN_VAZIO = { tipo: 'keyword', valor: '', completo: false };
	let tokens = $derived(parsearInput(inputValue));
	let ultimoToken = $derived(tokens[tokens.length - 1] ?? TOKEN_VAZIO);

	let sugestoesCtx = $derived({
		lojasMonitoradas,
		categoriasDisponiveis: engine.ctx.categoriasDisponiveis,
		marketplaces: MARKETPLACES?.suportados ?? [],
		buscasSalvas: engine.ctx.buscasSalvas
	});
	let sugestoesMap = $derived(gerarSugestoes(ultimoToken, sugestoesCtx, CFG));

	// Grupos com offset para mapear índice global (flat) → item renderizado.
	let grupos = $derived.by(() => {
		let offset = 0;
		const out = [];
		for (const [tipo, itens] of sugestoesMap) {
			out.push({ tipo, label: GRUPO_LABEL[tipo] ?? tipo, itens, offset });
			offset += itens.length;
		}
		return out;
	});
	let flat = $derived(grupos.flatMap((g) => g.itens));
	let mostrarDropdown = $derived(aberto && flat.length > 0);

	// ── Interação ────────────────────────────────────────────────────────────
	function onInput(e) {
		inputValue = e.currentTarget.value;
		highlightIdx = -1;
		aberto = true;
		// Só a parte keyword vai para a engine (debounce interno de 400ms).
		const kw = parsearInput(inputValue)
			.filter((t) => t.tipo === 'keyword')
			.map((t) => t.valor)
			.join(' ');
		engine.send({ type: 'DIGITAR', value: kw });
	}

	function fechar() {
		aberto = false;
		highlightIdx = -1;
	}

	/** Remove o token ativo (incompleto) do input, preservando os anteriores. */
	function removerTokenAtivo() {
		const restantes = parsearInput(inputValue);
		restantes.pop();
		inputValue = serializarTokens(restantes);
	}

	function selecionar(sug) {
		if (!sug) return;
		switch (sug.tipo) {
			case 'loja':
				engine.send({ type: 'ADICIONAR_LOJA', loja: sug.meta });
				removerTokenAtivo();
				break;
			case 'categoria':
				engine.send({ type: 'ADICIONAR_CATEGORIA', nome: sug.label, categoria: sug.meta });
				removerTokenAtivo();
				break;
			case 'marketplace': {
				const mkts = [...new Set([...engine.ctx.marketplacesFiltro, sug.meta.marketplace])];
				engine.send({ type: 'MUDAR_MARKETPLACES', marketplaces: mkts });
				removerTokenAtivo();
				break;
			}
			case 'busca_salva':
				engine.send({ type: 'CARREGAR_SALVA', config: sug.meta.config });
				inputValue = (sug.meta.config.keywords ?? [])[0] ?? '';
				break;
		}
		fechar();
		inputEl?.focus();
	}

	/** Enter sem sugestão destacada: resolve tokens digitados e executa a busca. */
	function executarBusca() {
		const ctx = tokensParaContexto(tokens, {
			lojasMonitoradas,
			categoriasDisponiveis: engine.ctx.categoriasDisponiveis,
			marketplaces: MARKETPLACES?.suportados ?? []
		});
		for (const loja of ctx.lojasResolvidas) {
			if (!engine.ctx.shopIds.includes(loja.id)) engine.send({ type: 'ADICIONAR_LOJA', loja });
		}
		for (const nome of ctx.categorias) {
			if (!engine.ctx.categorias.includes(nome)) engine.send({ type: 'ADICIONAR_CATEGORIA', nome });
		}
		if (ctx.marketplacesFiltro.length) {
			const merged = [...new Set([...engine.ctx.marketplacesFiltro, ...ctx.marketplacesFiltro])];
			engine.send({ type: 'MUDAR_MARKETPLACES', marketplaces: merged });
		}
		engine.send({ type: 'DIGITAR', value: ctx.keyword });
		fechar();
	}

	function onKeydown(e) {
		if (!aberto && (e.key === 'ArrowDown' || e.key === 'ArrowUp')) {
			aberto = true;
			return;
		}
		const n = flat.length;
		if (e.key === 'ArrowDown') {
			e.preventDefault();
			highlightIdx = n ? (highlightIdx + 1) % n : -1;
		} else if (e.key === 'ArrowUp') {
			e.preventDefault();
			highlightIdx = n ? (highlightIdx - 1 + n) % n : -1;
		} else if (e.key === 'Enter') {
			e.preventDefault();
			if (mostrarDropdown && highlightIdx >= 0) selecionar(flat[highlightIdx]);
			else executarBusca();
		} else if (e.key === 'Escape') {
			e.preventDefault();
			fechar();
		}
	}

	function onFocusout(e) {
		// Fecha ao sair do componente (não ao mover foco internamente).
		if (!e.currentTarget.contains(e.relatedTarget)) fechar();
	}
</script>

<div class="relative" onfocusout={onFocusout}>
	<span class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 opacity-50">🔍</span>
	<input
		bind:this={inputEl}
		type="text"
		value={inputValue}
		{placeholder}
		autocomplete="off"
		spellcheck="false"
		role="combobox"
		aria-expanded={mostrarDropdown}
		aria-controls="omnibox-listbox"
		aria-autocomplete="list"
		aria-activedescendant={highlightIdx >= 0 ? `omnibox-opt-${highlightIdx}` : undefined}
		class="w-full rounded-sm border border-input bg-background py-2.5 pl-9 pr-4 text-base text-foreground placeholder:text-muted-foreground focus-visible:border-ring focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/20"
		oninput={onInput}
		onfocus={() => (aberto = true)}
		onkeydown={onKeydown}
	/>

	<!-- Contagem para leitores de tela (Req 10.2) -->
	<span class="sr-only" aria-live="polite">
		{mostrarDropdown ? `${flat.length} ${flat.length === 1 ? 'sugestão' : 'sugestões'}` : ''}
	</span>

	{#if mostrarDropdown}
		<ul
			id="omnibox-listbox"
			role="listbox"
			aria-label="Sugestões de busca"
			class="absolute left-0 right-0 top-[calc(100%+4px)] z-50 max-h-80 overflow-y-auto rounded-md border border-border bg-popover p-1 shadow-md"
		>
			{#each grupos as grupo (grupo.tipo)}
				<li role="group" aria-label={grupo.label}>
					<div class="px-3 pb-1 pt-2 font-[var(--mono)] text-[0.6rem] uppercase tracking-wider text-muted-foreground">
						{grupo.label}
					</div>
					<ul>
						{#each grupo.itens as sug, i (grupo.tipo + ':' + sug.valor + i)}
							{@const idx = grupo.offset + i}
							<li id={`omnibox-opt-${idx}`} role="option" aria-selected={highlightIdx === idx}>
								<button
									type="button"
									tabindex="-1"
									class="flex w-full items-center gap-2 rounded-sm px-3 py-2 text-left text-sm transition-colors {highlightIdx ===
									idx
										? 'bg-accent text-accent-foreground'
										: 'hover:bg-accent'}"
									onmouseenter={() => (highlightIdx = idx)}
									onclick={() => selecionar(sug)}
								>
									<span aria-hidden="true">{sug.icone}</span>
									<span class="truncate font-semibold">{sug.label}</span>
								</button>
							</li>
						{/each}
					</ul>
				</li>
			{/each}
		</ul>
	{/if}
</div>
