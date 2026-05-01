import Link from 'next/link';
import { useRouter } from 'next/router';
import { Button } from '@voxire/components';
import { ASSETS, APP_URLS } from '@voxire/config';
import { ManageConsentLink } from '@voxire/consent';
import React from 'react';
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
        { to: '/changelog', label: 'Changelog' },
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
        { to: '/integrations', label: 'Integrations' },
        { to: 'https://status.rtylr.com', label: 'Status', external: true },
      ],
    },
    {
      title: 'COMPANY',
      links: [
        { to: '/about', label: 'About' },
        { to: 'https://voxire.com/careers', label: 'Careers', external: true },
        { to: '/partners', label: 'Partners' },
        { to: '/newsroom', label: 'Newsroom' },
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
      <div className={styles['footer-container']}>
        <div className={styles['footer-top']}>
          <div className={styles['footer-logo-section']}>
            <img src={ASSETS.logos.light} alt="Rtylr Logo" className={styles['footer-logo']} />
          </div>

          <div className={styles['footer-nav']}>
            {footerSections.map((section, index) => (
              <div key={index} className={styles['footer-section']}>
                <h3 className={styles['footer-section-title']}>{section.title}</h3>
                <nav aria-label={`${section.title} navigation`}>
                  <ul className={styles['footer-links']}>
                    {section.links.map((link, linkIndex) => (
                      <li key={linkIndex}>
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
              </div>
            ))}
          </div>

          <div className={styles['footer-cta']}>
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
          </div>
        </div>

        <div className={styles['footer-bottom']}>
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
        </div>
      </div>
    </footer>
  );
};

export default Footer;
