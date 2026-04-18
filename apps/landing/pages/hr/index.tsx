import React from 'react';
import { motion } from 'motion/react';
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
import { APP_URLS } from '@voxire/config';
import { navigateExternal } from '../../lib/navigation';
import { blurFadeIn, staggerBlur, viewportOnce } from '../../lib/animations';
import styles from './hr.module.scss';

// Schedule board mock — 5 days × a few staff
const DAYS = ['MON', 'TUE', 'WED', 'THU', 'FRI', 'SAT', 'SUN'];
const STAFF = [
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

const PILLARS = [
  {
    icon: Calendar,
    title: 'Schedules that publish themselves',
    copy: 'Drag, drop, publish. Staff get texts the moment a shift posts. Conflicts caught before they ship.',
  },
  {
    icon: Clock,
    title: 'Time & attendance, zero paperwork',
    copy: 'Clock in from the POS. Biometric, PIN, or geofenced. Auto-rounding and break enforcement built in.',
  },
  {
    icon: DollarSign,
    title: 'Payroll that runs itself',
    copy: 'Hours, tips, deductions, taxes — rolled up automatically. Direct deposit, pay stubs, filings handled.',
  },
  {
    icon: Palmtree,
    title: 'Time off, clearly',
    copy: 'One place for requests, approvals, and balances. Blackout dates, accrual rules, and coverage in one glance.',
  },
  {
    icon: UserPlus,
    title: 'Onboarding without friction',
    copy: 'E-sign offer letters, W-4s and I-9s. Self-service docs, handbook, and training from day one.',
  },
  {
    icon: FileCheck,
    title: 'Compliance that watches itself',
    copy: 'Overtime rules, break audits, minor labor, fair workweek. Flags before they become violations.',
  },
];

const HRPage = (): JSX.Element => {
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
          title="HR — Your team. Looked after. | rtylr"
          description="Scheduling, time & attendance, payroll, time off, onboarding, and compliance. One app that treats your staff like people."
        />

        <div className={styles.page}>
          {/* ── HERO — centered, warm, schedule-board visual below ─────────── */}
          <section className={styles.hero}>
            <motion.div
              className={styles.heroCopy}
              initial="hidden"
              animate="visible"
              variants={staggerBlur}
            >
              <motion.span className={styles.heroEyebrow} variants={blurFadeIn}>
                <HeartHandshake size={13} strokeWidth={2} />
                rtylr HR
              </motion.span>
              <motion.h1 className={styles.heroTitle} variants={blurFadeIn}>
                Your team.
                <br />
                <span className={styles.heroTitleSoft}>Looked after.</span>
              </motion.h1>
              <motion.p className={styles.heroSubtitle} variants={blurFadeIn}>
                Schedules, shifts, payroll, time off, onboarding, and compliance — in one app that
                treats your staff like people, not line items.
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

            {/* Schedule board mockup */}
            <motion.div
              className={styles.scheduleMock}
              initial={{ opacity: 0, y: 24 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.4, duration: 1.0, ease: [0.16, 1, 0.3, 1] }}
              aria-hidden="true"
            >
              <div className={styles.scheduleHeader}>
                <div className={styles.scheduleHeaderLeft}>
                  <Calendar size={14} strokeWidth={2} />
                  <span>Week of April 14 · Terrace</span>
                </div>
                <div className={styles.scheduleHeaderRight}>
                  <span className={styles.scheduleChip}>28 shifts</span>
                  <span className={`${styles.scheduleChip} ${styles.scheduleChipGood}`}>
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
                {STAFF.map((person) => (
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
                    {person.shifts.map((shift, i) => (
                      <div
                        key={i}
                        className={`${styles.scheduleShift} ${
                          shift === 'OFF' ? styles.scheduleShiftOff : ''
                        }`}
                        style={shift !== 'OFF' ? { background: `${person.color}26` } : undefined}
                      >
                        {shift}
                      </div>
                    ))}
                  </div>
                ))}
              </div>
              <div className={styles.scheduleFooter}>
                <span>Tips · Hours · Breaks · Auto-roll to payroll every Sunday night</span>
              </div>
            </motion.div>
          </section>

          {/* ── PILLAR GRID — 3×2 shared-border cells ──────────────────────── */}
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
                The entire employee
                <br />
                lifecycle, without the paperwork.
              </motion.h2>
            </motion.div>

            <motion.div
              className={styles.pillarsGrid}
              initial="hidden"
              whileInView="visible"
              viewport={viewportOnce}
              variants={staggerBlur}
            >
              {PILLARS.map(({ icon: Icon, title, copy }) => (
                <motion.div key={title} className={styles.pillarCard} variants={blurFadeIn}>
                  <div className={styles.pillarIcon}>
                    <Icon size={18} strokeWidth={1.8} />
                  </div>
                  <h3 className={styles.pillarTitle}>{title}</h3>
                  <p className={styles.pillarCopy}>{copy}</p>
                </motion.div>
              ))}
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

          {/* ── CTA ──────────────────────────────────────────────────────── */}
          <motion.section
            className={styles.ctaBlock}
            initial="hidden"
            whileInView="visible"
            viewport={viewportOnce}
            variants={staggerBlur}
          >
            <motion.h2 className={styles.ctaTitle} variants={blurFadeIn}>
              Your staff deserve
              <br />
              software built for them.
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
