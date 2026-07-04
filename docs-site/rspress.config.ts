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
						import panzoom from 'https://cdn.jsdelivr.net/npm/panzoom@9.4.3/+esm';
						mermaid.initialize({
							startOnLoad: false,
							theme: 'default',
							themeVariables: {
								background: '#ffffff',
								primaryColor: '#e8f4fd',
								primaryTextColor: '#1a1a1a',
								primaryBorderColor: '#4a90d9',
								secondaryColor: '#f0f7e6',
								secondaryTextColor: '#1a1a1a',
								secondaryBorderColor: '#5ba85b',
								tertiaryColor: '#fff4e6',
								tertiaryTextColor: '#1a1a1a',
								tertiaryBorderColor: '#d4860a',
								lineColor: '#333333',
								textColor: '#1a1a1a',
								mainBkg: '#e8f4fd',
								nodeBorder: '#4a90d9',
								clusterBkg: '#f8f9fa',
								titleColor: '#1a1a1a',
								edgeLabelBackground: '#ffffff',
								actorTextColor: '#1a1a1a',
								actorBkg: '#e8f4fd',
								actorBorder: '#4a90d9',
								signalColor: '#333333',
								signalTextColor: '#1a1a1a',
								labelBoxBkgColor: '#e8f4fd',
								labelBoxBorderColor: '#4a90d9',
								labelTextColor: '#1a1a1a',
								loopTextColor: '#1a1a1a',
								noteBkgColor: '#fff4e6',
								noteTextColor: '#1a1a1a',
								noteBorderColor: '#d4860a',
								activationBkgColor: '#dbeafe',
								activationBorderColor: '#4a90d9',
								sequenceNumberColor: '#ffffff'
							}
						});
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
								diagramDiv.className = 'mermaid-diagram';
								var mermaidDiv = document.createElement('div');
								mermaidDiv.className = 'mermaid';
								mermaidDiv.textContent = source;
								diagramDiv.appendChild(mermaidDiv);
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
								var zoomInBtn = document.createElement('button');
								zoomInBtn.className = 'mermaid-toggle';
								zoomInBtn.textContent = '+ Zoom';
								var zoomOutBtn = document.createElement('button');
								zoomOutBtn.className = 'mermaid-toggle';
								zoomOutBtn.textContent = '− Zoom';
								var resetBtn = document.createElement('button');
								resetBtn.className = 'mermaid-toggle';
								resetBtn.textContent = '↺ Reset';
								var downloadBtn = document.createElement('button');
								downloadBtn.className = 'mermaid-toggle';
								downloadBtn.textContent = '⬇ Download';
								downloadBtn.addEventListener('click', function() {
									var svg = diagramDiv.querySelector('svg');
									if (!svg) return;
									var svgClone = svg.cloneNode(true);
									svgClone.setAttribute('xmlns', 'http://www.w3.org/2000/svg');
									if (!svgClone.getAttribute('width')) {
										var bbox = svg.getBBox ? svg.getBBox() : null;
										if (bbox) {
											svgClone.setAttribute('width', bbox.width);
											svgClone.setAttribute('height', bbox.height);
										}
									}
									var blob = new Blob([svgClone.outerHTML], { type: 'image/svg+xml' });
									var url = URL.createObjectURL(blob);
									var a = document.createElement('a');
									a.href = url;
									a.download = 'diagrama.svg';
									a.click();
									URL.revokeObjectURL(url);
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
									mermaid.run({ nodes: [content] }).then(function() {
										var fsPz = panzoom(content.querySelector('svg') || content, { smoothScroll: false });
										modal.addEventListener('click', function(e) { if (e.target === modal) { fsPz.dispose(); modal.remove(); } });
										closeBtn.addEventListener('click', function() { fsPz.dispose(); });
									});
								});
								btnBar.appendChild(toggle);
								btnBar.appendChild(zoomInBtn);
								btnBar.appendChild(zoomOutBtn);
								btnBar.appendChild(resetBtn);
								btnBar.appendChild(downloadBtn);
								btnBar.appendChild(fullscreenBtn);
								wrapper.appendChild(btnBar);
								wrapper.appendChild(diagramDiv);
								wrapper.appendChild(sourceDiv);
								el.replaceWith(wrapper);
								// Store panzoom init function for after render
								wrapper._initPanzoom = function() {
									var svg = diagramDiv.querySelector('svg');
									if (!svg) return;
									var pz = panzoom(svg, { smoothScroll: false, maxZoom: 5, minZoom: 0.3 });
									zoomInBtn.addEventListener('click', function() { pz.smoothZoom(svg.clientWidth / 2, svg.clientHeight / 2, 1.3); });
									zoomOutBtn.addEventListener('click', function() { pz.smoothZoom(svg.clientWidth / 2, svg.clientHeight / 2, 0.7); });
									resetBtn.addEventListener('click', function() { pz.moveTo(0, 0); pz.zoomAbs(0, 0, 1); });
									wrapper._pz = pz;
								};
							});
							mermaid.run().then(function() {
								document.querySelectorAll('.mermaid-wrapper').forEach(function(w) {
									if (w._initPanzoom && !w._pz) w._initPanzoom();
								});
							});
						}
						var observer = new MutationObserver(function() { renderMermaidBlocks(); });
						observer.observe(document.body, { childList: true, subtree: true });
						setTimeout(renderMermaidBlocks, 800);
					`,
					append: true
				}
			]
		}
	}
});
