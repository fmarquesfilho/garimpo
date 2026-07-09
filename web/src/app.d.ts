// See https://svelte.dev/docs/kit/types#app.d.ts

import type { Auth, UserCredential } from 'firebase/auth';

// Global type augmentations for test environment
declare global {
	interface Window {
		__FIREBASE_AUTH_EMULATOR_HOST__?: string;
		__E2E_AUTH_USER__?: { uid: string; email: string; nome: string; foto: string | null };
		__E2E_ID_TOKEN__?: string;
		__MOCK_USER?: { uid: string; email: string; nome: string };
		__TEST_FORCE_AUTH?: boolean;
		__TEST_SIGN_IN__?: (email: string, password: string) => Promise<UserCredential>;
		__TEST_AUTH__?: Auth;
	}
}

export {};
