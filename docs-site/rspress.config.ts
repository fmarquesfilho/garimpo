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
	globalStyles: path.join(__dirname, 'theme', 'global.css'),
	builderConfig: {
		html: {
			tags: [
				{
					tag: 'script',
					attrs: { type: 'module' },
					children: `
						import mermaid from 'https://cdn.jsdelivr.net/npm/mermaid@11/dist/mermaid.esm.min.mjs';
						mermaid.initialize({ startOnLoad: false, theme: 'neutral' });
						function renderMermaidBlocks() {
							document.querySelectorAll('div.language-mermaid').forEach(function(el) {
								if (el.dataset.mermaidRendered) return;
								el.dataset.mermaidRendered = 'true';
								var code = el.querySelector('pre code');
								if (!code) return;
								var source = code.textContent;
								var wrapper = document.createElement('div');
								wrapper.className = 'mermaid-wrapper';
								var diagramDiv = document.createElement('div');
								diagramDiv.className = 'mermaid';
								diagramDiv.textContent = source;
								var sourceDiv = document.createElement('div');
								sourceDiv.className = 'mermaid-source hidden';
								sourceDiv.appendChild(el.cloneNode(true));
								var btnBar = document.createElement('div');
								btnBar.className = 'mermaid-btn-bar';
								var toggle = document.createElement('button');
								toggle.className = 'mermaid-toggle';
								toggle.textContent = '</> Código';
								toggle.addEventListener('click', function() {
									var showingSource = sourceDiv.classList.contains('hidden');
									sourceDiv.classList.toggle('hidden');
									diagramDiv.classList.toggle('hidden');
									toggle.textContent = showingSource ? '📊 Diagrama' : '</> Código';
								});
								var fullscreenBtn = document.createElement('button');
								fullscreenBtn.className = 'mermaid-toggle';
								fullscreenBtn.textContent = '⛶ Tela cheia';
								fullscreenBtn.addEventListener('click', function() {
									var modal = document.createElement('div');
									modal.className = 'mermaid-fullscreen';
									var closeBtn = document.createElement('button');
									closeBtn.className = 'mermaid-fullscreen-close';
									closeBtn.textContent = '✕ Fechar';
									closeBtn.addEventListener('click', function() { modal.remove(); });
									modal.addEventListener('click', function(e) { if (e.target === modal) modal.remove(); });
									var content = document.createElement('div');
									content.className = 'mermaid-fullscreen-content mermaid';
									content.textContent = source;
									modal.appendChild(closeBtn);
									modal.appendChild(content);
									document.body.appendChild(modal);
									mermaid.run({ nodes: [content] });
								});
								btnBar.appendChild(toggle);
								btnBar.appendChild(fullscreenBtn);
								wrapper.appendChild(btnBar);
								wrapper.appendChild(diagramDiv);
								wrapper.appendChild(sourceDiv);
								el.replaceWith(wrapper);
							});
							mermaid.run();
						}
						var observer = new MutationObserver(function() { renderMermaidBlocks(); });
						observer.observe(document.body, { childList: true, subtree: true });
						setTimeout(renderMermaidBlocks, 800);
					`,
					append: true
				},
				{
					tag: 'script',
					children: `
						(function initSidebarToggle() {
							var sidebar = document.querySelector('.rp-doc-layout__sidebar');
							if (!sidebar) { setTimeout(initSidebarToggle, 300); return; }
							if (document.querySelector('.sidebar-toggle-btn')) return;
							var overlay = document.createElement('div');
							overlay.className = 'sidebar-overlay';
							document.body.appendChild(overlay);
							var btn = document.createElement('button');
							btn.className = 'sidebar-toggle-btn';
							btn.innerHTML = '☰';
							btn.setAttribute('aria-label', 'Toggle sidebar');
							document.body.appendChild(btn);
							btn.addEventListener('click', function() { sidebar.classList.toggle('open'); overlay.classList.toggle('visible'); });
							overlay.addEventListener('click', function() { sidebar.classList.remove('open'); overlay.classList.remove('visible'); });
							sidebar.addEventListener('click', function(e) { if (e.target.closest('a')) { setTimeout(function() { sidebar.classList.remove('open'); overlay.classList.remove('visible'); }, 100); } });
						})();
					`,
					append: true
				}
			]
		}
	}
});
