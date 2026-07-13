/**
 * Teste arquitetural: valida que todos os componentes renderizados pela
 * pagina Descobrir (BuscaUnificada e descendentes) sao pure-renderers —
 * zero $state local. Qualquer estado deve viver na BuscaEngine.
 *
 * Se este teste falha, significa que alguem adicionou $state a um componente
 * da pagina Descobrir sem passar pela engine. Corrija removendo o $state e
 * delegando para engine.send(). Ver ADR-0033.
 */
import { describe, it, expect } from 'vitest';
import { readFileSync, existsSync } from 'fs';
import { resolve } from 'path';

const COMPONENTS_DIR = resolve(import.meta.dirname, '../lib/components');

// Componentes da pagina Descobrir que devem ser pure-renderers.
// BuscaUnificada importa estes diretamente + Omnibox importa Badge.
const PURE_RENDERER_COMPONENTS = [
	'Omnibox.svelte',
	'StoreCard.svelte',
	'BuscaUnificada.svelte',
	'MarketplaceFilter.svelte',
	'BuscasSalvasPanel.svelte'
];

// Componentes que legitimamente usam $state sao listados aqui para documentacao.
// (engine vive em .svelte.js, ui/ primitives tem outra regra)

describe('Arquitetura — Pure Renderers (ADR-0033)', () => {
	for (const component of PURE_RENDERER_COMPONENTS) {
		it(`${component} nao contem $state local`, () => {
			const filePath = resolve(COMPONENTS_DIR, component);
			if (!existsSync(filePath)) {
				// Componente pode ter sido movido/renomeado — skip com aviso
				return;
			}

			const content = readFileSync(filePath, 'utf-8');

			// Extrai apenas o bloco <script> para evitar falsos positivos no template
			const scriptMatch = content.match(/<script[^>]*>([\s\S]*?)<\/script>/);
			if (!scriptMatch) return; // sem script = puro por definicao

			const script = scriptMatch[1];

			// Procura por $state( — indica estado local
			const stateUsages = script.match(/\$state\s*\(/g);

			expect(
				stateUsages,
				`${component} contem $state() — proibido em pure-renderer.\n` +
					'Todo estado deve viver na BuscaEngine.\n' +
					'Use engine.send() para mutar e engine.* para ler.\n' +
					'Ref: docs/decisoes/0033-headless-ui-controller-omnibox.md'
			).toBeNull();
		});
	}

	it('lista de pure-renderers esta atualizada (BuscaUnificada imports)', () => {
		const buscaUnificada = readFileSync(resolve(COMPONENTS_DIR, 'BuscaUnificada.svelte'), 'utf-8');

		// Extrai imports de componentes locais (./ ou $lib/components/)
		const importRegex = /import\s+(\w+)\s+from\s+['"]\.\/([\w]+)\.svelte['"]/g;
		const imports = [];
		let match;
		while ((match = importRegex.exec(buscaUnificada)) !== null) {
			imports.push(match[2] + '.svelte');
		}

		// Verifica que todo componente importado esta na lista de pure-renderers
		// (exceto ui/ primitives que tem outra regra)
		for (const imp of imports) {
			if (imp === 'BuscaCard.svelte') continue; // renderizado pelo Panel, nao diretamente
			expect(
				PURE_RENDERER_COMPONENTS,
				`${imp} e importado por BuscaUnificada mas NAO esta na lista PURE_RENDERER_COMPONENTS.\n` +
					'Adicione-o a lista se for pure-renderer, ou justifique o $state.'
			).toContain(imp);
		}
	});
});
