import React from 'react';
import { Monitor, Moon, Sun } from 'lucide-react';
import type { ThemeMode } from '@voxire/utils';
import { useTranslation } from 'react-i18next';
import { useThemeStore } from '../../../../shared/stores/themeStore';
import { cn } from '../lib/utils';

export interface ModeToggleProps {
  className?: string;
}

interface Option {
  value: ThemeMode;
  icon: typeof Sun;
  label: string;
}

const ModeToggle = ({ className }: ModeToggleProps): React.JSX.Element => {
  const { t } = useTranslation();
  const mode = useThemeStore((s) => s.theme);
  const setMode = useThemeStore((s) => s.setTheme);

  const options: Option[] = [
    { value: 'light', icon: Sun, label: t('flow.shell.light') },
    { value: 'dark', icon: Moon, label: t('flow.shell.dark') },
    { value: 'system', icon: Monitor, label: t('flow.shell.system') },
  ];

  return (
    <div
      role="group"
      aria-label={t('flow.shell.appearance')}
      className={cn('grid grid-cols-3 gap-1', className)}
    >
      {options.map(({ value, icon: Icon, label }) => {
        const selected = mode === value;
        return (
          <button
            type="button"
            key={value}
            aria-pressed={selected}
            onClick={() => setMode(value)}
            className={cn(
              // The 44px floor is a literal so the density control cannot take
              // an interactive target below it at the compact multiplier.
              'flow-focus flex min-h-[44px] flex-col items-center gap-1 rounded-md py-2 text-flow-meta font-medium transition-colors'
            )}
            style={{
              background: selected ? 'var(--accent-muted)' : 'var(--flow-hover)',
              color: selected ? 'var(--accent-on-muted)' : 'var(--flow-text-muted)',
              border: selected ? '1px solid var(--accent-border)' : '1px solid transparent',
            }}
          >
            <Icon size={12} aria-hidden="true" />
            {label}
          </button>
        );
      })}
    </div>
  );
};

export default ModeToggle;
