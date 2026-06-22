// Configuração do Firebase para autenticação.
// Os valores de config são públicos (não são secrets) — são necessários para o
// SDK identificar o projeto no client. A segurança está na validação do token
// no servidor (Firebase Admin SDK / verificação de ID token).
import { initializeApp } from 'firebase/app';
import { getAuth, GoogleAuthProvider, signInWithPopup, signOut, onAuthStateChanged } from 'firebase/auth';
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

if (browser) {
	app = initializeApp(firebaseConfig);
	auth = getAuth(app);
}

// Store reativo do usuário logado (ou null)
function criarUserStore() {
	const { subscribe, set } = writable(null);

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
	await signInWithPopup(auth, provider);
}

/** Logout. */
export async function logout() {
	if (!auth) return;
	await signOut(auth);
}

/** Retorna o ID token JWT do usuário logado (para enviar ao backend). */
export async function getIdToken() {
	if (!auth?.currentUser) return null;
	return auth.currentUser.getIdToken();
}
