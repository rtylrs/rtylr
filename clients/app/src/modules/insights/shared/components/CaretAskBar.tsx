import React from 'react';
import { useCaretStore } from '@voxire/caret';
import { Sparkles } from 'lucide-react';
import { useTranslation } from 'react-i18next';

/** The embedded entry point into Caret AI, docked at the top of Overview. */
const CaretAskBar: React.FC = (): React.JSX.Element => {
  const { t } = useTranslation();
  const toggle = useCaretStore((s) => s.toggle);

  return (
    <button
      type="button"
      onClick={toggle}
      className="flex items-center gap-3 w-full rounded-2xl px-5 py-3.5 text-start transition-colors duration-150"
      style={{
        background: 'var(--hub-card)',
        border: '1px solid var(--hub-border)',
        color: 'var(--hub-text-muted)',
      }}
      onMouseEnter={(e) => {
        (e.currentTarget as HTMLElement).style.borderColor = 'var(--caret-brand)';
      }}
      onMouseLeave={(e) => {
        (e.currentTarget as HTMLElement).style.borderColor = 'var(--hub-border)';
      }}
    >
      <Sparkles size={16} className="shrink-0" style={{ color: 'var(--caret-brand)' }} />
      <span className="flex-1 text-sm truncate">
        {t('insights.dashboard.caretAskBar.placeholder')}
      </span>
      <kbd
        className="hidden sm:inline text-[10px] font-semibold px-1.5 py-1 rounded shrink-0"
        style={{ background: 'var(--hub-hover)', color: 'var(--hub-text-dim)' }}
      >
        ⌘J
      </kbd>
    </button>
  );
};

export default CaretAskBar;
