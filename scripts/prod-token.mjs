#!/usr/bin/env node
/**
 * prod-token.mjs — Gera um Firebase ID token para testar a API de produção.
 *
 * Uso:
 *   node scripts/prod-token.mjs                    # token do user padrão
 *   node scripts/prod-token.mjs <uid>              # token de um UID específico
 *   eval "TOKEN=$(node scripts/prod-token.mjs)"    # exporta como variável
 *
 * O token pode ser usado:
 *   curl -H "Authorization: Bearer $TOKEN" https://garimpei.app.br/api/destinos
 *
 * Requer: GOOGLE_APPLICATION_CREDENTIALS apontando para service account key,
 * ou gcloud auth application-default login configurado.
 */
import { initializeApp, cert, applicationDefault } from 'firebase-admin/app';
import { getAuth } from 'firebase-admin/auth';

const PROJECT_ID = 'garimpo-500114';
const API_KEY = 'AIzaSyA5sBUoVkNHiq58KUkmwbxIMLhvgTn7N8A';
const DEFAULT_UID = 'GZsXcFj0xwcOXAJVt06y0PQYa7W2'; // milenygsilva@gmail.com

const uid = process.argv[2] || DEFAULT_UID;

try {
	// Tenta usar Application Default Credentials
	initializeApp({ credential: applicationDefault(), projectId: PROJECT_ID });
} catch {
	console.error('❌ Precisa de Application Default Credentials.');
	console.error('   Rode: gcloud auth application-default login');
	process.exit(1);
}

try {
	// 1. Gera custom token (Firebase Admin SDK)
	const customToken = await getAuth().createCustomToken(uid);

	// 2. Troca custom token por ID token via REST API
	const resp = await fetch(
		`https://identitytoolkit.googleapis.com/v1/accounts:signInWithCustomToken?key=${API_KEY}`,
		{
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ token: customToken, returnSecureToken: true })
		}
	);
	const data = await resp.json();

	if (data.idToken) {
		console.log(data.idToken);
	} else {
		console.error('❌ Falha ao trocar custom token:', data.error?.message || JSON.stringify(data));
		process.exit(1);
	}
} catch (err) {
	console.error('❌ Erro:', err.message);
	process.exit(1);
}
