import React from 'react';
import ComponentCreator from '@docusaurus/ComponentCreator';

export default [
  {
    path: '/goclarc/docs',
    component: ComponentCreator('/goclarc/docs', '053'),
    routes: [
      {
        path: '/goclarc/docs',
        component: ComponentCreator('/goclarc/docs', '4bb'),
        routes: [
          {
            path: '/goclarc/docs',
            component: ComponentCreator('/goclarc/docs', 'a64'),
            routes: [
              {
                path: '/goclarc/docs',
                component: ComponentCreator('/goclarc/docs', '8e5'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/goclarc/docs/adapters/mongo',
                component: ComponentCreator('/goclarc/docs/adapters/mongo', '061'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/goclarc/docs/adapters/postgres',
                component: ComponentCreator('/goclarc/docs/adapters/postgres', '598'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/goclarc/docs/adapters/rtdb',
                component: ComponentCreator('/goclarc/docs/adapters/rtdb', '98e'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/goclarc/docs/commands/module',
                component: ComponentCreator('/goclarc/docs/commands/module', 'e98'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/goclarc/docs/commands/new',
                component: ComponentCreator('/goclarc/docs/commands/new', 'b6b'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/goclarc/docs/examples',
                component: ComponentCreator('/goclarc/docs/examples', '171'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/goclarc/docs/generated-code',
                component: ComponentCreator('/goclarc/docs/generated-code', '23f'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/goclarc/docs/getting-started',
                component: ComponentCreator('/goclarc/docs/getting-started', '487'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/goclarc/docs/installation',
                component: ComponentCreator('/goclarc/docs/installation', 'bdc'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/goclarc/docs/schema/field-types',
                component: ComponentCreator('/goclarc/docs/schema/field-types', '9d2'),
                exact: true,
                sidebar: "docsSidebar"
              },
              {
                path: '/goclarc/docs/schema/overview',
                component: ComponentCreator('/goclarc/docs/schema/overview', '4e1'),
                exact: true,
                sidebar: "docsSidebar"
              }
            ]
          }
        ]
      }
    ]
  },
  {
    path: '/goclarc/',
    component: ComponentCreator('/goclarc/', '0e4'),
    exact: true
  },
  {
    path: '*',
    component: ComponentCreator('*'),
  },
];
