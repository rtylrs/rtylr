import React, { useRef, useState } from 'react';
import { motion, AnimatePresence, type PanInfo } from 'motion/react';
import { useRouter } from 'next/router';
import { CheckSquare, LayoutGrid, GitBranch, Zap, MessageSquare, Calendar } from 'lucide-react';
import AppLayout from '../../components/Layout/AppLayout';
import MainLayout from '../../layouts/Main/MainLayout';
import HelmetContainer from '../../components/HelmetContainer';
import { APP_URLS } from '@voxire/config';
import { navigateExternal } from '../../lib/navigation';
import { blurFadeIn, staggerBlur, viewportOnce } from '../../lib/animations';
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

// Initial board state — hydrates from this once, then lives in component state
const INITIAL_COLUMNS: Column[] = [
  {
    id: 'backlog',
    title: 'Backlog',
    tint: '#8a8a8a',
    tasks: [
      { id: 't1', title: 'Seasonal menu rollout — Q3', assignee: 'MC', tag: 'Ops' },
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
        title: 'Friday dinner covers — staffing plan',
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
    title: 'Automations without Zapier',
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
    copy: 'Weekly checklists, monthly closes, new-location launches — reusable project templates.',
  },
];

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

const FlowPage = (): JSX.Element => {
  const router = useRouter();
  const [columns, setColumns] = useState<Column[]>(INITIAL_COLUMNS);

  // Refs keyed by column id — the draggable card walks these to find its
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
          {/* ── HERO ─────────────────────────────────────────────────── */}
          <section className={styles.hero}>
            <motion.div
              className={styles.heroCopy}
              initial="hidden"
              animate="visible"
              variants={staggerBlur}
            >
              <motion.span className={styles.heroEyebrow} variants={blurFadeIn}>
                rtylr Flow · try dragging the cards below
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

            {/* Live kanban — cards drag between columns */}
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
                    <span className={`${styles.boardView} ${styles.boardViewActive}`}>Board</span>
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

          {/* ── FEATURE GRID ──────────────────────────────────────────── */}
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
