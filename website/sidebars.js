/** @type {import('@docusaurus/plugin-content-docs').SidebarsConfig} */
const sidebars = {
  docsSidebar: [
    'intro',
    'installation',
    'getting-started',
    {
      type: 'category',
      label: 'CLI Reference',
      collapsed: false,
      items: ['commands/new', 'commands/module', 'commands/crypto'],
    },
    {
      type: 'category',
      label: 'Schema',
      collapsed: false,
      items: ['schema/overview', 'schema/field-types'],
    },
    {
      type: 'category',
      label: 'Database Adapters',
      collapsed: false,
      items: ['adapters/postgres', 'adapters/mongo', 'adapters/rtdb'],
    },
    'generated-code',
    'e2ee-architecture',
    'go-best-practices',
    'examples',
  ],
};

module.exports = sidebars;
