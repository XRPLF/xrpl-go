import type {ReactNode} from 'react';
import Layout from '@theme/Layout';
import {LinkButton} from '../components/LinkButton';
import config from '@site/docusaurus.config';

export default function Home(): ReactNode {
  return (
    <Layout
      title="Go SDK for the XRP Ledger"
      description="Build Go applications on the XRP Ledger. Query ledger data, subscribe to updates, and sign and submit transactions.">
      <main>
        <section className="hero-section">
          <img 
            src={`${config.baseUrl}/img/xrpl-go-logo.png`}
            alt="XRPL GO Logo"
            className="hero-logo"
          />
          <h1 className="hero-title">
            XRPL GO
          </h1>
          <p className="hero-description">
            Build Go applications on the XRP Ledger. Query ledger data,
            subscribe to updates, and sign and submit transactions.
          </p>
          <LinkButton href={`${config.baseUrl}/docs/intro`}>
            Get started
          </LinkButton>
        </section>
      </main>
    </Layout>
  );
}
