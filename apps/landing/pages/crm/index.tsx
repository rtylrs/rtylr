import React, { useMemo, useState } from 'react';
import { cn } from '../../lib/cn';
import { motion } from 'motion/react';
import { useTranslation } from 'react-i18next';
import { useRouter } from 'next/router';
import { Heart, Mail, Gift, Users, Star, Send, Target, ChevronRight, Plus } from 'lucide-react';
import AppLayout from '../../components/Layout/AppLayout';
import MainLayout from '../../layouts/Main/MainLayout';
import HelmetContainer from '../../components/HelmetContainer';
import DirectAnswer from '../../components/DirectAnswer/DirectAnswer';
import FaqSection from '../../components/FaqSection/FaqSection';
import { productWithBreadcrumbs } from '../../lib/structured-data';
import { useFaq } from '../../lib/faqs';
import { APP_URLS } from '@voxire/config';
import { navigateExternal } from '../../lib/navigation';
import { blurFadeIn, staggerBlur, viewportOnce } from '../../lib/animations';
import { SplitWords } from '../../components/motion/SplitWords';
import Lines from '../../components/text/Lines';
import styles from './crm.module.scss';

const VISITS = [
  { date: 'Apr 14', total: '$48.00', note: 'Terrace · 2 guests' },
  { date: 'Apr 08', total: '$112.50', note: 'Terrace · birthday' },
  { date: 'Mar 29', total: '$64.00', note: 'Take-out' },
  { date: 'Mar 17', total: '$28.00', note: 'Lunch · solo' },
];

const SEGMENTS = [
  { name: 'VIP · monthly+', size: 342, color: '#d6b88e' },
  { name: 'Birthday month', size: 1204, color: '#c9a48a' },
  { name: 'Lapsed 60+ days', size: 189, color: '#b89a8a' },
  { name: 'First-time', size: 88, color: '#a89680' },
  { name: 'High-tippers', size: 521, color: '#d6b88e' },
  { name: 'Takeout only', size: 412, color: '#c9a48a' },
];

// ── Interactive segment builder data ──────────────────────────────────────
// Each mock guest has the attributes the filter toggles operate on. The
// audience list updates live as the visitor plays with the checkboxes.

interface Guest {
  initials: string;
  name: string;
  avatar: string;
  visits: number;
  lifetime: number;
  birthdayThisMonth: boolean;
  highTipper: boolean;
  takeoutOnly: boolean;
  lapsed60: boolean;
  vip: boolean;
  lastVisit: string;
}

const GUESTS: Guest[] = [
  {
    initials: 'EA',
    name: 'Eliza Alvarez',
    avatar: '#d6b88e',
    visits: 24,
    lifetime: 1284,
    birthdayThisMonth: false,
    highTipper: true,
    takeoutOnly: false,
    lapsed60: false,
    vip: true,
    lastVisit: 'Apr 14',
  },
  {
    initials: 'JP',
    name: 'Juno Park',
    avatar: '#c9a48a',
    visits: 11,
    lifetime: 640,
    birthdayThisMonth: true,
    highTipper: false,
    takeoutOnly: false,
    lapsed60: false,
    vip: false,
    lastVisit: 'Apr 09',
  },
  {
    initials: 'MH',
    name: 'Marcus Hale',
    avatar: '#b89a8a',
    visits: 42,
    lifetime: 3120,
    birthdayThisMonth: false,
    highTipper: true,
    takeoutOnly: false,
    lapsed60: false,
    vip: true,
    lastVisit: 'Apr 12',
  },
  {
    initials: 'SN',
    name: 'Sasha Nomura',
    avatar: '#a89680',
    visits: 3,
    lifetime: 98,
    birthdayThisMonth: false,
    highTipper: false,
    takeoutOnly: true,
    lapsed60: false,
    vip: false,
    lastVisit: 'Apr 16',
  },
  {
    initials: 'DL',
    name: 'Dalia Lachman',
    avatar: '#d6b88e',
    visits: 18,
    lifetime: 980,
    birthdayThisMonth: true,
    highTipper: true,
    takeoutOnly: false,
    lapsed60: false,
    vip: true,
    lastVisit: 'Apr 03',
  },
  {
    initials: 'RK',
    name: 'Rashid Karim',
    avatar: '#c9a48a',
    visits: 6,
    lifetime: 240,
    birthdayThisMonth: false,
    highTipper: false,
    takeoutOnly: true,
    lapsed60: true,
    vip: false,
    lastVisit: 'Feb 08',
  },
  {
    initials: 'TO',
    name: 'Tamsin Okoye',
    avatar: '#b89a8a',
    visits: 29,
    lifetime: 1850,
    birthdayThisMonth: false,
    highTipper: true,
    takeoutOnly: false,
    lapsed60: false,
    vip: true,
    lastVisit: 'Apr 13',
  },
  {
    initials: 'PV',
    name: 'Priya Venkat',
    avatar: '#a89680',
    visits: 2,
    lifetime: 62,
    birthdayThisMonth: true,
    highTipper: false,
    takeoutOnly: true,
    lapsed60: false,
    vip: false,
    lastVisit: 'Apr 11',
  },
  {
    initials: 'LQ',
    name: 'Leo Quintana',
    avatar: '#d6b88e',
    visits: 9,
    lifetime: 410,
    birthdayThisMonth: false,
    highTipper: false,
    takeoutOnly: false,
    lapsed60: true,
    vip: false,
    lastVisit: 'Feb 01',
  },
  {
    initials: 'AR',
    name: 'Asha Rai',
    avatar: '#c9a48a',
    visits: 36,
    lifetime: 2410,
    birthdayThisMonth: true,
    highTipper: true,
    takeoutOnly: false,
    lapsed60: false,
    vip: true,
    lastVisit: 'Apr 15',
  },
  {
    initials: 'BC',
    name: 'Ben Cortez',
    avatar: '#b89a8a',
    visits: 14,
    lifetime: 720,
    birthdayThisMonth: false,
    highTipper: false,
    takeoutOnly: false,
    lapsed60: false,
    vip: false,
    lastVisit: 'Apr 10',
  },
  {
    initials: 'NW',
    name: 'Nadia Weiss',
    avatar: '#a89680',
    visits: 4,
    lifetime: 138,
    birthdayThisMonth: false,
    highTipper: false,
    takeoutOnly: true,
    lapsed60: true,
    vip: false,
    lastVisit: 'Jan 28',
  },
];

interface FilterDef {
  id: string;
  label: string;
  test: (g: Guest) => boolean;
}

const FILTERS: FilterDef[] = [
  { id: 'vip', label: 'VIP (20+ visits)', test: (g) => g.vip },
  { id: 'birthday', label: 'Birthday this month', test: (g) => g.birthdayThisMonth },
  { id: 'highTipper', label: 'High tipper', test: (g) => g.highTipper },
  { id: 'lapsed60', label: 'Lapsed 60+ days', test: (g) => g.lapsed60 },
  { id: 'takeoutOnly', label: 'Takeout only', test: (g) => g.takeoutOnly },
  { id: 'bigSpend', label: 'Lifetime spend > $500', test: (g) => g.lifetime > 500 },
];

const CAMPAIGNS_ICONS = [Mail, Gift, Target, Send];

const CRMPage = (): React.JSX.Element => {
  const faq = useFaq('crm');
  const { t } = useTranslation();
  const campaigns = t('crm.campaigns.items', { returnObjects: true }) as unknown as {
    title: string;
    copy: string;
  }[];
  const router = useRouter();

  // Segment-builder state, toggle attributes to watch the audience shrink/grow
  const [activeFilters, setActiveFilters] = useState<Set<string>>(new Set(['vip']));

  const toggleFilter = (id: string): void => {
    setActiveFilters((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  };

  const matchingGuests = useMemo(() => {
    if (activeFilters.size === 0) return GUESTS;
    const active = FILTERS.filter((f) => activeFilters.has(f.id));
    return GUESTS.filter((g) => active.every((f) => f.test(g)));
  }, [activeFilters]);

  const goRegister = (): void => {
    navigateExternal(`${APP_URLS.auth}/register?redirect=crm`);
  };

  const goDemo = (): void => {
    void router.push('/demo');
  };

  return (
    <AppLayout>
      <MainLayout>
        <HelmetContainer
          title={t('crm.meta.title')}
          description={t('crm.meta.description')}
          canonical="https://rtylr.com/crm/"
          structuredData={productWithBreadcrumbs({
            name: 'CRM',
            path: '/crm/',
            subCategory: 'Customer Relationship Management',
            description: t('crm.meta.schemaDescription'),
          })}
        />

        <div className={styles.page}>
          {/* ── HERO, copy left, profile card right ────────────────────── */}
          <section className={styles.hero}>
            <motion.div
              className={styles.heroCopy}
              initial="hidden"
              animate="visible"
              variants={staggerBlur}
            >
              <motion.span className={styles.heroEyebrow} variants={blurFadeIn}>
                <Heart size={13} strokeWidth={2} fill="currentColor" />
                {t('crm.hero.eyebrow')}
              </motion.span>
              <h1 className={styles.heroTitle}>
                <SplitWords as="span" text={t('crm.hero.titleA')} delay={0.1} />
                <br />
                <span className={styles.heroTitleSoft}>
                  <SplitWords as="span" text={t('crm.hero.titleB')} delay={0.35} />
                </span>
              </h1>
              <motion.p className={styles.heroSubtitle} variants={blurFadeIn}>
                {t('crm.hero.subtitle')}
              </motion.p>
              <motion.div className={styles.heroCtas} variants={blurFadeIn}>
                <button type="button" onClick={goRegister} className={styles.primaryBtn}>
                  Get started
                </button>
                <button type="button" onClick={goDemo} className={styles.outlineBtn}>
                  Book a demo
                </button>
              </motion.div>
            </motion.div>

            {/* Customer profile card mockup */}
            <motion.div
              className={styles.profileMock}
              initial={{ opacity: 0, y: 30, rotate: -1 }}
              animate={{ opacity: 1, y: 0, rotate: 0 }}
              transition={{ delay: 0.4, duration: 1.0, ease: [0.16, 1, 0.3, 1] }}
              aria-hidden="true"
            >
              <div className={styles.profileHeader}>
                <div
                  className={styles.profileAvatar}
                  style={{ background: 'linear-gradient(135deg, #d6b88e, #c9a48a)' }}
                >
                  EA
                </div>
                <div className={styles.profileMeta}>
                  <div className={styles.profileName}>Eliza Alvarez</div>
                  <div className={styles.profileHandle}>eliza.a@hey.com · +1 (512) ••• ••92</div>
                </div>
                <span className={styles.profileVip}>
                  <Star size={10} fill="currentColor" strokeWidth={0} />
                  VIP
                </span>
              </div>

              <div className={styles.profileStatsRow}>
                <div className={styles.profileStat}>
                  <span className={styles.profileStatValue}>24</span>
                  <span className={styles.profileStatLabel}>visits</span>
                </div>
                <div className={styles.profileStat}>
                  <span className={styles.profileStatValue}>$1,284</span>
                  <span className={styles.profileStatLabel}>lifetime</span>
                </div>
                <div className={styles.profileStat}>
                  <span className={styles.profileStatValue}>$53</span>
                  <span className={styles.profileStatLabel}>avg check</span>
                </div>
                <div className={styles.profileStat}>
                  <span className={styles.profileStatValue}>1,840</span>
                  <span className={styles.profileStatLabel}>points</span>
                </div>
              </div>

              <div className={styles.profileDivider}>RECENT VISITS</div>

              <ul className={styles.profileVisits}>
                {VISITS.map((v) => (
                  <li key={v.date}>
                    <span className={styles.visitDate}>{v.date}</span>
                    <span className={styles.visitNote}>{v.note}</span>
                    <span className={styles.visitTotal}>{v.total}</span>
                  </li>
                ))}
              </ul>

              <div className={styles.profileTags}>
                <span className={styles.profileTag}>VIP · monthly+</span>
                <span className={styles.profileTag}>Vegetarian</span>
                <span className={styles.profileTag}>Birthday: Oct</span>
                <span className={styles.profileTag}>High-tipper</span>
              </div>
            </motion.div>
          </section>

          <DirectAnswer label={t('crm.directAnswer.label')}>
            {t('crm.directAnswer.body')}
          </DirectAnswer>

          {/* ── SEGMENT BUBBLES ─────────────────────────────────────────── */}
          <motion.section
            className={styles.segmentsSection}
            initial="hidden"
            whileInView="visible"
            viewport={viewportOnce}
            variants={staggerBlur}
          >
            <div className={styles.segmentsHeader}>
              <motion.span className={styles.sectionEyebrow} variants={blurFadeIn}>
                LIVE SEGMENTS
              </motion.span>
              <motion.h2 className={styles.sectionTitle} variants={blurFadeIn}>
                How do you segment
                <br />
                your customer base?
              </motion.h2>
              <motion.p className={styles.sectionCopy} variants={blurFadeIn}>
                Build segments by visit velocity, lifetime spend, last-order date, tags, preferences
                , any combination. Re-run automatically as new orders land.
              </motion.p>
            </div>

            <motion.div className={styles.segmentBubbles} variants={blurFadeIn}>
              {SEGMENTS.map((seg) => {
                // Size based on segment population, with a min and max so the UI
                // stays readable across wildly different audience sizes.
                const scale = Math.min(1, Math.max(0.45, seg.size / 1400));
                const size = 90 + scale * 110;
                return (
                  <div
                    key={seg.name}
                    className={styles.segmentBubble}
                    style={{
                      width: `${size}px`,
                      height: `${size}px`,
                      background: `${seg.color}1a`,
                      borderColor: `${seg.color}66`,
                    }}
                  >
                    <span className={styles.segmentBubbleName}>{seg.name}</span>
                    <span className={styles.segmentBubbleSize}>{seg.size.toLocaleString()}</span>
                  </div>
                );
              })}
            </motion.div>
          </motion.section>

          {/* ── CAMPAIGNS GRID ──────────────────────────────────────────── */}
          <section className={styles.campaignsSection}>
            <motion.div
              className={styles.campaignsHeader}
              initial="hidden"
              whileInView="visible"
              viewport={viewportOnce}
              variants={staggerBlur}
            >
              <motion.span className={styles.sectionEyebrow} variants={blurFadeIn}>
                REACH OUT
              </motion.span>
              <motion.h2 className={styles.sectionTitle} variants={blurFadeIn}>
                Campaigns that
                <br />
                feel like a friend,
                <br />
                not a broadcast.
              </motion.h2>
            </motion.div>

            <motion.div
              className={styles.campaignsGrid}
              initial="hidden"
              whileInView="visible"
              viewport={viewportOnce}
              variants={staggerBlur}
            >
              {campaigns.map(({ title, copy }, i) => {
                const Icon = CAMPAIGNS_ICONS[i];
                return (
                <motion.div
                  key={title}
                  className={styles.campaignCard}
                  variants={blurFadeIn}
                  whileHover={{ y: -4 }}
                >
                  <div className={styles.campaignIcon}>
                    <Icon size={18} strokeWidth={1.8} />
                  </div>
                  <h3 className={styles.campaignTitle}>{title}</h3>
                  <p className={styles.campaignCopy}>{copy}</p>
                  <span className={styles.campaignLearn}>
                    Learn more
                    <ChevronRight size={14} strokeWidth={2} />
                  </span>
                </motion.div>
              );
              })}
            </motion.div>
          </section>

          {/* ── LIVE SEGMENT BUILDER ───────────────────────────────────── */}
          <motion.section
            className={styles.builderSection}
            initial="hidden"
            whileInView="visible"
            viewport={viewportOnce}
            variants={staggerBlur}
          >
            <motion.div className={styles.builderHeader} variants={blurFadeIn}>
              <span className={styles.sectionEyebrow}>TRY IT, LIVE</span>
              <h2 className={styles.sectionTitle}>
                Do segments stay
                <br />
                current on their own?
              </h2>
              <p className={styles.builderCopy}>
                This is a real segment builder. Toggle any attribute and the matching audience
                updates as you go.
              </p>
            </motion.div>

            <motion.div className={styles.builderGrid} variants={blurFadeIn}>
              <div className={styles.filterPanel}>
                <div className={styles.filterPanelHead}>
                  <Target size={14} strokeWidth={2} />
                  <span>Attributes</span>
                  {activeFilters.size > 0 && (
                    <button
                      type="button"
                      onClick={() => setActiveFilters(new Set())}
                      className={styles.filterClear}
                    >
                      Clear
                    </button>
                  )}
                </div>
                <ul className={styles.filterList}>
                  {FILTERS.map((f) => {
                    const active = activeFilters.has(f.id);
                    return (
                      <li key={f.id}>
                        <button
                          type="button"
                          onClick={() => toggleFilter(f.id)}
                          className={cn(styles.filterToggle, active && styles.filterToggleActive)}
                          aria-pressed={active}
                        >
                          <span className={styles.filterCheckbox}>
                            {active && <span className={styles.filterCheckDot} />}
                          </span>
                          {f.label}
                        </button>
                      </li>
                    );
                  })}
                </ul>
                <div className={styles.filterFoot}>
                  <button type="button" className={styles.filterAddMore}>
                    <Plus size={12} strokeWidth={2.2} />
                    Add another attribute
                  </button>
                </div>
              </div>

              <div className={styles.audiencePanel}>
                <div className={styles.audienceHead}>
                  <span className={styles.audienceCountBig}>{matchingGuests.length}</span>
                  <span className={styles.audienceCountLabel}>
                    {matchingGuests.length === 1 ? 'guest matches' : 'guests match'} ·{' '}
                    {GUESTS.length} total
                  </span>
                  <button
                    type="button"
                    className={styles.audienceCta}
                    onClick={goRegister}
                    disabled={matchingGuests.length === 0}
                  >
                    Send a campaign →
                  </button>
                </div>
                <ul className={styles.audienceList}>
                  {matchingGuests.length === 0 ? (
                    <li className={styles.audienceEmpty}>
                      No guests match this combination, loosen a filter.
                    </li>
                  ) : (
                    matchingGuests.map((g) => (
                      <motion.li
                        key={g.name}
                        className={styles.audienceItem}
                        layout
                        initial={{ opacity: 0, y: 6 }}
                        animate={{ opacity: 1, y: 0 }}
                        transition={{ duration: 0.22 }}
                      >
                        <div className={styles.audienceAvatar} style={{ background: g.avatar }}>
                          {g.initials}
                        </div>
                        <div className={styles.audienceMeta}>
                          <span className={styles.audienceName}>{g.name}</span>
                          <span className={styles.audienceSub}>
                            {g.visits} visits · ${g.lifetime.toLocaleString()} lifetime · last visit{' '}
                            {g.lastVisit}
                          </span>
                        </div>
                        {g.vip && (
                          <span className={styles.audienceTag}>
                            <Star size={10} fill="currentColor" strokeWidth={0} />
                            VIP
                          </span>
                        )}
                      </motion.li>
                    ))
                  )}
                </ul>
              </div>
            </motion.div>
          </motion.section>

          <FaqSection title={faq.title} faqs={faq.items} />

          {/* ── CTA ─────────────────────────────────────────────────────── */}
          <motion.section
            className={styles.ctaBlock}
            initial="hidden"
            whileInView="visible"
            viewport={viewportOnce}
            variants={staggerBlur}
          >
            <motion.h2 className={styles.ctaTitle} variants={blurFadeIn}>
              <Lines text={t('crm.cta.title')} />
            </motion.h2>
            <motion.div className={styles.ctaButtons} variants={blurFadeIn}>
              <button type="button" onClick={goRegister} className={styles.primaryBtn}>
                Get started
              </button>
              <button type="button" onClick={goDemo} className={styles.outlineBtn}>
                Book a demo
              </button>
            </motion.div>
          </motion.section>
        </div>
      </MainLayout>
    </AppLayout>
  );
};

export default CRMPage;
