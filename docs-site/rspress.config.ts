import { defineConfig } from '@rspress/core';
import path from 'node:path';
import fs from 'node:fs';

function loadSidebar() {
	const sidebarPath = path.join(__dirname, 'docs', '_sidebar.json');
	if (fs.existsSync(sidebarPath)) {
		return JSON.parse(fs.readFileSync(sidebarPath, 'utf-8'));
	}
	return [{ text: 'Introdução', items: [{ text: 'Home', link: '/' }] }];
}

export default defineConfig({
	root: path.join(__dirname, 'docs'),
	base: '/docs/',
	title: 'Garimpei Docs',
	description: 'Documentação técnica do Garimpei — curadoria inteligente para afiliados',
	icon: '/favicon.svg',
	logo: '/favicon.svg',
	logoText: 'Garimpei',
	markdown: {
		mdxRs: false,
		checkDeadLinks: true
	},
	head: [
		['script', { type: 'module' }, `
			import mermaid from 'https://cdn.jsdelivr.net/npm/mermaid@11/dist/mermaid.esm.min.mjs';
			mermaid.initialize({ startOnLoad: false, theme: 'neutral' });
			const observer = new MutationObserver(() => {
				document.querySelectorAll('pre code.language-mermaid').forEach(el => {
					const pre = el.parentElement;
					if (pre.dataset.mermaid) return;
					pre.dataset.mermaid = 'true';
					const div = document.createElement('div');
					div.className = 'mermaid';
					div.textContent = el.textContent;
					pre.replaceWith(div);
				});
				mermaid.run();
			});
			observer.observe(document.body, { childList: true, subtree: true });
			document.addEventListener('DOMContentLoaded', () => {
				document.querySelectorAll('pre code.language-mermaid').forEach(el => {
					const pre = el.parentElement;
					const div = document.createElement('div');
					div.className = 'mermaid';
					div.textContent = el.textContent;
					pre.replaceWith(div);
				});
				mermaid.run();
			});
		`]
	],
	route: {
		cleanUrls: true
	},
	themeConfig: {
		darkMode: true,
		hideNavbar: 'never',
		enableContentAnimation: true,
		enableScrollToTop: true,
		outline: {
			level: [2, 3]
		},
		footer: {
			message: '© 2026 Garimpei — Curadoria inteligente para afiliados'
		},
		socialLinks: [
			{
				icon: 'github',
				mode: 'link',
				content: 'https://github.com/fmarquesfilho/garimpo'
			}
		],
		sidebar: {
			'/': loadSidebar()
		},
		nav: [
			{ text: 'Início', link: '/' },
			{ text: 'Arquitetura', link: '/arquitetura' },
			{ text: 'Componentes', link: '/frontend/componentes' },
			{ text: 'ADRs', link: '/decisoes/0012-migracao-csharp-go-microservices' },
			{ text: 'Roadmap', link: '/gerado/ROADMAP' }
		]
	},
	globalStyles: path.join(__dirname, 'theme', 'global.css')
});
