import React from 'react';
import { motion } from 'motion/react';
import { useRouter } from 'next/router';
import { CheckSquare, LayoutGrid, GitBranch, Zap, MessageSquare, Calendar } from 'lucide-react';
import AppLayout from '../../components/Layout/AppLayout';
import MainLayout from '../../layouts/Main/MainLayout';
import HelmetContainer from '../../components/HelmetContainer';
import { APP_URLS } from '@voxire/config';
import { navigateExternal } from '../../lib/navigation';
import { blurFadeIn, staggerBlur, viewportOnce } from '../../lib/animations';
import styles from './flow.module.scss';

// Kanban columns for the hero mockup
const COLUMNS: {
  title: string;
  count: number;
  tint: string;
  cards: { title: string; tag?: string; assignee: string; priority?: 'high' | 'med' }[];
}[] = [
  {
    title: 'Backlog',
    count: 12,
    tint: '#8a8a8a',
    cards: [
      { title: 'Seasonal menu rollout — Q3', assignee: 'MC', tag: 'Ops' },
      { title: 'New supplier onboarding: Rossini', assignee: 'JK', tag: 'ERP' },
    ],
  },
  {
    title: 'In progress',
    count: 5,
    tint: '#d6b88e',
    cards: [
      {
        title: 'Friday dinner covers — staffing plan',
        assignee: 'SP',
        tag: 'HR',
        priority: 'high',
      },
      { title: 'Redo brunch prep checklist', assignee: 'AR', tag: 'Kitchen' },
      { title: 'Email blast: Mother\u2019s Day brunch', assignee: 'MC', tag: 'CRM' },
    ],
  },
  {
    title: 'Review',
    count: 2,
    tint: '#8ca39d',
    cards: [{ title: 'Espresso machine repair SLA', assignee: 'JK', tag: 'Vendor' }],
  },
  {
    title: 'Done',
    count: 18,
    tint: '#6bcf7f',
    cards: [
      { title: 'Q1 sales-tax filing', assignee: 'SP', tag: 'Finance' },
      { title: 'POS firmware v4.2 rollout', assignee: 'AR', tag: 'Ops' },
    ],
  },
];

const FEATURES = [
  {
    icon: LayoutGrid,
    title: 'Boards for every team',
    copy: 'Kanban, list, timeline, or calendar. Each team picks their view. Same data underneath.',
  },
  {
    icon: GitBranch,
    title: 'Tasks that link to reality',
    copy: 'Attach a task to a customer, a supplier, an invoice, or a shift. When it moves, everything updates.',
  },
  {
    icon: Zap,
    title: 'Automations without zapier',
    copy: 'When a PO is approved → task on purchasing. When payroll closes → reminder on HR. Built in.',
  },
  {
    icon: MessageSquare,
    title: 'Conversations in context',
    copy: 'Comments, mentions, and files live on the task. No more "which Slack thread was this in?"',
  },
  {
    icon: Calendar,
    title: 'Deadlines that actually land',
    copy: 'Due-date enforcement. Roll-up views by owner. Escalation before things slip.',
  },
  {
    icon: CheckSquare,
    title: 'Templates for repeatable ops',
    copy: 'Weekly checklists, monthly closes, new-location launches — all as reusable project templates.',
  },
];

const FlowPage = (): JSX.Element => {
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
          title="Flow — Projects, tasks, and team workspaces. | rtylr"
          description="Kanban boards, automations, templates, and tasks linked to everything else in rtylr. Operations that don't slip."
        />

        <div className={styles.page}>
          {/* ── HERO — centered copy + kanban board mockup below ──────────── */}
          <section className={styles.hero}>
            <motion.div
              className={styles.heroCopy}
              initial="hidden"
              animate="visible"
              variants={staggerBlur}
            >
              <motion.span className={styles.heroEyebrow} variants={blurFadeIn}>
                rtylr Flow
              </motion.span>
              <motion.h1 className={styles.heroTitle} variants={blurFadeIn}>
                Work that moves.
                <br />
                <span className={styles.heroTitleSoft}>Operations that don&apos;t slip.</span>
              </motion.h1>
              <motion.p className={styles.heroSubtitle} variants={blurFadeIn}>
                Kanban, lists, timelines, and automations. Tasks that link to the customer,
                supplier, invoice, or shift they&apos;re actually about. The work OS for every team,
                every workflow.
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

            {/* Kanban board mockup */}
            <motion.div
              className={styles.board}
              initial={{ opacity: 0, y: 40 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.4, duration: 1.0, ease: [0.16, 1, 0.3, 1] }}
              aria-hidden="true"
            >
              <div className={styles.boardHeader}>
                <div className={styles.boardCrumbs}>
                  <span className={styles.boardCrumb}>Workspaces</span>
                  <span className={styles.boardCrumbSep}>/</span>
                  <span className={styles.boardCrumb}>Operations</span>
                  <span className={styles.boardCrumbSep}>/</span>
                  <span className={styles.boardCrumbActive}>Weekly sprint</span>
                </div>
                <div className={styles.boardViews}>
                  <span className={`${styles.boardView} ${styles.boardViewActive}`}>Board</span>
                  <span className={styles.boardView}>List</span>
                  <span className={styles.boardView}>Timeline</span>
                </div>
              </div>

              <div className={styles.boardColumns}>
                {COLUMNS.map((col) => (
                  <div key={col.title} className={styles.column}>
                    <div className={styles.columnHeader}>
                      <span className={styles.columnTitle}>
                        <span className={styles.columnDot} style={{ background: col.tint }} />
                        {col.title}
                      </span>
                      <span className={styles.columnCount}>{col.count}</span>
                    </div>
                    <div className={styles.columnCards}>
                      {col.cards.map((card) => (
                        <div key={card.title} className={styles.taskCard}>
                          <div className={styles.taskTitle}>{card.title}</div>
                          <div className={styles.taskMeta}>
                            {card.tag && (
                              <span
                                className={styles.taskTag}
                                style={{
                                  color: col.tint,
                                  borderColor: `${col.tint}40`,
                                  background: `${col.tint}10`,
                                }}
                              >
                                {card.tag}
                              </span>
                            )}
                            {card.priority === 'high' && (
                              <span className={styles.taskPriority}>High</span>
                            )}
                            <span className={styles.taskAssignee}>{card.assignee}</span>
                          </div>
                        </div>
                      ))}
                      <div className={styles.taskGhost}>+ Add task</div>
                    </div>
                  </div>
                ))}
              </div>
            </motion.div>
          </section>

          {/* ── FEATURE GRID — 3×2 bordered cells ─────────────────────── */}
          <section className={styles.featuresSection}>
            <motion.div
              className={styles.featuresHeader}
              initial="hidden"
              whileInView="visible"
              viewport={viewportOnce}
              variants={staggerBlur}
            >
              <motion.span className={styles.sectionEyebrow} variants={blurFadeIn}>
                WHAT FLOW DOES
              </motion.span>
              <motion.h2 className={styles.sectionTitle} variants={blurFadeIn}>
                The work OS your team
                <br />
                will actually open.
              </motion.h2>
            </motion.div>

            <motion.div
              className={styles.featuresGrid}
              initial="hidden"
              whileInView="visible"
              viewport={viewportOnce}
              variants={staggerBlur}
            >
              {FEATURES.map(({ icon: Icon, title, copy }) => (
                <motion.div
                  key={title}
                  className={styles.featureCard}
                  variants={blurFadeIn}
                  whileHover={{ y: -4 }}
                >
                  <div className={styles.featureIcon}>
                    <Icon size={18} strokeWidth={1.8} />
                  </div>
                  <h3 className={styles.featureTitle}>{title}</h3>
                  <p className={styles.featureCopy}>{copy}</p>
                </motion.div>
              ))}
            </motion.div>
          </section>

          {/* ── CONTEXT STORY — tasks linked to real objects ──────────── */}
          <motion.section
            className={styles.contextSection}
            initial="hidden"
            whileInView="visible"
            viewport={viewportOnce}
            variants={staggerBlur}
          >
            <motion.div className={styles.contextInner} variants={blurFadeIn}>
              <span className={styles.sectionEyebrow}>TASKS WITH CONTEXT</span>
              <h2 className={styles.contextTitle}>
                A task that knows the invoice,
                <br />
                the customer, and the shift.
              </h2>
              <p className={styles.contextCopy}>
                In rtylr Flow, tasks are first-class citizens linked to other objects in the
                platform. &ldquo;Follow up on invoice #4821&rdquo; opens the invoice. &ldquo;Cover
                Sam&apos;s Friday closing shift&rdquo; jumps to the schedule. No tab-switching, no
                copy-paste, no &ldquo;wait, which one?&rdquo;
              </p>
              <div className={styles.linkedBadges}>
                <span className={styles.linkedBadge}>Linked · Invoice</span>
                <span className={styles.linkedBadge}>Linked · Customer</span>
                <span className={styles.linkedBadge}>Linked · Shift</span>
                <span className={styles.linkedBadge}>Linked · Supplier</span>
                <span className={styles.linkedBadge}>Linked · Menu item</span>
                <span className={styles.linkedBadge}>Linked · Location</span>
              </div>
            </motion.div>
          </motion.section>

          {/* ── CTA ──────────────────────────────────────────────────── */}
          <motion.section
            className={styles.ctaBlock}
            initial="hidden"
            whileInView="visible"
            viewport={viewportOnce}
            variants={staggerBlur}
          >
            <motion.h2 className={styles.ctaTitle} variants={blurFadeIn}>
              Stop losing work
              <br />
              between tools.
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

export default FlowPage;
