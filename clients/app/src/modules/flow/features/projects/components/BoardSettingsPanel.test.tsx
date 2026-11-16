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
      groupBy="none"
      onGroupByChange={vi.fn()}
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

  it('disables the card-field toggles and names the missing field', () => {
    renderPanel();

    expect(screen.getByRole('switch', { name: 'flow.board.cardField.loggedTime' })).toBeDisabled();
    expect(screen.getByText('flow.board.cardFieldsBlocked')).toBeInTheDocument();
  });
});

describe('BoardSettingsPanel: group by', () => {
  it('changes the grouping', () => {
    const onGroupByChange = vi.fn();
    renderPanel({ onGroupByChange });

    fireEvent.click(screen.getByRole('radio', { name: 'flow.board.groupBy.assignee' }));

    expect(onGroupByChange).toHaveBeenCalledWith('assignee');
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

  it('still lets a viewer change their own grouping, which is their preference alone', () => {
    const onGroupByChange = vi.fn();
    renderPanel({ canManage: false, onGroupByChange });

    fireEvent.click(screen.getByRole('radio', { name: 'flow.board.groupBy.priority' }));

    expect(onGroupByChange).toHaveBeenCalledWith('priority');
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
    const { container } = renderPanel();
    const panel = container.querySelector('[role="dialog"] > div');
    expect(panel?.className).toContain('max-h-[calc(100dvh-64px)]');
    expect(panel?.className).toContain('overflow-y-auto');
  });

  it('does not close on Escape while shut', () => {
    const onClose = vi.fn();
    renderPanel({ open: false, onClose });
    fireEvent.keyDown(document.body, { key: 'Escape' });
    expect(onClose).not.toHaveBeenCalled();
  });
});
