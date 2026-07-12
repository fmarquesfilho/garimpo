import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vite';

// Alvo do proxy de API no dev server. Default: backend Go local (:8080).
// Sobrescreva com DEV_API_PROXY=https://garimpei.app.br para "jogar" contra a
// API de produção sem subir o backend local (o proxy roda server-side → sem CORS).
// Requer VITE_API_BASE='' para o browser chamar a mesma origem (localhost).
const DEV_API_PROXY = process.env.DEV_API_PROXY || 'http://localhost:8080';

export default defineConfig({
	plugins: [tailwindcss(), sveltekit()],
	server: {
		proxy: {
			'/api': { target: DEV_API_PROXY, changeOrigin: true },
			'/internal': { target: DEV_API_PROXY, changeOrigin: true }
		}
	},
	preview: {
		proxy: {
			'/api': {
				target: 'http://localhost:8090',
				changeOrigin: true
			},
			'/internal': {
				target: 'http://localhost:8090',
				changeOrigin: true
			}
		}
	}
});
