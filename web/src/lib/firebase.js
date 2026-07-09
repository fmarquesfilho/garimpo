// Configuração do Firebase para autenticação.
// Os valores de config são públicos (não são secrets) — são necessários para o
// SDK identificar o projeto no client. A segurança está na validação do token
// no servidor (Firebase Admin SDK / verificação de ID token).
import { initializeApp } from 'firebase/app';
import {
	getAuth,
	GoogleAuthProvider,
	signInWithPopup,
	signOut,
	onAuthStateChanged,
	connectAuthEmulator,
	signInWithEmailAndPassword
} from 'firebase/auth';
import { writable } from 'svelte/store';
import { browser } from '$app/environment';

// Config do projeto Firebase — preencha com os valores do console Firebase:
// https://console.firebase.google.com/project/garimpo-500114/settings/general
const firebaseConfig = {
	apiKey: 'AIzaSyA5sBUoVkNHiq58KUkmwbxIMLhvgTn7N8A',
	authDomain: 'garimpo-500114.firebaseapp.com',
	projectId: 'garimpo-500114'
};

let app;
let auth;

/**
 * Modo de teste E2E local (sem Firebase, sem emulador): o harness Playwright
 * injeta `window.__E2E_AUTH_USER__` via addInitScript ANTES do boot. Nunca é
 * definido em produção — o app real segue pelo caminho normal do Firebase.
 * Permite rodar os E2E localmente antes do push.
 */
const testeUser = browser && window.__E2E_AUTH_USER__ ? window.__E2E_AUTH_USER__ : null;

if (browser && !testeUser) {
	app = initializeApp(firebaseConfig);
	auth = getAuth(app);

	// Conectar ao Auth Emulator em ambiente de teste (E2E)
	if (window.__FIREBASE_AUTH_EMULATOR_HOST__) {
		connectAuthEmulator(auth, `http://${window.__FIREBASE_AUTH_EMULATOR_HOST__}`, { disableWarnings: true });
	}

	// Expor auth para testes E2E (permite signIn via page.evaluate)
	if (window.__FIREBASE_AUTH_EMULATOR_HOST__) {
		window.__TEST_AUTH__ = auth;
		window.__TEST_SIGN_IN__ = (email, password) => signInWithEmailAndPassword(auth, email, password);
	}
}

// Store reativo do usuário logado (ou null)
function criarUserStore() {
	// No modo de teste, começa já autenticado com a conta de teste injetada.
	const { subscribe, set } = writable(testeUser);

	if (browser && auth) {
		onAuthStateChanged(auth, (user) => {
			set(user ? { uid: user.uid, email: user.email, nome: user.displayName, foto: user.photoURL } : null);
		});
	}

	return { subscribe };
}

export const usuario = criarUserStore();

/** Login com Google via popup. */
export async function login() {
	if (!auth) return;
	const provider = new GoogleAuthProvider();
	// Força seleção de conta — mesmo que já tenha sessão Google,
	// mostra a tela de escolher conta (permite trocar de usuário).
	provider.setCustomParameters({ prompt: 'select_account' });
	await signInWithPopup(auth, provider);
}

/** Logout — limpa sessão e força reload para estado limpo. */
export async function logout() {
	if (!auth) return;
	await signOut(auth);
	// Força reload para garantir que nenhum cache/state residual mantenha o usuário.
	// O signOut limpa o IndexedDB do Firebase, mas stores reativos e cache
	// de fetch podem manter dados stale. Reload é a forma mais segura.
	window.location.href = '/';
}

/** Retorna o ID token JWT do usuário logado (para enviar ao backend). */
export async function getIdToken() {
	if (testeUser) return browser && window.__E2E_ID_TOKEN__ ? window.__E2E_ID_TOKEN__ : 'e2e-fake-token';
	if (!auth?.currentUser) return null;
	try {
		// Timeout de 5s para evitar que token refresh pendure a UI
		const token = await Promise.race([
			auth.currentUser.getIdToken(),
			new Promise((_, reject) => setTimeout(() => reject(new Error('token timeout')), 5000))
		]);
		return token;
	} catch {
		return null;
	}
}
