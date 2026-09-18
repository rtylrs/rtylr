import React, { useMemo, useState } from 'react';
import { cn } from '../../lib/cn';
import { motion } from 'motion/react';
import { useTranslation } from 'react-i18next';
import { useRouter } from 'next/router';
import {
  CreditCard,
  Truck,
  Mail,
  Calendar,
  BookOpen,
  Globe,
  Printer,
  Link as LinkIcon,
  Search,
} from 'lucide-react';
import AppLayout from '../../components/Layout/AppLayout';
import MainLayout from '../../layouts/Main/MainLayout';
import HelmetContainer from '../../components/HelmetContainer';
import { itemListSchema } from '../../lib/structured-data';
import { APP_URLS } from '@voxire/config';
import { navigateExternal } from '../../lib/navigation';
import { blurFadeIn, staggerBlur, viewportOnce } from '../../lib/animations';
import Lines from '../../components/text/Lines';
import styles from './integrations.module.scss';

// ── Integration catalog ───────────────────────────────────────────────────
interface Integration {
  name: string;
  category: 'Payments' | 'Accounting' | 'Delivery' | 'Comms' | 'Hardware' | 'Productivity';
  summary: string;
  status: 'live' | 'beta' | 'soon';
}

const INTEGRATIONS: Integration[] = [
  {
    name: 'Stripe',
    category: 'Payments',
    summary: 'Card processing, Apple Pay, Google Pay, ACH.',
    status: 'soon',
  },
  {
    name: 'Square',
    category: 'Payments',
    summary: 'In-person and card-present processing.',
    status: 'soon',
  },
  {
    name: 'Apple Pay',
    category: 'Payments',
    summary: 'Native tap-to-pay at checkout.',
    status: 'soon',
  },
  { name: 'Google Pay', category: 'Payments', summary: 'Mobile wallet payments.', status: 'soon' },
  {
    name: 'QuickBooks',
    category: 'Accounting',
    summary: 'Nightly chart-of-accounts export.',
    status: 'soon',
  },
  {
    name: 'Xero',
    category: 'Accounting',
    summary: 'Bi-directional bridge for small business.',
    status: 'soon',
  },
  {
    name: 'Sage',
    category: 'Accounting',
    summary: 'Journal-entry export on a schedule.',
    status: 'soon',
  },
  {
    name: 'Uber Eats',
    category: 'Delivery',
    summary: 'Menu sync, order intake, ticket firing.',
    status: 'soon',
  },
  {
    name: 'DoorDash',
    category: 'Delivery',
    summary: 'Orders flow straight into POS.',
    status: 'soon',
  },
  {
    name: 'Deliveroo',
    category: 'Delivery',
    summary: 'EU and UK delivery marketplace.',
    status: 'soon',
  },
  {
    name: 'Mailchimp',
    category: 'Comms',
    summary: 'Export CRM segments for email campaigns.',
    status: 'soon',
  },
  {
    name: 'Resend',
    category: 'Comms',
    summary: 'Transactional email relay for receipts.',
    status: 'soon',
  },
  {
    name: 'Twilio SMS',
    category: 'Comms',
    summary: 'Send reservation reminders and alerts.',
    status: 'soon',
  },
  {
    name: 'Slack',
    category: 'Comms',
    summary: 'Route critical alerts to a channel.',
    status: 'soon',
  },
  {
    name: 'Google Workspace',
    category: 'Comms',
    summary: 'SSO plus calendar sync for shifts.',
    status: 'soon',
  },
  {
    name: 'Microsoft 365',
    category: 'Comms',
    summary: 'SSO and identity for Scale customers.',
    status: 'soon',
  },
  {
    name: 'Star Micronics',
    category: 'Hardware',
    summary: 'Thermal receipt printers.',
    status: 'soon',
  },
  {
    name: 'Epson',
    category: 'Hardware',
    summary: 'Kitchen and receipt printer series.',
    status: 'soon',
  },
  {
    name: 'Zebra',
    category: 'Hardware',
    summary: 'Barcode scanners and label printers.',
    status: 'soon',
  },
  { name: 'Ingenico', category: 'Hardware', summary: 'EMV card readers.', status: 'soon' },
  {
    name: 'OpenTable',
    category: 'Productivity',
    summary: 'Reservations flow into your host stand.',
    status: 'soon',
  },
  {
    name: 'Resy',
    category: 'Productivity',
    summary: 'Reservations and guest data.',
    status: 'soon',
  },
  {
    name: 'Notion',
    category: 'Productivity',
    summary: 'Task sync with Flow for cross-tool teams.',
    status: 'soon',
  },
  {
    name: 'Zapier',
    category: 'Productivity',
    summary: 'Trigger any app from 6,000+ services.',
    status: 'soon',
  },
];

const CATEGORIES: {
  id: Integration['category'] | 'all';
  key: string;
  icon: React.FC<React.SVGProps<SVGSVGElement>>;
}[] = [
  // The id is the filter value and matches Integration['category']; the
  // visible label comes from integrations.categories.<key>.
  { id: 'all', key: 'all', icon: Globe },
  { id: 'Payments', key: 'payments', icon: CreditCard },
  { id: 'Accounting', key: 'accounting', icon: BookOpen },
  { id: 'Delivery', key: 'delivery', icon: Truck },
  { id: 'Comms', key: 'comms', icon: Mail },
  { id: 'Hardware', key: 'hardware', icon: Printer },
  { id: 'Productivity', key: 'productivity', icon: Calendar },
];

const IntegrationsPage = (): React.JSX.Element => {
  const { t } = useTranslation();
  const router = useRouter();
  const [query, setQuery] = useState('');
  const [activeCategory, setActiveCategory] = useState<(typeof CATEGORIES)[number]['id']>('all');

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    return INTEGRATIONS.filter((i) => {
      if (activeCategory !== 'all' && i.category !== activeCategory) return false;
      if (!q) return true;
      return (
        i.name.toLowerCase().includes(q) ||
        i.summary.toLowerCase().includes(q) ||
        i.category.toLowerCase().includes(q)
      );
    });
  }, [query, activeCategory]);

  const goRegister = (): void => {
    navigateExternal(`${APP_URLS.auth}/register`);
  };

  const goContact = (): void => {
    void router.push('/contact');
  };

  return (
    <AppLayout>
      <MainLayout>
        <HelmetContainer
          title={t('integrations.meta.title')}
          description={t('integrations.meta.description')}
          canonical="https://rtylr.com/integrations/"
          structuredData={itemListSchema({
            name: 'Integrations',
            path: '/integrations/',
            description:
              'Every service rtylr connects to, payments, accounting, delivery, comms, hardware, and productivity.',
            // Generated from the array the page renders, so the list in
            // the markup and the list on screen cannot drift apart.
            items: INTEGRATIONS.map((i) => i.name),
          })}
        />

        <div className={styles.page}>
          {/* HERO */}
          <section className={styles.hero}>
            <motion.div
              className={styles.heroInner}
              initial="hidden"
              animate="visible"
              variants={staggerBlur}
            >
              <motion.span className={styles.heroEyebrow} variants={blurFadeIn}>
                <LinkIcon size={13} strokeWidth={2} />
                {t('integrations.hero.eyebrow')}
              </motion.span>
              <motion.h1 className={styles.heroTitle} variants={blurFadeIn}>
                <Lines text={t('integrations.hero.title')} />
              </motion.h1>
              <motion.p className={styles.heroSubtitle} variants={blurFadeIn}>
                {t('integrations.hero.subtitle')}
              </motion.p>

              <motion.div className={styles.searchRow} variants={blurFadeIn}>
                <div className={styles.searchBox}>
                  <Search size={16} strokeWidth={2} className={styles.searchIcon} />
                  <input
                    type="search"
                    placeholder={t('integrations.hero.searchPlaceholder')}
                    value={query}
                    onChange={(e) => setQuery(e.target.value)}
                    className={styles.searchInput}
                  />
                </div>
              </motion.div>

              <motion.div className={styles.categoryTabs} variants={blurFadeIn}>
                {CATEGORIES.map(({ id, key, icon: Icon }) => (
                  <button
                    key={id}
                    type="button"
                    className={cn(
                      styles.categoryTab,
                      activeCategory === id && styles.categoryTabActive
                    )}
                    onClick={() => setActiveCategory(id)}
                  >
                    <Icon width={14} height={14} strokeWidth={1.8} />
                    {t(`integrations.categories.${key}`)}
                  </button>
                ))}
              </motion.div>
            </motion.div>
          </section>

          {/* GRID */}
          <section className={styles.gridSection}>
            <motion.div
              className={styles.grid}
              initial="hidden"
              whileInView="visible"
              viewport={viewportOnce}
              variants={staggerBlur}
              key={`${activeCategory}-${query}`}
            >
              {filtered.length === 0 ? (
                <div className={styles.empty}>
                  No integrations match. Try clearing the filter, or email{' '}
                  <a href="mailto:support@voxire.com">support@voxire.com</a> to request one.
                </div>
              ) : (
                filtered.map((integration) => (
                  <motion.div
                    key={integration.name}
                    className={styles.card}
                    variants={blurFadeIn}
                    whileHover={{ y: -3 }}
                  >
                    <div className={styles.cardHeader}>
                      <div className={styles.cardLogo}>
                        {integration.name
                          .split(' ')
                          .map((w) => w[0])
                          .slice(0, 2)
                          .join('')}
                      </div>
                      <span
                        className={cn(styles.statusPill, styles[`status_${integration.status}`])}
                      >
                        {t(`integrations.status.${integration.status}`)}
                      </span>
                    </div>
                    <div className={styles.cardName}>{integration.name}</div>
                    <div className={styles.cardCategory}>{integration.category}</div>
                    <p className={styles.cardSummary}>{integration.summary}</p>
                  </motion.div>
                ))
              )}
            </motion.div>
          </section>

          {/* API CTA */}
          <motion.section
            className={styles.apiSection}
            initial="hidden"
            whileInView="visible"
            viewport={viewportOnce}
            variants={staggerBlur}
          >
            <motion.div className={styles.apiInner} variants={blurFadeIn}>
              <div className={styles.apiSide}>
                <span className={styles.sectionEyebrow}>{t('integrations.api.eyebrow')}</span>
                <h2 className={styles.apiTitle}>{t('integrations.api.title')}</h2>
                <p className={styles.apiCopy}>
                  {t('integrations.api.copy')}
                </p>
                <div className={styles.apiLinks}>
                  <a
                    href="https://docs.rtylr.com"
                    target="_blank"
                    rel="noopener noreferrer"
                    className={styles.apiPrimary}
                  >
                    {t('integrations.api.docs')}
                  </a>
                  <button type="button" onClick={goContact} className={styles.apiSecondary}>
                    {t('integrations.api.contact')}
                  </button>
                </div>
              </div>
              <div className={styles.apiSample} aria-hidden="true">
                <div className={styles.apiSampleHead}>
                  <span className={styles.apiSampleMethod}>POST</span>
                  <span className={styles.apiSampleUrl}>/v1/orders</span>
                </div>
                <pre className={styles.apiSampleBody}>{`{
  "location_id": "loc_terrace",
  "items": [
    { "sku": "PIZ-MARG", "qty": 1 },
    { "sku": "WIN-RED-02", "qty": 1 }
  ],
  "customer": { "email": "m@rtylr.com" },
  "source": "integrations.delivery.uber"
}`}</pre>
              </div>
            </motion.div>
          </motion.section>

          {/* BOTTOM CTA */}
          <motion.section
            className={styles.ctaBlock}
            initial="hidden"
            whileInView="visible"
            viewport={viewportOnce}
            variants={staggerBlur}
          >
            <motion.h2 className={styles.ctaTitle} variants={blurFadeIn}>
              Connect everything
              <br />
              you already run on.
            </motion.h2>
            <motion.div className={styles.ctaButtons} variants={blurFadeIn}>
              <button type="button" onClick={goRegister} className={styles.ctaPrimary}>
                Get started
              </button>
              <button type="button" onClick={goContact} className={styles.ctaSecondary}>
                Request an integration
              </button>
            </motion.div>
          </motion.section>
        </div>
      </MainLayout>
    </AppLayout>
  );
};

export default IntegrationsPage;
