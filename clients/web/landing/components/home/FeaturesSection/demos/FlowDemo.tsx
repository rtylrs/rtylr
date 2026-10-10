import React from 'react';
import { LayoutGroup, motion } from 'motion/react';
import { useTranslation } from 'react-i18next';
import { cn } from '../../../../lib/cn';
import { PortalFrame, type PortalView } from './portal/PortalFrame';
import { useDemoStep } from './portal/usePortalDemo';
import styles from './portal/PortalDemo.module.scss';
import local from './FlowDemo.module.scss';

type Tag = 'design' | 'web' | 'ops' | 'sales';

interface Task {
  id: number; // index into home.features.demos.flow.tasks
  tag: Tag;
  who: string;
  color: string;
  day: number;
}

const TASKS: Task[] = [
  { id: 0, tag: 'web', who: 'MK', color: '#8b5cf6', day: 1 },
  { id: 1, tag: 'ops', who: 'SA', color: '#10b981', day: 0 },
  { id: 2, tag: 'design', who: 'LR', color: '#f59e0b', day: 3 },
  { id: 3, tag: 'web', who: 'JD', color: '#3b82f6', day: 2 },
  { id: 4, tag: 'web', who: 'MK', color: '#8b5cf6', day: 4 },
  { id: 5, tag: 'sales', who: 'SA', color: '#10b981', day: 4 },
  { id: 6, tag: 'web', who: 'JD', color: '#3b82f6', day: 1 },
];

// Column per task at each step. Task 2 ("Design pricing page") walks from
// In progress to Review to Done; everything else stays put.
const BASE = [0, 1, 1, 2, 0, 3, 3];
const MOVING = 2;
const MOVING_PATH = [1, 2, 3];

export const FlowDemo = ({
  isActive,
  onNavigate,
}: {
  isActive: boolean;
  onNavigate?: (view: PortalView) => void;
}): React.JSX.Element => {
  const { t } = useTranslation();
  const step = useDemoStep(isActive, MOVING_PATH.length, 1800);
  const columns = t('home.features.demos.flow.columns', { returnObjects: true }) as string[];
  const titles = t('home.features.demos.flow.tasks', { returnObjects: true }) as string[];
  const days = t('home.features.demos.flow.days', { returnObjects: true }) as string[];
  const views = t('home.features.demos.flow.views', { returnObjects: true }) as string[];

  const columnOf = (task: Task): number => (task.id === MOVING ? MOVING_PATH[step] : BASE[task.id]);

  return (
    <PortalFrame
      onNavigate={onNavigate}
      app="flow"
      crumb={t('home.features.demos.flow.crumb', { returnObjects: true }) as string[]}
      actions={
        <span className={styles.segmented}>
          {views.map((v, i) => (
            <span key={v} className={cn(styles.segment, i === 0 && styles.segmentOn)}>
              {v}
            </span>
          ))}
        </span>
      }
    >
      <LayoutGroup>
        <div className={local.board}>
          {columns.map((name, col) => {
            const tasks = TASKS.filter((task) => columnOf(task) === col);
            return (
              <div key={name} className={local.column}>
                <div className={local.columnHead}>
                  <span className={local.dot} data-col={col} />
                  {name}
                  <span className={styles.muted}>{tasks.length}</span>
                </div>
                {tasks.map((task) => (
                  <motion.div
                    key={task.id}
                    layout
                    layoutId={`flow-task-${task.id}`}
                    transition={{ type: 'spring', stiffness: 260, damping: 30 }}
                    className={cn(local.task, task.id === MOVING && local.taskMoving)}
                  >
                    <div className={col === 3 ? local.taskDone : undefined}>{titles[task.id]}</div>
                    <div className={styles.spread}>
                      <span className={styles.pill}>
                        {t(`home.features.demos.flow.tags.${task.tag}`)}
                      </span>
                      <span className={local.taskMeta}>
                        <span className={styles.muted}>{days[task.day]}</span>
                        <span className={styles.avatar} style={{ background: task.color }}>
                          {task.who}
                        </span>
                      </span>
                    </div>
                  </motion.div>
                ))}
              </div>
            );
          })}
        </div>
      </LayoutGroup>
    </PortalFrame>
  );
};
