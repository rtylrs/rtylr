import React, { useMemo, useState } from 'react';
import { motion } from 'motion/react';
import { useTranslation } from 'react-i18next';
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
// Icons only. Titles, descriptions and article lists resolve from
// support.topics, matched by position.
const TOPIC_ICONS = [Rocket, CreditCard, Users, Settings, Zap, KeyRound];

// Icon, href and target only. Title, detail and cta resolve from
// support.channels, matched by position. One channel has no href: live
// chat opens inside the workspace rather than at a URL.
const CHANNEL_META: { icon: typeof Mail; href: string | null; external?: boolean }[] = [
  { icon: MessageSquare, href: null },
  { icon: Mail, href: 'mailto:support@voxire.com' },
  { icon: Clock, href: 'https://status.rtylr.com', external: true },
];

const SupportPage = (): React.JSX.Element => {
  const { t } = useTranslation();
  const topics = t('support.topics', { returnObjects: true }) as unknown as {
    title: string;
    description: string;
    articles: string[];
  }[];
  const channels = t('support.channels', { returnObjects: true }) as unknown as {
    title: string;
    detail: string;
    cta: string;
  }[];
  const [query, setQuery] = useState('');

  // Flatten article list for client-side search
  const results = useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!q) return [];
    const hits: { topic: string; article: string }[] = [];
    for (const topic of topics) {
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
  }, [query, topics]);

  return (
    <AppLayout>
      <MainLayout>
        <HelmetContainer
          title={t('support.meta.title')}
          description={t('support.meta.description')}
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
              {t('support.hero.eyebrow')}
            </motion.span>
              <motion.h1 className={styles.heroTitle} variants={blurFadeIn}>
              {t('support.hero.title')}
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
                  placeholder={t('support.hero.searchPlaceholder')}
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
              {t('support.topicsTitle')}
            </motion.h2>
            <motion.div
              className={styles.topicsGrid}
              initial="hidden"
              whileInView="visible"
              viewport={viewportOnce}
              variants={staggerBlur}
            >
              {TOPIC_ICONS.map((Icon, ti) => (
                <motion.div key={topics[ti].title} className={styles.topicCard} variants={blurFadeIn}>
                  <div className={styles.topicIcon}>
                    <Icon size={20} strokeWidth={1.8} />
                  </div>
                  <h3 className={styles.topicTitle}>{topics[ti].title}</h3>
                  <p className={styles.topicDescription}>{topics[ti].description}</p>
                  <ul className={styles.topicArticles}>
                    {topics[ti].articles.slice(0, 4).map((article) => (
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
              <span className={styles.sectionEyebrow}>{t('support.channelsSection.eyebrow')}</span>
              <h2 className={styles.sectionTitle}>{t('support.channelsSection.title')}</h2>
            </motion.div>
            <motion.div className={styles.channelsGrid} variants={blurFadeIn}>
              {CHANNEL_META.map(({ icon: Icon, href, external }, ci) => (
                <div key={channels[ci].title} className={styles.channelCard}>
                  <Icon size={18} strokeWidth={1.8} className={styles.channelIcon} />
                  <div className={styles.channelTitle}>{channels[ci].title}</div>
                  <div className={styles.channelDetail}>{channels[ci].detail}</div>
                  {href ? (
                    <a
                      href={href}
                      className={styles.channelCta}
                      target={external ? '_blank' : undefined}
                      rel={external ? 'noopener noreferrer' : undefined}
                    >
                      {channels[ci].cta} →
                    </a>
                  ) : (
                    <span className={styles.channelCta}>{channels[ci].cta}</span>
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
              <h2 className={styles.docsTitle}>{t('support.docs.title')}</h2>
              <p className={styles.docsCopy}>
                Full API reference, webhook events, and SDK usage, at{' '}
                <a href="https://docs.rtylr.com" target="_blank" rel="noopener noreferrer">
                  docs.rtylr.com
                </a>
                . Public for everyone, no login required.
              </p>
              <Link href="/contact" className={styles.docsLinkBtn}>
              {t('support.docs.link')}
            </Link>
            </motion.div>
          </motion.section>
        </div>
      </MainLayout>
    </AppLayout>
  );
};

export default SupportPage;
