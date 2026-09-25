import React from 'react';
import { motion } from 'motion/react';
import { useTranslation } from 'react-i18next';
import { useRouter } from 'next/router';
import {
  Handshake,
  Puzzle,
  Store,
  GraduationCap,
  TrendingUp,
  BadgeCheck,
  BookOpen,
  DollarSign,
  Users,
  Zap,
} from 'lucide-react';
import AppLayout from '../../components/Layout/AppLayout';
import MainLayout from '../../layouts/Main/MainLayout';
import HelmetContainer from '../../components/HelmetContainer';
import { blurFadeIn, staggerBlur, viewportOnce } from '../../lib/animations';
import { webPageSchema } from '../../lib/structured-data';
import Lines from '../../components/text/Lines';
import styles from './partners.module.scss';

// Icons only. Every string resolves from partners.tracks, matched by
// position; the track number is derived from that position.
const TRACK_ICONS = [Store, Puzzle, GraduationCap, Handshake];



const PartnersPage = (): React.JSX.Element => {
  const { t } = useTranslation();
  const tracks = t('partners.tracks', { returnObjects: true }) as unknown as {
    name: string;
    summary: string;
    who: string;
    benefits: string[];
  }[];
  const tiers = t('partners.tiers', { returnObjects: true }) as unknown as {
    name: string;
    commitment: string;
    desc: string;
  }[];
  const router = useRouter();

  const goApply = (): void => {
    void router.push('/contact');
  };

  return (
    <AppLayout>
      <MainLayout>
        <HelmetContainer
          title={t('partners.meta.title')}
          description={t('partners.meta.description')}
          canonical="https://rtylr.com/partners/"
          structuredData={webPageSchema({
            name: 'Partner Program',
            path: '/partners/',
            description:
              'Build with rtylr. Implementation partners, technology partners, solution partners, and hardware partners.',
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
                <Handshake size={13} strokeWidth={2} />
                {t('partners.hero.eyebrow')}
              </motion.span>
              <motion.h1 className={styles.heroTitle} variants={blurFadeIn}>
                <Lines text={t('partners.hero.title')} />
              </motion.h1>
              <motion.p className={styles.heroSubtitle} variants={blurFadeIn}>
                Agencies that set it up. Software that plugs in. Hardware that ships with it.
                Trainers that teach it. We share revenue, tools, and co-marketing with the people
                who help our customers succeed.
              </motion.p>
              <motion.div className={styles.heroCtas} variants={blurFadeIn}>
                <button type="button" onClick={goApply} className={styles.primaryBtn}>
                {t('partners.hero.apply')}
              </button>
                <a href="mailto:partners@voxire.com" className={styles.outlineBtn}>
                {t('partners.hero.email')}
              </a>
              </motion.div>
            </motion.div>
          </section>

          {/* TRACKS */}
          <section className={styles.tracksSection}>
            <motion.div
              className={styles.tracksHeader}
              initial="hidden"
              whileInView="visible"
              viewport={viewportOnce}
              variants={staggerBlur}
            >
              <motion.span className={styles.sectionEyebrow} variants={blurFadeIn}>
                {t('partners.tracksSection.eyebrow')}
              </motion.span>
              <motion.h2 className={styles.sectionTitle} variants={blurFadeIn}>
                {t('partners.tracksSection.title')}
              </motion.h2>
            </motion.div>

            <motion.div
              className={styles.tracksList}
              initial="hidden"
              whileInView="visible"
              viewport={viewportOnce}
              variants={staggerBlur}
            >
              {TRACK_ICONS.map((Icon, i) => (
                <motion.div key={tracks[i].name} className={styles.track} variants={blurFadeIn}>
                  <div className={styles.trackHeader}>
                    <span className={styles.trackNumber}>{String(i + 1).padStart(2, '0')}</span>
                    <div className={styles.trackIcon}>
                      <Icon size={18} strokeWidth={1.8} />
                    </div>
                    <div className={styles.trackTitles}>
                      <h3 className={styles.trackName}>{tracks[i].name}</h3>
                      <p className={styles.trackWho}>{tracks[i].who}</p>
                    </div>
                  </div>
                  <p className={styles.trackSummary}>{tracks[i].summary}</p>
                  <ul className={styles.trackBenefits}>
                    {tracks[i].benefits.map((b) => (
                      <li key={b}>
                        <BadgeCheck size={14} strokeWidth={2} />
                        <span>{b}</span>
                      </li>
                    ))}
                  </ul>
                </motion.div>
              ))}
            </motion.div>
          </section>

          {/* TIERS */}
          <motion.section
            className={styles.tiersSection}
            initial="hidden"
            whileInView="visible"
            viewport={viewportOnce}
            variants={staggerBlur}
          >
            <motion.div className={styles.tiersHeader} variants={blurFadeIn}>
              <span className={styles.sectionEyebrow}>{t('partners.tiersSection.eyebrow')}</span>
              <h2 className={styles.sectionTitle}>{t('partners.tiersSection.title')}</h2>
            </motion.div>
            <motion.div className={styles.tiersGrid} variants={blurFadeIn}>
              {tiers.map((tier, i) => (
                <div key={tier.name} className={styles.tierCard}>
                  <div className={styles.tierBadge}>{tier.name}</div>
                  <div className={styles.tierCommitment}>{tier.commitment}</div>
                  <p className={styles.tierDesc}>{tier.desc}</p>
                  <div className={styles.tierProgress}>
                    <div
                      className={styles.tierProgressBar}
                      style={{ width: `${(i + 1) * 33.33}%` }}
                    />
                  </div>
                </div>
              ))}
            </motion.div>
          </motion.section>

          {/* PERKS STRIP */}
          <motion.section
            className={styles.perksSection}
            initial="hidden"
            whileInView="visible"
            viewport={viewportOnce}
            variants={staggerBlur}
          >
            <motion.ul className={styles.perksList} variants={blurFadeIn}>
              <li>
                <TrendingUp size={18} strokeWidth={1.8} />
                <div>
                  <strong>Revenue share up to 30%</strong>
                  <span>on first-year subscriptions you influence.</span>
                </div>
              </li>
              <li>
                <BookOpen size={18} strokeWidth={1.8} />
                <div>
                  <strong>Full product training</strong>
                  <span>self-paced or live-led, certification-track available.</span>
                </div>
              </li>
              <li>
                <Users size={18} strokeWidth={1.8} />
                <div>
                  <strong>Co-selling motions</strong>
                  <span>warm handoff with our account team, not a cold form.</span>
                </div>
              </li>
              <li>
                <Zap size={18} strokeWidth={1.8} />
                <div>
                  <strong>Early API access</strong>
                  <span>preview builds and beta endpoints for integration partners.</span>
                </div>
              </li>
              <li>
                <DollarSign size={18} strokeWidth={1.8} />
                <div>
                  <strong>Transparent reporting</strong>
                  <span>every referral, every commission, visible in the partner portal.</span>
                </div>
              </li>
            </motion.ul>
          </motion.section>

          {/* CTA */}
          <motion.section
            className={styles.ctaBlock}
            initial="hidden"
            whileInView="visible"
            viewport={viewportOnce}
            variants={staggerBlur}
          >
            <motion.h2 className={styles.ctaTitle} variants={blurFadeIn}>
              {t('partners.cta.title')}
            </motion.h2>
            <motion.p className={styles.ctaCopy} variants={blurFadeIn}>
              Applications are reviewed weekly. We&apos;ll respond within 5 business days with a
              next step.
            </motion.p>
            <motion.div className={styles.ctaButtons} variants={blurFadeIn}>
              <button type="button" onClick={goApply} className={styles.primaryBtn}>
                {t('partners.cta.apply')}
              </button>
              <a href="mailto:partners@voxire.com" className={styles.outlineBtn}>
                partners@voxire.com
              </a>
            </motion.div>
          </motion.section>
        </div>
      </MainLayout>
    </AppLayout>
  );
};

export default PartnersPage;
