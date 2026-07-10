# Git

- Nunca fazer `git push` automaticamente. Sempre perguntar antes de pushar.
- Commits podem ser feitos normalmente, mas o push é decisão do usuário.
- Ao terminar uma sequência de commits, informar que está pronto para push e aguardar confirmação.
- Nunca usar `--no-verify` no push. O pre-push hook existe para pegar erros antes do CI.
- Se o pre-push falha, corrigir o código — não bypassar o hook.
