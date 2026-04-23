// @ts-check
const { themes } = require('prism-react-renderer');

/** @type {import('@docusaurus/types').Config} */
const config = {
  title: 'goclarc',
  tagline: 'NestJS-style scaffolding for Go Clean Architecture APIs',
  favicon: 'img/favicon.ico',

  url: 'https://deBilla.github.io',
  baseUrl: '/goclarc/',

  organizationName: 'deBilla',
  projectName: 'goclarc',
  deploymentBranch: 'gh-pages',
  trailingSlash: false,

  onBrokenLinks: 'throw',
  onBrokenMarkdownLinks: 'warn',

  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },

  presets: [
    [
      'classic',
      /** @type {import('@docusaurus/preset-classic').Options} */
      ({
        docs: {
          sidebarPath: require.resolve('./sidebars.js'),
          editUrl: 'https://github.com/deBilla/goclarc/edit/main/website/',
        },
        blog: false,
        theme: {
          customCss: require.resolve('./src/css/custom.css'),
        },
      }),
    ],
  ],

  themeConfig:
    /** @type {import('@docusaurus/preset-classic').ThemeConfig} */
    ({
      image: 'img/goclarc-social.png',
      colorMode: {
        defaultMode: 'dark',
        disableSwitch: false,
        respectPrefersColorScheme: true,
      },
      navbar: {
        title: 'goclarc',
        logo: {
          alt: 'goclarc logo',
          src: 'img/logo.svg',
        },
        items: [
          {
            type: 'docSidebar',
            sidebarId: 'docsSidebar',
            position: 'left',
            label: 'Docs',
          },
          {
            to: '/docs/commands/new',
            label: 'CLI Reference',
            position: 'left',
          },
          {
            href: 'https://github.com/deBilla/goclarc',
            label: 'GitHub',
            position: 'right',
          },
        ],
      },
      footer: {
        style: 'dark',
        links: [
          {
            title: 'Docs',
            items: [
              { label: 'Getting Started', to: '/docs/getting-started' },
              { label: 'Schema Reference', to: '/docs/schema/overview' },
              { label: 'CLI Reference', to: '/docs/commands/new' },
            ],
          },
          {
            title: 'Adapters',
            items: [
              { label: 'PostgreSQL', to: '/docs/adapters/postgres' },
              { label: 'MongoDB', to: '/docs/adapters/mongo' },
              { label: 'Firebase RTDB', to: '/docs/adapters/rtdb' },
            ],
          },
          {
            title: 'More',
            items: [
              { label: 'GitHub', href: 'https://github.com/deBilla/goclarc' },
              { label: 'Examples', to: '/docs/examples' },
              { label: 'Contributing', href: 'https://github.com/deBilla/goclarc/blob/main/CONTRIBUTING.md' },
            ],
          },
        ],
        copyright: `Copyright © ${new Date().getFullYear()} deBilla. MIT License.`,
      },
      prism: {
        theme: themes.github,
        darkTheme: themes.dracula,
        additionalLanguages: ['go', 'yaml', 'bash', 'sql'],
      },
      algolia: undefined,
    }),
};

module.exports = config;
