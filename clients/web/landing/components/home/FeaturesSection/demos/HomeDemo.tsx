import React, { useState } from 'react';
import { motion } from 'motion/react';
import { useTranslation } from 'react-i18next';
import { Check } from 'lucide-react';
import { cn } from '../../../../lib/cn';
import { PortalFrame, type AppId, type PortalView } from './portal/PortalFrame';
import { money, useCountUp } from './portal/usePortalDemo';
import styles from './portal/PortalDemo.module.scss';
import local from './HomeDemo.module.scss';

// Activity rows link to the app they came from; colours match the rail.
const ACTIVITY: { app: AppId; color: string; when: string }[] = [
  { app: 'finance', color: '#f59e0b', when: '2h' },
  { app: 'hr', color: '#10b981', when: '10:30' },
  { app: 'erp', color: '#64c5b9', when: '1h' },
  { app: 'crm', color: '#8b5cf6', when: '5h' },
];

/**
 * The portal's home screen for the hero: greeting, three numbers, today's
 * tasks (tickable) and recent activity (each row opens its app).
 */
export const HomeDemo = ({
  isActive,
  onNavigate,
}: {
  isActive: boolean;
  onNavigate?: (view: PortalView) => void;
}): React.JSX.Element => {
  const { t } = useTranslation();
  const tasks = t('home.hero.portal.home.tasks', { returnObjects: true }) as string[];
  const activity = t('home.hero.portal.home.activity', { returnObjects: true }) as string[];
  const [done, setDone] = useState<boolean[]>(() => tasks.map((_, i) => i === 0));
  const revenue = useCountUp(48200, isActive);

  const toggle = (i: number): void => setDone((d) => d.map((v, k) => (k === i ? !v : v)));
  const openTasks = done.filter((d) => !d).length;

  return (
    <PortalFrame
      onNavigate={onNavigate}
      app="home"
      crumb={[t('home.hero.portal.apps.home')]}
    >
      <div className={local.layout}>
        <div>
          {/* Not a heading: this sits under the page's h1, and a mock screen
              shouldn't add to the document outline. */}
          <div className={local.greeting}>{t('home.hero.portal.home.greeting')}</div>
          <div className={styles.muted}>{t('home.hero.portal.home.date')}</div>
        </div>

        <div className={styles.kpis}>
          <div className={styles.kpi}>
            <div className={styles.kpiLabel}>{t('home.hero.portal.home.kpis.revenue')}</div>
            <div className={styles.kpiValue}>{money(revenue)}</div>
            <div className={styles.delta}>+12.4%</div>
          </div>
          <div className={styles.kpi}>
            <div className={styles.kpiLabel}>{t('home.hero.portal.home.kpis.tasks')}</div>
            <div className={styles.kpiValue}>{openTasks}</div>
            <div className={styles.muted}>{t('home.hero.portal.home.kpis.tasksNote')}</div>
          </div>
          <div className={styles.kpi}>
            <div className={styles.kpiLabel}>{t('home.hero.portal.home.kpis.overdue')}</div>
            <div className={styles.kpiValue}>3</div>
            <div className={cn(styles.delta, styles.deltaDown)}>{money(6450)}</div>
          </div>
        </div>

        <div className={local.columns}>
          <div className={cn(styles.card, local.panel)}>
            <span className={local.panelTitle}>{t('home.hero.portal.home.today')}</span>
            {tasks.map((task, i) => (
              <button
                key={task}
                type="button"
                className={local.task}
                onClick={() => toggle(i)}
                aria-pressed={done[i]}
              >
                <motion.span
                  className={cn(local.check, done[i] && local.checkOn)}
                  animate={{ scale: done[i] ? [0.8, 1] : 1 }}
                  transition={{ duration: 0.2 }}
                >
                  {done[i] && <Check size={10} strokeWidth={3.5} />}
                </motion.span>
                <span className={done[i] ? local.taskDone : undefined}>{task}</span>
              </button>
            ))}
          </div>

          <div className={cn(styles.card, local.panel)}>
            <span className={local.panelTitle}>{t('home.hero.portal.home.recent')}</span>
            {ACTIVITY.map((a, i) => (
              <button
                key={a.app}
                type="button"
                className={local.activity}
                onClick={() => onNavigate?.(a.app)}
              >
                <span className={local.activityDot} style={{ background: a.color }} />
                <span className={local.activityText}>{activity[i]}</span>
                <span className={styles.muted}>{a.when}</span>
              </button>
            ))}
          </div>
        </div>
      </div>
    </PortalFrame>
  );
};
