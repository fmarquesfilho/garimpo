import { defineConfig } from '@rspress/core';
import path from 'node:path';

export default defineConfig({
  root: path.join(__dirname, 'docs'),
  title: 'Garimpei',
  description: 'Documentação do Garimpei — curadoria inteligente para afiliados Shopee',
  icon: '/favicon.svg',
  markdown: {
    mdxRs: false,
    checkDeadLinks: false,
  },
  themeConfig: {
    socialLinks: [
      { icon: 'github', mode: 'link', content: 'https://github.com/fmarquesfilho/garimpo' },
    ],
    sidebar: {
      '/': [
        {
          text: 'Visão Geral',
          items: [
            { text: 'Introdução', link: '/' },
            { text: 'Visão e Negócio', link: '/visao-e-negocio' },
          ],
        },
        {
          text: 'Arquitetura',
          items: [
            { text: 'Visão Geral', link: '/arquitetura' },
            { text: 'Fluxos e Modelo', link: '/fluxos-e-modelo' },
            { text: 'Operação Shopee', link: '/operacao-shopee' },
          ],
        },
        {
          text: 'Operação',
          items: [
            { text: 'Manual do Usuário', link: '/manual-do-usuario' },
            { text: 'Qualidade e Testes', link: '/qualidade-e-testes' },
            { text: 'Dados e IA', link: '/dados-e-ia' },
          ],
        },
        {
          text: 'Decisões (ADRs)',
          items: [
            { text: 'ADR-0012: Migração C# + Go', link: '/decisoes/0012-migracao-csharp-go-microservices' },
            { text: 'ADR-0013: WhatsApp Meta', link: '/decisoes/0013-whatsapp-meta-cloud-api' },
            { text: 'ADR-0014: Analyzer Python', link: '/decisoes/0014-analyzer-python-fastapi' },
            { text: 'ADR-0015: Multi-tenant', link: '/decisoes/0015-multi-tenant-credenciais' },
          ],
        },
        {
          text: 'API',
          items: [
            { text: 'OpenAPI Spec', link: '/api' },
          ],
        },
      ],
    },
  },
});
