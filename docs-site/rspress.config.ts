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
		`<script type="module">
			import mermaid from 'https://cdn.jsdelivr.net/npm/mermaid@11/dist/mermaid.esm.min.mjs';
			mermaid.initialize({ startOnLoad: false, theme: 'neutral' });
			function renderMermaidBlocks() {
				document.querySelectorAll('pre code.language-mermaid').forEach(el => {
					const pre = el.parentElement;
					if (pre.dataset.mermaid) return;
					pre.dataset.mermaid = 'true';
					const source = el.textContent;
					const wrapper = document.createElement('div');
					wrapper.className = 'mermaid-wrapper';
					const diagramDiv = document.createElement('div');
					diagramDiv.className = 'mermaid';
					diagramDiv.textContent = source;
					const sourceDiv = document.createElement('div');
					sourceDiv.className = 'mermaid-source hidden';
					const sourcePre = pre.cloneNode(true);
					delete sourcePre.dataset.mermaid;
					sourceDiv.appendChild(sourcePre);
					const toggle = document.createElement('button');
					toggle.className = 'mermaid-toggle';
					toggle.textContent = '</> Código';
					toggle.addEventListener('click', () => {
						const showingSource = sourceDiv.classList.contains('hidden');
						sourceDiv.classList.toggle('hidden');
						diagramDiv.classList.toggle('hidden');
						toggle.textContent = showingSource ? '📊 Diagrama' : '</> Código';
					});
					wrapper.appendChild(toggle);
					wrapper.appendChild(diagramDiv);
					wrapper.appendChild(sourceDiv);
					pre.replaceWith(wrapper);
				});
				mermaid.run();
			}
			const observer = new MutationObserver(() => renderMermaidBlocks());
			observer.observe(document.body, { childList: true, subtree: true });
			setTimeout(renderMermaidBlocks, 500);
		</script>`,
		`<script>
			document.addEventListener('DOMContentLoaded', () => {
				const sidebar = document.querySelector('.rspress-sidebar');
				if (!sidebar) return;
				const overlay = document.createElement('div');
				overlay.className = 'sidebar-overlay';
				document.body.appendChild(overlay);
				const btn = document.createElement('button');
				btn.className = 'sidebar-toggle-btn';
				btn.innerHTML = '☰';
				btn.setAttribute('aria-label', 'Toggle sidebar');
				document.body.appendChild(btn);
				btn.addEventListener('click', () => { sidebar.classList.toggle('open'); overlay.classList.toggle('visible'); });
				overlay.addEventListener('click', () => { sidebar.classList.remove('open'); overlay.classList.remove('visible'); });
				sidebar.addEventListener('click', (e) => { if (e.target.closest('a')) { sidebar.classList.remove('open'); overlay.classList.remove('visible'); } });
			});
		</script>`
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
