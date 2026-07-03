import { defineConfig } from '@rspress/core';
import path from 'node:path';
import fs from 'node:fs';

// Sidebar gerado automaticamente pelo docs/sync script.
// Se o arquivo não existir (primeira vez), usa um fallback mínimo.
function loadSidebar() {
  const sidebarPath = path.join(__dirname, 'docs', '_sidebar.json');
  if (fs.existsSync(sidebarPath)) {
    return JSON.parse(fs.readFileSync(sidebarPath, 'utf-8'));
  }
  return [{ text: 'Introdução', items: [{ text: 'Home', link: '/' }] }];
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
      '/': loadSidebar(),
    },
  },
});
