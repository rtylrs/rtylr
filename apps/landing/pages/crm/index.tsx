import React from 'react';
import { motion } from 'motion/react';
import { useRouter } from 'next/router';
import { Heart, Mail, Gift, Users, Star, Send, Target, ChevronRight } from 'lucide-react';
import AppLayout from '../../components/Layout/AppLayout';
import MainLayout from '../../layouts/Main/MainLayout';
import HelmetContainer from '../../components/HelmetContainer';
import { APP_URLS } from '@voxire/config';
import { navigateExternal } from '../../lib/navigation';
import { blurFadeIn, staggerBlur, viewportOnce } from '../../lib/animations';
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

const CAMPAIGNS = [
  {
    icon: Mail,
    title: 'Email that reads personal',
    copy: 'Pre-built templates. Personalisation tokens. Auto-sent on triggers like birthday, lapse, or first-visit.',
  },
  {
    icon: Gift,
    title: 'Loyalty that means something',
    copy: 'Points, tiers, rewards. Auto-applied at checkout. Built into POS, not stapled onto it.',
  },
  {
    icon: Target,
    title: 'Segments that make sense',
    copy: 'Slice by visit frequency, average spend, last order, dietary preferences. Save as a segment, send a campaign.',
  },
  {
    icon: Send,
    title: 'SMS that doesn\u2019t feel like spam',
    copy: 'Short-code, opt-in, transactional or broadcast. Throttled to protect your sender reputation.',
  },
];

const CRMPage = (): JSX.Element => {
  const router = useRouter();

  const goRegister = (): void => {
    navigateExternal(`${APP_URLS.auth}/register`);
  };

  const goDemo = (): void => {
    void router.push('/demo');
  };

  return (
    <AppLayout>
      <MainLayout>
        <HelmetContainer
          title="CRM — Every customer, remembered. | rtylr"
          description="Customer profiles, loyalty, segments, and campaigns. Every POS purchase enriches the profile automatically."
        />

        <div className={styles.page}>
          {/* ── HERO — copy left, profile card right ────────────────────── */}
          <section className={styles.hero}>
            <motion.div
              className={styles.heroCopy}
              initial="hidden"
              animate="visible"
              variants={staggerBlur}
            >
              <motion.span className={styles.heroEyebrow} variants={blurFadeIn}>
                <Heart size={13} strokeWidth={2} fill="currentColor" />
                rtylr CRM
              </motion.span>
              <motion.h1 className={styles.heroTitle} variants={blurFadeIn}>
                Every customer,
                <br />
                <span className={styles.heroTitleSoft}>remembered.</span>
              </motion.h1>
              <motion.p className={styles.heroSubtitle} variants={blurFadeIn}>
                Every order at the POS enriches the profile. Every birthday, every visit, every
                preference — stored, segmented, ready for the right message at the right moment.
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
                Every slice of your
                <br />
                customer base, one click away.
              </motion.h2>
              <motion.p className={styles.sectionCopy} variants={blurFadeIn}>
                Build segments by visit velocity, lifetime spend, last-order date, tags, preferences
                — any combination. Re-run automatically as new orders land.
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
              {CAMPAIGNS.map(({ icon: Icon, title, copy }) => (
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
              ))}
            </motion.div>
          </section>

          {/* ── INTEGRATED STORY ────────────────────────────────────────── */}
          <motion.section
            className={styles.integratedSection}
            initial="hidden"
            whileInView="visible"
            viewport={viewportOnce}
            variants={staggerBlur}
          >
            <motion.div className={styles.integratedInner} variants={blurFadeIn}>
              <Users size={22} strokeWidth={1.8} className={styles.integratedIcon} />
              <h2 className={styles.integratedTitle}>
                The profile gets richer with every tap of the POS.
              </h2>
              <p className={styles.integratedCopy}>
                When a guest pays, rtylr CRM catches their card, email, or phone number
                automatically. Every item they ordered, every modifier, every server who served them
                — it all flows into the profile. No CRM data entry. No forgotten customers. No
                &ldquo;do I know this person&rdquo; moments at the door.
              </p>
            </motion.div>
          </motion.section>

          {/* ── CTA ─────────────────────────────────────────────────────── */}
          <motion.section
            className={styles.ctaBlock}
            initial="hidden"
            whileInView="visible"
            viewport={viewportOnce}
            variants={staggerBlur}
          >
            <motion.h2 className={styles.ctaTitle} variants={blurFadeIn}>
              Stop forgetting
              <br />
              your best customers.
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
