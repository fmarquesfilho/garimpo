import { defineConfig } from '@rspress/core';
import path from 'node:path';
import fs from 'node:fs';

// Gera sidebar das ADRs dinamicamente a partir do filesystem
function getADRItems() {
  const decisoesDir = path.join(__dirname, 'docs', 'decisoes');
  if (!fs.existsSync(decisoesDir)) return [];
  return fs.readdirSync(decisoesDir)
    .filter(f => f.endsWith('.md'))
    .sort()
    .map(f => {
      const name = f.replace('.md', '');
      const num = name.split('-')[0];
      const title = name.replace(/^\d+-/, '').replace(/-/g, ' ');
      return { text: `${num}: ${title.charAt(0).toUpperCase() + title.slice(1)}`, link: `/decisoes/${name}` };
    });
}

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
          text: 'Frontend',
          items: [
            { text: 'Componentes UI', link: '/frontend/componentes' },
            { text: 'Linting & Qualidade', link: '/frontend/linting' },
            { text: 'Impacto da Migração', link: '/frontend/impacto-migracao-ui' },
          ],
        },
        {
          text: 'Projeto',
          items: [
            { text: 'Quadro Kanban', link: '/gerado/BOARD' },
            { text: 'Roadmap', link: '/gerado/ROADMAP' },
            { text: 'Sprint Atual', link: '/projeto/sprint' },
            { text: 'Entidades (ER)', link: '/gerado/ENTIDADES' },
            { text: 'Variáveis de Ambiente', link: '/gerado/env-vars' },
          ],
        },
        {
          text: 'Decisões (ADRs)',
          items: getADRItems(),
        },
        {
          text: 'Guias',
          items: [
            { text: 'Configurar WhatsApp Meta', link: '/guias/configurar-whatsapp-meta' },
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
