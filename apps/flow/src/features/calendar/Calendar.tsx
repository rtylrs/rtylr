import React, { useMemo, useState } from 'react';
import { motion, AnimatePresence } from 'motion/react';
import { ChevronLeft, ChevronRight, X } from 'lucide-react';
import { useNavigate } from '@tanstack/react-router';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { useTasks, useCreateTask } from '../tasks/useTasks';
import { useLabels } from '../tasks/useLabels';
import { useFlowStore } from '../../shared/stores/flowStore';
import PriorityIcon from '../../shared/components/PriorityIcon';
import type { ApiTask } from '../../shared/services/flowApi';

const ACCENT_VAR = 'var(--accent)';

interface CreateModal {
  date: string;
}

const Calendar: React.FC = () => {
  const { t, i18n } = useTranslation();
  const { data: apiTasks = [] } = useTasks({});
  const tasks = apiTasks as ApiTask[];
  const navigate = useNavigate();
  const { projects } = useFlowStore();
  const { data: labels = [] } = useLabels();
  const createTask = useCreateTask();
  const [cursor, setCursor] = useState(() => new Date());
  const [createModal, setCreateModal] = useState<CreateModal | null>(null);
  const [newTitle, setNewTitle] = useState('');
  const [newProjectId, setNewProjectId] = useState('');
  const [newLabelId, setNewLabelId] = useState('');

  React.useEffect(() => {
    if (!newLabelId && labels.length > 0) setNewLabelId(String(labels[0].id));
  }, [labels, newLabelId]);

  const { days, monthLabel } = useMemo(() => {
    const first = new Date(cursor.getFullYear(), cursor.getMonth(), 1);
    const startDay = first.getDay();
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
      monthLabel: cursor.toLocaleString(i18n.language, { month: 'long', year: 'numeric' }),
    };
  }, [cursor, i18n.language]);

  const tasksByDate = useMemo(() => {
    const map: Record<string, ApiTask[]> = {};
    tasks.forEach((t) => {
      if (!t.dueDate) return;
      const key = t.dueDate.split('T')[0];
      if (!map[key]) map[key] = [];
      map[key].push(t);
    });
    return map;
  }, [tasks]);

  const today = new Date().toISOString().split('T')[0];

  const openCreate = (date: Date, inMonth: boolean) => {
    if (!inMonth) return;
    const iso = date.toISOString().split('T')[0];
    setCreateModal({ date: iso });
    setNewTitle('');
    setNewProjectId(projects[0]?.id ?? '');
    setNewLabelId(labels[0] ? String(labels[0].id) : '');
  };

  const submitCreate = () => {
    if (!newTitle.trim()) {
      toast.error(t('errors.generic'));
      return;
    }
    const projectId = newProjectId || projects[0]?.id;
    // The guard used to also require `activeWorkspaceId`, which the removed
    // workspace layer leaves permanently empty, so Create could never succeed
    // and always blamed the project, which the modal had pre-filled. It also
    // required a label, so a business with no labels could not create a task
    // here at all; a label is optional on the create contract.
    if (!projectId) {
      toast.error(t('toasts.task.selectProjectFirst'));
      return;
    }
    createTask.mutate(
      {
        projectId,
        title: newTitle.trim(),
        dueDate: createModal!.date,
        ...(newLabelId ? { labels: [newLabelId] } : {}),
      },
      {
        onSuccess: () => {
          toast.success(t('toasts.task.created'));
          setCreateModal(null);
        },
        onError: () => toast.error(t('toasts.task.createFailed')),
      }
    );
  };

  return (
    <div className="flex flex-col h-full">
      <HelmetContainer title={`${t('flow.calendar.heading')} · Flow`} />
      <div
        className="flex items-center justify-between px-6 py-4 shrink-0"
        style={{ borderBottom: '1px solid var(--flow-border)' }}
      >
        <div>
          <h1 className="text-[15px] font-semibold" style={{ color: 'var(--flow-text)' }}>
            {t('flow.calendar.heading')}
          </h1>
          <p className="text-[11px]" style={{ color: 'var(--flow-text-dim)' }}>
            {t('flow.calendar.subtitle')}
          </p>
        </div>
        <div className="flex items-center gap-2">
          <button
            type="button"
            onClick={() => setCursor(new Date(cursor.getFullYear(), cursor.getMonth() - 1, 1))}
            aria-label={t('flow.calendar.prevMonthLabel')}
            className="h-7 w-7 rounded-sm flex items-center justify-center hover:bg-[var(--flow-hover)]"
            style={{ color: 'var(--flow-text-muted)' }}
          >
            <ChevronLeft size={13} />
          </button>
          <span
            className="text-[12px] font-semibold w-[128px] text-center"
            style={{ color: 'var(--flow-text)' }}
          >
            {monthLabel}
          </span>
          <button
            type="button"
            onClick={() => setCursor(new Date(cursor.getFullYear(), cursor.getMonth() + 1, 1))}
            aria-label={t('flow.calendar.nextMonthLabel')}
            className="h-7 w-7 rounded-sm flex items-center justify-center hover:bg-[var(--flow-hover)]"
            style={{ color: 'var(--flow-text-muted)' }}
          >
            <ChevronRight size={13} />
          </button>
          <button
            type="button"
            onClick={() => setCursor(new Date())}
            className="text-[11px] px-2 py-1 rounded-sm hover:bg-[var(--flow-hover)] font-medium"
            style={{ color: 'var(--flow-text-muted)' }}
          >
            {t('flow.calendar.today')}
          </button>
        </div>
      </div>

      <div className="flex-1 overflow-auto p-4">
        <div className="grid grid-cols-7 gap-px" style={{ background: 'var(--flow-border)' }}>
          {days.slice(0, 7).map((c) => (
            <div
              key={c.date.toISOString()}
              className="py-1.5 text-center text-[10px] font-semibold uppercase"
              style={{ background: 'var(--flow-bg)', color: 'var(--flow-text-dim)' }}
            >
              {new Intl.DateTimeFormat(i18n.language, { weekday: 'short' }).format(c.date)}
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
                onClick={() => openCreate(c.date, c.inMonth)}
                className="min-h-[100px] p-1.5 cursor-pointer group"
                style={{
                  background: 'var(--flow-bg)',
                  opacity: c.inMonth ? 1 : 0.4,
                }}
              >
                <div className="flex items-center justify-between mb-1">
                  <span
                    className="text-[10px] font-semibold h-5 w-5 flex items-center justify-center rounded-full"
                    style={{
                      background: isToday ? ACCENT_VAR : 'transparent',
                      color: isToday ? '#0a0a0a' : 'var(--flow-text-muted)',
                    }}
                  >
                    {c.date.getDate()}
                  </span>
                  {c.inMonth && (
                    <button
                      type="button"
                      onClick={(e) => {
                        e.stopPropagation();
                        openCreate(c.date, c.inMonth);
                      }}
                      aria-label={t('flow.calendar.addTaskLabel', {
                        date: c.date.toLocaleDateString(i18n.language, {
                          month: 'long',
                          day: 'numeric',
                          year: 'numeric',
                        }),
                      })}
                      className="flex h-[44px] w-[44px] shrink-0 items-center justify-center rounded-sm text-[9px] opacity-0 group-hover:opacity-60 group-focus-within:opacity-60 focus:opacity-100 transition-opacity focus:outline-none focus-visible:ring-1 focus-visible:ring-[var(--accent)] md:h-8 md:w-8"
                      style={{ color: 'var(--flow-text-dim)' }}
                    >
                      {t('flow.calendar.addHint')}
                    </button>
                  )}
                </div>
                <div className="space-y-0.5">
                  {dayTasks.slice(0, 3).map((t) => (
                    <button
                      type="button"
                      key={t.id}
                      onClick={(e) => {
                        e.stopPropagation();
                        navigate({ to: `/task/${t.id}` as string });
                      }}
                      className="w-full text-left px-1.5 py-0.5 rounded-sm text-[10px] truncate flex items-center gap-1 hover:bg-[var(--flow-hover)]"
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
                    <button
                      type="button"
                      onClick={(e) => e.stopPropagation()}
                      className="text-[9px] px-1 hover:underline"
                      style={{ color: ACCENT_VAR }}
                    >
                      {t('flow.calendar.moreCount', { count: dayTasks.length - 3 })}
                    </button>
                  )}
                </div>
              </motion.div>
            );
          })}
        </div>
      </div>

      {/* Create task modal */}
      <AnimatePresence>
        {createModal && (
          <>
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              className="fixed inset-0 z-40"
              style={{ background: 'rgba(0,0,0,0.4)' }}
              onClick={() => setCreateModal(null)}
            />
            <motion.div
              initial={{ opacity: 0, scale: 0.96, y: 8 }}
              animate={{ opacity: 1, scale: 1, y: 0 }}
              exit={{ opacity: 0, scale: 0.96 }}
              transition={{ duration: 0.15 }}
              className="fixed z-50 top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 rounded-xl shadow-2xl p-5 w-[360px]"
              style={{ background: 'var(--flow-card)', border: '1px solid var(--flow-border)' }}
            >
              <div className="flex items-center justify-between mb-4">
                <h2 className="text-[13px] font-semibold" style={{ color: 'var(--flow-text)' }}>
                  {t('flow.calendar.newTaskFor')}{' '}
                  <span style={{ color: ACCENT_VAR }}>
                    {new Date(createModal.date + 'T00:00:00').toLocaleDateString(i18n.language, {
                      month: 'short',
                      day: 'numeric',
                    })}
                  </span>
                </h2>
                <button
                  type="button"
                  onClick={() => setCreateModal(null)}
                  aria-label={t('common.close')}
                  className="h-6 w-6 rounded-sm flex items-center justify-center hover:bg-[var(--flow-hover)]"
                  style={{ color: 'var(--flow-text-dim)' }}
                >
                  <X size={12} />
                </button>
              </div>

              <input
                autoFocus
                value={newTitle}
                onChange={(e) => setNewTitle(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && submitCreate()}
                placeholder={t('flow.calendar.titlePlaceholder')}
                className="w-full rounded-lg px-3 py-2 text-[13px] outline-none mb-3"
                style={{
                  background: 'var(--flow-input-bg)',
                  border: '1px solid var(--flow-border)',
                  color: 'var(--flow-text)',
                }}
              />

              {projects.length > 0 && (
                <select
                  value={newProjectId}
                  onChange={(e) => setNewProjectId(e.target.value)}
                  className="w-full rounded-lg px-3 py-2 text-[12px] outline-none mb-4"
                  style={{
                    background: 'var(--flow-input-bg)',
                    border: '1px solid var(--flow-border)',
                    color: 'var(--flow-text)',
                  }}
                >
                  {projects.map((p) => (
                    <option key={p.id} value={p.id}>
                      {p.name}
                    </option>
                  ))}
                </select>
              )}

              <select
                aria-label={t('common.settings.labels')}
                value={newLabelId}
                onChange={(e) => setNewLabelId(e.target.value)}
                disabled={labels.length === 0}
                className="w-full rounded-lg px-3 py-2 text-[12px] outline-none mb-4 disabled:opacity-50"
                style={{
                  background: 'var(--flow-input-bg)',
                  border: '1px solid var(--flow-border)',
                  color: 'var(--flow-text)',
                }}
              >
                {labels.length === 0 ? (
                  <option value="">{t('common.noData')}</option>
                ) : (
                  labels.map((label) => (
                    <option key={label.id} value={label.id}>
                      {label.name}
                    </option>
                  ))
                )}
              </select>

              <div className="flex justify-end gap-2">
                <button
                  type="button"
                  onClick={() => setCreateModal(null)}
                  className="px-3 py-1.5 rounded-sm text-[12px]"
                  style={{ color: 'var(--flow-text-muted)' }}
                >
                  {t('flow.calendar.cancel')}
                </button>
                <button
                  type="button"
                  onClick={submitCreate}
                  disabled={!newTitle.trim() || createTask.isPending}
                  className="px-3 py-1.5 rounded-sm text-[12px] font-semibold disabled:opacity-40"
                  style={{ background: ACCENT_VAR, color: '#0a0a0a' }}
                >
                  {createTask.isPending ? t('flow.calendar.creating') : t('flow.calendar.create')}
                </button>
              </div>
            </motion.div>
          </>
        )}
      </AnimatePresence>
    </div>
  );
};

export default Calendar;
