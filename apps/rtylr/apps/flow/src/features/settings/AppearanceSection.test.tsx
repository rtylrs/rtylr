import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { PALETTES, appearanceTokens } from '@voxire/utils';
import type { AppearancePrefs } from '@voxire/utils';

/**
 * Behaviour of the Appearance panel, not its markup.
 *
 * The load-bearing test here is the audition pair: hovering an option must move
 * the preview pane and must not write to the store. Those two assertions are the
 * whole "audition on hover, commit on click" model, and either one passing alone
 * proves nothing. Forcing the hover handler to commit is expected to fail both
 * the store assertion and the mutation probe that guards it.
 *
 * Design: docs/superpowers/specs/2026-08-22-flow-appearance-design.md (rtylr)
 */

// FlowSettings statically imports TimeSection, whose import chain reaches the
// app's real i18next bootstrap and createOneKitClient at module load - so these
// mocks keep the real exports and override only what the panel needs, matching
// FlowSettings.test.tsx.
vi.mock('react-i18next', async (importOriginal) => {
  const actual = await importOriginal<typeof import('react-i18next')>();
  return { ...actual, useTranslation: () => ({ t: (key: string) => key }) };
});

vi.mock('@voxire/components', () => ({ HelmetContainer: () => null }));
vi.mock('sonner', () => ({ toast: { error: vi.fn(), success: vi.fn() } }));

vi.mock('@voxire/config', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@voxire/config')>();
  return {
    ...actual,
    useFeatureFlag: () => false,
    FLAG_KEYS: { APP_FLOW_PREVIEW_ROUTES: 'app.flow.previewRoutes' },
  };
});

vi.mock('@voxire/api', async (importOriginal) => {
  const actual = await importOriginal<typeof import('@voxire/api')>();
  return { ...actual, unwrap: () => Promise.resolve({}) };
});

vi.mock('../../shared/services/flowApi', () => ({
  flowApi: {
    getUserSettings: () => Promise.resolve({}),
    updateUserSettings: () => Promise.resolve({}),
  },
}));

vi.mock('../../shared/hooks/queries/useFlowTeam', () => ({
  useFlowTeam: () => ({ data: [], isLoading: false }),
  useFlowGrantAccess: () => ({ mutate: vi.fn(), isPending: false }),
  useFlowDeactivateMember: () => ({ mutate: vi.fn(), isPending: false }),
}));

vi.mock('../../shared/hooks/queries/useFlowSubscription', () => ({
  useFlowSubscription: () => ({ data: undefined }),
  centsToDisplay: (cents: number) => `$${(cents / 100).toFixed(2)}`,
  computeSeatProration: () => ({
    daysRemaining: 0,
    daysInMonth: 30,
    proratedCents: 0,
    nextMonthName: 'January',
  }),
}));

vi.mock('../../features/tasks/useLabels', () => ({
  useLabels: () => ({ data: [], isLoading: false }),
  useCreateLabel: () => ({ mutate: vi.fn(), isPending: false }),
  useDeleteLabel: () => ({ mutate: vi.fn(), isPending: false }),
}));

vi.mock('@voxire/auth', () => ({ getCookie: () => null }));

vi.mock('@tanstack/react-router', () => ({
  getRouteApi: () => ({ useSearch: () => ({ section: undefined }) }),
}));

// The workspace whose colour the accent falls back to while `accent` is null.
// This comes from the store rather than a query hook: the workspaces API was
// stubbed out when Flow moved to the v2 contracts, so a mock of the old
// useWorkspaces hook would be mocking a module that no longer exists.
vi.mock('../../shared/stores/flowStore', () => ({
  useFlowStore: () => ({
    activeWorkspaceId: 'ws-1',
    workspaces: [{ id: 'ws-1', name: 'Studio', color: '#8b5cf6' }],
  }),
}));

/**
 * A stand-in for the real appearance store. The setters are spies because the
 * point of most of these tests is which setter was called with what, and the
 * state is mutable so a test can start from a chosen accent.
 */
const setters = vi.hoisted(() => ({
  setMode: vi.fn(),
  setPalette: vi.fn(),
  setAccent: vi.fn(),
  setDensity: vi.fn(),
  setTextSize: vi.fn(),
  setCorners: vi.fn(),
  setReduceMotion: vi.fn(),
  reset: vi.fn(),
}));

const storeState = vi.hoisted(() => ({
  prefs: {
    mode: 'dark',
    palette: 'midnight',
    accent: null,
    density: 'cozy',
    textSize: 'm',
    corners: 'soft',
    reduceMotion: false,
  } as {
    mode: string;
    palette: string;
    accent: string | null;
    density: string;
    textSize: string;
    corners: string;
    reduceMotion: boolean;
  },
}));

vi.mock('../../shared/stores/appearanceStore', () => ({
  FLOW_APPEARANCE_STORAGE_KEY: 'flow-appearance',
  useAppearanceStore: () => ({ ...storeState.prefs, ...setters }),
}));

import { AppearanceSection } from './FlowSettings';

const WORKSPACE_ACCENT = '#8b5cf6';

function everySetter(): ReturnType<typeof vi.fn>[] {
  return Object.values(setters);
}

function renderPanel(): ReturnType<typeof render> {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={queryClient}>
      <AppearanceSection />
    </QueryClientProvider>
  );
}

function previewToken(name: string): string {
  const pane = screen.getByTestId('appearance-preview');
  return pane.style.getPropertyValue(name).trim();
}

function paletteCard(id: string): HTMLElement {
  return screen.getByRole('button', { name: new RegExp(`appearance\\.palettes\\.${id}`) });
}

beforeEach(() => {
  for (const spy of everySetter()) spy.mockClear();
  storeState.prefs = {
    mode: 'dark',
    palette: 'midnight',
    accent: null,
    density: 'cozy',
    textSize: 'm',
    corners: 'soft',
    reduceMotion: false,
  };
});

describe('Appearance panel commits', () => {
  it('writes the chosen mode', () => {
    renderPanel();
    fireEvent.click(screen.getByRole('button', { name: 'flow.shell.light' }));
    expect(setters.setMode).toHaveBeenCalledWith('light');
  });

  it('writes the chosen palette', () => {
    renderPanel();
    fireEvent.click(paletteCard('graphite'));
    expect(setters.setPalette).toHaveBeenCalledWith('graphite');
  });

  it('offers only the resolved mode’s palette family', () => {
    renderPanel();
    // Dark mode: the four dark presets, and none of the light ones.
    for (const preset of PALETTES) {
      const query = screen.queryByRole('button', {
        name: new RegExp(`appearance\\.palettes\\.${preset.id}`),
      });
      if (preset.mode === 'dark') expect(query).toBeInTheDocument();
      else expect(query).not.toBeInTheDocument();
    }
  });

  it('writes the chosen accent swatch', () => {
    renderPanel();
    fireEvent.click(screen.getByRole('button', { name: 'appearance.accents.iris' }));
    expect(setters.setAccent).toHaveBeenCalledWith('#8b8cf5');
  });

  it('writes the chosen density and corners', () => {
    renderPanel();
    fireEvent.click(
      screen.getByRole('button', { name: 'flow.settings.appearance.density.compact' })
    );
    fireEvent.click(screen.getByRole('button', { name: 'flow.settings.appearance.corners.round' }));
    expect(setters.setDensity).toHaveBeenCalledWith('compact');
    expect(setters.setCorners).toHaveBeenCalledWith('round');
  });

  it('toggles reduce motion through a switch that reports its state', () => {
    renderPanel();
    const toggle = screen.getByRole('switch', { name: 'flow.settings.appearance.groups.motion' });
    expect(toggle).toHaveAttribute('aria-checked', 'false');
    fireEvent.click(toggle);
    expect(setters.setReduceMotion).toHaveBeenCalledWith(true);
  });
});

describe('Appearance panel accent input', () => {
  it('writes a valid hex, normalised', () => {
    renderPanel();
    fireEvent.change(screen.getByLabelText('flow.settings.appearance.accentHex'), {
      target: { value: '#A3D84B' },
    });
    expect(setters.setAccent).toHaveBeenCalledWith('#a3d84b');
  });

  it('rejects invalid input without throwing and without writing', () => {
    renderPanel();
    const field = screen.getByLabelText('flow.settings.appearance.accentHex');
    expect(() => {
      fireEvent.change(field, { target: { value: '#' } });
      fireEvent.change(field, { target: { value: 'not-a-colour' } });
      fireEvent.change(field, { target: { value: '#12345' } });
      fireEvent.change(field, { target: { value: '#zzzzzz' } });
    }).not.toThrow();

    expect(setters.setAccent).not.toHaveBeenCalled();
    expect(field).toHaveAttribute('aria-invalid', 'true');
    expect(screen.getByText('flow.settings.appearance.accentInvalid')).toBeInTheDocument();
  });

  it('resets to the brand default by writing null', () => {
    storeState.prefs.accent = '#f0805a';
    renderPanel();
    fireEvent.click(
      screen.getByRole('button', { name: /flow\.settings\.appearance\.accentReset/ })
    );
    expect(setters.setAccent).toHaveBeenCalledWith(null);
  });

  it('derives from the supplied fallback accent while the accent is null', () => {
    renderPanel();
    expect(previewToken('--accent')).toBe(WORKSPACE_ACCENT);
  });

  it('lets a stored accent override the supplied fallback', () => {
    storeState.prefs.accent = '#a3d84b';
    renderPanel();
    expect(previewToken('--accent')).toBe('#a3d84b');
  });
});

describe('Appearance panel audition', () => {
  /** The tokens the pane should show for a palette, from the same pure function. */
  function tokensForPalette(id: string): Record<string, string> {
    const prefs = { ...storeState.prefs, palette: id } as unknown as AppearancePrefs;
    return appearanceTokens(prefs, {
      resolvedMode: 'dark',
      workspaceAccent: WORKSPACE_ACCENT,
    });
  }

  it('shows the applied palette when nothing is being auditioned', () => {
    renderPanel();
    expect(previewToken('--flow-bg')).toBe(tokensForPalette('midnight')['--flow-bg']);
  });

  // The core of the audition model: hovering moves the pane and writes nothing.
  it('previews a hovered palette without writing to the store', () => {
    renderPanel();
    fireEvent.pointerEnter(paletteCard('ink'));

    expect(previewToken('--flow-bg')).toBe(tokensForPalette('ink')['--flow-bg']);
    expect(previewToken('--flow-bg')).not.toBe(tokensForPalette('midnight')['--flow-bg']);
    for (const spy of everySetter()) expect(spy).not.toHaveBeenCalled();
  });

  // There is deliberately NO unit test here for "travelling between two cards
  // does not revert the preview", even though that is a real fixed defect.
  //
  // It cannot be tested at this level. fireEvent.pointerLeave on a card is
  // routed to the grid's handler by React's enter/leave simulation, so the test
  // environment cannot tell a per-card clear from a per-grid one. A test written
  // for it passed with the defect deliberately reintroduced, which makes it
  // worse than no test: it reports a guard that does not exist.
  //
  // Proven by probe rather than assumed. Removing the grid's onPointerLeave
  // fails the next test down. Restoring the card's onPointerLeave fails nothing.
  //
  // The evidence for the fix is a browser measurement: sampling one pixel at a
  // time across the 8px gap between two palette cards showed the applied
  // palette for eight pixels of travel. If you are changing this, verify it the
  // same way; the suite will not stop you.

  it('snaps back to the applied palette when the pointer leaves', () => {
    renderPanel();
    const card = paletteCard('ink');
    fireEvent.pointerEnter(card);
    fireEvent.pointerLeave(card);
    expect(previewToken('--flow-bg')).toBe(tokensForPalette('midnight')['--flow-bg']);
  });

  // focusIn / focusOut rather than .focus(): React listens for the bubbling
  // pair, and testing-library wraps a dispatched event in act() where a direct
  // .focus() call would leave the state update outside it.
  it('auditions on focus too, so the keyboard path is not a dead end', () => {
    renderPanel();
    const card = paletteCard('moss');

    fireEvent.focusIn(card);
    expect(previewToken('--flow-bg')).toBe(tokensForPalette('moss')['--flow-bg']);
    for (const spy of everySetter()) expect(spy).not.toHaveBeenCalled();

    fireEvent.focusOut(card);
    expect(previewToken('--flow-bg')).toBe(tokensForPalette('midnight')['--flow-bg']);
  });

  it('previews a hovered accent swatch without writing to the store', () => {
    renderPanel();
    fireEvent.pointerEnter(screen.getByRole('button', { name: 'appearance.accents.coral' }));
    expect(previewToken('--accent')).toBe('#f0805a');
    for (const spy of everySetter()) expect(spy).not.toHaveBeenCalled();
  });

  it('previews a hovered mode without writing to the store', () => {
    renderPanel();
    fireEvent.pointerEnter(screen.getByRole('button', { name: 'flow.shell.light' }));
    // Light mode falls back to the light family's first preset.
    expect(previewToken('--flow-bg')).toBe('#f4f4f5');
    for (const spy of everySetter()) expect(spy).not.toHaveBeenCalled();
  });

  it('previews a hovered density without writing to the store', () => {
    renderPanel();
    fireEvent.pointerEnter(
      screen.getByRole('button', { name: 'flow.settings.appearance.density.roomy' })
    );
    // A unitless multiplier, not a length: --spacing keeps its rem base so
    // spacing still follows a browser font-size setting.
    expect(previewToken('--flow-u')).toBe('1.25');
    for (const spy of everySetter()) expect(spy).not.toHaveBeenCalled();
  });

  it('never writes appearance tokens to documentElement', () => {
    renderPanel();
    fireEvent.pointerEnter(paletteCard('ink'));
    // The provider is the single writer. A second writer on the root would
    // outrank every stylesheet rule and could not be undone by one.
    expect(document.documentElement.style.getPropertyValue('--flow-bg')).toBe('');
    expect(document.documentElement.style.getPropertyValue('--accent')).toBe('');
  });
});

describe('Appearance panel text size gate', () => {
  it('renders the control disabled, with the reason, and cannot be activated', () => {
    renderPanel();
    const large = screen.getByRole('button', { name: 'flow.settings.appearance.textSize.l' });

    // aria-disabled, not native disabled, and the distinction is the point: a
    // natively disabled button leaves the tab order, which makes the
    // aria-describedby below unannounceable and hides the reason from exactly
    // the users who most need it. Asserting reachability, not just inertness.
    expect(large).toHaveAttribute('aria-disabled', 'true');
    expect(large).not.toBeDisabled();
    expect(large.tabIndex).toBeGreaterThanOrEqual(0);
    expect(large).toHaveAttribute('aria-describedby', 'appearance-text-size-gate');
    expect(screen.getByText('flow.settings.appearance.textSizeGate')).toBeInTheDocument();

    fireEvent.click(large);
    fireEvent.pointerEnter(large);
    fireEvent.keyDown(large, { key: 'Enter' });
    fireEvent.keyDown(large, { key: ' ' });
    expect(setters.setTextSize).not.toHaveBeenCalled();
    // The audition is wired off too, so the pane never previews a value the
    // user has no way of selecting.
    expect(previewToken('--flow-fs')).toBe('1');
  });

  it('still shows the concept: all three steps are present', () => {
    renderPanel();
    for (const step of ['s', 'm', 'l']) {
      expect(
        screen.getByRole('button', { name: `flow.settings.appearance.textSize.${step}` })
      ).toBeInTheDocument();
    }
  });
});

describe('Appearance panel disclosure', () => {
  it('states that the settings are per-device', () => {
    renderPanel();
    expect(screen.getByText('flow.settings.appearance.perDevice')).toBeInTheDocument();
  });
});
