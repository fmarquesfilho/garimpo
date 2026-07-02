import { defineConfig } from '@rspress/core';
import path from 'node:path';

export default defineConfig({
  root: path.join(__dirname, 'docs'),
  base: '/docs/',
  title: 'Garimpei',
  description: 'Documentação do Garimpei — curadoria inteligente para afiliados Shopee',
  icon: '/favicon.svg',
  markdown: {
    mdxRs: false,
    checkDeadLinks: true,
  },
  route: {
    cleanUrls: true,
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
            { text: '0001: Nome Garimpei', link: '/decisoes/0001-nome-garimpei' },
            { text: '0002: Só Nicho', link: '/decisoes/0002-so-nicho' },
            { text: '0003: Deploy GCP', link: '/decisoes/0003-deploy-gcp' },
            { text: '0004: Página Descobrir', link: '/decisoes/0004-pagina-descobrir' },
            { text: '0005: Canais Telegram/WhatsApp', link: '/decisoes/0005-canais-telegram-whatsapp' },
            { text: '0006: Categorias Plural', link: '/decisoes/0006-categorias-plural' },
            { text: '0007: Persistência Favoritos', link: '/decisoes/0007-persistencia-favoritos' },
            { text: '0008: Alertas Desabilitados', link: '/decisoes/0008-alertas-desabilitados' },
            { text: '0009: Adoção Chi Router', link: '/decisoes/0009-adocao-chi-router' },
            { text: '0010: Error Handling', link: '/decisoes/0010-error-handling' },
            { text: '0011: Repository Pattern', link: '/decisoes/0011-repository-pattern' },
            { text: '0012: Migração C# + Go', link: '/decisoes/0012-migracao-csharp-go-microservices' },
            { text: '0013: WhatsApp Meta', link: '/decisoes/0013-whatsapp-meta-cloud-api' },
            { text: '0014: Analyzer Python', link: '/decisoes/0014-analyzer-python-fastapi' },
            { text: '0015: Multi-tenant', link: '/decisoes/0015-multi-tenant-credenciais' },
            { text: '0016: Multi-marketplace', link: '/decisoes/0016-multi-marketplace' },
            { text: '0017: Coupon Monitoring', link: '/decisoes/0017-coupon-monitoring' },
            { text: '0018: Collector Unificado', link: '/decisoes/0018-collector-unificado' },
            { text: '0019: Categorias Marketplace', link: '/decisoes/0019-categorias-marketplace' },
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
