import React, { useMemo, useState } from 'react';
import { motion } from 'motion/react';
import { ChevronLeft, ChevronRight } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { useFlowStore, Task } from '../../shared/stores/flowStore';
import { useTasks } from '../tasks/useTasks';
import { isFlowLocal } from '../../shared/lib/flowLocalMode';
import PriorityIcon from '../../shared/components/PriorityIcon';

const Calendar: React.FC = () => {
  const storeTasks = useFlowStore((s) => s.tasks);
  const { data: apiTasks } = useTasks({});
  const tasks = (isFlowLocal() ? storeTasks : (apiTasks ?? storeTasks)) as Task[];
  const navigate = useNavigate();
  const [cursor, setCursor] = useState(() => new Date());

  const { days, monthLabel } = useMemo(() => {
    const first = new Date(cursor.getFullYear(), cursor.getMonth(), 1);
    const startDay = first.getDay(); // Sun=0
    const daysInMonth = new Date(cursor.getFullYear(), cursor.getMonth() + 1, 0).getDate();
    const cells: { date: Date; inMonth: boolean }[] = [];
    for (let i = startDay - 1; i >= 0; i--) {
      cells.push({ date: new Date(cursor.getFullYear(), cursor.getMonth(), -i), inMonth: false });
    }
    for (let d = 1; d <= daysInMonth; d++) {
      cells.push({ date: new Date(cursor.getFullYear(), cursor.getMonth(), d), inMonth: true });
    }
    while (cells.length % 7 !== 0) {
      const last = cells[cells.length - 1].date;
      cells.push({
        date: new Date(last.getFullYear(), last.getMonth(), last.getDate() + 1),
        inMonth: false,
      });
    }
    return {
      days: cells,
      monthLabel: cursor.toLocaleString('default', { month: 'long', year: 'numeric' }),
    };
  }, [cursor]);

  const tasksByDate = useMemo(() => {
    const map: Record<string, Task[]> = {};
    tasks.forEach((t) => {
      if (!t.dueDate) return;
      if (!map[t.dueDate]) map[t.dueDate] = [];
      map[t.dueDate].push(t);
    });
    return map;
  }, [tasks]);

  const today = new Date().toISOString().split('T')[0];

  return (
    <div className="flex flex-col h-full">
      <div
        className="flex items-center justify-between px-6 py-4 shrink-0"
        style={{ borderBottom: '1px solid var(--flow-border)' }}
      >
        <div>
          <h1 className="text-[15px] font-semibold" style={{ color: 'var(--flow-text)' }}>
            Calendar
          </h1>
          <p className="text-[11px]" style={{ color: 'var(--flow-text-dim)' }}>
            Tasks organized by due date
          </p>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={() => setCursor(new Date(cursor.getFullYear(), cursor.getMonth() - 1, 1))}
            className="h-7 w-7 rounded flex items-center justify-center hover:bg-[var(--flow-hover)]"
            style={{ color: 'var(--flow-text-muted)' }}
          >
            <ChevronLeft size={13} />
          </button>
          <span
            className="text-[12px] font-semibold w-32 text-center"
            style={{ color: 'var(--flow-text)' }}
          >
            {monthLabel}
          </span>
          <button
            onClick={() => setCursor(new Date(cursor.getFullYear(), cursor.getMonth() + 1, 1))}
            className="h-7 w-7 rounded flex items-center justify-center hover:bg-[var(--flow-hover)]"
            style={{ color: 'var(--flow-text-muted)' }}
          >
            <ChevronRight size={13} />
          </button>
          <button
            onClick={() => setCursor(new Date())}
            className="text-[11px] px-2 py-1 rounded hover:bg-[var(--flow-hover)] font-medium"
            style={{ color: 'var(--flow-text-muted)' }}
          >
            Today
          </button>
        </div>
      </div>

      <div className="flex-1 overflow-auto p-4">
        <div className="grid grid-cols-7 gap-px" style={{ background: 'var(--flow-border)' }}>
          {['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'].map((d) => (
            <div
              key={d}
              className="py-1.5 text-center text-[10px] font-semibold uppercase"
              style={{ background: 'var(--flow-bg)', color: 'var(--flow-text-dim)' }}
            >
              {d}
            </div>
          ))}
          {days.map((c, i) => {
            const iso = c.date.toISOString().split('T')[0];
            const dayTasks = tasksByDate[iso] || [];
            const isToday = iso === today;
            return (
              <motion.div
                key={i}
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                transition={{ duration: 0.1, delay: i * 0.002 }}
                className="min-h-[100px] p-1.5"
                style={{
                  background: 'var(--flow-bg)',
                  opacity: c.inMonth ? 1 : 0.4,
                }}
              >
                <div className="flex items-center justify-between mb-1">
                  <span
                    className="text-[10px] font-semibold h-5 w-5 flex items-center justify-center rounded-full"
                    style={{
                      background: isToday ? '#64C5B9' : 'transparent',
                      color: isToday ? '#0a0a0a' : 'var(--flow-text-muted)',
                    }}
                  >
                    {c.date.getDate()}
                  </span>
                </div>
                <div className="space-y-0.5">
                  {dayTasks.slice(0, 3).map((t) => (
                    <button
                      key={t.id}
                      onClick={() => navigate(`/task/${t.id}`)}
                      className="w-full text-left px-1.5 py-0.5 rounded text-[10px] truncate flex items-center gap-1 hover:bg-[var(--flow-hover)]"
                      style={{
                        background: 'var(--flow-card)',
                        color: 'var(--flow-text)',
                      }}
                    >
                      <PriorityIcon priority={t.priority} size={9} />
                      <span className="truncate">{t.title}</span>
                    </button>
                  ))}
                  {dayTasks.length > 3 && (
                    <span className="text-[9px] px-1" style={{ color: 'var(--flow-text-dim)' }}>
                      +{dayTasks.length - 3} more
                    </span>
                  )}
                </div>
              </motion.div>
            );
          })}
        </div>
      </div>
    </div>
  );
};

export default Calendar;
