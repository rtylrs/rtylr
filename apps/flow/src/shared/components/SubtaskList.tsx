import React, { useState } from 'react';
import { Plus, Check } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { Subtask } from '../stores/flowStore';

interface SubtaskListProps {
  subtasks: Subtask[];
  onToggle: (id: string) => void;
  onAdd: (title: string) => void;
  disabled?: boolean;
}

export const SubtaskList: React.FC<SubtaskListProps> = ({
  subtasks,
  onToggle,
  onAdd,
  disabled,
}) => {
  const { t } = useTranslation();
  const [newTitle, setNewTitle] = useState('');
  const completed = subtasks.filter((s) => s.completed).length;
  const pct = subtasks.length > 0 ? Math.round((completed / subtasks.length) * 100) : 0;

  const submit = () => {
    if (disabled || !newTitle.trim()) return;
    onAdd(newTitle.trim());
    setNewTitle('');
  };

  return (
    <div>
      {subtasks.length > 0 && (
        <>
          <div className="flex items-center gap-2 mb-2">
            <span className="text-meta" style={{ color: 'var(--flow-text-muted)' }}>
              {t('flow.taskDetail.subtaskCount', { done: completed, total: subtasks.length })}
            </span>
            <div
              className="flex-1 h-1 rounded-full overflow-hidden"
              style={{ background: 'var(--flow-hover)' }}
            >
              <div
                className="h-full transition-all"
                style={{ width: `${pct}%`, background: 'var(--accent)' }}
              />
            </div>
            <span className="text-meta" style={{ color: 'var(--flow-text-dim)' }}>
              {pct}%
            </span>
          </div>
          <div className="space-y-1">
            {subtasks.map((s) => (
              <button
                type="button"
                key={s.id}
                onClick={() => onToggle(s.id)}
                disabled={disabled}
                className="w-full flex items-center gap-2 px-2 py-1 rounded-sm hover:bg-[var(--flow-hover)] transition-colors text-left disabled:opacity-60 disabled:cursor-not-allowed"
              >
                <span
                  className="h-4 w-4 rounded-sm flex items-center justify-center shrink-0"
                  style={{
                    border: '1px solid var(--flow-border)',
                    background: s.completed ? 'var(--accent)' : 'transparent',
                  }}
                >
                  {s.completed && <Check size={10} color="#0a0a0a" strokeWidth={3} />}
                </span>
                <span
                  className="text-meta flex-1"
                  style={{
                    color: s.completed ? 'var(--flow-text-dim)' : 'var(--flow-text)',
                    textDecoration: s.completed ? 'line-through' : undefined,
                  }}
                >
                  {s.title}
                </span>
              </button>
            ))}
          </div>
        </>
      )}

      <div className="flex items-center gap-2 mt-2">
        <Plus size={12} style={{ color: 'var(--flow-text-dim)' }} />
        <input
          value={newTitle}
          onChange={(e) => setNewTitle(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && submit()}
          placeholder={t('flow.taskDetail.subtaskAddPlaceholder')}
          disabled={disabled}
          className="flex-1 bg-transparent outline-none text-meta disabled:opacity-60"
          style={{ color: 'var(--flow-text)' }}
        />
      </div>
    </div>
  );
};

export default SubtaskList;
