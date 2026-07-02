# Garimpei — Frontend

SPA SvelteKit 5 servida via Cloudflare Pages.

## Stack

- **SvelteKit 2** (adapter-static → SPA)
- **Svelte 5** (runes mode)
- **Vite 8**
- **Firebase Auth** (login Google)
- **TipTap** (editor WYSIWYG para templates)

## Setup

```bash
npm install
npm run dev         # → http://localhost:5173
```

Em dev, o frontend aponta para `http://localhost:8080` por padrão.
Para apontar para a API C#:

```bash
VITE_API_BASE=http://localhost:5000 npm run dev
```

## Páginas

| Rota | Descrição |
|------|-----------|
| `/` | Landing page (exige login) |
| `/publicar` | Curadoria + publicação (4 fontes: busca, quedas, novos, favoritos) |
| `/lojas` | Monitoramento de lojas (novidades, evolução) |
| `/oportunidades` | Feed unificado de variações de preço |
| `/canais` | Configurar destinos (Telegram/WhatsApp) |
| `/publicacoes` | Histórico de publicações |
| `/coletas` | Histórico de coletas |
| `/estatisticas` | Dashboard analytics |
| `/configurar` | Onboarding + configuração de conta (multi-step) |
| `/admin` | Painel admin |

## Testes

```bash
npx vitest run       # Unit (109 testes, <2s)
npm test             # E2E (Playwright)
npm run lint         # ESLint + Stylelint
npm run lint:dead    # Knip (dead code)
```

## Build

```bash
npm run build        # Gera static site em build/
npm run preview      # Preview do build local
```

Deploy: Cloudflare Pages (automático via GitHub Actions).

## API Client

Toda comunicação com o backend está centralizada em `src/lib/api.js`.
Usa `getIdToken()` do Firebase Auth para autenticação (Bearer JWT).
