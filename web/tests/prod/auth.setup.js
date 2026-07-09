/**
 * Auth setup — faz login real via Firebase (email/senha) e salva o estado.
 *
 * Roda uma vez antes de todos os testes de produção. Os testes subsequentes
 * reutilizam o storageState salvo (cookies + localStorage com token Firebase).
 *
 * Credenciais vêm de .env.e2e.local (gitignored).
 */
import { test as setup, expect } from '@playwright/test';

const authFile = 'tests/.auth/prod-user.json';

setup('login via Firebase (email/senha)', async ({ page }) => {
	const email = process.env.E2E_EMAIL;
	const password = process.env.E2E_PASSWORD;
	const baseURL = process.env.E2E_BASE_URL || 'https://garimpei.app.br';

	if (!email || !password) {
		throw new Error(
			'E2E_EMAIL e E2E_PASSWORD são obrigatórios.\n' +
				'Copie .env.e2e para .env.e2e.local e preencha as credenciais.\n' +
				'Criar usuário de teste: Firebase Console → Authentication → Add user'
		);
	}

	await page.goto(baseURL);

	// Aguarda a página carregar e o Firebase SDK inicializar
	await page.waitForLoadState('networkidle');

	// Faz login via Firebase signInWithEmailAndPassword (executado no contexto do browser)
	const loginResult = await page.evaluate(
		async ({ email, password }) => {
			// Aguarda o Firebase Auth estar disponível
			const { initializeApp } = await import('https://www.gstatic.com/firebasejs/11.8.1/firebase-app.js');
			const { getAuth, signInWithEmailAndPassword } =
				await import('https://www.gstatic.com/firebasejs/11.8.1/firebase-auth.js');

			const app = initializeApp({
				apiKey: 'AIzaSyA5sBUoVkNHiq58KUkmwbxIMLhvgTn7N8A',
				authDomain: 'garimpo-500114.firebaseapp.com',
				projectId: 'garimpo-500114'
			});
			const auth = getAuth(app);
			const cred = await signInWithEmailAndPassword(auth, email, password);
			const token = await cred.user.getIdToken();
			return { uid: cred.user.uid, email: cred.user.email, token };
		},
		{ email, password }
	);

	// Verifica que o login funcionou
	expect(loginResult.uid).toBeTruthy();
	expect(loginResult.token).toBeTruthy();

	// Recarrega a página para que o app reconheça o usuário logado
	await page.reload();
	await page.waitForLoadState('networkidle');

	// Verifica que a UI mostra conteúdo autenticado (sem botão de login)
	await expect(page.getByPlaceholder(/Buscar produto/i)).toBeVisible({ timeout: 15000 });

	// Salva o estado do browser (localStorage com tokens Firebase + cookies)
	await page.context().storageState({ path: authFile });
});
