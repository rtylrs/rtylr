import React, { useMemo, useState } from 'react';
import { motion } from 'motion/react';
import Link from 'next/link';
import {
  Search,
  Book,
  MessageSquare,
  Mail,
  Clock,
  CreditCard,
  Users,
  Settings,
  KeyRound,
  Zap,
  FileQuestion,
  Rocket,
} from 'lucide-react';
import AppLayout from '../../components/Layout/AppLayout';
import MainLayout from '../../layouts/Main/MainLayout';
import HelmetContainer from '../../components/HelmetContainer';
import { webPageSchema } from '../../lib/structured-data';
import { blurFadeIn, staggerBlur, viewportOnce } from '../../lib/animations';
import styles from './support.module.scss';

// ── Searchable article index ───────────────────────────────────────────────
const TOPICS = [
  {
    icon: Rocket,
    title: 'Getting started',
    description: 'Create your workspace, configure the first location, invite your team.',
    articles: [
      'Create your rtylr workspace',
      'Add your first location',
      'Invite team members and assign apps',
      'Connect your POS hardware',
    ],
  },
  {
    icon: CreditCard,
    title: 'Billing and subscriptions',
    description: 'Plans, invoices, seat proration, upgrades, and cancellation.',
    articles: [
      'How per-seat proration is calculated',
      'Upgrade from Free to Pro or Scale',
      'Download invoices and billing history',
      'Update your payment method',
      'Cancel your subscription',
    ],
  },
  {
    icon: Users,
    title: 'Team and permissions',
    description: 'Roles, app access, seat management, and invites.',
    articles: [
      'Assign roles to team members',
      'Give someone access to a specific app',
      'Remove a member and what it means for billing',
      'Set up SSO (Scale plan)',
    ],
  },
  {
    icon: Settings,
    title: 'Configuration',
    description: 'Set up each app for your business type and workflow.',
    articles: [
      'Configure POS for a restaurant',
      'Configure POS for retail',
      'Set up payroll policies in HR',
      'Customize invoice templates in Finance',
      'Build your first CRM segment',
    ],
  },
  {
    icon: Zap,
    title: 'Integrations',
    description: 'Payments, hardware, email, accounting, and API access.',
    articles: [
      'Connect Stripe for card processing',
      'Connect Square Reader hardware',
      'Export data to QuickBooks or Xero',
      'Generate an API key',
      'Webhook reference',
    ],
  },
  {
    icon: KeyRound,
    title: 'Security and account',
    description: 'Passwords, MFA, device sessions, audit logs.',
    articles: [
      'Enable multi-factor authentication',
      'Reset your password',
      'Review active sessions',
      'Download audit logs',
    ],
  },
];

const CHANNELS = [
  {
    icon: MessageSquare,
    title: 'Live chat',
    detail: 'In-app, 9am–9pm EST, Mon–Fri',
    href: null,
    cta: 'Start a chat from your workspace',
  },
  {
    icon: Mail,
    title: 'Email support',
    detail: 'We respond within 24h on weekdays',
    href: 'mailto:support@voxire.com',
    cta: 'support@voxire.com',
  },
  {
    icon: Clock,
    title: 'Status and uptime',
    detail: 'Live incidents and scheduled maintenance',
    href: 'https://status.rtylr.com',
    cta: 'status.rtylr.com',
    external: true,
  },
];

const SupportPage = (): React.JSX.Element => {
  const [query, setQuery] = useState('');

  // Flatten article list for client-side search
  const results = useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!q) return [];
    const hits: { topic: string; article: string }[] = [];
    for (const topic of TOPICS) {
      for (const article of topic.articles) {
        if (
          article.toLowerCase().includes(q) ||
          topic.title.toLowerCase().includes(q) ||
          topic.description.toLowerCase().includes(q)
        ) {
          hits.push({ topic: topic.title, article });
        }
      }
    }
    return hits.slice(0, 10);
  }, [query]);

  return (
    <AppLayout>
      <MainLayout>
        <HelmetContainer
          title="Help Center · Guides, Billing & Support | rtylr"
          description="Getting started, billing, team, configuration, integrations, security, everything you need to use rtylr."
          canonical="https://rtylr.com/support/"
          structuredData={webPageSchema({
            name: 'Help Center',
            path: '/support/',
            description:
              'Getting started, billing, team, configuration, integrations, security, everything you need to use rtylr.',
          })}
        />

        <div className={styles.page}>
          {/* HERO with search */}
          <section className={styles.hero}>
            <motion.div
              className={styles.heroInner}
              initial="hidden"
              animate="visible"
              variants={staggerBlur}
            >
              <motion.span className={styles.heroEyebrow} variants={blurFadeIn}>
                HELP CENTER
              </motion.span>
              <motion.h1 className={styles.heroTitle} variants={blurFadeIn}>
                How can we help?
              </motion.h1>
              <motion.p className={styles.heroSubtitle} variants={blurFadeIn}>
                Search our docs, or browse the topics below. Stuck? Email{' '}
                <a href="mailto:support@voxire.com">support@voxire.com</a> and we&apos;ll get back
                to you within a business day.
              </motion.p>

              <motion.div className={styles.searchBox} variants={blurFadeIn}>
                <Search size={16} strokeWidth={2} className={styles.searchIcon} />
                <input
                  type="search"
                  placeholder="Search articles, topics, or settings…"
                  value={query}
                  onChange={(e) => setQuery(e.target.value)}
                  className={styles.searchInput}
                />
                {query && (
                  <button type="button" className={styles.searchClear} onClick={() => setQuery('')}>
                    Clear
                  </button>
                )}
              </motion.div>

              {query && (
                <motion.div
                  className={styles.searchResults}
                  initial={{ opacity: 0, y: 8 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ duration: 0.2 }}
                >
                  {results.length === 0 ? (
                    <div className={styles.searchEmpty}>
                      <FileQuestion size={18} strokeWidth={1.8} />
                      No matches. Try a different term or email{' '}
                      <a href="mailto:support@voxire.com">support@voxire.com</a>.
                    </div>
                  ) : (
                    <ul>
                      {results.map((hit, i) => (
                        <li key={`${hit.topic}-${i}`}>
                          <span className={styles.resultTopic}>{hit.topic}</span>
                          <span className={styles.resultArticle}>{hit.article}</span>
                        </li>
                      ))}
                    </ul>
                  )}
                </motion.div>
              )}
            </motion.div>
          </section>

          {/* TOPICS */}
          <section className={styles.topicsSection}>
            <motion.h2
              className={styles.sectionTitle}
              initial="hidden"
              whileInView="visible"
              viewport={viewportOnce}
              variants={blurFadeIn}
            >
              Browse by topic
            </motion.h2>
            <motion.div
              className={styles.topicsGrid}
              initial="hidden"
              whileInView="visible"
              viewport={viewportOnce}
              variants={staggerBlur}
            >
              {TOPICS.map(({ icon: Icon, title, description, articles }) => (
                <motion.div key={title} className={styles.topicCard} variants={blurFadeIn}>
                  <div className={styles.topicIcon}>
                    <Icon size={20} strokeWidth={1.8} />
                  </div>
                  <h3 className={styles.topicTitle}>{title}</h3>
                  <p className={styles.topicDescription}>{description}</p>
                  <ul className={styles.topicArticles}>
                    {articles.slice(0, 4).map((article) => (
                      <li key={article}>{article}</li>
                    ))}
                  </ul>
                </motion.div>
              ))}
            </motion.div>
          </section>

          {/* CONTACT CHANNELS */}
          <motion.section
            className={styles.channelsSection}
            initial="hidden"
            whileInView="visible"
            viewport={viewportOnce}
            variants={staggerBlur}
          >
            <motion.div className={styles.channelsHeader} variants={blurFadeIn}>
              <span className={styles.sectionEyebrow}>CAN&apos;T FIND IT?</span>
              <h2 className={styles.sectionTitle}>Three ways to reach us.</h2>
            </motion.div>
            <motion.div className={styles.channelsGrid} variants={blurFadeIn}>
              {CHANNELS.map(({ icon: Icon, title, detail, href, cta, external }) => (
                <div key={title} className={styles.channelCard}>
                  <Icon size={18} strokeWidth={1.8} className={styles.channelIcon} />
                  <div className={styles.channelTitle}>{title}</div>
                  <div className={styles.channelDetail}>{detail}</div>
                  {href ? (
                    <a
                      href={href}
                      className={styles.channelCta}
                      target={external ? '_blank' : undefined}
                      rel={external ? 'noopener noreferrer' : undefined}
                    >
                      {cta} →
                    </a>
                  ) : (
                    <span className={styles.channelCta}>{cta}</span>
                  )}
                </div>
              ))}
            </motion.div>
          </motion.section>

          {/* DOCS LINK */}
          <motion.section
            className={styles.docsCta}
            initial="hidden"
            whileInView="visible"
            viewport={viewportOnce}
            variants={staggerBlur}
          >
            <motion.div className={styles.docsInner} variants={blurFadeIn}>
              <Book size={22} strokeWidth={1.8} className={styles.docsIcon} />
              <h2 className={styles.docsTitle}>Looking for developer docs?</h2>
              <p className={styles.docsCopy}>
                Full API reference, webhook events, and SDK usage, at{' '}
                <a href="https://docs.rtylr.com" target="_blank" rel="noopener noreferrer">
                  docs.rtylr.com
                </a>
                . Public for everyone, no login required.
              </p>
              <Link href="/contact" className={styles.docsLinkBtn}>
                Still need a human? Contact us →
              </Link>
            </motion.div>
          </motion.section>
        </div>
      </MainLayout>
    </AppLayout>
  );
};

export default SupportPage;
