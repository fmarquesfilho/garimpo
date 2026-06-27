# Negócio, Releases e Conformidade — Garimpei

Documento vivo. Três frentes que andam juntas agora que o produto evoluiu e há
beta testers à mão: (A) montar o plano de negócio, (B) organizar releases sem
prometer datas, e (C) maturar o multi-tenant para a empresa da Mileny operar em
conformidade com a LGPD.

> Contexto: ~3 semanas de desenvolvimento, MVP multi-tenant implementado, beta
> testers disponíveis naturalmente. O estágio é **pré-product-market-fit** — isso
> dita as ferramentas certas (validação > planejamento extenso).

---

## A. Plano de negócio

### A.1 Comece com Lean Canvas, não com business plan tradicional

No seu estágio, a ferramenta certa é o **Lean Canvas** (Ash Maurya), não o Business
Model Canvas completo nem um plano de 30 páginas. O Lean Canvas é mais útil quando você ainda não tem clientes; o BMC é mais útil quando você está escalando algo já comprovado. Enquanto o Lean Canvas foca em testar hipóteses, o Business Model Canvas destaca os sistemas que sustentam um negócio estável ou em crescimento — ou seja, BMC fica para depois do PMF.

O Lean Canvas adapta o BMC trocando quatro blocos: Ash Maurya substituiu Relacionamento com Clientes, Atividades-Chave, Parcerias-Chave e Recursos-Chave por Problema, Solução, Métricas-Chave e Vantagem Injusta. E preenche-se numa ordem específica, do mais certo ao mais arriscado: (1) Problema + Segmentos de Cliente juntos, (2) Proposta de Valor Única, (3) Solução, (4) Canais, (5) Receita + Custos juntos, (6) Métricas-Chave, (7) Vantagem Injusta.

Cuidados que valem para o Garimpei: seja implacavelmente específico no cliente — "donas de operação de afiliados de skincare coreano no Instagram/Telegram" é melhor que "afiliados"; a Proposta de Valor é um benefício, não lista de features ("escolher o produto do dia em minutos, com dado" e não "tem ranking com IA"); e a Solução deve ser o mínimo que resolve a dor mais urgente, evitando feature creep.

### A.2 Rascunho de Lean Canvas (ponto de partida — preencha/valide)

| Bloco | Hipótese inicial (a validar com beta) |
|---|---|
| **Problema** | (1) Selecionar o produto do dia é manual, demorado e sem dado (o gargalo do `MODELO.md`); (2) sem feedback do que converteu; (3) risco de divulgar falsificação (origem). |
| **Segmento** | Operadoras de marketing de afiliados Shopee em nicho de beleza/skincare, publicando em Telegram/WhatsApp. *Early adopter:* a própria Mileny + a rede dela. |
| **Proposta de Valor Única** | "De centenas de produtos da Shopee aos poucos que valem divulgar — com origem, comissão e histórico — em minutos, não horas." |
| **Solução** | Curadoria por teor + monitoramento de lojas + publicação com template + (futuro) atribuição via `conversionReport`. |
| **Canais** | Boca a boca na comunidade de afiliados; conteúdo da própria Mileny como prova; grupos de Telegram do nicho. |
| **Receita** | Assinatura mensal (ver A.4). Possível plano grátis limitado para entrada. |
| **Custos** | GCP (Cloud Run escala a zero + BigQuery free tier — hoje ~R$0–50/mês); seu tempo; eventual WhatsApp BSP. |
| **Métricas-Chave** | Ativação (1ª loja monitorada + 1ª publicação), retenção semanal, nº de publicações/semana, e — quando houver — receita-por-hora do usuário. |
| **Vantagem Injusta** | O dado proprietário: cruzar a curadoria com o resultado real ao longo do tempo — nenhuma ferramenta paga sabe o que converte para a audiência dela, no nicho dela. Esse histórico não é copiável. |

### A.3 Métricas que importam no estágio SaaS

Foque poucas. O framework **AARRR** (Aquisição, Ativação, Retenção, Referência,
Receita) dá a visão de funil. SaaS depende fortemente de retenção, experiência e engajamento de longo prazo, não só da venda inicial; priorize Customer Lifetime Value (CLTV) sobre CAC e acompanhe churn, MRR e CAC. Para o seu volume, comece com um cálculo manual de churn mensal — uma calculadora simples de churn já ajuda a acompanhar a % que sai por mês sem precisar de software complexo.

Atenção a uma armadilha de pricing: a escolha do modelo de cobrança molda a arquitetura mais cedo do que se imagina — cobrança por uso exige infraestrutura de medição, integração de billing e dashboards de uso; planeje isso no estágio 2, não no 6. Como sua coleta já gera eventos por tenant, medir uso (lojas monitoradas, publicações, coletas) é barato — guarde isso para suportar pricing por uso/tier depois.

### A.4 Pricing (hipóteses para testar no beta)

Não decida pricing no abstrato — é uma das hipóteses mais arriscadas do Canvas.
Estruturas plausíveis para validar:
- **Por tiers de uso:** Free (1 loja, sem agendamento), Pro (N lojas, agendamento,
  alertas), Business (lojas ilimitadas, múltiplos destinos, conversões).
- **Ancorado em valor:** preço como fração da comissão que a ferramenta ajuda a
  ganhar (receita-por-hora é a métrica-rainha do `MODELO.md`).
- **Beta gratuito → conversão:** beta testers entram de graça; ao fim, oferta de
  fundador (desconto vitalício) para os que ficam. Gera os primeiros depoimentos.

### A.5 Quando migrar para Business Model Canvas

Quando tiver tração inicial, usuários validados e precisar detalhar operações, pricing e parcerias, o BMC passa a ser o melhor formato. Até lá, o Lean Canvas revisado a cada poucas semanas é suficiente.

---

## B. Plano de releases

### B.1 Roadmap por horizontes, não por datas (Now / Next / Later)

Para um produto que reprioriza a cada feedback, **datas são ficção**. Use
**Now / Next / Later**: os roadmaps públicos mais eficazes em 2026 evitam datas específicas e usam horizontes — Now (em desenvolvimento ativo), Next (priorizado e escopado, mas não em desenvolvimento), Later (no radar, ainda não comprometido). Startups em estágio inicial se beneficiam muito desse formato porque as prioridades mudam semana a semana conforme aprendem com feedback; ele comunica direção a investidores e conselheiros sem fazer promessas de entrega impossíveis de cumprir.

Regras operacionais do formato:
- Now: no máximo 3–5 itens (força foco); Next: 5–10 itens (planejados, não comprometidos), que sobem para Now quando o trabalho atual termina.
- A coluna Now deve ser deliberadamente pequena para sinalizar foco real e proteger contra excesso de compromisso; reveja cada horizonte numa cadência recorrente (semanal/quinzenal).
- Faça poda ativa: arquive itens parados há mais de ~3 meses; quando decidir não fazer algo, publique como "Não planejado" com uma frase de justificativa — honestidade gera confiança.

> Esse roadmap é **gerado** a partir do backlog YAML (ver `03`) — o
> `docs/gerado/ROADMAP.md` é a saída, ninguém edita à mão.

### B.2 Roadmap ≠ plano de release

Mantenha os dois distintos: roadmap = direção e resultados, de longo prazo e adaptativo, dono é o produto; plano de release = entrega, features e capacidade, de curto prazo e sensível a sequência/dependências, dono é quem entrega. Você acumula os dois papéis, mas separe os artefatos: o **roadmap** (Now/Next/Later) responde "estamos construindo a coisa certa?"; o **plano de release** (o sprint atual no board) responde "o que entregamos a seguir e em que ordem, dadas as dependências?".

### B.3 Escopo de MVP com MoSCoW

Para decidir o que entra em cada release, **MoSCoW** funciona bem no início:
Must-have (sem isto o produto não funciona), Should-have (importante mas não vital), Could-have (bom ter, baixo impacto se faltar), Won't-have (adiado) — ótimo nas fases iniciais para focar primeiro na funcionalidade crítica.

Evite a "fábrica de features": o foco em features pode virar comportamento de "feature factory", onde velocidade de entrega importa mais que valor para o cliente; organize por Now/Next/Later e conecte cada item a um resultado.

### B.4 Programa de beta (aproveitar os testers agora)

Estrutura enxuta para extrair sinal dos beta testers:
1. **Caminho de retorno curto** (o `DEPLOY.md` já sugere): um grupo de Telegram +
   um label `feedback` que vira tarefa no backlog (ver `03`, seção 7).
2. **Foco de validação por ciclo:** cada semana de beta testa **uma** hipótese do
   Canvas (ex.: "operadoras conseguem ativar sozinhas?" → mede ativação).
3. **Instrumente ativação e retenção** desde já (eventos por tenant) — sem isso o
   beta vira opinião, não dado. A IA acelera a execução do código, mas não diz se a feature vale a pena; a qualidade das decisões de produto continua sendo o gargalo upstream.
4. **Status visível:** quando um beta tester vê o pedido dele sair de "Next" para
   "Now", a confiança aumenta e o feedback melhora — é o ciclo virtuoso do Now/Next/Later.

---

## C. Multi-tenant + LGPD

### C.1 Onde o multi-tenant está e o que falta

Do `BACKLOG.md`, o **MVP multi-tenant** já existe: package `internal/tenant`
(Config, Store, MemoryStore), criptografia AES-256-GCM dos secrets, 6 endpoints de
onboarding, página `/configurar` (wizard), e exclusão de conta. **Antes dos beta
testers entrarem de verdade**, faltam (já mapeados, aqui priorizados):

| Pendência | Por que é bloqueante | Risco se ignorar |
|---|---|---|
| **`BigQueryTenantStore`** (hoje é `MemoryStore`) | Configs somem a cada cold start do Cloud Run | Tenant perde credenciais; experiência quebrada |
| **`ScopedStore` por `owner_uid`** em todas as queries | Isolamento de dados entre tenants | **Vazamento entre clientes** — incidente LGPD |
| **Resolver credenciais do tenant no middleware** | Coletas precisam usar o token Shopee do tenant certo | Coleta usa credencial errada / falha |
| Billing por tenant | Não-MVP | — (depois) |

> O **`ScopedStore` é o item de maior risco**: multi-tenancy sem isolamento por
> linha é a falha clássica. A maioria dos produtos SaaS modernos nasce multi-tenant porque o custo de mudar depois é alto — construir single-tenant e depois passar 18 meses fazendo retrofit de multi-tenancy é um erro comum. Você já nasceu multi-tenant; só falta fechar o isolamento.

**Padrão recomendado:** *row-level security* lógica — toda query do BigQuery
filtra por `owner_uid` no `ScopedStore`, e o `owner_uid` vem **sempre** do token
Firebase verificado no middleware, **nunca** de parâmetro do cliente. Teste de
regressão dedicado: "tenant A não enxerga dado de tenant B".

### C.2 LGPD — o que a empresa da Mileny precisa

A LGPD se aplica a qualquer tratamento de dados de pessoas no Brasil. Pontos que
afetam o Garimpei diretamente (com base na regulação vigente em 2026):

**Encarregado (DPO).** Diferente do GDPR, que limita a obrigação a categorias específicas, a LGPD exige que todos os controladores nomeiem um Encarregado, com requisitos simplificados apenas para agentes de pequeno porte sob a Resolução CD/ANPD 2/2022. O encarregado deve ser pessoa natural, e sua identidade e contato precisam ser publicamente divulgados e registrados na ANPD. Para vocês: agentes de pequeno porte (≤ R$ 4,8 mi de receita) são dispensados a menos que façam tratamento de alto risco, como dados de crianças ou IA — então a obrigação plena pode não pegar agora, mas **nomear o Fernando como encarregado inicial** (já no backlog) e publicar um canal de contato é barato e recomendado. Não nomeie um encarregado de fachada: a ANPD espera alguém genuinamente acessível e que conheça as atividades de tratamento.

**Base legal por atividade.** São dez bases legais (consentimento, obrigação legal, execução de contrato, legítimo interesse, etc.); todas são igualmente válidas e o controlador deve identificar uma por atividade de tratamento. Mapeie: cadastro/login (execução de contrato), envio de alertas (execução de contrato/legítimo interesse), dados da operação (legítimo interesse). Atenção a apoiar-se em consentimento para coletas comerciais — é o ponto mais escrutinado.

**Mapeamento de dados (RoPA).** A ausência de mapeamento de dados é o indicador mais confiável de compliance imaturo. O Art. 37 exige Registro das Operações de Tratamento, especialmente quando o tratamento se apoia em legítimo interesse. Faça um inventário simples: que dado pessoal existe (email, UID Firebase, secrets do tenant), onde fica (BigQuery, Secret Manager), para quê, e por quanto tempo.

**Política de privacidade acessível** (já no backlog como pendente). Deve conter, no
mínimo: identidade do controlador, contato do Encarregado, finalidades do tratamento, tipos de dados coletados, base legal de cada atividade, compartilhamento com terceiros, transferências internacionais, prazo de retenção, direitos do titular e descrição das medidas de segurança.

**Direitos do titular em até 15 dias.** É preciso implementar sistemas operacionais que cumpram pedidos de titulares em 15 dias e notifiquem violações em 72 horas. Você já tem exclusão de conta (bom — direito de eliminação); falta cobrir acesso/portabilidade e ter um canal para os pedidos.

**Transferência internacional — ponto técnico importante.** Para transferir dados para fora do Brasil use as cláusulas-padrão da Resolução 19/2024 (texto exato), consentimento específico destacado, ou outra base; nenhum país tem decisão de adequação. Implicação para a arquitetura: **mantenha o BigQuery e o Cloud Run em `southamerica-east1`** (dados ficam no Brasil — você já usa essa região). Avalie onde Firebase Auth, Telegram e Shopee processam dados; se houver fluxo para fora, documente a base de transferência.

**Notificação de incidente.** A aplicação recente contra Meta e X mostra que a ANPD suspende tratamentos quando faltam salvaguardas adequadas. Tenha um plano mínimo de resposta a incidentes (quem avisa, em quanto tempo, para a ANPD e titulares).

**Sanções (dimensione o risco).** Multas chegam a 2% da receita no Brasil, limitadas a R$ 50 milhões por infração, além de bloqueio de dados e suspensão do tratamento — proporcional à receita, então hoje o risco financeiro é baixo, mas suspensão de tratamento mataria a operação.

### C.3 Checklist LGPD priorizado para o Garimpei

**Agora (antes do beta crescer):**
- [ ] `ScopedStore` por `owner_uid` (isolamento) + teste de regressão A≠B
- [ ] `BigQueryTenantStore` (persistir configs; secrets já criptografados ✅)
- [ ] Política de privacidade publicada (conteúdo mínimo acima)
- [ ] Encarregado nomeado (Fernando) + canal de contato divulgado
- [ ] Manter dados em região Brasil (`southamerica-east1`) ✅ já é o caso
- [ ] Aceite de termos com timestamp ✅ já existe · Exclusão de conta ✅ já existe

**Next:**
- [ ] Inventário/RoPA simples (planilha basta no início)
- [ ] Atender direitos de acesso/portabilidade (além da exclusão)
- [ ] Mapear base legal por atividade de tratamento
- [ ] DPA com fornecedores-chave (Google/GCP, provedor WhatsApp) — verificar se há Acordos de Processamento de Dados com os fornecedores, com as cláusulas mínimas da LGPD

**Later:**
- [ ] Plano formal de resposta a incidentes (72h)
- [ ] Logs de acesso a dados pessoais (já está no backlog)
- [ ] Reavaliar obrigações se passar de pequeno porte ou tratar dado sensível

> A LGPD não é "banner de cookies + política": muitos acreditam, de forma equivocada, que a conformidade se resume a banner de cookies, política de privacidade ou selo, ignorando a necessidade de gestão contínua e efetiva. Trate como processo, não entregável único. Para uma operação pequena, conformidade bem-feita também é **vantagem competitiva** ao vender para clientes brasileiros.

---

### Aviso

Este documento organiza requisitos e técnicas; não é aconselhamento jurídico. Para
a política de privacidade, o registro do encarregado na ANPD e os termos de uso,
vale uma revisão com advogado especializado em proteção de dados — há serviços de
DPO terceirizado acessíveis para pequenos negócios no Brasil.
