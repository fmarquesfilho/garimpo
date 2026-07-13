# CLI / Execução de comandos

- Nunca usar `npx` no diretório web/. Sempre usar `bunx`.
- Nunca usar `npm run` no diretório web/. Sempre usar `bun run`.
- Comandos corretos: `bunx vitest run`, `bunx playwright test`, `bunx svelte-check`, `bunx eslint`.
- Para scripts do package.json: `bun run test:unit`, `bun run build`, `bun run check`.
- Fora do web/ (Go, Python, .NET): usar as ferramentas nativas (`go test`, `dotnet test`, `mise run`).
- Mise tasks (`mise run test:*`) são o wrapper preferido — delegam para bun internamente.
- Ao executar Playwright, sempre fazer `bun run build` antes se o teste usa webServer com preview (build estático).
