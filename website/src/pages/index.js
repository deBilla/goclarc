import React from 'react';
import clsx from 'clsx';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Layout from '@theme/Layout';
import Heading from '@theme/Heading';
import styles from './index.module.css';

const features = [
  {
    title: 'Schema-Driven',
    emoji: '📋',
    description: (
      <>
        Define your domain once in a simple YAML file. goclarc reads your field
        definitions and generates fully typed Go structs, DTOs, and database params
        — no hand-writing boilerplate ever again.
      </>
    ),
  },
  {
    title: 'Multi-Database',
    emoji: '🗄️',
    description: (
      <>
        One schema, any backend. Generate PostgreSQL (sqlc + pgx), MongoDB
        (mongo-driver/v2), or Firebase RTDB repositories from the same YAML.
        Mix adapters across modules in a single project.
      </>
    ),
  },
  {
    title: 'Clean Architecture',
    emoji: '🏛️',
    description: (
      <>
        Every generated module follows strict Handler → Service → Repository
        layering with interface-first design. Drop the generated code into
        production and start adding real business logic immediately.
      </>
    ),
  },
];

function Feature({ emoji, title, description }) {
  return (
    <div className={clsx('col col--4')}>
      <div className="text--center padding-horiz--md padding-vert--lg">
        <div style={{ fontSize: '3rem', marginBottom: '1rem' }}>{emoji}</div>
        <Heading as="h3">{title}</Heading>
        <p>{description}</p>
      </div>
    </div>
  );
}

function HomepageHeader() {
  const { siteConfig } = useDocusaurusContext();
  return (
    <header className={clsx('hero hero--primary', styles.heroBanner)}>
      <div className="container">
        <Heading as="h1" className="hero__title">
          {siteConfig.title}
        </Heading>
        <p className="hero__subtitle">{siteConfig.tagline}</p>

        <div className={styles.installBlock}>
          <code className={styles.installCommand}>
            go install github.com/deBilla/goclarc@latest
          </code>
        </div>

        <div className={styles.buttons}>
          <Link
            className="button button--primary button--lg"
            to="/docs/getting-started"
          >
            Get Started →
          </Link>
          <Link
            className="button button--secondary button--lg"
            href="https://github.com/deBilla/goclarc"
            style={{ marginLeft: '1rem' }}
          >
            GitHub
          </Link>
        </div>
      </div>
    </header>
  );
}

function QuickStartSection() {
  return (
    <section className={styles.quickStart}>
      <div className="container">
        <Heading as="h2" className="text--center">Up and running in 30 seconds</Heading>
        <div className={styles.codeGrid}>
          <div className={styles.codeStep}>
            <span className={styles.stepNum}>1</span>
            <pre className={styles.codeBlock}>{`goclarc new my-api \\
  --module-path github.com/you/my-api`}</pre>
            <p>Scaffold a full project skeleton with Gin, config, middleware, and error handling.</p>
          </div>
          <div className={styles.codeStep}>
            <span className={styles.stepNum}>2</span>
            <pre className={styles.codeBlock}>{`# schemas/user.yaml
module: user
fields:
  - name: id
    type: uuid
    primary: true
    auto: true
  - name: email
    type: string
    required: true`}</pre>
            <p>Write a concise YAML schema defining your domain fields and types.</p>
          </div>
          <div className={styles.codeStep}>
            <span className={styles.stepNum}>3</span>
            <pre className={styles.codeBlock}>{`goclarc module user \\
  --db postgres \\
  --schema schemas/user.yaml`}</pre>
            <p>Get entity, DTO, repository, service, handler, and routes — all typed and wired.</p>
          </div>
        </div>
      </div>
    </section>
  );
}

export default function Home() {
  const { siteConfig } = useDocusaurusContext();
  return (
    <Layout
      title={`${siteConfig.title} — NestJS for Go`}
      description="Scaffold production-ready Go Clean Architecture APIs from a YAML schema. Supports PostgreSQL, MongoDB, and Firebase RTDB."
    >
      <HomepageHeader />
      <main>
        <section className={styles.features}>
          <div className="container">
            <div className="row">
              {features.map((props, idx) => (
                <Feature key={idx} {...props} />
              ))}
            </div>
          </div>
        </section>
        <QuickStartSection />
      </main>
    </Layout>
  );
}
