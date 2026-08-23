import React, { useMemo, useState } from 'react';
import { cn } from '../../lib/cn';
import { motion } from 'motion/react';
import { useTranslation } from 'react-i18next';
import { useRouter } from 'next/router';
import {
  Calendar,
  Clock,
  DollarSign,
  HeartHandshake,
  UserPlus,
  Palmtree,
  FileCheck,
} from 'lucide-react';
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
import styles from './hr.module.scss';

// Schedule board mock, 7 days × 4 staff. State-driven so each cell is editable.
const DAYS = ['MON', 'TUE', 'WED', 'THU', 'FRI', 'SAT', 'SUN'];

interface StaffMember {
  name: string;
  role: string;
  color: string;
  shifts: string[];
}

const INITIAL_STAFF: StaffMember[] = [
  {
    name: 'Maya Chen',
    role: 'Server',
    color: '#d6b88e',
    shifts: ['9–5', '9–5', 'OFF', '11–7', '11–7', '4–12', '4–12'],
  },
  {
    name: 'Jordan Kim',
    role: 'Bar',
    color: '#c9a48a',
    shifts: ['4–12', '4–12', '4–12', 'OFF', 'OFF', '3–11', '3–11'],
  },
  {
    name: 'Alex Rivera',
    role: 'Kitchen',
    color: '#b89a8a',
    shifts: ['8–4', '8–4', '8–4', '8–4', '8–4', 'OFF', 'OFF'],
  },
  {
    name: 'Sam Patel',
    role: 'Manager',
    color: '#a89680',
    shifts: ['9–6', 'OFF', '9–6', '9–6', '9–6', '11–7', 'OFF'],
  },
];

// Click-to-cycle through these shift options; role-appropriate for each person
// but we keep it simple and use the same ring everywhere.
const SHIFT_CYCLE = ['OFF', '8–4', '9–5', '9–6', '11–7', '3–11', '4–12'];

const PILLARS_ICONS = [Calendar, Clock, DollarSign, Palmtree, UserPlus, FileCheck];

const HRPage = (): React.JSX.Element => {
  const faq = useFaq('hr');
  const { t } = useTranslation();
  const pillars = t('hr.pillars.items', { returnObjects: true }) as unknown as {
    title: string;
    copy: string;
  }[];
  const router = useRouter();
  const [staff, setStaff] = useState<StaffMember[]>(INITIAL_STAFF);

  // Click a shift cell → cycle to the next value in SHIFT_CYCLE.
  const cycleShift = (staffIdx: number, dayIdx: number): void => {
    setStaff((prev) => {
      const updated = [...prev];
      const member = { ...updated[staffIdx] };
      member.shifts = [...member.shifts];
      const current = member.shifts[dayIdx];
      const pos = SHIFT_CYCLE.indexOf(current);
      const next = SHIFT_CYCLE[(pos + 1) % SHIFT_CYCLE.length];
      member.shifts[dayIdx] = next;
      updated[staffIdx] = member;
      return updated;
    });
  };

  const totalShifts = useMemo(
    () => staff.reduce((acc, s) => acc + s.shifts.filter((x) => x !== 'OFF').length, 0),
    [staff]
  );

  const goRegister = (): void => {
    navigateExternal(`${APP_URLS.auth}/register?redirect=hr`);
  };

  const goDemo = (): void => {
    void router.push('/demo');
  };

  return (
    <AppLayout>
      <MainLayout>
        <HelmetContainer
          title={t('hr.meta.title')}
          description={t('hr.meta.description')}
          canonical="https://rtylr.com/hr/"
          structuredData={productWithBreadcrumbs({
            name: 'HR',
            path: '/hr/',
            subCategory: 'Human Resources and Payroll',
            description: t('hr.meta.schemaDescription'),
          })}
        />

        <div className={styles.page}>
          {/* ── HERO, centered, warm, schedule-board visual below ─────────── */}
          <section className={styles.hero}>
            <motion.div
              className={styles.heroCopy}
              initial="hidden"
              animate="visible"
              variants={staggerBlur}
            >
              <motion.span className={styles.heroEyebrow} variants={blurFadeIn}>
                <HeartHandshake size={13} strokeWidth={2} />
                {t('hr.hero.eyebrow')}
              </motion.span>
              <h1 className={styles.heroTitle}>
                <SplitWords as="span" text={t('hr.hero.titleA')} delay={0.1} />
                <br />
                <span className={styles.heroTitleSoft}>
                  <SplitWords as="span" text={t('hr.hero.titleB')} delay={0.35} />
                </span>
              </h1>
              <motion.p className={styles.heroSubtitle} variants={blurFadeIn}>
                {t('hr.hero.subtitle')}
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

            {/* Live schedule board, click any cell to cycle its shift */}
            <motion.div
              className={styles.scheduleMock}
              initial={{ opacity: 0, y: 24 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.4, duration: 1.0, ease: [0.16, 1, 0.3, 1] }}
            >
              <div className={styles.scheduleHeader}>
                <div className={styles.scheduleHeaderLeft}>
                  <Calendar size={14} strokeWidth={2} />
                  <span>Week of April 14 · Terrace · click any cell to reassign</span>
                </div>
                <div className={styles.scheduleHeaderRight}>
                  <motion.span
                    key={totalShifts}
                    className={styles.scheduleChip}
                    initial={{ opacity: 0, scale: 0.9 }}
                    animate={{ opacity: 1, scale: 1 }}
                    transition={{ duration: 0.2 }}
                  >
                    {totalShifts} shifts
                  </motion.span>
                  <span className={cn(styles.scheduleChip, styles.scheduleChipGood)}>
                    No conflicts
                  </span>
                </div>
              </div>
              <div className={styles.scheduleGrid}>
                <div className={styles.scheduleRow}>
                  <div className={styles.scheduleStaff} />
                  {DAYS.map((day) => (
                    <div key={day} className={styles.scheduleDayLabel}>
                      {day}
                    </div>
                  ))}
                </div>
                {staff.map((person, staffIdx) => (
                  <div key={person.name} className={styles.scheduleRow}>
                    <div className={styles.scheduleStaff}>
                      <div className={styles.scheduleAvatar} style={{ background: person.color }}>
                        {person.name
                          .split(' ')
                          .map((n) => n[0])
                          .join('')}
                      </div>
                      <div className={styles.scheduleStaffText}>
                        <div className={styles.scheduleName}>{person.name}</div>
                        <div className={styles.scheduleRole}>{person.role}</div>
                      </div>
                    </div>
                    {person.shifts.map((shift, dayIdx) => (
                      <motion.button
                        type="button"
                        key={`${person.name}-${dayIdx}`}
                        className={cn(
                          styles.scheduleShift,
                          shift === 'OFF' && styles.scheduleShiftOff
                        )}
                        style={shift !== 'OFF' ? { background: `${person.color}26` } : undefined}
                        onClick={() => cycleShift(staffIdx, dayIdx)}
                        whileTap={{ scale: 0.94 }}
                        aria-label={`Change shift for ${person.name} on ${DAYS[dayIdx]}`}
                      >
                        <motion.span
                          key={shift}
                          initial={{ opacity: 0, y: -4 }}
                          animate={{ opacity: 1, y: 0 }}
                          transition={{ duration: 0.18 }}
                        >
                          {shift}
                        </motion.span>
                      </motion.button>
                    ))}
                  </div>
                ))}
              </div>
              <div className={styles.scheduleFooter}>
                <span>Tips · Hours · Breaks · Auto-roll to payroll every Sunday night</span>
              </div>
            </motion.div>
          </section>

          <DirectAnswer label={t('hr.directAnswer.label')}>
            {t('hr.directAnswer.body')}
          </DirectAnswer>

          {/* ── PILLAR GRID, 3×2 shared-border cells ──────────────────────── */}
          <section className={styles.pillarsSection}>
            <motion.div
              className={styles.pillarsHeader}
              initial="hidden"
              whileInView="visible"
              viewport={viewportOnce}
              variants={staggerBlur}
            >
              <motion.span className={styles.sectionEyebrow} variants={blurFadeIn}>
                SIX PILLARS
              </motion.span>
              <motion.h2 className={styles.sectionTitle} variants={blurFadeIn}>
                What does rtylr HR
                <br />
                handle for a shift team?
              </motion.h2>
            </motion.div>

            <motion.div
              className={styles.pillarsGrid}
              initial="hidden"
              whileInView="visible"
              viewport={viewportOnce}
              variants={staggerBlur}
            >
              {pillars.map(({ title, copy }, i) => {
                const Icon = PILLARS_ICONS[i];
                return (
                <motion.div key={title} className={styles.pillarCard} variants={blurFadeIn}>
                  <div className={styles.pillarIcon}>
                    <Icon size={18} strokeWidth={1.8} />
                  </div>
                  <h3 className={styles.pillarTitle}>{title}</h3>
                  <p className={styles.pillarCopy}>{copy}</p>
                </motion.div>
              );
              })}
            </motion.div>
          </section>

          {/* ── QUOTE / SENTIMENT STRIP ────────────────────────────────────── */}
          <motion.section
            className={styles.quoteSection}
            initial="hidden"
            whileInView="visible"
            viewport={viewportOnce}
            variants={staggerBlur}
          >
            <motion.p className={styles.quoteText} variants={blurFadeIn}>
              &ldquo;The first time my team saw payroll hit on a Monday morning without me opening a
              spreadsheet, I knew rtylr was different.&rdquo;
            </motion.p>
            <motion.div className={styles.quoteAttribution} variants={blurFadeIn}>
              <div
                className={styles.quoteAvatar}
                style={{ background: 'linear-gradient(135deg, #c9a48a, #d6b88e)' }}
              >
                LA
              </div>
              <div>
                <div className={styles.quoteName}>Laila Ahmed</div>
                <div className={styles.quoteRole}>Owner · Byblos Kitchen</div>
              </div>
            </motion.div>
          </motion.section>

          <FaqSection title={faq.title} faqs={faq.items} />

          {/* ── CTA ──────────────────────────────────────────────────────── */}
          <motion.section
            className={styles.ctaBlock}
            initial="hidden"
            whileInView="visible"
            viewport={viewportOnce}
            variants={staggerBlur}
          >
            <motion.h2 className={styles.ctaTitle} variants={blurFadeIn}>
              <Lines text={t('hr.cta.title')} />
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

export default HRPage;
