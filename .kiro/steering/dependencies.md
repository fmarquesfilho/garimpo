# Dependências

- Não adicionar dependências npm/pip/go que não tiveram release nos últimos 3 meses.
- Verificar com `curl -s "https://registry.npmjs.org/{pkg}" | jq '.time.modified'` antes de instalar.
- Preferir soluções inline (componentes próprios) quando o escopo é pequeno e a dependência é de alto risco de abandono.
- Exceções: pacotes core (svelte, vite, playwright, opentelemetry) que podem ter releases menos frequentes mas são mantidos por organizações grandes.
- Usar versões exatas (pinned) em package.json, go.mod e requirements.txt.
- Sempre usar `bun` como package manager no frontend (web/). Nunca usar npm ou yarn. Comandos: `bun install`, `bun add`, `bun run`, `bunx`.
