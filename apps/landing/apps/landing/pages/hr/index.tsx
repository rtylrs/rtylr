import React, { useMemo, useState } from 'react';
import { cn } from '../../lib/cn';
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
import DirectAnswer from '../../components/DirectAnswer/DirectAnswer';
import FaqSection from '../../components/FaqSection/FaqSection';
import { productWithBreadcrumbs } from '../../lib/structured-data';
import { HR_FAQS } from '../../lib/faqs';
import { APP_URLS } from '@voxire/config';
import { navigateExternal } from '../../lib/navigation';
import { blurFadeIn, staggerBlur, viewportOnce } from '../../lib/animations';
import { SplitWords } from '../../components/motion/SplitWords';
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
    copy: 'Hours, tips, deductions, taxes, rolled up automatically. Direct deposit, pay stubs, filings handled.',
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
          title="HR · Your team. Looked after. | rtylr"
          description="Scheduling, time & attendance, payroll, time off, onboarding, and compliance. One app that treats your staff like people."
          canonical="https://rtylr.com/hr/"
          structuredData={productWithBreadcrumbs({
            name: 'HR',
            path: '/hr/',
            subCategory: 'Human Resources and Payroll',
            description:
              'Scheduling, payroll, compliance, and time-off management for modern teams.',
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
                rtylr HR
              </motion.span>
              <h1 className={styles.heroTitle}>
                <SplitWords as="span" text={'Your team.'} delay={0.1} />
                <br />
                <span className={styles.heroTitleSoft}>
                  <SplitWords as="span" text={'Looked after.'} delay={0.35} />
                </span>
              </h1>
              <motion.p className={styles.heroSubtitle} variants={blurFadeIn}>
                Drag-to-publish schedules. Clock-in from the POS. Payroll that rolls up tips, hours,
                and deductions on its own. Compliance flags before they become violations. Software
                that treats staff like people, not line items in a spreadsheet.
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

          <DirectAnswer label="What is rtylr HR?">
            rtylr HR is the people app inside rtylr, for teams with shift workers, hourly staff or both.
            It covers employee records, scheduling, shift swaps, time off, payroll calculation and
            compliance documents. A person is created once: the schedule, the pay run and their POS login
            all read that same record, so a new hire is not entered in four different places.
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

          <FaqSection title="Questions about rtylr HR" faqs={HR_FAQS} />

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
