import React from 'react';
import { Monitor, Moon, Sun } from 'lucide-react';
import { resolvePalette } from '@voxire/utils';
import type { ThemeMode } from '@voxire/utils';
import { useTranslation } from 'react-i18next';
import { useAppearanceStore } from '../stores/appearanceStore';
import { cn } from '../lib/utils';

/**
 * `compact` is the account menu's quick toggle: three small tiles, icon over
 * label. `cards` is the settings panel's version: the same three options with a
 * miniature render of the ground each one produces.
 */
export type ModeCardsVariant = 'compact' | 'cards';

export interface ModeCardsProps {
  variant?: ModeCardsVariant;
  /**
   * Called with a mode while it is hovered or focused, and with null when it is
   * not. Lets a preview pane audition an option without a store write, which is
   * what keeps the applied state the single source of truth. Omit it and the
   * control does nothing extra.
   */
  onPreview?: (mode: ThemeMode | null) => void;
  className?: string;
}

interface Option {
  value: ThemeMode;
  icon: typeof Sun;
  label: string;
}

/**
 * Flow's light and dark mode switcher, in one place.
 *
 * There were two of these: an inline set of cards in the settings panel and an
 * inline set of tiles in the account menu, each with its own copy of the option
 * list and its own idea of what "selected" looks like. The spec keeps both entry
 * points, because the quick toggle is a genuinely useful shortcut, and requires
 * that they render one component reading one store.
 *
 * The preview grounds come from the palette data in `@voxire/utils` rather than
 * from literals here, so a light preview shows the light preset the user
 * actually has selected instead of a stand in for it, and so this file holds no
 * colour of its own.
 *
 * Design: docs/superpowers/specs/2026-08-22-flow-appearance-design.md, section 5.
 */
const ModeCards = ({
  variant = 'cards',
  onPreview,
  className,
}: ModeCardsProps): React.JSX.Element => {
  const { t } = useTranslation();
  const mode = useAppearanceStore((s) => s.mode);
  const setMode = useAppearanceStore((s) => s.setMode);
  const palette = useAppearanceStore((s) => s.palette);

  const options: Option[] = [
    { value: 'light', icon: Sun, label: t('flow.shell.light') },
    { value: 'dark', icon: Moon, label: t('flow.shell.dark') },
    { value: 'system', icon: Monitor, label: t('flow.shell.system') },
  ];

  // Resolved against the stored palette id so each preview is the ground this
  // user would actually get. A stored id belonging to the other family falls
  // back to that family's first preset, which resolvePalette handles.
  const lightGround = resolvePalette(palette, 'light').bg;
  const darkGround = resolvePalette(palette, 'dark').bg;
  const ground: Record<ThemeMode, string> = {
    light: lightGround,
    dark: darkGround,
    system: `linear-gradient(135deg, ${lightGround} 50%, ${darkGround} 50%)`,
  };

  const isCards = variant === 'cards';

  return (
    <div
      role="group"
      aria-label={t('flow.shell.appearance')}
      className={cn('grid grid-cols-3', isCards ? 'gap-3' : 'gap-1', className)}
    >
      {options.map(({ value, icon: Icon, label }) => {
        const selected = mode === value;
        return (
          <button
            type="button"
            key={value}
            aria-pressed={selected}
            onClick={() => setMode(value)}
            onPointerEnter={onPreview ? () => onPreview(value) : undefined}
            onPointerLeave={onPreview ? () => onPreview(null) : undefined}
            onFocus={onPreview ? () => onPreview(value) : undefined}
            onBlur={onPreview ? () => onPreview(null) : undefined}
            className={cn(
              // The 44px floor is a literal so the density control cannot take
              // an interactive target below it at the compact step.
              'flow-focus flex min-h-[44px] flex-col items-center font-medium transition-colors',
              isCards
                ? 'gap-2 rounded-xl border-2 p-3 text-[12px]'
                : 'gap-1 rounded-md py-2 text-[10px]'
            )}
            style={
              isCards
                ? {
                    borderColor: selected ? 'var(--accent)' : 'var(--flow-border)',
                    background: 'var(--flow-card)',
                    color: selected ? 'var(--accent)' : 'var(--flow-text-muted)',
                  }
                : {
                    background: selected ? 'var(--accent-muted)' : 'var(--flow-hover)',
                    color: selected ? 'var(--accent)' : 'var(--flow-text-muted)',
                    border: selected
                      ? '1px solid var(--accent-border)'
                      : '1px solid transparent',
                  }
            }
          >
            {isCards ? (
              <span
                aria-hidden="true"
                className="flex w-full items-center justify-center rounded-lg py-3"
                style={{ background: ground[value] }}
              >
                <Icon size={16} style={{ color: 'var(--accent)' }} />
              </span>
            ) : (
              <Icon size={12} aria-hidden="true" />
            )}
            {label}
          </button>
        );
      })}
    </div>
  );
};

export default ModeCards;
