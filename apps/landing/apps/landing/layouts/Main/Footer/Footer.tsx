import Link from 'next/link';
import { useRouter } from 'next/router';
import { Button } from '@voxire/components';
import { ASSETS, APP_URLS } from '@voxire/config';
import { ManageConsentLink } from '@voxire/consent';
import React from 'react';
import { motion } from 'motion/react';
import { fadeInUp, staggerContainer, viewportOnce } from '../../../lib/animations';
import { navigateExternal } from '../../../lib/navigation';
import styles from './Footer.module.scss';

const Footer: React.FC = () => {
  const router = useRouter();
  const isDemoPage = router.pathname === '/demo';

  const handleSignupClick = () => {
    navigateExternal(`${APP_URLS.auth}/register`);
  };

  const handleBookDemoClick = () => {
    router.push('/demo');
  };

  const footerSections = [
    {
      title: 'PRODUCT',
      links: [
        { to: '/pos', label: 'POS' },
        { to: '/erp', label: 'ERP' },
        { to: '/hr', label: 'HR' },
        { to: '/crm', label: 'CRM' },
        { to: '/finance', label: 'Finance' },
        { to: '/flow', label: 'Flow' },
        { to: '/pricing', label: 'Pricing' },
        { to: '/hardware', label: 'Hardware' },
      ],
    },
    {
      title: 'RESOURCES',
      links: [
        { to: '/blog', label: 'Blog' },
        {
          to: 'https://docs.rtylr.com',
          label: 'Documentation',
          external: true,
        },
        { to: '/integrations', label: 'Integrations' },
        { to: '/support', label: 'Help Center' },
        { to: 'https://docs.rtylr.com/changelog', label: 'Changelog', external: true },
      ],
    },
    {
      title: 'DEVELOPERS',
      links: [
        {
          to: 'https://docs.rtylr.com',
          label: 'Documentation',
          external: true,
        },
        {
          to: 'https://docs.rtylr.com/api',
          label: 'API Reference',
          external: true,
        },
        { to: 'https://status.rtylr.com', label: 'Status', external: true },
      ],
    },
    {
      title: 'COMPANY',
      links: [
        { to: '/about', label: 'About' },
        { to: 'https://voxire.com/careers', label: 'Careers', external: true },
        { to: '/partners', label: 'Partners' },
        { to: '/contact', label: 'Contact' },
      ],
    },
  ];

  const socialLinks: Array<{
    href: string;
    icon: React.ComponentType;
    label: string;
  }> = [];

  return (
    <footer className={styles.footer} role="contentinfo">
      <motion.div
        className={styles['footer-container']}
        initial="hidden"
        whileInView="visible"
        viewport={viewportOnce}
        variants={staggerContainer}
      >
        <div className={styles['footer-top']}>
          <motion.div className={styles['footer-logo-section']} variants={fadeInUp}>
            <img
              src={ASSETS.logos.light}
              alt="Rtylr Logo"
              className={styles['footer-logo']}
              loading="lazy"
              decoding="async"
            />
          </motion.div>

          <div className={styles['footer-nav']}>
            {footerSections.map((section) => (
              <motion.div
                key={section.title}
                className={styles['footer-section']}
                variants={fadeInUp}
              >
                <h3 className={styles['footer-section-title']}>{section.title}</h3>
                <nav aria-label={`${section.title} navigation`}>
                  <ul className={styles['footer-links']}>
                    {section.links.map((link) => (
                      <li key={link.to}>
                        {link.external ? (
                          <a
                            href={link.to}
                            target="_blank"
                            rel="noopener noreferrer"
                            className={styles['footer-link']}
                          >
                            {link.label}
                          </a>
                        ) : (
                          <Link href={link.to} className={styles['footer-link']}>
                            {link.label}
                          </Link>
                        )}
                      </li>
                    ))}
                  </ul>
                </nav>
              </motion.div>
            ))}
          </div>

          <motion.div className={styles['footer-cta']} variants={fadeInUp}>
            <Button
              onClick={handleSignupClick}
              variant="dark-outline"
              className={styles['cta-button']}
            >
              Start for free
            </Button>
            {!isDemoPage && (
              <Button onClick={handleBookDemoClick} variant="soft" className={styles['cta-button']}>
                Book a demo
              </Button>
            )}
          </motion.div>
        </div>

        <motion.div className={styles['footer-bottom']} variants={fadeInUp}>
          <div className={styles['footer-legal']}>
            <Link href="/legal/terms" className={styles['legal-link']}>
              TERMS OF USE
            </Link>
            <Link href="/legal/privacy-policy" className={styles['legal-link']}>
              PRIVACY POLICY
            </Link>
            <Link href="/legal/security" className={styles['legal-link']}>
              SECURITY
            </Link>
            <Link href="/trust" className={styles['legal-link']}>
              TRUST CENTER
            </Link>
            <ManageConsentLink className={styles['legal-link']}>
              COOKIE PREFERENCES
            </ManageConsentLink>
            <a
              href="https://voxire.com"
              target="_blank"
              rel="noopener noreferrer"
              className={styles['legal-link']}
            >
              A VOXIRE COMPANY
            </a>
          </div>
          <p className={styles.copyright}>© RTYLR {new Date().getFullYear()}</p>
        </motion.div>
      </motion.div>
    </footer>
  );
};

export default Footer;
