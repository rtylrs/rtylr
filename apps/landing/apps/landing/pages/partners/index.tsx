import React from 'react';
import { motion } from 'motion/react';
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
import styles from './partners.module.scss';

const TRACKS = [
  {
    icon: Store,
    name: 'Implementation Partners',
    summary:
      'Consultancies and agencies that set up rtylr for end customers, POS configuration, menu migration, HR onboarding.',
    who: 'For ops consultancies, restaurant-tech specialists, and digital agencies.',
    benefits: [
      'Referral fees on first-year revenue',
      'Co-selling with our account team',
      'Dedicated partner portal with pre-sales collateral',
      'rtylr Certified Implementer accreditation',
    ],
  },
  {
    icon: Puzzle,
    name: 'Technology Partners',
    summary:
      'Software that integrates with rtylr, delivery marketplaces, payment processors, loyalty engines, BI tools.',
    who: 'For SaaS companies adjacent to commerce, POS, HR, or finance.',
    benefits: [
      'Listed in the integrations directory',
      'Co-marketing on launch',
      'Joint webinars with our customers',
      'Priority partner support channel',
    ],
  },
  {
    icon: GraduationCap,
    name: 'Solution Partners',
    summary:
      'Training, enablement, and managed services built around rtylr, franchise rollouts, staff training, compliance audits.',
    who: 'For training companies, franchise enablement teams, and vertical specialists.',
    benefits: [
      'Revenue share on multi-location deals',
      'Rtylr Certified Trainer program',
      'Early access to new apps and APIs',
      'Feature placement in customer newsletters',
    ],
  },
  {
    icon: Handshake,
    name: 'Hardware Partners',
    summary:
      'Printers, card readers, scanners, kiosk terminals, hardware that rtylr ships to the end customer.',
    who: 'For hardware OEMs and authorized distributors.',
    benefits: [
      'Direct inclusion in the rtylr hardware catalog',
      'Certification program for plug-and-play pairing',
      'Joint customer fulfillment operations',
      'SDK access for native driver integration',
    ],
  },
];

const TIER_LEVELS = [
  {
    name: 'Registered',
    desc: 'You\u2019re in the program. Directory listing, referral tracking, basic co-marketing assets.',
    commitment: '0 closed deals',
  },
  {
    name: 'Certified',
    desc: 'You\u2019ve completed the product training and closed your first deals. Higher referral rates.',
    commitment: '3+ closed deals',
  },
  {
    name: 'Premier',
    desc: 'Dedicated partner manager, pre-launch access, co-selling motions, feature on the home page.',
    commitment: '15+ closed deals',
  },
];

const PartnersPage = (): React.JSX.Element => {
  const router = useRouter();

  const goApply = (): void => {
    void router.push('/contact');
  };

  return (
    <AppLayout>
      <MainLayout>
        <HelmetContainer
          title="Partner Program | rtylr"
          description="Build with rtylr. Implementation partners, technology partners, solution partners, and hardware partners, terms, tiers, and how to apply."
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
                PARTNER PROGRAM
              </motion.span>
              <motion.h1 className={styles.heroTitle} variants={blurFadeIn}>
                Build with rtylr.
                <br />
                <span className={styles.heroTitleSoft}>Grow with us.</span>
              </motion.h1>
              <motion.p className={styles.heroSubtitle} variants={blurFadeIn}>
                Agencies that set it up. Software that plugs in. Hardware that ships with it.
                Trainers that teach it. We share revenue, tools, and co-marketing with the people
                who help our customers succeed.
              </motion.p>
              <motion.div className={styles.heroCtas} variants={blurFadeIn}>
                <button type="button" onClick={goApply} className={styles.primaryBtn}>
                  Apply to the program
                </button>
                <a href="mailto:partners@voxire.com" className={styles.outlineBtn}>
                  Email partners@voxire.com
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
                FOUR TRACKS
              </motion.span>
              <motion.h2 className={styles.sectionTitle} variants={blurFadeIn}>
                Pick the track
                <br />
                that fits you.
              </motion.h2>
            </motion.div>

            <motion.div
              className={styles.tracksList}
              initial="hidden"
              whileInView="visible"
              viewport={viewportOnce}
              variants={staggerBlur}
            >
              {TRACKS.map(({ icon: Icon, name, summary, who, benefits }, i) => (
                <motion.div key={name} className={styles.track} variants={blurFadeIn}>
                  <div className={styles.trackHeader}>
                    <span className={styles.trackNumber}>{String(i + 1).padStart(2, '0')}</span>
                    <div className={styles.trackIcon}>
                      <Icon size={18} strokeWidth={1.8} />
                    </div>
                    <div className={styles.trackTitles}>
                      <h3 className={styles.trackName}>{name}</h3>
                      <p className={styles.trackWho}>{who}</p>
                    </div>
                  </div>
                  <p className={styles.trackSummary}>{summary}</p>
                  <ul className={styles.trackBenefits}>
                    {benefits.map((b) => (
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
              <span className={styles.sectionEyebrow}>HOW YOU LEVEL UP</span>
              <h2 className={styles.sectionTitle}>Three tiers. Earned, not paid.</h2>
            </motion.div>
            <motion.div className={styles.tiersGrid} variants={blurFadeIn}>
              {TIER_LEVELS.map((tier, i) => (
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
              Want in?
            </motion.h2>
            <motion.p className={styles.ctaCopy} variants={blurFadeIn}>
              Applications are reviewed weekly. We&apos;ll respond within 5 business days with a
              next step.
            </motion.p>
            <motion.div className={styles.ctaButtons} variants={blurFadeIn}>
              <button type="button" onClick={goApply} className={styles.primaryBtn}>
                Start an application
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
