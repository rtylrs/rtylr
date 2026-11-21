import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, params?: Record<string, unknown>) =>
      params ? `${key}:${Object.values(params).join(',')}` : key,
    i18n: { language: 'en-US' },
  }),
}));

import BoardSettingsPanel from './BoardSettingsPanel';
import { DEFAULT_BOARD_VIEW_SETTINGS } from '../boardViewSettings';

const COLUMNS = [
  { id: '1', name: 'Backlog' },
  { id: '2', name: 'In progress' },
  { id: '3', name: 'Done' },
];

function renderPanel(props: Partial<React.ComponentProps<typeof BoardSettingsPanel>> = {}) {
  return render(
    <BoardSettingsPanel
      open
      onClose={vi.fn()}
      columns={COLUMNS}
      canManage
      settings={DEFAULT_BOARD_VIEW_SETTINGS}
      onSettingsChange={vi.fn()}
      onResetSettings={vi.fn()}
      onManageStatuses={vi.fn()}
      {...props}
    />
  );
}

beforeEach(() => {
  vi.clearAllMocks();
});

describe('BoardSettingsPanel: columns are statuses', () => {
  it('sends column changes to the statuses panel instead of a second editor', () => {
    const onManageStatuses = vi.fn();
    renderPanel({ onManageStatuses });

    expect(screen.getByText('flow.board.columnsAreStatuses')).toBeInTheDocument();
    expect(screen.queryByLabelText('flow.board.columnName')).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'flow.status.openSettings' }));

    expect(onManageStatuses).toHaveBeenCalledOnce();
  });
});

describe('BoardSettingsPanel: what has no field to persist', () => {
  it('disables the prompts-for-hours choice and names the missing field', () => {
    renderPanel();

    expect(
      screen.getByRole('switch', { name: 'flow.board.promptsForHoursFor:Done' })
    ).toBeDisabled();
    expect(screen.getByText('flow.board.promptsForHoursBlocked')).toBeInTheDocument();
  });
});

describe('BoardSettingsPanel: your view', () => {
  it('changes the grouping, including labels, cycles and milestones', () => {
    const onSettingsChange = vi.fn();
    renderPanel({ onSettingsChange });

    fireEvent.click(screen.getByRole('radio', { name: 'flow.board.groupBy.assignee' }));
    fireEvent.click(screen.getByRole('radio', { name: 'flow.board.groupBy.label' }));
    fireEvent.click(screen.getByRole('radio', { name: 'flow.board.groupBy.cycle' }));
    fireEvent.click(screen.getByRole('radio', { name: 'flow.board.groupBy.milestone' }));

    expect(onSettingsChange.mock.calls.map(([patch]) => patch.groupBy)).toEqual([
      'assignee',
      'label',
      'cycle',
      'milestone',
    ]);
  });

  it('starts the card toggles on the card as it is today', () => {
    renderPanel();

    expect(screen.getByRole('switch', { name: 'flow.board.cardField.key' })).toBeChecked();
    expect(screen.getByRole('switch', { name: 'flow.board.cardField.assignee' })).toBeChecked();
    expect(screen.getByRole('switch', { name: 'flow.board.cardField.estimate' })).not.toBeChecked();
    expect(screen.getByRole('switch', { name: 'flow.board.cardField.cycle' })).not.toBeChecked();
  });

  it('turns one card field on and keeps the others', () => {
    const onSettingsChange = vi.fn();
    renderPanel({ onSettingsChange });

    fireEvent.click(screen.getByRole('switch', { name: 'flow.board.cardField.estimate' }));

    expect(onSettingsChange).toHaveBeenCalledWith({
      cardFields: { ...DEFAULT_BOARD_VIEW_SETTINGS.cardFields, estimate: true },
    });
  });

  it('offers 1, 7, 14 and 30 days or never for done cards, starting at 14', () => {
    const onSettingsChange = vi.fn();
    renderPanel({ onSettingsChange });
    const group = screen.getByRole('radiogroup', { name: 'flow.board.hideDoneHeading' });

    expect(
      screen.getByRole('radio', { name: 'flow.board.hideDone.days14' }).getAttribute('aria-checked')
    ).toBe('true');
    expect(group.querySelectorAll('[role="radio"]')).toHaveLength(5);
    fireEvent.click(screen.getByRole('radio', { name: 'flow.board.hideDone.never' }));

    expect(onSettingsChange).toHaveBeenCalledWith({ hideDoneAfterDays: null });
  });

  it('switches the columns between fixed and flexible', () => {
    const onSettingsChange = vi.fn();
    renderPanel({ onSettingsChange });

    fireEvent.click(screen.getByRole('radio', { name: 'flow.board.columnWidth.flexible' }));

    expect(onSettingsChange).toHaveBeenCalledWith({ columnWidth: 'flexible' });
  });

  it('resets the view', () => {
    const onResetSettings = vi.fn();
    renderPanel({ onResetSettings });

    fireEvent.click(screen.getByRole('button', { name: 'flow.board.resetView' }));

    expect(onResetSettings).toHaveBeenCalledOnce();
  });

  it('says which settings are yours and which your team sees', () => {
    renderPanel();

    expect(screen.getByText('flow.board.scopeExplained')).toBeInTheDocument();
  });
});

describe('BoardSettingsPanel: a viewer', () => {
  it('shows the panel read-only with the reason rather than hiding it', () => {
    renderPanel({ canManage: false });

    expect(screen.getByText('flow.board.readOnlyReason')).toBeInTheDocument();
    expect(
      screen.getByRole('switch', { name: 'flow.board.promptsForHoursFor:Done' })
    ).toBeDisabled();
  });

  it('still lets a viewer change their own view, which is their preference alone', () => {
    const onSettingsChange = vi.fn();
    renderPanel({ canManage: false, onSettingsChange });

    fireEvent.click(screen.getByRole('radio', { name: 'flow.board.groupBy.priority' }));
    fireEvent.click(screen.getByRole('switch', { name: 'flow.board.cardField.milestone' }));

    expect(onSettingsChange).toHaveBeenCalledWith({ groupBy: 'priority' });
    expect(onSettingsChange).toHaveBeenCalledWith({
      cardFields: { ...DEFAULT_BOARD_VIEW_SETTINGS.cardFields, milestone: true },
    });
  });
});

describe('BoardSettingsPanel: dismissal and height', () => {
  it('closes on Escape when the key lands on the body', async () => {
    // The reported defect. onKeyDown sat on the overlay div, which has no
    // tabIndex and so cannot be focused, and the panel never took focus on
    // open, so the keydown landed on document.body and Esc did nothing.
    const onClose = vi.fn();
    renderPanel({ onClose });
    await waitFor(() => expect(document.activeElement).not.toBe(document.body));
    fireEvent.keyDown(document.body, { key: 'Escape' });
    expect(onClose).toHaveBeenCalledTimes(1);
  });

  it('moves focus into the panel on open', async () => {
    const { container } = renderPanel();
    const panel = container.querySelector('[role="dialog"] > div');
    expect(panel).not.toBeNull();
    await waitFor(() => expect(panel?.contains(document.activeElement)).toBe(true));
  });

  it('bounds the panel height and lets it scroll', () => {
    // Measured at 1242px in a 900px viewport, centred by the overlay so its
    // own title sat at -171px, with overflow-y visible and nothing to scroll.
    // Now a side panel: the full viewport height, scrolling inside.
    const { container } = renderPanel();
    const panel = container.querySelector('[role="dialog"] > div');
    expect(panel?.className).toContain('max-h-dvh');
    expect(panel?.className).toContain('overflow-y-auto');
  });

  it('does not close on Escape while shut', () => {
    const onClose = vi.fn();
    renderPanel({ open: false, onClose });
    fireEvent.keyDown(document.body, { key: 'Escape' });
    expect(onClose).not.toHaveBeenCalled();
  });
});
