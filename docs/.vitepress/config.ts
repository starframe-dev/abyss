import { defineConfig } from 'vitepress'

export default defineConfig({
  title: 'Abyss',
  description: 'TUI browser that renders web pages as ANSI color art',
  lang: 'en-US',
  base: '/abyss/docs/',

  locales: {
    root: {
      label: 'English',
      lang: 'en-US',
      link: '/',
      themeConfig: {
        nav: [
          { text: 'Guide', link: '/guide/getting-started' },
          { text: 'GitHub', link: 'https://github.com/starframe-dev/abyss' },
        ],
        sidebar: {
          '/guide/': [
            {
              text: 'Guide',
              items: [
                { text: 'Getting Started', link: '/guide/getting-started' },
                { text: 'Architecture', link: '/guide/architecture' },
                { text: 'Supported CSS', link: '/guide/supported-css' },
              ],
            },
          ],
        },
      },
    },
    ru: {
      label: 'Русский',
      lang: 'ru-RU',
      link: '/ru/',
      themeConfig: {
        nav: [
          { text: 'Руководство', link: '/ru/guide/getting-started' },
          { text: 'GitHub', link: 'https://github.com/starframe-dev/abyss' },
        ],
        sidebar: {
          '/ru/guide/': [
            {
              text: 'Руководство',
              items: [
                { text: 'Начало работы', link: '/ru/guide/getting-started' },
                { text: 'Архитектура', link: '/ru/guide/architecture' },
                { text: 'Поддерживаемый CSS', link: '/ru/guide/supported-css' },
              ],
            },
          ],
        },
      },
    },
  },

  themeConfig: {
    socialLinks: [{ icon: 'github', link: 'https://github.com/starframe-dev/abyss' }],
  },
})
