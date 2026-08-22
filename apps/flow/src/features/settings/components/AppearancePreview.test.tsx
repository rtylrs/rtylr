import React from 'react';
import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi, afterEach } from 'vitest';
import { APPEARANCE_DEFAULTS, appearanceTokens } from '@voxire/utils';
import AppearancePreview from './AppearancePreview';

vi.mock('react-i18next', () => ({ useTranslation: () => ({ t: (key: string) => key }) }));

/**
 * The pane's own contract, independent of the panel that drives it: an override
 * wins over the applied preferences, and the tokens land on its wrapper rather
 * than on documentElement.
 *
 * That second assertion is the one worth keeping. Writing to the root from here
 * would fight the provider, and an inline custom property on the root outranks
 * every stylesheet rule, so the damage would not be undoable from CSS.
 */

const WORKSPACE_ACCENT = '#8b5cf6';

function renderPane(override?: Partial<typeof APPEARANCE_DEFAULTS>): HTMLElement {
  render(
    <AppearancePreview
      prefs={APPEARANCE_DEFAULTS}
      previewOverride={override}
      workspaceAccent={WORKSPACE_ACCENT}
      prefersDark={false}
    />
  );
  return screen.getByTestId('appearance-preview');
}

afterEach(() => {
  document.documentElement.removeAttribute('style');
});

describe('AppearancePreview', () => {
  it('shows the applied preferences when no override is given', () => {
    const pane = renderPane();
    const expected = appearanceTokens(APPEARANCE_DEFAULTS, {
      resolvedMode: 'dark',
      workspaceAccent: WORKSPACE_ACCENT,
    });
    expect(pane.style.getPropertyValue('--flow-bg')).toBe(expected['--flow-bg']);
    expect(pane.style.getPropertyValue('--accent')).toBe(WORKSPACE_ACCENT);
  });

  it('lets an override win over the applied preferences', () => {
    const pane = renderPane({ palette: 'ink', accent: '#a3d84b', corners: 'round' });
    expect(pane.style.getPropertyValue('--flow-bg')).toBe('#080b12');
    expect(pane.style.getPropertyValue('--accent')).toBe('#a3d84b');
    expect(pane.style.getPropertyValue('--flow-r')).toBe('14px');
  });

  it('resolves a system override against the supplied system preference', () => {
    // prefersDark is false above, so `system` has to land on the light family.
    const pane = renderPane({ mode: 'system' });
    expect(pane.style.getPropertyValue('--flow-bg')).toBe('#f4f4f5');
  });

  it('writes its tokens to its own subtree and never to documentElement', () => {
    renderPane({ palette: 'ink' });
    expect(document.documentElement.style.getPropertyValue('--flow-bg')).toBe('');
    expect(document.documentElement.style.getPropertyValue('--accent')).toBe('');
    expect(document.documentElement.style.getPropertyValue('--flow-r')).toBe('');
  });

  it('is hidden from assistive technology, since every option it reflects is already named', () => {
    const pane = renderPane();
    expect(pane).toHaveAttribute('aria-hidden', 'true');
  });
});
