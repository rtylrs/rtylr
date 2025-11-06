import React from 'react';
import styles from './IntegrationsSection.module.scss';

const integrations = [
  { name: 'Stripe', logo: '💳' },
  { name: 'Square', logo: '📱' },
  { name: 'PayPal', logo: '🔵' },
  { name: 'QuickBooks', logo: '📊' },
  { name: 'Shopify', logo: '🛒' },
  { name: 'WooCommerce', logo: '🛍️' },
];

const IntegrationsSection: React.FC = () => {
  return (
    <section className={styles.integrationsSection}>
      <div className={styles.container}>
        <h2 className={styles.sectionTitle}>Integrations</h2>
        <p className={styles.sectionDescription}>
          Connect with the tools you already use
        </p>
        <div className={styles.integrationsGrid}>
          {integrations.map((integration, index) => (
            <div key={index} className={styles.integrationItem}>
              <div className={styles.integrationLogo}>{integration.logo}</div>
              <div className={styles.integrationName}>{integration.name}</div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
};

export default IntegrationsSection;

