import React from 'react';
import { Monitor, Moon, Sun } from 'lucide-react';
import type { ThemeMode } from '@voxire/utils';
import { useTranslation } from 'react-i18next';
import { useAppearanceStore } from '../stores/appearanceStore';
import { cn } from '../lib/utils';

export interface ModeToggleProps {
  className?: string;
}

interface Option {
  value: ThemeMode;
  icon: typeof Sun;
  label: string;
}

/**
 * The account menu's quick mode toggle: three small tiles, icon over label.
 *
 * This started life as a `ModeCards` component with a `variant` prop, on the
 * theory that the settings panel and the account menu were two renderings of one
 * control and should be deduplicated into one. They are not, and the settings
 * panel is the reason.
 *
 * The panel's version renders a real miniature of each mode built from the full
 * derived token set, and it draws two of them side by side for `system`, because
 * what `system` shows is exactly "whichever of these the OS is asking for". This
 * one draws an icon on a flat ground in a 24px tile. Sharing a body between those
 * two would be a component with two implementations and a boolean to pick, which
 * is not less duplication than two components, only less honest about it.
 *
 * What genuinely wanted sharing was the option list and the store wiring, and
 * both are shared already: this file and the panel read the same
 * `useAppearanceStore`, so there is one source of truth for the mode. That was
 * the actual defect, because previously each had its own store.
 *
 * Design: docs/superpowers/specs/2026-08-22-flow-appearance-design.md, section 5.
 */
const ModeToggle = ({ className }: ModeToggleProps): React.JSX.Element => {
  const { t } = useTranslation();
  const mode = useAppearanceStore((s) => s.mode);
  const setMode = useAppearanceStore((s) => s.setMode);

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
              'flow-focus flex min-h-[44px] flex-col items-center gap-1 rounded-md py-2 text-meta font-medium transition-colors'
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
