<script>
	import { buscarCandidatos, compararEstrategias } from '$lib/api.js';
	import { quadro } from '$lib/board.js';
	import CandidateCard from '$lib/components/CandidateCard.svelte';
	import StrategyToggle from '$lib/components/StrategyToggle.svelte';

	let modo = $state('nicho'); // nicho | diversificada | comparar
	let busca = $state(''); // keyword da Shopee
	let categoria = $state('cosméticos'); // rótulo carimbado (para o nicho)
	let pisoComissao = $state(0.07);
	let quantos = $state(9);
	let vendasMin = $state(5); // piso de credibilidade (filtra produto-fantasma)
	let notaMinima = $state(0);

	let carregando = $state(true);
	let erro = $state(null);
	let lista = $state([]); // para nicho/diversificada
	let pares = $state(null); // { nicho:[], diversificada:[] } para comparar
	let fonteAtiva = $state('');

	async function carregar() {
		carregando = true;
		erro = null;
		pares = null;
		const filtros = {
			keyword: busca.trim(),
			categoria,
			comissaoMin: pisoComissao,
			vendasMin,
			notaMin: notaMinima
		};
		try {
			if (modo === 'comparar') {
				const r = await compararEstrategias({ top: 6, ...filtros });
				pares = r;
				fonteAtiva = r.fonte ?? '';
			} else {
				const r = await buscarCandidatos({ estrategia: modo, top: quantos, ...filtros });
				lista = (r.candidatos ?? []).map((c) => ({ ...c, estrategia: modo }));
				fonteAtiva = r.fonte ?? '';
			}
		} catch (e) {
			erro = e.message;
		} finally {
			carregando = false;
		}
	}

	// recarrega quando muda qualquer controle, com debounce — busca e número de
	// vendas mudam a cada tecla, e cada keyword nova é uma chamada real à Shopee.
	let timer;
	$effect(() => {
		// dependências rastreadas:
		modo;
		busca;
		categoria;
		pisoComissao;
		quantos;
		vendasMin;
		notaMinima;
		clearTimeout(timer);
		timer = setTimeout(carregar, 350);
		return () => clearTimeout(timer);
	});

	function selecionar(c) {
		quadro.selecionar(c);
	}
</script>

<section class="intro">
	<p class="rotulo">a peneira do dia</p>
	<h1>O que vale a pena garimpar hoje</h1>
	<p class="sub">
		Produtos elegíveis (comissão no piso ou acima), ordenados pelo <strong>teor</strong> — o quanto
		cada um rende pelo esforço. Escolha a estratégia e mande os melhores pro quadro.
	</p>
</section>

<div class="controles">
	<StrategyToggle bind:valor={modo} />
</div>

<div class="filtros">
	<label class="campo busca">
		<span class="rotulo">buscar na shopee</span>
		<input type="search" bind:value={busca} placeholder="perfume, sérum, batom…" class="entrada" />
	</label>
	<label class="campo">
		<span class="rotulo">categoria</span>
		<select bind:value={categoria} class="dado">
			<option value="cosméticos">cosméticos</option>
			<option value="perfumaria">perfumaria</option>
			<option value="bem-estar">bem-estar</option>
			<option value="">sem rótulo</option>
		</select>
	</label>
	<label class="campo">
		<span class="rotulo">comissão mín.</span>
		<select bind:value={pisoComissao} class="dado">
			<option value={0.05}>5%</option>
			<option value={0.07}>7%</option>
			<option value={0.1}>10%</option>
			<option value={0.15}>15%</option>
		</select>
	</label>
	<label class="campo">
		<span class="rotulo">vendas mín.</span>
		<input type="number" min="0" step="1" bind:value={vendasMin} class="entrada num" />
	</label>
	<label class="campo">
		<span class="rotulo">nota mín.</span>
		<select bind:value={notaMinima} class="dado">
			<option value={0}>todas</option>
			<option value={4}>4,0+</option>
			<option value={4.5}>4,5+</option>
		</select>
	</label>
	{#if modo !== 'comparar'}
		<label class="campo">
			<span class="rotulo">quantos</span>
			<select bind:value={quantos} class="dado">
				<option value={6}>6</option>
				<option value={9}>9</option>
				<option value={12}>12</option>
			</select>
		</label>
	{/if}
</div>

<p class="nota-curadoria">
	Comissão alta com zero venda costuma ser produto-fantasma. O piso de vendas e a nota mínima
	deixam na peneira só o que já tem tração.{#if fonteAtiva}<span class="fonte dado"> · fonte: {fonteAtiva}</span>{/if}
</p>

{#if carregando}
	<p class="aviso">Garimpando os melhores produtos…</p>
{:else if erro}
	<div class="erro">
		<p><strong>Não consegui falar com a API.</strong></p>
		<p>{erro}</p>
		<p class="dica dado">Confira se o servidor está rodando: <code>go run ./cmd/garimpo-api</code> (porta 8080).</p>
	</div>
{:else if modo === 'comparar' && pares}
	<div class="comparacao">
		<div class="coluna">
			<h2 class="tit-col rosa">Nicho</h2>
			<div class="empilhado">
				{#each pares.nicho as c, i (c.id)}
					<CandidateCard candidato={{ ...c, estrategia: 'nicho' }} posicao={i + 1} onselecionar={selecionar} />
				{/each}
			</div>
		</div>
		<div class="coluna">
			<h2 class="tit-col ardosia">Diversificada</h2>
			<div class="empilhado">
				{#each pares.diversificada as c, i (c.id)}
					<CandidateCard candidato={{ ...c, estrategia: 'diversificada' }} posicao={i + 1} onselecionar={selecionar} />
				{/each}
			</div>
		</div>
	</div>
{:else if lista.length === 0}
	<div class="vazio">
		{#if busca.trim() === ''}
			<p>Comece por uma busca.</p>
			<p class="dica">Digite o que quer divulgar — perfume, sérum, batom — pra peneirar a Shopee.</p>
		{:else}
			<p>Nada na peneira para “{busca}”.</p>
			<p class="dica">Tente outro termo, ou afrouxe os pisos de comissão, vendas e nota.</p>
		{/if}
	</div>
{:else}
	<div class="grade">
		{#each lista as c, i (c.id)}
			<CandidateCard candidato={c} posicao={i + 1} destaque={i === 0} onselecionar={selecionar} />
		{/each}
	</div>
{/if}

<style>
	.intro {
		max-width: 40rem;
		margin-bottom: var(--r8);
	}
	h1 {
		font-size: clamp(2rem, 6vw, 3.2rem);
		margin: var(--r2) 0 var(--r4);
	}
	.sub {
		color: var(--tinta-suave);
		font-size: 1.05rem;
		margin: 0;
	}
	.controles {
		display: flex;
		flex-wrap: wrap;
		align-items: flex-end;
		gap: var(--r4);
		margin-bottom: var(--r4);
	}
	.filtros {
		display: flex;
		flex-wrap: wrap;
		align-items: flex-end;
		gap: var(--r4);
		margin-bottom: var(--r4);
		padding: var(--r4);
		background: var(--nevoa);
		border: 1px solid var(--linha);
		border-radius: var(--raio);
	}
	.campo {
		display: flex;
		flex-direction: column;
		gap: 5px;
	}
	.campo.busca {
		flex: 1 1 220px;
	}
	.entrada {
		font-family: var(--ui);
		font-size: 0.95rem;
		padding: 9px 12px;
		border-radius: 10px;
		border: 1px solid var(--linha);
		background: var(--porcelana);
		color: var(--tinta);
		width: 100%;
	}
	.entrada::placeholder {
		color: var(--tinta-suave);
		opacity: 0.7;
	}
	.entrada.num {
		font-family: var(--mono);
		width: 5.5rem;
	}
	.fonte {
		opacity: 0.8;
	}
	.nota-curadoria {
		margin: 0 0 var(--r8);
		font-size: 0.85rem;
		color: var(--tinta-suave);
		max-width: 60ch;
		border-left: 2px solid var(--ouro-claro);
		padding-left: var(--r3);
	}
	select {
		font-family: var(--mono);
		font-size: 0.9rem;
		padding: 9px 12px;
		border-radius: 10px;
		border: 1px solid var(--linha);
		background: var(--porcelana);
		color: var(--tinta);
	}
	.grade {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
		gap: var(--r6);
	}
	.grade :global(.destaque) {
		grid-column: 1 / -1;
	}
	@media (min-width: 720px) {
		.grade :global(.destaque) {
			grid-column: span 2;
		}
	}
	.comparacao {
		display: grid;
		grid-template-columns: 1fr;
		gap: var(--r8);
	}
	@media (min-width: 800px) {
		.comparacao {
			grid-template-columns: 1fr 1fr;
		}
	}
	.empilhado {
		display: flex;
		flex-direction: column;
		gap: var(--r4);
	}
	.tit-col {
		font-size: 1.3rem;
		margin-bottom: var(--r4);
		padding-bottom: var(--r2);
		border-bottom: 2px solid;
	}
	.tit-col.rosa {
		color: var(--rosa);
	}
	.tit-col.ardosia {
		color: var(--ardosia);
	}
	.aviso {
		color: var(--tinta-suave);
		font-style: italic;
	}
	.vazio,
	.erro {
		background: var(--nevoa);
		border: 1px solid var(--linha);
		border-radius: var(--raio);
		padding: var(--r8);
		text-align: center;
	}
	.erro {
		border-color: color-mix(in srgb, var(--alerta) 30%, var(--linha));
	}
	.erro p {
		margin: var(--r2) 0;
	}
	.dica {
		color: var(--tinta-suave);
		font-size: 0.85rem;
	}
	code {
		background: var(--ouro-fundo);
		padding: 2px 6px;
		border-radius: 6px;
	}
</style>
