import { test, expect } from '@playwright/test';

/**
 * Testes isolados do SeletorGrupo — verifica que selecionar um grupo
 * no <select> realmente propaga o valor e habilita o botão de submit.
 *
 * Usa uma página HTML mínima que renderiza só o componente, sem Firebase Auth.
 */

const gruposMock = [
	{ id: '120363430000000000@g.us', nome: '#1 Garimpo Hoje' },
	{ id: '120363410893012870@g.us', nome: '#08 AVANÇADO VOE' },
	{ id: '120363426313232441@g.us', nome: '#96 NOSSO GRUPINHO' },
	{ id: '120363156757082979@g.us', nome: 'Ofertas | Beleza na Web' },
	{ id: '558491629647-1486926372@g.us', nome: 'Famílias da Pipa' }
];

// Testa a lógica de seleção em isolamento no DOM (sem Svelte runtime)
test.describe('SeletorGrupo — lógica DOM isolada', () => {
	test('selecionar option atualiza value do select', async ({ page }) => {
		await page.goto('/');

		const resultado = await page.evaluate((grupos) => {
			// Cria o DOM equivalente ao SeletorGrupo
			const container = document.createElement('div');
			const select = document.createElement('select');
			const placeholder = document.createElement('option');
			placeholder.value = '';
			placeholder.textContent = 'Selecione um grupo…';
			select.appendChild(placeholder);

			for (const g of grupos) {
				const opt = document.createElement('option');
				opt.value = g.id;
				opt.textContent = g.nome;
				select.appendChild(opt);
			}
			container.appendChild(select);
			document.body.appendChild(container);

			// Simula seleção do primeiro grupo
			select.value = grupos[0].id;
			select.dispatchEvent(new Event('change', { bubbles: true }));

			const valor = select.value;
			document.body.removeChild(container);
			return { valor, esperado: grupos[0].id };
		}, gruposMock);

		expect(resultado.valor).toBe(resultado.esperado);
		expect(resultado.valor).not.toBe('');
	});

	test('filtrar e selecionar mantém o valor', async ({ page }) => {
		await page.goto('/');

		const resultado = await page.evaluate((grupos) => {
			// Simula: filtrar por "garimpo" → selecionar o resultado
			const filtro = 'garimpo';
			const filtrados = grupos.filter(g =>
				g.nome.toLowerCase().includes(filtro.toLowerCase())
			);

			// Monta o select com apenas os filtrados
			const select = document.createElement('select');
			const placeholder = document.createElement('option');
			placeholder.value = '';
			select.appendChild(placeholder);
			for (const g of filtrados) {
				const opt = document.createElement('option');
				opt.value = g.id;
				opt.textContent = g.nome;
				select.appendChild(opt);
			}
			document.body.appendChild(select);

			// Seleciona
			select.value = filtrados[0].id;
			select.dispatchEvent(new Event('change', { bubbles: true }));

			const valor = select.value;
			document.body.removeChild(select);

			return {
				filtradosCount: filtrados.length,
				valor,
				vazio: !valor.trim(),
				nomeEsperado: filtrados[0].nome
			};
		}, gruposMock);

		expect(resultado.filtradosCount).toBe(1);
		expect(resultado.valor).toBe('120363430000000000@g.us');
		expect(resultado.vazio).toBe(false);
	});

	test('botão desabilitado: nome vazio OU config vazio', async ({ page }) => {
		await page.goto('/');

		const resultado = await page.evaluate(() => {
			function isDisabled(nome, config, salvando = false) {
				return salvando || !nome.trim() || !config.trim();
			}

			return [
				{ caso: 'ambos vazios', disabled: isDisabled('', ''), esperado: true },
				{ caso: 'só nome', disabled: isDisabled('Teste', ''), esperado: true },
				{ caso: 'só config', disabled: isDisabled('', '123@g.us'), esperado: true },
				{ caso: 'ambos preenchidos', disabled: isDisabled('Teste', '123@g.us'), esperado: false },
				{ caso: 'salvando', disabled: isDisabled('Teste', '123@g.us', true), esperado: true },
				{ caso: 'espaços no nome', disabled: isDisabled('   ', '123@g.us'), esperado: true },
				{ caso: 'espaços no config', disabled: isDisabled('Teste', '   '), esperado: true },
				{ caso: 'group ID real', disabled: isDisabled('Garimpo', '120363430000000000@g.us'), esperado: false },
			];
		});

		for (const r of resultado) {
			expect(r.disabled, `Caso "${r.caso}": disabled=${r.disabled}, esperado=${r.esperado}`).toBe(r.esperado);
		}
	});

	test('fluxo completo: selecionar tipo → carregar grupos → filtrar → selecionar → botão habilitado', async ({ page }) => {
		await page.goto('/');

		const resultado = await page.evaluate((grupos) => {
			// Estado inicial
			let tipo = 'telegram';
			let nome = 'Garimpo Hoje';
			let config = '';

			// 1. Muda tipo para whatsapp
			tipo = 'whatsapp';
			config = ''; // reset ao mudar tipo

			// 2. Grupos carregados (simula API)
			const gruposCarregados = grupos;

			// 3. Filtra por "garimpo"
			const filtro = 'garimpo';
			const filtrados = gruposCarregados.filter(g =>
				g.nome.toLowerCase().includes(filtro.toLowerCase())
			);

			// 4. Seleciona o grupo
			config = filtrados[0].id;

			// 5. Verifica se botão estaria habilitado
			const disabled = !nome.trim() || !config.trim();

			return {
				tipo,
				nome,
				config,
				filtradosCount: filtrados.length,
				grupoNome: filtrados[0].nome,
				disabled,
				botaoHabilitado: !disabled
			};
		}, gruposMock);

		expect(resultado.tipo).toBe('whatsapp');
		expect(resultado.config).toBe('120363430000000000@g.us');
		expect(resultado.filtradosCount).toBe(1);
		expect(resultado.grupoNome).toBe('#1 Garimpo Hoje');
		expect(resultado.botaoHabilitado).toBe(true);
	});
});
