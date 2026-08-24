import React, { useState, useEffect, useRef } from 'react';
import { motion, AnimatePresence } from 'motion/react';
import { useTranslation } from 'react-i18next';

interface Task {
  id: string;
  tagKey: 'ops' | 'erp' | 'hr' | 'kitchen' | 'crm' | 'vendor' | 'finance';
  assignee: string;
  priority?: 'high' | 'med';
}

interface Column {
  id: string;
  titleKey: 'backlog' | 'progress' | 'review' | 'done';
  tint: string;
  tasks: Task[];
}

// title comes from home.features.demos.flow.tasks[id], tag from .tags[tagKey].
const INITIAL: Column[] = [
  {
    id: 'backlog',
    titleKey: 'backlog',
    tint: '#8a8a8a',
    tasks: [
      { id: 't1', tagKey: 'ops', assignee: 'MC' },
      { id: 't2', tagKey: 'erp', assignee: 'JK' },
    ],
  },
  {
    id: 'progress',
    titleKey: 'progress',
    tint: '#d6b88e',
    tasks: [
      {
        id: 't3',
        tagKey: 'hr',
        assignee: 'SP',
        priority: 'high',
      },
      { id: 't4', tagKey: 'kitchen', assignee: 'AR' },
      { id: 't5', tagKey: 'crm', assignee: 'MC' },
    ],
  },
  {
    id: 'review',
    titleKey: 'review',
    tint: '#8ca39d',
    tasks: [{ id: 't6', tagKey: 'vendor', assignee: 'JK' }],
  },
  {
    id: 'done',
    titleKey: 'done',
    tint: '#6bcf7f',
    tasks: [
      { id: 't7', tagKey: 'finance', assignee: 'SP' },
      { id: 't8', tagKey: 'ops', assignee: 'AR' },
    ],
  },
];

const COL_ORDER = ['backlog', 'progress', 'review', 'done'];

export const FlowDemo = ({ isActive }: { isActive: boolean }): React.JSX.Element => {
  const { t } = useTranslation();
  const [cols, setCols] = useState<Column[]>(INITIAL);
  const tickRef = useRef(0);

  const moveTask = (): void => {
    setCols((prev) => {
      const moveable = prev
        .filter((c) => c.id !== 'done' && c.tasks.length > 0)
        .flatMap((c) => c.tasks.map((task) => ({ task, colId: c.id })));
      if (!moveable.length) return prev;
      const pick = moveable[tickRef.current % moveable.length];
      const fromIdx = COL_ORDER.indexOf(pick.colId);
      const toId = COL_ORDER[Math.min(fromIdx + 1, COL_ORDER.length - 1)];
      if (toId === pick.colId) return prev;
      return prev.map((c) => {
        if (c.id === pick.colId)
          return { ...c, tasks: c.tasks.filter((task) => task.id !== pick.task.id) };
        if (c.id === toId) return { ...c, tasks: [pick.task, ...c.tasks] };
        return c;
      });
    });
    tickRef.current += 1;
  };

  useEffect(() => {
    if (!isActive) return;
    const timer = setInterval(moveTask, 2800);
    return () => clearInterval(timer);
  }, [isActive]);

  return (
    <div
      style={{
        position: 'absolute',
        inset: 0,
        background: 'var(--bg-surface-a)',
        fontFamily: 'inherit',
        fontSize: 13,
        display: 'flex',
        flexDirection: 'column',
        overflow: 'hidden',
      }}
    >
      {/* Header */}
      <div
        style={{
          padding: '14px 16px 12px',
          borderBottom: '1px solid var(--hairline)',
          display: 'flex',
          alignItems: 'center',
          gap: 8,
          flexShrink: 0,
        }}
      >
        <span style={{ color: 'rgba(19,19,19,0.55)', fontSize: 12 }}>
          {t('home.features.demos.flow.breadcrumb.workspaces')}
        </span>
        <span style={{ color: 'rgba(19,19,19,0.3)', fontSize: 12 }}>/</span>
        <span style={{ color: 'rgba(19,19,19,0.55)', fontSize: 12 }}>
          {t('home.features.demos.flow.breadcrumb.operations')}
        </span>
        <span style={{ color: 'rgba(19,19,19,0.3)', fontSize: 12 }}>/</span>
        <span style={{ color: 'var(--ink)', fontSize: 12, fontWeight: 500 }}>
          {t('home.features.demos.flow.breadcrumb.sprint')}
        </span>
        <div style={{ marginInlineStart: 'auto', display: 'flex', gap: 6 }}>
          {(['board', 'list', 'timeline'] as const).map((v, i) => (
            <div
              key={v}
              style={{
                padding: '4px 10px',
                borderRadius: 6,
                fontSize: 11,
                fontWeight: 500,
                background: i === 0 ? 'var(--bg-surface-b)' : 'transparent',
                color: i === 0 ? 'var(--ink)' : 'rgba(19,19,19,0.45)',
                border: i === 0 ? '1px solid var(--hairline-strong)' : '1px solid transparent',
              }}
            >
              {t(`home.features.demos.flow.views.${v}`)}
            </div>
          ))}
        </div>
      </div>

      {/* Board */}
      <div
        style={{
          flex: 1,
          display: 'grid',
          gridTemplateColumns: 'repeat(4, 1fr)',
          gap: 10,
          padding: '14px 14px',
          overflow: 'hidden',
          minHeight: 0,
        }}
      >
        {cols.map((col) => (
          <div
            key={col.id}
            style={{
              background: 'var(--bg-surface-b)',
              border: '1px solid var(--hairline)',
              borderRadius: 10,
              padding: '12px 10px',
              display: 'flex',
              flexDirection: 'column',
              overflow: 'hidden',
              minHeight: 0,
            }}
          >
            {/* Column header */}
            <div
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 7,
                marginBottom: 10,
                flexShrink: 0,
              }}
            >
              <div
                style={{
                  width: 7,
                  height: 7,
                  borderRadius: '50%',
                  background: col.tint,
                  flexShrink: 0,
                }}
              />
              <span style={{ fontSize: 11, fontWeight: 600, color: 'var(--ink)' }}>
                {t(`home.features.demos.flow.columns.${col.titleKey}`)}
              </span>
              <span
                style={{
                  marginInlineStart: 'auto',
                  fontSize: 10,
                  color: 'rgba(19,19,19,0.5)',
                  background: 'var(--hairline)',
                  borderRadius: 999,
                  padding: '1px 6px',
                }}
              >
                {col.tasks.length}
              </span>
            </div>

            {/* Cards */}
            <div
              style={{
                flex: 1,
                display: 'flex',
                flexDirection: 'column',
                gap: 7,
                overflowY: 'hidden',
              }}
            >
              <AnimatePresence mode="popLayout">
                {col.tasks.map((task) => (
                  <motion.div
                    key={task.id}
                    layout
                    initial={{ opacity: 0, scale: 0.95 }}
                    animate={{ opacity: 1, scale: 1 }}
                    exit={{ opacity: 0, scale: 0.95 }}
                    transition={{ duration: 0.22 }}
                    style={{
                      background: 'var(--bg-surface-a)',
                      border: '1px solid var(--hairline)',
                      borderRadius: 7,
                      padding: '9px 10px',
                      flexShrink: 0,
                    }}
                  >
                    <div
                      style={{
                        fontSize: 11,
                        color: 'var(--ink)',
                        lineHeight: 1.4,
                        marginBottom: 7,
                        fontWeight: 500,
                      }}
                    >
                      {t(`home.features.demos.flow.tasks.${task.id}`)}
                    </div>
                    <div
                      style={{ display: 'flex', alignItems: 'center', gap: 5, flexWrap: 'wrap' }}
                    >
                      <span
                        style={{
                          background: 'var(--hairline)',
                          border: '1px solid var(--hairline-strong)',
                          borderRadius: 4,
                          padding: '1px 6px',
                          fontSize: 9,
                          color: 'rgba(19,19,19,0.6)',
                        }}
                      >
                        {t(`home.features.demos.flow.tags.${task.tagKey}`)}
                      </span>
                      {task.priority === 'high' && (
                        <span
                          style={{
                            background: 'rgba(239,68,68,0.12)',
                            color: '#dc2626',
                            borderRadius: 4,
                            padding: '1px 6px',
                            fontSize: 9,
                            fontWeight: 600,
                          }}
                        >
                          {t('home.features.demos.flow.priorityHigh')}
                        </span>
                      )}
                      <div
                        style={{
                          marginInlineStart: 'auto',
                          width: 18,
                          height: 18,
                          borderRadius: '50%',
                          background: 'var(--hairline-strong)',
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                          fontSize: 8,
                          fontWeight: 700,
                          color: 'rgba(19,19,19,0.7)',
                        }}
                      >
                        {task.assignee}
                      </div>
                    </div>
                  </motion.div>
                ))}
              </AnimatePresence>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};
