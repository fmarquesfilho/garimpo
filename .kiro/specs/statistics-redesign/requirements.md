# Spec: Redesign da Página de Estatísticas

## Problema atual (observado em produção 27/06)

A página mostra informações redundantes e sem análise:
1. **"4 lojas monitoradas"** — mas a evolução só mostra 2 (inconsistente)
2. **Lista de lojas** — redundante (já está em Configurações → Lojas)
3. **Últimas publicações** — deveria estar na página Publicações, não aqui
4. **Evolução de preço** — cards pouco informativos, sem conclusão acionável
5. **13 erros de envio** — sem contexto de quando/porquê

## O que deveria ser (princípio)

Estatísticas = **análises que ajudam a tomar decisões**, não listagem de dados.

## Decisões a tomar

- [ ] O que remover: lista de lojas, últimas publicações (mover para suas páginas)
- [ ] O que manter: evolução de preço (mas reformulada com conclusões)
- [ ] O que adicionar: receita por canal (quando conversões estiverem ativas), performance por publicação
- [ ] Inconsistência 4 vs 2 lojas: investigar (pode ser buscas por keyword contadas como lojas)
- [ ] Mover "últimas publicações" para /publicacoes como seção colapsável

## Status

Spec aberta — **não implementar agora**. Aguardar dados de conversão e estabilização da coleta de lojas para ter base estatística real. Enquanto isso, a página mantém o resumo básico (cards) + evolução.

## Dependências
- Spec de conversões (para mostrar performance real)
- 2+ semanas de coletas acumuladas (para evolução ter significância)
- Definição de quais perguntas a Mileny quer responder com dados
