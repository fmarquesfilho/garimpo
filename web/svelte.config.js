import adapter from '@sveltejs/adapter-static';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	kit: {
		// Site 100% estático: nginx serve os arquivos e faz proxy de /api -> Go.
		// fallback dá um index para qualquer rota (comporta o app como SPA no refresh).
		adapter: adapter({
			pages: 'build',
			assets: 'build',
			fallback: '200.html',
			precompress: false,
			strict: false
		})
	}
};

export default config;
