# CI / Testes

- Nunca adicionar testes E2E de integração real (test:e2e:*) ao CI (GitHub Actions).
- Esses testes dependem de APIs externas (Shopee) e serviços Go locais — são exclusivamente para validação manual local.
- O CI deve rodar apenas: testes unitários, drift checks, lint, build, e testes com mocks/InMemory DB.
- Testes E2E de integração real ficam disponíveis via `mise run test:e2e:*` para uso local do desenvolvedor.
- Ao editar .github/workflows/*.yml, sempre validar YAML com `yq '.' .github/workflows/ci.yml > /dev/null` antes de commitar.
- O deploy backend usa lógica conservadora: deploya sempre, EXCETO quando é seguro pular (apenas web/ mudou). Na dúvida, deploya.
