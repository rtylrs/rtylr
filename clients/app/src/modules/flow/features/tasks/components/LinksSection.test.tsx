import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, opts?: Record<string, string>) => (opts?.name ? `${key}:${opts.name}` : key),
    i18n: { language: 'en' },
  }),
}));

const addMutate = vi.hoisted(() => vi.fn());
const removeMutate = vi.hoisted(() => vi.fn());
const links = vi.hoisted(() => ({ data: [] as Array<Record<string, unknown>> }));

vi.mock('../useTaskLinks', async () => {
  const actual = await vi.importActual<typeof import('../useTaskLinks')>('../useTaskLinks');
  return {
    isWebUrl: actual.isWebUrl,
    useTaskLinks: () => ({ data: links.data }),
    useAddTaskLink: () => ({ mutate: addMutate, isPending: false }),
    useRemoveTaskLink: () => ({ mutate: removeMutate, isPending: false }),
  };
});

import LinksSection from './LinksSection';

const PR = {
  id: '9',
  taskId: '1',
  url: 'https://github.com/rtylrs/rtylr/pull/1615',
  kind: 'github_pr',
  title: 'Developer project switch',
  ref: 'rtylrs/rtylr#1615',
  state: 'merged',
  addedBy: '5',
  createdAt: '2026-11-17T08:00:00Z',
};

describe('LinksSection', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    links.data = [];
  });

  it('shows a link card with its title, reference, state and who added it', () => {
    links.data = [PR];
    render(<LinksSection taskId="1" canEdit peopleNames={{ '5': 'Abed' }} />);
    const anchor = screen.getByRole('link', { name: 'Developer project switch' });
    expect(anchor).toHaveAttribute('href', PR.url);
    expect(anchor).toHaveAttribute('rel', 'noopener noreferrer');
    expect(screen.getByText('rtylrs/rtylr#1615')).toBeInTheDocument();
    expect(screen.getByText('flow.taskDetail.linkState.merged')).toBeInTheDocument();
    expect(screen.getByText('flow.taskDetail.linkAddedBy:Abed')).toBeInTheDocument();
  });

  it('adds a pasted URL', () => {
    render(<LinksSection taskId="1" canEdit peopleNames={{}} />);
    fireEvent.change(screen.getByLabelText('flow.taskDetail.addLink'), {
      target: { value: ' https://www.figma.com/design/abc ' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'flow.taskDetail.addLink' }));
    expect(addMutate).toHaveBeenCalledWith(
      { taskId: '1', url: 'https://www.figma.com/design/abc' },
      expect.anything()
    );
  });

  it('rejects text that is not a web link without calling the server', () => {
    render(<LinksSection taskId="1" canEdit peopleNames={{}} />);
    fireEvent.change(screen.getByLabelText('flow.taskDetail.addLink'), {
      target: { value: 'javascript:alert(1)' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'flow.taskDetail.addLink' }));
    expect(addMutate).not.toHaveBeenCalled();
    expect(screen.getByText('flow.taskDetail.invalidLink')).toBeInTheDocument();
  });

  it('lets an editor remove any link, and hides editing for read-only viewers', () => {
    links.data = [PR];
    const { unmount } = render(<LinksSection taskId="1" canEdit peopleNames={{}} />);
    fireEvent.click(
      screen.getByRole('button', { name: 'flow.taskDetail.removeLink:Developer project switch' })
    );
    expect(removeMutate).toHaveBeenCalledWith({ taskId: '1', linkId: '9' });
    unmount();

    render(<LinksSection taskId="1" canEdit={false} peopleNames={{}} />);
    expect(screen.queryByRole('button', { name: /removeLink/ })).toBeNull();
    expect(screen.queryByLabelText('flow.taskDetail.addLink')).toBeNull();
  });
});
