import React, { useEffect, useRef, useState } from 'react';
import { cn } from '../../lib/cn';
import { motion, AnimatePresence, type PanInfo } from 'motion/react';
import { useTranslation } from 'react-i18next';
import { useRouter } from 'next/router';
import {
  CheckSquare,
  LayoutGrid,
  GitBranch,
  Zap,
  MessageSquare,
  Calendar,
  ShoppingCart,
  Package,
  Users,
  UserCircle2,
  DollarSign,
  Workflow as WorkflowIcon,
  ArrowRight,
  Play,
  Pause,
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
import styles from './flow.module.scss';

// ── Board data types ───────────────────────────────────────────────────────

type Priority = 'high' | 'med';

interface Task {
  id: string;
  title: string;
  tag?: string;
  assignee: string;
  priority?: Priority;
}

interface Column {
  id: string;
  title: string;
  tint: string;
  tasks: Task[];
}

// Initial board state, hydrates from this once, then lives in component state
const INITIAL_COLUMNS: Column[] = [
  {
    id: 'backlog',
    title: 'Backlog',
    tint: '#8a8a8a',
    tasks: [
      { id: 't1', title: 'Seasonal menu rollout, Q3', assignee: 'MC', tag: 'Ops' },
      { id: 't2', title: 'New supplier onboarding: Rossini', assignee: 'JK', tag: 'ERP' },
    ],
  },
  {
    id: 'progress',
    title: 'In progress',
    tint: '#d6b88e',
    tasks: [
      {
        id: 't3',
        title: 'Friday dinner covers, staffing plan',
        assignee: 'SP',
        tag: 'HR',
        priority: 'high',
      },
      { id: 't4', title: 'Redo brunch prep checklist', assignee: 'AR', tag: 'Kitchen' },
      { id: 't5', title: 'Email blast: Mother\u2019s Day brunch', assignee: 'MC', tag: 'CRM' },
    ],
  },
  {
    id: 'review',
    title: 'Review',
    tint: '#8ca39d',
    tasks: [{ id: 't6', title: 'Espresso machine repair SLA', assignee: 'JK', tag: 'Vendor' }],
  },
  {
    id: 'done',
    title: 'Done',
    tint: '#6bcf7f',
    tasks: [
      { id: 't7', title: 'Q1 sales-tax filing', assignee: 'SP', tag: 'Finance' },
      { id: 't8', title: 'POS firmware v4.2 rollout', assignee: 'AR', tag: 'Ops' },
    ],
  },
];

// ── Automation builder rules ──────────────────────────────────────────────
// Each rule shows a source event → action. The middle section cycles through
// these with an animated connecting wire.

interface Rule {
  source: {
    app: string;
    appIcon: React.ComponentType<{ size?: number; strokeWidth?: number }>;
    appColor: string;
    trigger: string;
    detail: string;
  };
  action: {
    app: string;
    appIcon: React.ComponentType<{ size?: number; strokeWidth?: number }>;
    appColor: string;
    label: string;
    detail: string;
  };
}

const RULES: Rule[] = [
  {
    source: {
      app: 'ERP',
      appIcon: Package,
      appColor: '#a2d9af',
      trigger: 'Purchase order approved',
      detail: 'PO #4821 · Rossini supplies · $2,480',
    },
    action: {
      app: 'Flow',
      appIcon: WorkflowIcon,
      appColor: '#d6b88e',
      label: 'Create task on purchasing',
      detail: 'Assigned to head of purchasing · due tomorrow',
    },
  },
  {
    source: {
      app: 'HR',
      appIcon: Users,
      appColor: '#f3c08a',
      trigger: 'Shift marked as "no-show"',
      detail: 'Maya Chen · Friday 11-7',
    },
    action: {
      app: 'Flow',
      appIcon: WorkflowIcon,
      appColor: '#d6b88e',
      label: 'Escalate to manager',
      detail: 'Coverage task opened · SMS sent to on-call',
    },
  },
  {
    source: {
      app: 'POS',
      appIcon: ShoppingCart,
      appColor: '#b3c5e8',
      trigger: 'Daily sales close',
      detail: '$12,840 processed · drawer $240 over',
    },
    action: {
      app: 'Finance',
      appIcon: DollarSign,
      appColor: '#a2d9af',
      label: 'Book journal entry',
      detail: 'Cash + card reconciled · variance flagged for review',
    },
  },
  {
    source: {
      app: 'CRM',
      appIcon: UserCircle2,
      appColor: '#c9a48a',
      trigger: 'Guest hits VIP threshold',
      detail: 'Eliza Alvarez · 20th lifetime visit',
    },
    action: {
      app: 'CRM',
      appIcon: UserCircle2,
      appColor: '#c9a48a',
      label: 'Send thank-you email + next-visit comp',
      detail: 'Template: vip-reward · drafted for approval',
    },
  },
  {
    source: {
      app: 'Finance',
      appIcon: DollarSign,
      appColor: '#a2d9af',
      trigger: 'Invoice overdue 14 days',
      detail: 'INV-0294 · Byblos Catering · $6,400',
    },
    action: {
      app: 'Flow',
      appIcon: WorkflowIcon,
      appColor: '#d6b88e',
      label: 'Task the AR owner',
      detail: 'Second reminder scheduled · escalation in 7d',
    },
  },
  {
    source: {
      app: 'ERP',
      appIcon: Package,
      appColor: '#a2d9af',
      trigger: 'Stock drops below reorder point',
      detail: 'Olive Oil (5L) · 6 units left',
    },
    action: {
      app: 'ERP',
      appIcon: Package,
      appColor: '#a2d9af',
      label: 'Draft PO to preferred supplier',
      detail: 'Rossini · quantity 24 · awaiting approval',
    },
  },
];

// Three icons, because the strip renders three cards. The array used to
// hold six, and three of them had never been on the page.
const FEATURES_ICONS = [LayoutGrid, GitBranch, Zap];

// ── Draggable task card ────────────────────────────────────────────────────
// On drag end, it reads the element rects of every column to decide which one
// the card was released over, and calls the parent callback to update state.

interface DragTaskCardProps {
  task: Task;
  columnTint: string;
  columnIds: string[];
  columnRefs: React.RefObject<Map<string, HTMLDivElement | null>>;
  onDropInColumn: (taskId: string, targetColumnId: string) => void;
}

const DragTaskCard = ({
  task,
  columnTint,
  columnIds,
  columnRefs,
  onDropInColumn,
}: DragTaskCardProps): React.JSX.Element => {
  const handleDragEnd = (_: PointerEvent | MouseEvent | TouchEvent, info: PanInfo): void => {
    const x = info.point.x;
    const y = info.point.y;
    const refs = columnRefs.current;
    if (!refs) return;
    for (const colId of columnIds) {
      const el = refs.get(colId);
      if (!el) continue;
      const rect = el.getBoundingClientRect();
      if (x >= rect.left && x <= rect.right && y >= rect.top && y <= rect.bottom) {
        onDropInColumn(task.id, colId);
        return;
      }
    }
  };

  return (
    <motion.div
      layout
      layoutId={task.id}
      drag
      dragMomentum={false}
      dragElastic={0.18}
      dragSnapToOrigin
      whileDrag={{ scale: 1.03, zIndex: 10, boxShadow: '0 20px 40px rgba(0,0,0,0.45)' }}
      whileHover={{ y: -1 }}
      onDragEnd={handleDragEnd}
      className={styles.taskCard}
      transition={{ type: 'spring', stiffness: 500, damping: 40 }}
    >
      <div className={styles.taskTitle}>{task.title}</div>
      <div className={styles.taskMeta}>
        {task.tag && (
          <span
            className={styles.taskTag}
            style={{
              color: columnTint,
              borderColor: `${columnTint}40`,
              background: `${columnTint}10`,
            }}
          >
            {task.tag}
          </span>
        )}
        {task.priority === 'high' && <span className={styles.taskPriority}>High</span>}
        <span className={styles.taskAssignee}>{task.assignee}</span>
      </div>
    </motion.div>
  );
};

// ── Page ──────────────────────────────────────────────────────────────────

const FlowPage = (): React.JSX.Element => {
  const faq = useFaq('flow');
  const { t } = useTranslation();
  const features = t('flow.features.items', { returnObjects: true }) as unknown as {
    title: string;
    copy: string;
  }[];
  const router = useRouter();
  const [columns, setColumns] = useState<Column[]>(INITIAL_COLUMNS);
  const [ruleIdx, setRuleIdx] = useState(0);
  const [autoPlay, setAutoPlay] = useState(true);

  // Cycle through automation examples every 3s while auto-play is on
  useEffect(() => {
    if (!autoPlay) return;
    const t = setInterval(() => {
      setRuleIdx((i) => (i + 1) % RULES.length);
    }, 3200);
    return () => clearInterval(t);
  }, [autoPlay]);

  const selectRule = (idx: number): void => {
    setRuleIdx(idx);
    setAutoPlay(false);
  };

  const rule = RULES[ruleIdx];

  // Refs keyed by column id, the draggable card walks these to find its
  // drop target on release.
  const columnRefs = useRef<Map<string, HTMLDivElement | null>>(new Map());
  const columnIds = columns.map((c) => c.id);

  const moveTask = (taskId: string, targetColumnId: string): void => {
    setColumns((prev) => {
      let moving: Task | null = null;
      const without = prev.map((col) => {
        const found = col.tasks.find((t) => t.id === taskId);
        if (found) {
          moving = found;
          return { ...col, tasks: col.tasks.filter((t) => t.id !== taskId) };
        }
        return col;
      });
      if (!moving) return prev;
      return without.map((col) =>
        col.id === targetColumnId
          ? { ...col, tasks: [...col.tasks, moving as unknown as Task] }
          : col
      );
    });
  };

  const totalTasks = columns.reduce((acc, c) => acc + c.tasks.length, 0);

  const goRegister = (): void => {
    navigateExternal(`${APP_URLS.auth}/register?redirect=flow`);
  };

  const goDemo = (): void => {
    void router.push('/demo');
  };

  return (
    <AppLayout>
      <MainLayout>
        <HelmetContainer
          title={t('flow.meta.title')}
          description={t('flow.meta.description')}
          canonical="https://rtylr.com/flow/"
          structuredData={productWithBreadcrumbs({
            name: 'Flow',
            path: '/flow/',
            subCategory: 'Project and Task Management',
            description: t('flow.meta.schemaDescription'),
          })}
        />

        <div className={styles.page}>
          {/* ── HERO ─────────────────────────────────────────────────── */}
          <section className={styles.hero}>
            <motion.div
              className={styles.heroCopy}
              initial="hidden"
              animate="visible"
              variants={staggerBlur}
            >
              <motion.span className={styles.heroEyebrow} variants={blurFadeIn}>
                {t('flow.hero.eyebrow')}
              </motion.span>
              <h1 className={styles.heroTitle}>
                <SplitWords as="span" text={t('flow.hero.titleA')} delay={0.1} />
                <br />
                <span className={styles.heroTitleSoft}>
                  <SplitWords as="span" text={t('flow.hero.titleB')} delay={0.35} />
                </span>
              </h1>
              <motion.p className={styles.heroSubtitle} variants={blurFadeIn}>
                {t('flow.hero.subtitle')}
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

            {/* Live kanban, cards drag between columns */}
            <motion.div
              className={styles.board}
              initial={{ opacity: 0, y: 40 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.4, duration: 1.0, ease: [0.16, 1, 0.3, 1] }}
            >
              <div className={styles.boardHeader}>
                <div className={styles.boardCrumbs}>
                  <span className={styles.boardCrumb}>Workspaces</span>
                  <span className={styles.boardCrumbSep}>/</span>
                  <span className={styles.boardCrumb}>Operations</span>
                  <span className={styles.boardCrumbSep}>/</span>
                  <span className={styles.boardCrumbActive}>Weekly sprint</span>
                </div>
                <div className={styles.boardHeaderRight}>
                  <span className={styles.boardCount}>{totalTasks} tasks</span>
                  <div className={styles.boardViews}>
                    <span className={cn(styles.boardView, styles.boardViewActive)}>Board</span>
                    <span className={styles.boardView}>List</span>
                    <span className={styles.boardView}>Timeline</span>
                  </div>
                </div>
              </div>

              <div className={styles.boardColumns}>
                {columns.map((col) => (
                  <div
                    key={col.id}
                    ref={(el) => {
                      columnRefs.current.set(col.id, el);
                    }}
                    className={styles.column}
                  >
                    <div className={styles.columnHeader}>
                      <span className={styles.columnTitle}>
                        <span className={styles.columnDot} style={{ background: col.tint }} />
                        {col.title}
                      </span>
                      <span className={styles.columnCount}>{col.tasks.length}</span>
                    </div>
                    <div className={styles.columnCards}>
                      <AnimatePresence>
                        {col.tasks.map((task) => (
                          <DragTaskCard
                            key={task.id}
                            task={task}
                            columnTint={col.tint}
                            columnIds={columnIds}
                            columnRefs={columnRefs}
                            onDropInColumn={moveTask}
                          />
                        ))}
                      </AnimatePresence>
                      <div className={styles.taskGhost}>+ Add task</div>
                    </div>
                  </div>
                ))}
              </div>
            </motion.div>
          </section>

          <DirectAnswer label={t('flow.directAnswer.label')}>
            {t('flow.directAnswer.body')}
          </DirectAnswer>

          {/* ── AUTOMATION BUILDER ─────────────────────────────────────── */}
          <section className={styles.automationSection}>
            <motion.div
              className={styles.automationHeader}
              initial="hidden"
              whileInView="visible"
              viewport={viewportOnce}
              variants={staggerBlur}
            >
              <motion.span className={styles.sectionEyebrow} variants={blurFadeIn}>
                AUTOMATIONS · NO ZAPIER REQUIRED
              </motion.span>
              <motion.h2 className={styles.sectionTitle} variants={blurFadeIn}>
                How do automations
                <br />
                work in Flow?
              </motion.h2>
              <motion.p className={styles.automationCopy} variants={blurFadeIn}>
                Every app in rtylr emits events. Every app listens. Click an example, or let it
                auto-play, to see what real rules look like.
              </motion.p>
            </motion.div>

            <motion.div
              className={styles.automationBoard}
              initial="hidden"
              whileInView="visible"
              viewport={viewportOnce}
              variants={staggerBlur}
            >
              <AnimatePresence mode="wait">
                <motion.div
                  key={ruleIdx}
                  className={styles.automationWire}
                  initial={{ opacity: 0, y: 14 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0, y: -8 }}
                  transition={{ duration: 0.35, ease: [0.16, 1, 0.3, 1] }}
                >
                  {/* Trigger card */}
                  <div className={styles.ruleCard}>
                    <div className={styles.ruleCardHead}>
                      <span
                        className={styles.ruleChip}
                        style={{
                          color: rule.source.appColor,
                          background: `${rule.source.appColor}18`,
                          borderColor: `${rule.source.appColor}44`,
                        }}
                      >
                        <rule.source.appIcon size={11} strokeWidth={2} />
                        {rule.source.app}
                      </span>
                      <span className={styles.ruleCardLabel}>WHEN</span>
                    </div>
                    <div className={styles.ruleCardTrigger}>{rule.source.trigger}</div>
                    <div className={styles.ruleCardDetail}>{rule.source.detail}</div>
                  </div>

                  {/* Animated wire between cards */}
                  <div className={styles.wireTrack} aria-hidden="true">
                    <svg className={styles.wireSvg} viewBox="0 0 100 40" preserveAspectRatio="none">
                      <line x1="0" y1="20" x2="100" y2="20" className={styles.wireLine} />
                    </svg>
                    <motion.span
                      className={styles.wirePulse}
                      initial={{ left: '0%' }}
                      animate={{ left: '100%' }}
                      transition={{
                        duration: 1.4,
                        repeat: Infinity,
                        ease: 'easeInOut',
                      }}
                    />
                    <span className={styles.wireArrow}>
                      <ArrowRight size={14} strokeWidth={2} />
                    </span>
                  </div>

                  {/* Action card */}
                  <div className={styles.ruleCard}>
                    <div className={styles.ruleCardHead}>
                      <span
                        className={styles.ruleChip}
                        style={{
                          color: rule.action.appColor,
                          background: `${rule.action.appColor}18`,
                          borderColor: `${rule.action.appColor}44`,
                        }}
                      >
                        <rule.action.appIcon size={11} strokeWidth={2} />
                        {rule.action.app}
                      </span>
                      <span className={styles.ruleCardLabel}>THEN</span>
                    </div>
                    <div className={styles.ruleCardTrigger}>{rule.action.label}</div>
                    <div className={styles.ruleCardDetail}>{rule.action.detail}</div>
                  </div>
                </motion.div>
              </AnimatePresence>

              {/* Rule picker + play controls */}
              <div className={styles.rulePicker}>
                <button
                  type="button"
                  onClick={() => setAutoPlay((v) => !v)}
                  className={styles.playToggle}
                  aria-label={autoPlay ? 'Pause examples' : 'Play examples'}
                  title={autoPlay ? 'Pause' : 'Play'}
                >
                  {autoPlay ? (
                    <Pause size={11} strokeWidth={2.5} fill="currentColor" />
                  ) : (
                    <Play size={11} strokeWidth={2.5} fill="currentColor" />
                  )}
                </button>
                <div className={styles.ruleDots}>
                  {RULES.map((_, i) => (
                    <button
                      key={i}
                      type="button"
                      onClick={() => selectRule(i)}
                      className={cn(styles.ruleDot, i === ruleIdx && styles.ruleDotActive)}
                      aria-label={`Example ${i + 1}`}
                    />
                  ))}
                </div>
                <span className={styles.ruleCounter}>
                  {ruleIdx + 1} / {RULES.length}
                </span>
              </div>
            </motion.div>

            {/* Small features strip, 3 up, not a full grid */}
            <motion.div
              className={styles.featuresStrip}
              initial="hidden"
              whileInView="visible"
              viewport={viewportOnce}
              variants={staggerBlur}
            >
              {features.map(({ title, copy }, i) => {
                const Icon = FEATURES_ICONS[i];
                return (
                  <motion.div
                    key={title}
                    className={styles.featureStripCard}
                    variants={blurFadeIn}
                  >
                    <div className={styles.featureIcon}>
                      <Icon size={16} strokeWidth={1.8} />
                    </div>
                    <div>
                      <h3 className={styles.featureTitle}>{title}</h3>
                      <p className={styles.featureCopy}>{copy}</p>
                    </div>
                  </motion.div>
                );
              })}
            </motion.div>
          </section>

          {/* ── CONTEXT STORY ─────────────────────────────────────────── */}
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
                What does a task in Flow
                <br />
                already know about?
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

          <FaqSection title={faq.title} faqs={faq.items} />

          {/* ── CTA ──────────────────────────────────────────────────── */}
          <motion.section
            className={styles.ctaBlock}
            initial="hidden"
            whileInView="visible"
            viewport={viewportOnce}
            variants={staggerBlur}
          >
            <motion.h2 className={styles.ctaTitle} variants={blurFadeIn}>
              <Lines text={t('flow.cta.title')} />
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
