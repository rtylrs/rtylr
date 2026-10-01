import Link from 'next/link';
import { useRouter } from 'next/router';
import { Button } from '@voxire/components';
import { ASSETS, APP_URLS, resolveAuthUrl } from '@voxire/config';
import { ManageConsentLink } from '@voxire/consent';
import React from 'react';
import { motion } from 'motion/react';
import { useTranslation } from 'react-i18next';
import { fadeInUp, staggerContainer, viewportOnce } from '../../../lib/animations';
import { navigateExternal } from '../../../lib/navigation';
import { localeFromPath, localisePath, navPath } from '../../../lib/locales';
import { track, EV_SIGNUP_CLICK, EV_CTA_CLICK } from '../../../lib/analytics';
import styles from './Footer.module.scss';

const Footer: React.FC = () => {
  const router = useRouter();
  // First t() call sites in this repo. The provider has been mounted in
  // _app.tsx since the site launched and nothing consumed it.
  const { t } = useTranslation();
  const isDemoPage = router.pathname === '/demo';
  // The legal/trust links point at real routes in this locale now, not just
  // a translated label over an English-only destination.
  const locale = localeFromPath(router.asPath).code;

  const handleSignupClick = () => {
    track(EV_SIGNUP_CLICK, { source: 'footer' });
    navigateExternal(`${resolveAuthUrl()}/register`);
  };

  const handleBookDemoClick = () => {
    track(EV_CTA_CLICK, { cta: 'demo', page: router.pathname });
    router.push('/demo');
  };

  const footerSections = [
    {
      id: 'product',
      title: t('footer.group.product'),
      links: [
        { to: navPath('/pos', locale), label: t('footer.link.pos') },
        { to: navPath('/erp', locale), label: t('footer.link.erp') },
        { to: '/hr', label: t('footer.link.hr') },
        { to: '/crm', label: t('footer.link.crm') },
        { to: '/finance', label: t('footer.link.finance') },
        { to: '/flow', label: t('footer.link.flow') },
        { to: navPath('/pricing', locale), label: t('footer.link.pricing') },
        { to: '/hardware', label: t('footer.link.hardware') },
      ],
    },
    {
      id: 'resources',
      title: t('footer.group.resources'),
      links: [
        { to: '/blog', label: t('footer.link.blog') },
        {
          to: 'https://docs.rtylr.com',
          label: t('footer.link.documentation'),
          external: true,
        },
        { to: '/integrations', label: t('footer.link.integrations') },
        { to: '/support', label: t('footer.link.helpCenter') },
      ],
    },
    {
      id: 'developers',
      title: t('footer.group.developers'),
      links: [
        {
          to: 'https://docs.rtylr.com',
          label: t('footer.link.documentation'),
          external: true,
        },
        { to: 'https://status.rtylr.com', label: t('footer.link.status'), external: true },
      ],
    },
    {
      id: 'company',
      title: t('footer.group.company'),
      links: [
        { to: '/about', label: t('footer.link.about') },
        { to: 'https://voxire.com/careers', label: t('footer.link.careers'), external: true },
        { to: '/partners', label: t('footer.link.partners') },
        { to: navPath('/contact', locale), label: t('footer.link.contact') },
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
              alt={t('footer.logoAlt')}
              className={styles['footer-logo']}
              loading="lazy"
              decoding="async"
            />
          </motion.div>

          <div className={styles['footer-nav']}>
            {footerSections.map((section) => (
              <motion.div key={section.id} className={styles['footer-section']} variants={fadeInUp}>
                {/*
                  A <p>, not an <h3>. These are labels for groups of links,
                  not sections of the document, and as headings they broke
                  the outline on every page whose main content has no <h2>:
                  /demo/ and /404/ went h1 straight to h3 purely because of
                  the footer. The nav below already carries the accessible
                  name via aria-label, so nothing is lost by demoting this.
                */}
                <p className={styles['footer-section-title']}>{section.title}</p>
                <nav aria-label={t('footer.navLabel', { group: section.title })}>
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
              {t('common.cta.startFree')}
            </Button>
            {!isDemoPage && (
              <Button onClick={handleBookDemoClick} variant="soft" className={styles['cta-button']}>
                {t('common.cta.bookDemo')}
              </Button>
            )}
          </motion.div>
        </div>

        <motion.div className={styles['footer-bottom']} variants={fadeInUp}>
          <div className={styles['footer-legal']}>
            <Link href={localisePath('/legal/terms/', locale)} className={styles['legal-link']}>
              {t('footer.legal.terms')}
            </Link>
            <Link
              href={localisePath('/legal/privacy-policy/', locale)}
              className={styles['legal-link']}
            >
              {t('footer.legal.privacy')}
            </Link>
            <Link href={localisePath('/legal/security/', locale)} className={styles['legal-link']}>
              {t('footer.legal.security')}
            </Link>
            <Link href={localisePath('/trust/', locale)} className={styles['legal-link']}>
              {t('footer.legal.trust')}
            </Link>
            <ManageConsentLink className={styles['legal-link']}>
              {t('footer.legal.cookies')}
            </ManageConsentLink>
            <a
              href="https://voxire.com"
              target="_blank"
              rel="noopener noreferrer"
              className={styles['legal-link']}
            >
              {t('footer.legal.parent')}
            </a>
          </div>
          <p className={styles.copyright}>© RTYLR {new Date().getFullYear()}</p>
        </motion.div>
      </motion.div>
    </footer>
  );
};

export default Footer;
