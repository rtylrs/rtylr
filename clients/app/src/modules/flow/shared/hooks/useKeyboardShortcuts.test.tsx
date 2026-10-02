import React from 'react';
import { render, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

const navigateSpy = vi.hoisted(() => vi.fn());
vi.mock('@tanstack/react-router', () => ({ useNavigate: () => navigateSpy }));

const flagState = vi.hoisted(() => ({ previewRoutes: false }));
vi.mock('@voxire/config', () => ({
  FLAG_KEYS: { APP_FLOW_PREVIEW_ROUTES: 'APP_FLOW_PREVIEW_ROUTES' },
  useFeatureFlag: () => flagState.previewRoutes,
}));

const storeState = vi.hoisted(() => ({
  setCommandBarOpen: vi.fn(),
  setQuickCreateOpen: vi.fn(),
  commandBarOpen: false,
  quickCreateOpen: false,
  activeWorkspaceId: 'ws-1',
}));
vi.mock('../stores/flowStore', () => ({ useFlowStore: () => storeState }));

import { useKeyboardShortcuts } from './useKeyboardShortcuts';
import { useCommandPalette } from '../../../../shared/stores/commandPaletteStore';
import { useUserStore } from '../../../../shared/stores/userStore';
import { createEffectiveAccess } from '../../../../../test/effectiveAccess';

function Host(): React.JSX.Element {
  useKeyboardShortcuts();
  return (
    <div>
      <input aria-label="field" />
      {/* A descendant that swallows the key, which is what took `c` from the
          shortcut the command bar advertises. */}
      <div data-testid="greedy" onKeyDown={(e) => e.stopPropagation()}>
        <button type="button">inside</button>
      </div>
    </div>
  );
}

beforeEach(() => {
  navigateSpy.mockClear();
  storeState.setCommandBarOpen.mockClear();
  storeState.setQuickCreateOpen.mockClear();
  storeState.commandBarOpen = false;
  storeState.quickCreateOpen = false;
  flagState.previewRoutes = false;
  useCommandPalette.setState({ open: false });
  useUserStore.setState({
    user: { effectiveAccess: createEffectiveAccess('flow', { isOwner: true }) },
  });
});

describe('useKeyboardShortcuts', () => {
  it('opens quick create on c, the shortcut the command bar prints', () => {
    render(<Host />);
    fireEvent.keyDown(document.body, { key: 'c' });
    expect(storeState.setQuickCreateOpen).toHaveBeenCalledWith(true);
  });

  it('still honours c when a descendant stops propagation', () => {
    // The reported defect: the command bar prints C for New task and the key
    // went somewhere else. A bubble-phase window listener loses to anything
    // below it that calls stopPropagation.
    const { getByText } = render(<Host />);
    fireEvent.keyDown(getByText('inside'), { key: 'c' });
    expect(storeState.setQuickCreateOpen).toHaveBeenCalledWith(true);
  });

  it('leaves typing alone', () => {
    const { getByLabelText } = render(<Host />);
    fireEvent.keyDown(getByLabelText('field'), { key: 'c' });
    expect(storeState.setQuickCreateOpen).not.toHaveBeenCalled();
  });

  it('opens the command bar on slash', () => {
    render(<Host />);
    fireEvent.keyDown(document.body, { key: '/' });
    expect(useCommandPalette.getState().open).toBe(true);
  });

  it.each(['t', 'r', 's'])(
    'does not navigate g %s into a preview route while the flag is off',
    (key) => {
      // The command bar already filters these three. The chords were the
      // remaining way into "Preview route unavailable".
      render(<Host />);
      fireEvent.keyDown(document.body, { key: 'g' });
      fireEvent.keyDown(document.body, { key });
      expect(navigateSpy).not.toHaveBeenCalled();
    }
  );

  it.each([
    ['t', '/flow/timeline'],
    ['r', '/flow/roadmap'],
    ['s', '/flow/insights'],
  ])('navigates g %s once the flag is on', (key, to) => {
    flagState.previewRoutes = true;
    render(<Host />);
    fireEvent.keyDown(document.body, { key: 'g' });
    fireEvent.keyDown(document.body, { key });
    expect(navigateSpy).toHaveBeenCalledWith({ to });
  });

  it.each([
    ['i', '/flow/inbox'],
    ['m', '/flow/my-issues'],
    ['y', '/flow/cycles'],
    ['x', '/flow/triage'],
    ['f', '/flow/today'],
  ])('navigates g %s regardless of the flag', (key, to) => {
    render(<Host />);
    fireEvent.keyDown(document.body, { key: 'g' });
    fireEvent.keyDown(document.body, { key });
    expect(navigateSpy).toHaveBeenCalledWith({ to });
  });

  it('does not reopen quick create when it is already open', () => {
    storeState.quickCreateOpen = true;
    render(<Host />);
    fireEvent.keyDown(document.body, { key: 'c' });
    expect(storeState.setQuickCreateOpen).not.toHaveBeenCalled();
  });
});
