import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import type { ApiProjectColumn } from '../../../shared/services/flowApi';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, params?: Record<string, unknown>) =>
      params ? `${key}:${Object.values(params).join(',')}` : key,
    i18n: { language: 'en-US' },
  }),
}));

const createColumn = vi.hoisted(() => vi.fn());
const updateColumn = vi.hoisted(() => vi.fn());
const deleteColumn = vi.hoisted(() => vi.fn());

vi.mock('../useColumns', () => ({
  useCreateColumn: () => ({ mutate: createColumn, isPending: false }),
  useUpdateColumn: () => ({ mutate: updateColumn, isPending: false }),
  useDeleteColumn: () => ({ mutate: deleteColumn, isPending: false }),
}));

import BoardSettingsPanel from './BoardSettingsPanel';

function column(over: Partial<ApiProjectColumn> = {}): ApiProjectColumn {
  return {
    id: 1,
    projectId: 'p1',
    name: 'Backlog',
    color: '#6b7280',
    sortOrder: 0,
    createdAt: '',
    updatedAt: '',
    ...over,
  };
}

const COLUMNS = [
  column({ id: 1, name: 'Backlog', sortOrder: 0 }),
  column({ id: 2, name: 'In progress', sortOrder: 1 }),
  column({ id: 3, name: 'Done', sortOrder: 2 }),
];

function renderPanel(props: Partial<React.ComponentProps<typeof BoardSettingsPanel>> = {}) {
  return render(
    <BoardSettingsPanel
      open
      onClose={vi.fn()}
      projectId="p1"
      workspaceId="w1"
      columns={COLUMNS}
      canManage
      groupBy="none"
      onGroupByChange={vi.fn()}
      {...props}
    />
  );
}

beforeEach(() => {
  vi.clearAllMocks();
});

describe('BoardSettingsPanel: columns', () => {
  it('lists every column in sort order', () => {
    renderPanel();

    const names = screen
      .getAllByLabelText('flow.board.columnName')
      .map((i) => (i as HTMLInputElement).value);
    expect(names).toEqual(['Backlog', 'In progress', 'Done']);
  });

  it('renames a column on blur', () => {
    renderPanel();

    const first = screen.getAllByLabelText('flow.board.columnName')[0];
    fireEvent.change(first, { target: { value: 'Icebox' } });
    fireEvent.blur(first);

    expect(updateColumn).toHaveBeenCalledWith({
      id: '1',
      name: 'Icebox',
      color: '#6b7280',
      sortOrder: 0,
    });
  });

  it('does not send a rename when the name is unchanged', () => {
    renderPanel();

    fireEvent.blur(screen.getAllByLabelText('flow.board.columnName')[0]);

    expect(updateColumn).not.toHaveBeenCalled();
  });

  it('refuses to save a blank name and restores the previous one', () => {
    renderPanel();

    const first = screen.getAllByLabelText('flow.board.columnName')[0];
    fireEvent.change(first, { target: { value: '   ' } });
    fireEvent.blur(first);

    expect(updateColumn).not.toHaveBeenCalled();
    expect((first as HTMLInputElement).value).toBe('Backlog');
  });

  it('adds a column at the end', () => {
    renderPanel();

    fireEvent.change(screen.getByLabelText('flow.board.newColumnName'), {
      target: { value: 'Blocked' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'flow.board.addColumn' }));

    expect(createColumn).toHaveBeenCalledWith({
      name: 'Blocked',
      sortOrder: 3,
    });
  });

  it('labels each remove control by the column it removes, not by the verb', () => {
    renderPanel();

    expect(
      screen.getByRole('button', { name: 'flow.board.removeColumnLabel:Backlog' })
    ).toBeInTheDocument();
  });

  it('asks before deleting, then deletes', () => {
    renderPanel();

    fireEvent.click(screen.getByRole('button', { name: 'flow.board.removeColumnLabel:Backlog' }));
    expect(deleteColumn).not.toHaveBeenCalled();

    fireEvent.click(screen.getByRole('button', { name: 'flow.board.confirmRemoveColumn' }));
    expect(deleteColumn).toHaveBeenCalledWith('1');
  });

  it('protects the last column, because a board with none cannot hold a task', () => {
    renderPanel({ columns: [COLUMNS[0]] });

    expect(
      screen.getByRole('button', { name: 'flow.board.removeColumnLabel:Backlog' })
    ).toBeDisabled();
    expect(screen.getByText('flow.board.lastColumnProtected')).toBeInTheDocument();
  });
});

describe('BoardSettingsPanel: reordering', () => {
  it('swaps sort order with the neighbour above', () => {
    renderPanel();

    fireEvent.click(screen.getByRole('button', { name: 'flow.board.moveUpLabel:In progress' }));

    expect(updateColumn).toHaveBeenCalledWith({
      id: '2',
      name: 'In progress',
      color: '#6b7280',
      sortOrder: 0,
    });
    expect(updateColumn).toHaveBeenCalledWith({
      id: '1',
      name: 'Backlog',
      color: '#6b7280',
      sortOrder: 1,
    });
  });

  it('cannot move the first column up or the last down', () => {
    renderPanel();

    expect(screen.getByRole('button', { name: 'flow.board.moveUpLabel:Backlog' })).toBeDisabled();
    expect(screen.getByRole('button', { name: 'flow.board.moveDownLabel:Done' })).toBeDisabled();
  });

  it('reorders from the keyboard with Alt+ArrowDown, without drag', () => {
    renderPanel();

    fireEvent.keyDown(screen.getAllByLabelText('flow.board.columnName')[0], {
      key: 'ArrowDown',
      altKey: true,
    });

    expect(updateColumn).toHaveBeenCalledWith({
      id: '1',
      name: 'Backlog',
      color: '#6b7280',
      sortOrder: 1,
    });
  });

  it('announces the move for a screen reader', () => {
    renderPanel();

    fireEvent.click(screen.getByRole('button', { name: 'flow.board.moveUpLabel:In progress' }));

    const live = screen.getByRole('status');
    expect(live).toHaveAttribute('aria-live', 'polite');
    expect(live).toHaveTextContent('flow.board.movedAnnouncement');
  });
});

describe('BoardSettingsPanel: what has no field to persist', () => {
  it('disables the WIP limit and names the missing field', () => {
    renderPanel();

    const wip = screen.getAllByLabelText('flow.board.wipLimitFor:Backlog')[0];
    expect(wip).toBeDisabled();
    expect(screen.getByText('flow.board.wipLimitBlocked')).toBeInTheDocument();
  });

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
    expect(screen.getAllByLabelText('flow.board.columnName')[0]).toBeDisabled();
    expect(screen.queryByRole('button', { name: 'flow.board.addColumn' })).not.toBeInTheDocument();
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
