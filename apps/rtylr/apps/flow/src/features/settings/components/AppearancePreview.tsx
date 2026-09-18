import React from 'react';
import { useTranslation } from 'react-i18next';
import { appearanceTokens, resolveMode } from '@voxire/utils';
import type { AppearancePrefs } from '@voxire/utils';
import { tokenStyle } from './AppearanceMiniature';

/**
 * The audition pane.
 *
 * Shows the applied preferences by default, and whatever option is currently
 * hovered or focused instead. That is the whole model: there is no draft tier and
 * no Apply button, because the panel commits on click and a second copy of the
 * preferences would eventually show a palette the user is not actually getting.
 * `previewOverride` is therefore transient by construction - the panel sets it
 * from pointer and focus events and clears it on leave and blur, and nothing here
 * ever writes to the store.
 *
 * The tokens come from the same `appearanceTokens()` the provider calls, and land
 * on this component's own wrapper rather than on documentElement. That
 * distinction is the reason a second token writer is allowed to exist at all: an
 * inline custom property on the root outranks every stylesheet rule, so a second
 * writer there would reintroduce exactly the conflict the single-writer
 * architecture removes. Scoped to a container, the same values are just a
 * subtree's theme.
 *
 * `--flow-u` and `--flow-r` are consumed here directly, so density and corners
 * demonstrate themselves on more than one shape. `--flow-fs` deliberately is not:
 * the text-size control is disabled, so wiring the pane to it would make the one
 * control that cannot work look like the one control that does.
 *
 * Touch gets no audition affordance. There is no hover to have, and tapping is
 * safe because every change is instant and reversible.
 *
 * Design: docs/superpowers/specs/2026-08-22-flow-appearance-design.md (rtylr),
 * section 5.
 */

interface AppearancePreviewProps {
  /** The applied preferences. */
  prefs: AppearancePrefs;
  /** The hovered or focused option, or undefined when nothing is auditioned. */
  previewOverride?: Partial<AppearancePrefs>;
  /** Accent to derive from when the effective `accent` is null. */
  workspaceAccent: string;
  /** What `system` resolves to right now. */
  prefersDark: boolean;
}

const AppearancePreview = ({
  prefs,
  previewOverride,
  workspaceAccent,
  prefersDark,
}: AppearancePreviewProps): React.JSX.Element => {
  const { t } = useTranslation();

  const effective: AppearancePrefs = { ...prefs, ...previewOverride };
  const tokens = appearanceTokens(effective, {
    resolvedMode: resolveMode(effective.mode, prefersDark),
    workspaceAccent,
  });

  const railItems = [
    { key: 'inbox', label: t('nav.inbox'), active: true },
    { key: 'tasks', label: t('nav.tasks'), active: false },
    { key: 'projects', label: t('nav.projects'), active: false },
  ];

  return (
    // Decorative relative to the controls: every option it reflects is already
    // named and announced by the control that owns it, so a screen reader gains
    // nothing from a second, wordless copy of the same state.
    <div
      aria-hidden="true"
      data-testid="appearance-preview"
      className="flex overflow-hidden"
      style={tokenStyle(tokens, {
        background: 'var(--flow-bg)',
        border: '1px solid var(--flow-border)',
        borderRadius: 'var(--flow-r)',
        // --flow-u is a unitless multiplier, so it is not a length and cannot be
        // used as one: `gap: var(--flow-u)` is `gap: 1`, which is invalid and
        // gets dropped, collapsing the pane to zero padding. Turn it into a
        // length once, here, exactly the way globals.css derives --spacing.
        ['--pane-u' as string]: 'calc(0.25rem * var(--flow-u, 1))',
      })}
    >
      <div
        className="flex w-[92px] shrink-0 flex-col"
        style={{
          background: 'var(--flow-surface)',
          borderInlineEnd: '1px solid var(--flow-border)',
          gap: 'var(--pane-u)',
          padding: 'calc(var(--pane-u) * 2)',
        }}
      >
        {railItems.map((item) => (
          <span
            key={item.key}
            className="flex items-center truncate text-meta font-medium"
            style={{
              background: item.active ? 'var(--accent-muted)' : 'transparent',
              borderRadius: 'var(--flow-r)',
              color: item.active ? 'var(--accent-on-muted)' : 'var(--flow-text-muted)',
              gap: 'var(--pane-u)',
              padding: 'var(--pane-u) calc(var(--pane-u) * 1.5)',
            }}
          >
            <span
              className="block h-1.5 w-1.5 shrink-0 rounded-full"
              style={{ background: item.active ? 'var(--accent)' : 'var(--flow-text-dim)' }}
            />
            {item.label}
          </span>
        ))}
      </div>

      <div
        className="flex min-w-0 flex-1 flex-col"
        style={{ gap: 'calc(var(--pane-u) * 2)', padding: 'calc(var(--pane-u) * 2.5)' }}
      >
        <div
          className="flex min-w-0 items-center"
          style={{
            background: 'var(--flow-card)',
            border: '1px solid var(--flow-border)',
            borderRadius: 'var(--flow-r)',
            gap: 'calc(var(--pane-u) * 1.5)',
            padding: 'calc(var(--pane-u) * 2)',
          }}
        >
          <span
            className="block h-2 w-2 shrink-0 rounded-full"
            style={{ background: 'var(--accent)' }}
          />
          <span className="min-w-0 flex-1 truncate text-meta" style={{ color: 'var(--flow-text)' }}>
            {t('flow.settings.appearance.preview.taskTitle')}
          </span>
          <span
            className="shrink-0 text-meta font-semibold"
            style={{
              background: 'var(--accent-muted)',
              border: '1px solid var(--accent-border)',
              borderRadius: 'var(--flow-r)',
              color: 'var(--accent-on-muted)',
              padding: 'calc(var(--pane-u) * 0.5) var(--pane-u)',
            }}
          >
            {t('flow.settings.appearance.preview.chip')}
          </span>
          <span
            className="block h-5 w-5 shrink-0 rounded-full"
            style={{ background: 'var(--accent)', border: '1px solid var(--accent-border)' }}
          />
        </div>

        <div
          className="h-1.5 w-full overflow-hidden rounded-full"
          style={{ background: 'var(--flow-hover)', border: '1px solid var(--flow-border)' }}
        >
          <span className="block h-full w-[62%]" style={{ background: 'var(--accent)' }} />
        </div>

        <span
          className="inline-flex w-fit items-center text-meta font-semibold"
          style={{
            background: 'var(--accent)',
            borderRadius: 'var(--flow-r)',
            color: 'var(--flow-on-accent)',
            padding: 'var(--pane-u) calc(var(--pane-u) * 2.5)',
          }}
        >
          {t('flow.settings.appearance.preview.action')}
        </span>
      </div>
    </div>
  );
};

export default AppearancePreview;
