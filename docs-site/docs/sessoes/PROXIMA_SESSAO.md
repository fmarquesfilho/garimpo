# Próxima Sessão — Docs Site UX + Sync Scheduler

**Data:** 2026-07-05
**Prioridade:** Alta

---

## 1. Corrigir docs-site (Mermaid + Sidebar)

O site de documentação (`garimpei.app.br/docs`) tem 4 problemas na renderização
dos diagramas Mermaid e na navegação mobile/tablet.

### Contexto técnico

- Stack: Rspress (baseado em Rsbuild/React), deploy via Cloudflare Pages
- Config: `docs-site/rspress.config.ts`
- Estilos: `docs-site/theme/global.css`
- Scripts injetados via `builderConfig.html.tags` (única forma que funciona no Rspress)
- Mermaid 11 carregado via CDN (`jsdelivr`)
- Classes DOM do Rspress: sidebar = `.rp-doc-layout__sidebar`, code blocks = `div.language-mermaid`

### Problemas a resolver

**1. Diagrama com fundo preto / texto escuro (ilegível em dark mode)**
- Causa: Mermaid `theme: 'neutral'` herda cores do CSS do container (dark mode)
- Fix: usar `mermaid.initialize({ theme: 'default', themeVariables: { background: '#ffffff' } })` 
  OU forçar `.mermaid-wrapper { background: white }` e `.mermaid svg { color: #333 }`
  OU usar `theme: 'forest'` que tem bom contraste em ambos os modos

**2. Diagramas não são escaláveis (sem zoom/pan)**
- Os SVGs são estáticos — não dá pra dar zoom
- Fix: integrar `svg-pan-zoom` (npm) ou `panzoom` no wrapper do diagrama rendered
- Alternativa: CSS `overflow: auto` + `transform: scale()` com controles +/-

**3. PDF pré-gerado dos diagramas (asset estático)**
- Gerar no build time usando `@mermaid-js/mermaid-cli` (mmdc)
- Cada diagram source → SVG file → asset linkável
- Adicionar botão "⬇ PDF" ao lado dos outros botões
- Script: `npx mmdc -i input.mmd -o output.svg` no build

**4. Sidebar não colapsa em nenhum tamanho de tela**
- O Rspress já tem sidebar responsiva nativa (hamburger ☰ no mobile)
- O CSS custom em `global.css` pode estar conflitando (`transform: translateX(-100%)`)
- Fix: REMOVER todo o CSS custom de sidebar collapse e o script JS de toggle
- Em vez disso, usar a config nativa do Rspress: `themeConfig.sidebar` com `collapsed: true`
- Se precisar de toggle em tablet (768-1024px), usar apenas media query no CSS
  sem JavaScript — ex: `@media (max-width: 1024px) { .rp-doc-layout__sidebar { display: none } }`
  com um checkbox hack ou o próprio menu do Rspress

### Abordagem recomendada

1. Remover TODO o JS de sidebar toggle e CSS custom de collapse (que não funciona)
2. Testar se o Rspress nativo já tem sidebar responsiva — se sim, usar
3. Para Mermaid: forçar fundo branco no wrapper, usar theme com bom contraste
4. Para zoom: avaliar `panzoom` (2KB) vs controles CSS simples
5. Para PDF: adicionar step no CI ou no `npm run build` do docs-site

---

## 2. Sync Scheduler (T-0028)

Auto-sync: buscas ativas no PostgreSQL → cron jobs no scheduler.

---

## Resultados da sessão 2026-07-04

(ver commits anteriores para detalhes completos)
