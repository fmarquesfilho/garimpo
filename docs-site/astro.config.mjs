import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

export default defineConfig({
  site: 'https://garimpei.app.br',
  base: '/docs',
  integrations: [
    starlight({
      title: 'Garimpei — Documentação',
      defaultLocale: 'root',
      locales: {
        root: { label: 'Português', lang: 'pt-BR' },
      },
      sidebar: [
        { label: 'Comece aqui', link: '/' },
        { label: 'Visão e negócio', link: '/01-visao-e-negocio/' },
        { label: 'Arquitetura', link: '/02-arquitetura/' },
        { label: 'Fluxos e modelo', link: '/03-fluxos-e-modelo/' },
        { label: 'Integração Shopee', link: '/04-operacao-shopee/' },
        { label: 'Manual do usuário', link: '/05-manual-do-usuario/' },
        { label: 'Qualidade e testes', link: '/06-qualidade-e-testes/' },
        { label: 'Dados e IA', link: '/07-dados-e-ia/' },
        {
          label: 'Gerado',
          items: [
            { label: 'Referência da API', link: '/gerado/api/' },
            { label: 'Modelo de dados (ER)', link: '/gerado/entidades/' },
            { label: 'Variáveis de ambiente', link: '/gerado/env-vars/' },
            { label: 'Quadro (Kanban)', link: '/gerado/board/' },
            { label: 'Roadmap', link: '/gerado/roadmap/' },
          ],
        },
        { label: 'Decisões (ADRs)', autogenerate: { directory: 'decisoes' } },
      ],
    }),
  ],
  legacy: {
    collections: true,
  },
});
