# ADR 0024 — Uso do `uv` como gerenciador de pacotes Python

**Status:** aceite  
**Data:** 2026-07-06  

## Contexto

A atualização para novas versões do Python (ex: 3.14) via bot do Renovate pode causar falhas e timeouts nas pipelines de CI (GitHub Actions). Isso ocorre porque novas versões frequentemente não possuem pacotes pré-compilados (*wheels*) disponíveis imediatamente para dependências pesadas, como `pandas` ou `google-cloud-bigquery`. Nestes cenários, o `pip` recorre à compilação local (build from source), excedendo o limite de tempo estipulado (`timeout-minutes`).

Para contornar o problema e modernizar o stack, consideramos mudar o instalador de pacotes para uma alternativa mais eficiente.

## Decisão

Adotar o **`uv`** (escrito em Rust pela Astral) como o gerenciador e instalador padrão para pacotes Python no CI e local. 
- O `uv` substitui o `pip install` na pipeline. 
- Foi incrementada a margem de segurança de timeout nos jobs que envolvem Python (de 5 para 10 minutos).

### Exemplo no CI

Em vez de:
```bash
pip install -r services/analyzer/requirements.txt
```
Passamos a utilizar a seguinte configuração, que inclui também o bypass de compatibilidade do PyO3 para versões do Python lançadas antes da biblioteca atualizar (ex: erro no pydantic-core em Python 3.14+):
```yaml
env:
  PYO3_USE_ABI3_FORWARD_COMPATIBILITY: "1"
run: |
  uv pip install --system -r services/analyzer/requirements.txt
```

## Consequências

### Se aceitar
- **Rapidez:** O `uv` é extremamente rápido (10x a 100x mais veloz que o `pip`), o que atenua drasticamente eventuais gargalos de compilação sem pacotes *wheels* nativos.
- O `uv` já constava previamente nas ferramentas gerenciadas no `mise.toml` local, facilitando a adoção imediata sem atritos de instalação adicional nas pipelines que usam o `mise-action`.
- Maior estabilidade perante os pull requests de atualização lançados pelo Renovate.

### Custo
- Padronização de comandos (`uv pip` em vez de apenas `pip`) por parte da equipe nos ambientes de desenvolvimento.
