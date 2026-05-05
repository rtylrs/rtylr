import React, { useState, useEffect, useRef } from 'react';
import { motion, AnimatePresence } from 'motion/react';

interface Task {
  id: string;
  title: string;
  tag: string;
  assignee: string;
  priority?: 'high' | 'med';
}

interface Column {
  id: string;
  title: string;
  tint: string;
  tasks: Task[];
}

const INITIAL: Column[] = [
  {
    id: 'backlog',
    title: 'Backlog',
    tint: '#8a8a8a',
    tasks: [
      { id: 't1', title: 'Seasonal menu rollout, Q3', tag: 'Ops', assignee: 'MC' },
      { id: 't2', title: 'New supplier: Rossini', tag: 'ERP', assignee: 'JK' },
    ],
  },
  {
    id: 'progress',
    title: 'In progress',
    tint: '#d6b88e',
    tasks: [
      {
        id: 't3',
        title: 'Friday covers, staffing plan',
        tag: 'HR',
        assignee: 'SP',
        priority: 'high',
      },
      { id: 't4', title: 'Redo brunch prep checklist', tag: 'Kitchen', assignee: 'AR' },
      { id: 't5', title: "Email blast: Mother's Day", tag: 'CRM', assignee: 'MC' },
    ],
  },
  {
    id: 'review',
    title: 'Review',
    tint: '#8ca39d',
    tasks: [{ id: 't6', title: 'Espresso machine SLA', tag: 'Vendor', assignee: 'JK' }],
  },
  {
    id: 'done',
    title: 'Done',
    tint: '#6bcf7f',
    tasks: [
      { id: 't7', title: 'Q1 sales-tax filing', tag: 'Finance', assignee: 'SP' },
      { id: 't8', title: 'POS firmware v4.2', tag: 'Ops', assignee: 'AR' },
    ],
  },
];

const COL_ORDER = ['backlog', 'progress', 'review', 'done'];

export const FlowDemo = ({ isActive }: { isActive: boolean }): React.JSX.Element => {
  const [cols, setCols] = useState<Column[]>(INITIAL);
  const tickRef = useRef(0);

  const moveTask = (): void => {
    setCols((prev) => {
      const moveable = prev
        .filter((c) => c.id !== 'done' && c.tasks.length > 0)
        .flatMap((c) => c.tasks.map((t) => ({ task: t, colId: c.id })));
      if (!moveable.length) return prev;
      const pick = moveable[tickRef.current % moveable.length];
      const fromIdx = COL_ORDER.indexOf(pick.colId);
      const toId = COL_ORDER[Math.min(fromIdx + 1, COL_ORDER.length - 1)];
      if (toId === pick.colId) return prev;
      return prev.map((c) => {
        if (c.id === pick.colId)
          return { ...c, tasks: c.tasks.filter((t) => t.id !== pick.task.id) };
        if (c.id === toId) return { ...c, tasks: [pick.task, ...c.tasks] };
        return c;
      });
    });
    tickRef.current += 1;
  };

  useEffect(() => {
    if (!isActive) return;
    const t = setInterval(moveTask, 2800);
    return () => clearInterval(t);
  }, [isActive]);

  return (
    <div
      style={{
        position: 'absolute',
        inset: 0,
        background: '#0c0c0f',
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
          borderBottom: '1px solid rgba(255,255,255,0.07)',
          display: 'flex',
          alignItems: 'center',
          gap: 8,
          flexShrink: 0,
        }}
      >
        <span style={{ color: 'rgba(255,255,255,0.45)', fontSize: 12 }}>Workspaces</span>
        <span style={{ color: 'rgba(255,255,255,0.25)', fontSize: 12 }}>/</span>
        <span style={{ color: 'rgba(255,255,255,0.45)', fontSize: 12 }}>Operations</span>
        <span style={{ color: 'rgba(255,255,255,0.25)', fontSize: 12 }}>/</span>
        <span style={{ color: '#e8e6e1', fontSize: 12, fontWeight: 500 }}>Weekly sprint</span>
        <div style={{ marginLeft: 'auto', display: 'flex', gap: 6 }}>
          {['Board', 'List', 'Timeline'].map((v, i) => (
            <div
              key={v}
              style={{
                padding: '4px 10px',
                borderRadius: 6,
                fontSize: 11,
                fontWeight: 500,
                background: i === 0 ? 'rgba(255,255,255,0.08)' : 'transparent',
                color: i === 0 ? '#e8e6e1' : 'rgba(255,255,255,0.35)',
                border: i === 0 ? '1px solid rgba(255,255,255,0.12)' : '1px solid transparent',
              }}
            >
              {v}
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
              background: 'rgba(255,255,255,0.02)',
              border: '1px solid rgba(255,255,255,0.06)',
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
              <span style={{ fontSize: 11, fontWeight: 600, color: '#e8e6e1' }}>{col.title}</span>
              <span
                style={{
                  marginLeft: 'auto',
                  fontSize: 10,
                  color: 'rgba(255,255,255,0.35)',
                  background: 'rgba(255,255,255,0.06)',
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
                      background: '#161616',
                      border: '1px solid rgba(255,255,255,0.07)',
                      borderRadius: 7,
                      padding: '9px 10px',
                      flexShrink: 0,
                    }}
                  >
                    <div
                      style={{
                        fontSize: 11,
                        color: '#e8e6e1',
                        lineHeight: 1.4,
                        marginBottom: 7,
                        fontWeight: 500,
                      }}
                    >
                      {task.title}
                    </div>
                    <div
                      style={{ display: 'flex', alignItems: 'center', gap: 5, flexWrap: 'wrap' }}
                    >
                      <span
                        style={{
                          background: 'rgba(255,255,255,0.06)',
                          border: '1px solid rgba(255,255,255,0.09)',
                          borderRadius: 4,
                          padding: '1px 6px',
                          fontSize: 9,
                          color: 'rgba(255,255,255,0.5)',
                        }}
                      >
                        {task.tag}
                      </span>
                      {task.priority === 'high' && (
                        <span
                          style={{
                            background: 'rgba(239,68,68,0.12)',
                            color: '#f87171',
                            borderRadius: 4,
                            padding: '1px 6px',
                            fontSize: 9,
                            fontWeight: 600,
                          }}
                        >
                          High
                        </span>
                      )}
                      <div
                        style={{
                          marginLeft: 'auto',
                          width: 18,
                          height: 18,
                          borderRadius: '50%',
                          background: 'rgba(255,255,255,0.1)',
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                          fontSize: 8,
                          fontWeight: 700,
                          color: 'rgba(255,255,255,0.6)',
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
