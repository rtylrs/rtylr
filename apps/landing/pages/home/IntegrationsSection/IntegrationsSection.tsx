import React from 'react';
import { motion } from 'framer-motion';
import {
  fadeInUp,
  staggerContainer,
  scaleIn,
  viewportOnce,
} from '../../../lib/animations';
import styles from './IntegrationsSection.module.scss';

const integrations = [
  {
    name: 'Stripe',
    icon: (
      <svg viewBox="0 0 24 24" fill="currentColor" width="32" height="32">
        <path d="M13.976 9.15c-2.172-.806-3.356-1.426-3.356-2.409 0-.831.683-1.305 1.901-1.305 2.227 0 4.515.858 6.09 1.631l.89-5.494C18.252.975 15.697 0 12.165 0 9.667 0 7.589.654 6.104 1.872 4.56 3.147 3.757 4.992 3.757 7.218c0 4.039 2.467 5.76 6.476 7.219 2.585.92 3.445 1.574 3.445 2.583 0 .98-.84 1.545-2.354 1.545-1.875 0-4.965-.921-6.99-2.109l-.9 5.555C5.175 22.99 8.385 24 11.714 24c2.641 0 4.843-.624 6.328-1.813 1.664-1.305 2.525-3.236 2.525-5.732 0-4.128-2.524-5.851-6.591-7.305z" />
      </svg>
    ),
    color: '#635BFF',
  },
  {
    name: 'Square',
    icon: (
      <svg viewBox="0 0 24 24" fill="currentColor" width="32" height="32">
        <path d="M4.01 0A4.01 4.01 0 000 4.01v15.98A4.01 4.01 0 004.01 24h15.98A4.01 4.01 0 0024 19.99V4.01A4.01 4.01 0 0019.99 0zm2.04 6.72h11.9c.56 0 1.02.46 1.02 1.02v8.52c0 .56-.46 1.02-1.02 1.02H6.05c-.56 0-1.02-.46-1.02-1.02V7.74c0-.56.46-1.02 1.02-1.02zm2.87 2.75a.51.51 0 00-.51.51v4.04c0 .28.23.51.51.51h6.16c.28 0 .51-.23.51-.51V9.98a.51.51 0 00-.51-.51z" />
      </svg>
    ),
    color: '#3E4348',
  },
  {
    name: 'PayPal',
    icon: (
      <svg viewBox="0 0 24 24" fill="currentColor" width="32" height="32">
        <path d="M7.076 21.337H2.47a.641.641 0 01-.633-.74L4.944.901C5.026.382 5.474 0 5.998 0h7.46c2.57 0 4.578.543 5.69 1.81 1.01 1.15 1.304 2.42 1.012 4.287-.023.143-.047.288-.077.437-.983 5.05-4.349 6.797-8.647 6.797h-2.19c-.524 0-.968.382-1.05.9l-1.12 7.106zm14.146-14.42a3.35 3.35 0 00-.607-.541c-.013.076-.026.175-.041.254-.93 4.778-4.005 7.201-9.138 7.201h-2.19a.563.563 0 00-.556.479l-1.187 7.527h-.506l-.24 1.516a.56.56 0 00.554.647h3.882c.46 0 .85-.334.922-.788.06-.26.76-4.852.816-5.09a.932.932 0 01.92-.789h.58c3.76 0 6.705-1.528 7.565-5.946.36-1.847.174-3.388-.774-4.47z" />
      </svg>
    ),
    color: '#003087',
  },
  {
    name: 'QuickBooks',
    icon: (
      <svg viewBox="0 0 24 24" fill="currentColor" width="32" height="32">
        <path d="M12 0C5.372 0 0 5.372 0 12s5.372 12 12 12 12-5.372 12-12S18.628 0 12 0zm-1.2 17.4H8.4c-1.988 0-3.6-1.612-3.6-3.6V10.2c0-1.988 1.612-3.6 3.6-3.6h2.4v2.4H8.4c-.663 0-1.2.537-1.2 1.2v3.6c0 .663.537 1.2 1.2 1.2h2.4v2.4zm8.4-3.6c0 1.988-1.612 3.6-3.6 3.6h-2.4v-2.4h2.4c.663 0 1.2-.537 1.2-1.2v-3.6c0-.663-.537-1.2-1.2-1.2h-2.4V6.6h2.4c1.988 0 3.6 1.612 3.6 3.6v3.6z" />
      </svg>
    ),
    color: '#2CA01C',
  },
  {
    name: 'Shopify',
    icon: (
      <svg viewBox="0 0 24 24" fill="currentColor" width="32" height="32">
        <path d="M15.337 23.979l7.216-1.561s-2.604-17.613-2.625-17.73c-.018-.116-.114-.192-.211-.192s-1.929-.136-1.929-.136-1.275-1.274-1.439-1.411c-.045-.037-.075-.058-.121-.074l-.914 21.104zM11.71 11.305s-.81-.424-1.774-.424c-1.447 0-1.504.903-1.504 1.141 0 1.232 3.228 1.705 3.228 4.609 0 2.283-1.447 3.746-3.398 3.746-2.34 0-3.537-1.447-3.537-1.447l.631-2.072s1.233 1.054 2.267 1.054c.675 0 .96-.532.96-.919 0-1.608-2.648-1.68-2.648-4.344 0-2.232 1.607-4.402 4.83-4.402 1.236 0 1.85.356 1.85.356l-.905 2.702zM15.244 3.474c-.1 0-.192.048-.288.144-.77.77-1.715 3.514-1.715 3.514l-3.29 0c0 0 .192-1.824.577-3.034.108-.34-.432-.606-.432-.606S9.43 1.684 8.982 1.684c-.204 0-.48.192-.48.192C5.993 4.376 5.337 8.31 5.337 8.31l5.465 2.045s.33-2.065.495-3.127C12.02 3.45 15.052 3.474 15.244 3.474z" />
      </svg>
    ),
    color: '#96BF48',
  },
  {
    name: 'WooCommerce',
    icon: (
      <svg viewBox="0 0 24 24" fill="currentColor" width="32" height="32">
        <path d="M2.227 4.857A2.228 2.228 0 000 7.084v9.643c0 .949.6 1.759 1.441 2.073l4.722 4.33-.82-4.33h14.43A2.228 2.228 0 0022 16.573V7.084a2.228 2.228 0 00-2.227-2.227H2.227zM4.146 7.48c.384.025.672.18.862.465.456.678.688 1.942.7 3.79-.186.602-.464 1.083-.835 1.443-.335.329-.614.417-.835.265-.22-.153-.392-.558-.517-1.217-.124-.602-.166-.95-.126-1.043.04-.08.06-.19.06-.327 0-.373-.048-.886-.145-1.54-.052-.394.086-.662.416-.804.103-.042.23-.051.38-.042l.04.01zm4.486 0c.383.025.672.18.862.465.455.678.688 1.942.699 3.79-.186.602-.463 1.083-.834 1.443-.335.329-.614.417-.836.265-.22-.153-.391-.558-.516-1.217-.124-.602-.167-.95-.127-1.043.04-.08.06-.19.06-.327 0-.373-.047-.886-.145-1.54-.052-.394.087-.662.417-.804.103-.042.23-.051.38-.042l.04.01zm4.598.153c.267-.04.524.072.77.337.267.286.463.736.587 1.35l.135.667c.164.832.216 1.502.155 2.01-.082.638-.297 1.023-.648 1.154-.162.06-.348.03-.555-.093-.207-.12-.397-.34-.573-.66a6.282 6.282 0 01-.577-1.443l-.01-.04c-.164-.71-.226-1.34-.185-1.887.063-.81.32-1.29.77-1.35.043-.007.087-.012.13-.045z" />
      </svg>
    ),
    color: '#96588A',
  },
];

const IntegrationsSection: React.FC = () => {
  return (
    <section className={styles.integrationsSection}>
      <div className={styles.container}>
        <motion.div
          className={styles.header}
          initial="hidden"
          whileInView="visible"
          viewport={viewportOnce}
          variants={fadeInUp}
        >
          <span className={styles.badge}>Integrations</span>
          <h2 className={styles.sectionTitle}>
            Works with your
            <br />
            favorite tools
          </h2>
          <p className={styles.sectionDescription}>
            Connect Rtylr to the platforms and payment processors you already
            rely on. Set up in minutes, not days.
          </p>
        </motion.div>

        <motion.div
          className={styles.integrationsGrid}
          initial="hidden"
          whileInView="visible"
          viewport={viewportOnce}
          variants={staggerContainer}
        >
          {integrations.map((integration, index) => (
            <motion.div
              key={index}
              className={styles.integrationCard}
              variants={scaleIn}
              whileHover={{
                y: -6,
                transition: { duration: 0.25, ease: 'easeOut' },
              }}
            >
              <div
                className={styles.iconWrapper}
                style={
                  {
                    '--brand-color': integration.color,
                  } as React.CSSProperties
                }
              >
                {integration.icon}
              </div>
              <span className={styles.integrationName}>
                {integration.name}
              </span>
              <div className={styles.statusDot} />
            </motion.div>
          ))}
        </motion.div>

        <motion.p
          className={styles.footnote}
          initial="hidden"
          whileInView="visible"
          viewport={viewportOnce}
          variants={fadeInUp}
        >
          More integrations coming soon
        </motion.p>
      </div>
    </section>
  );
};

export default IntegrationsSection;
