# Roadmap

> Gerado automaticamente. Não edite — rode `mise run docs:board`.

## 🔵 Now (em andamento)

- **T-0034** Descobrir: corrigir filtros + badge drift + remover nota mínima · E-001 · M

## 🟡 Next (próximo sprint)

- **T-0002** Persistir conversões da Shopee no BigQuery · conversoes · M
- **T-0024** Testar publicação WhatsApp via Meta Cloud API · produto · P
- **T-0027** Publisher multi-tenant: tokens do tenant em vez de env vars globais · multi-tenant · M
- **T-0030** Onboarding: adicionar WhatsApp Meta Cloud API + finalizar fluxo · multi-tenant · M
- **T-0033** Validar multi-marketplace + coupon monitoring com banco limpo · E-003 · M
- **T-0041** UI polish: format:check no CI + dark mode audit + Lighthouse · qualidade · P
- **T-0045** Encriptar credenciais do tenant (Shopee, Telegram, WhatsApp) · seguranca · M
- **T-0046** Validar todos os 11 fluxos end-to-end em produção · qualidade · M
- **T-0047** Observabilidade: structured logging + métricas de negócio · operacao · M
- **T-0048** Limpeza final: código morto, warnings, deps desatualizadas · qualidade · S
- **T-0049** Reconciliação: sincronizar buscas existentes no PG com Scheduler · backend · S
- **T-0050** Testar envio real: Telegram + WhatsApp (Publisher end-to-end) · qualidade · S
- **T-0051** Documentação final: consolidar estado pós-migração · documentacao · S
- **T-0052** OpenAPI spec integrada: Swagger com todas as APIs (C# + Go gRPC) · documentacao · M
- **T-0053** Avaliar alternativas ao GitHub Actions (CI/CD + hosting de git) · operacao · S

## ⚪ Later (radar)

- **T-0007** Recomendação personalizada baseada em histórico · dados-ia · G
- **T-0031** Compartilhamento de credenciais Shopee entre tenants · multi-tenant · P

## ✅ Concluídas

- **T-0001** Docs: site Starlight + geradores + consolidação · docs-migration · G
- **T-0003** Backlog como código (schema + gen-board) · docs-migration · M
- **T-0004** ScopedStore por owner_uid (multi-tenant) · multi-tenant · G
- **T-0005** Alertas configuráveis por usuário · produto · M
- **T-0006** Integrar docs-check e docs-board na CI · qualidade · P
- **T-0008** Refactor: tratamento de erros idiomático + telemetria · qualidade · G
- **T-0009** Setup mono-repo (Go + C# + protos) + Docker Compose · migracao-arch · M
- **T-0010** Proto definitions + shopee-collector gRPC server · migracao-arch · M
- **T-0011** Publisher gRPC server (extract publish package) · migracao-arch · M
- **T-0012** Scheduler gRPC server (serviço Go separado) · migracao-arch · M
- **T-0013** C# Web App — auth middleware + health + CI · migracao-arch · M
- **T-0014** PostgreSQL schema + EF Core migrations (dados transacionais) · migracao-arch · M
- **T-0015** Multi-tenant em C# (EF Core + PostgreSQL) · migracao-arch · G
- **T-0016** Curadoria controller + scoring port em C# · migracao-arch · G
- **T-0017** Routing split (Cloudflare Worker v1→v2) · migracao-arch · P
- **T-0018** Migrar handlers de publicação para C# · migracao-arch · G
- **T-0019** Migrar handlers de lojas/buscas para C# · migracao-arch · G
- **T-0020** PostgreSQL como fonte primária + BigQuery analytics-only · migracao-arch · G
- **T-0021** Cloud Run multi-container deploy (C# + sidecars Go) · migracao-arch · M
- **T-0022** Descomissionar monólito Go · migracao-arch · M
- **T-0023** Migrar WhatsApp sender de Maytapi para Meta Cloud API · migracao-arch · M
- **T-0025** Serviço Analyzer Python (FastAPI + BigQuery) · migracao-arch · M
- **T-0026** Portar endpoints restantes do Go legado para C# · migracao-arch · G
- **T-0028** Configurar coleta no scheduler (popular snapshots do zero) · produto · M
- **T-0029** Deploy da nova API C# em produção (com endpoints portados) · migracao-arch · P
- **T-0032** Migrar docs-site de Astro Starlight para Rspress · docs-migration · P
- **T-0035** Pós-migração: resolver bugs de deploy e garantir funcionalidade completa · E-001 · M
- **T-0036** Avaliar migração de scripts bash para Taskfile ou mise · E-001 · M
- **T-0037** Biblioteca de componentes UI com Bits UI + Design Tokens · E-001 · L
- **T-0038** Migração UI fase 2: consumir compostos restantes · E-001 · M
- **T-0039** Dark mode com tokens CSS e detecção automática · E-001 · M
- **T-0040** Migrar frontend para shadcn-svelte + Tailwind CSS v4 · E-001 · G
- **T-0042** Sistema de layout consistente + auditoria dark mode · E-001 · M
- **T-0043** Fix: deploy Cloud Run não criava nova revision + binding JSON quebrado · migracao-arch · M
- **T-0044** Testar fluxo de variação de preços end-to-end · migracao-arch · M
