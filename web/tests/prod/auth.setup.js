/**
 * Auth setup — obtém token Firebase real e salva para os testes.
 *
 * Estratégia: faz signInWithEmailAndPassword via REST API do Firebase (Identity Toolkit),
 * obtém idToken + refreshToken, e salva num arquivo JSON que os testes leem.
 * Cada teste injeta o token via page.evaluate no contexto do browser.
 *
 * Isso bypassa o problema do IndexedDB (Playwright storageState não captura
 * tokens Firebase que ficam em IndexedDB).
 */
import { test as setup } from '@playwright/test';
import { writeFileSync } from 'fs';
import { resolve, dirname } from 'path';
import { fileURLToPath } from 'url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const tokenFile = resolve(__dirname, '../.auth/prod-token.json');

setup('obter token Firebase via REST', async () => {
	const email = process.env.E2E_EMAIL;
	const password = process.env.E2E_PASSWORD;

	if (!email || !password) {
		throw new Error(
			'E2E_EMAIL e E2E_PASSWORD são obrigatórios.\n' + 'Copie .env.e2e para .env.e2e.local e preencha as credenciais.'
		);
	}

	// Firebase Auth REST API (Identity Toolkit v1)
	const apiKey = 'AIzaSyA5sBUoVkNHiq58KUkmwbxIMLhvgTn7N8A';
	const resp = await fetch(`https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword?key=${apiKey}`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ email, password, returnSecureToken: true })
	});

	if (!resp.ok) {
		const err = await resp.json();
		throw new Error(`Firebase login falhou: ${err.error?.message || resp.status}`);
	}

	const data = await resp.json();

	// Salva token para uso nos testes
	writeFileSync(
		tokenFile,
		JSON.stringify(
			{
				idToken: data.idToken,
				refreshToken: data.refreshToken,
				uid: data.localId,
				email: data.email,
				expiresAt: Date.now() + parseInt(data.expiresIn) * 1000
			},
			null,
			2
		)
	);
});
