import { defineConfig } from '@playwright/test';

export default defineConfig({
	testDir: 'tests',
	timeout: 60000,
	retries: 1,
	use: {
		baseURL: 'http://localhost:4173',
		headless: true
	},
	webServer: {
		command: 'npm run preview',
		port: 4173,
		reuseExistingServer: true,
		timeout: 60000
	},
	projects: [
		{ name: 'chromium', use: { browserName: 'chromium' } }
	]
});
