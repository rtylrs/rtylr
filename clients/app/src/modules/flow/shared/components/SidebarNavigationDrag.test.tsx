// @vitest-environment happy-dom
import React from 'react';
import { afterEach, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';
import { createEvent, fireEvent, render } from '@testing-library/react';
import path from 'node:path';
import { initTestI18nFrom } from '../../../../../test/i18nTestInstance';

vi.mock('@tanstack/react-router', () => ({
  Link: ({ children, ...props }: React.PropsWithChildren<Record<string, unknown>>) => (
    <a href="/" {...props}>
      {children}
    </a>
  ),
  useNavigate: () => vi.fn(),
  useLocation: () => ({ pathname: '/' }),
}));

import { SidebarNavigation } from './SidebarNavigation';
import {
  createDefaultSidebarPreferences,
  type SidebarPreferences,
} from '../lib/sidebarPreferences';

beforeAll(() => initTestI18nFrom(path.resolve(import.meta.dirname, '../../locales')));

const update = vi.fn();
const preferences = createDefaultSidebarPreferences();

const renderNav = () =>
  render(
    <SidebarNavigation
      preferences={preferences}
      update={update}
      showPreviewRoutes={false}
      unreadNotifications={0}
      myIssuesCount={0}
      todayFocusCount={0}
      projects={[]}
      onNewProject={vi.fn()}
      sidebarRef={{ current: null }}
    />
  );

const dataTransfer = { setData: vi.fn(), effectAllowed: '', dropEffect: '' };

// happy-dom has no layout: every draggable row is 40px tall, stacked in DOM order.
let restoreRect: () => void;
let clock = 0;
beforeEach(() => {
  vi.clearAllMocks();
  // The drag ignores pointer moves for 140ms after the preview changes; step the
  // clock past that on every read so each move in a test counts.
  clock = 0;
  vi.spyOn(performance, 'now').mockImplementation(() => (clock += 500));
  const original = HTMLElement.prototype.getBoundingClientRect;
  HTMLElement.prototype.getBoundingClientRect = function (this: HTMLElement): DOMRect {
    const rows = [...document.querySelectorAll('nav [draggable="true"]')];
    const at = rows.indexOf(this);
    const top = at < 0 ? 0 : at * 40;
    return { top, height: 40, bottom: top + 40, left: 0, right: 200, width: 200 } as DOMRect;
  };
  restoreRect = () => {
    HTMLElement.prototype.getBoundingClientRect = original;
  };
});
afterEach(() => {
  restoreRect();
  vi.restoreAllMocks();
});

/** The top-level navigation rows, as labelled, in the order they are drawn. */
const topRows = (): string[] =>
  [...document.querySelectorAll('nav')[0].querySelectorAll('[draggable="true"]')].map((row) =>
    (row.textContent ?? '').trim()
  );
const row = (label: string): HTMLElement => {
  const found = [
    ...document.querySelectorAll('nav')[0].querySelectorAll('[draggable="true"]'),
  ].find((el) => (el.textContent ?? '').trim() === label);
  if (!found) throw new Error(`no row "${label}"`);
  return found as HTMLElement;
};
// happy-dom builds drag events as plain Events, which carry no pointer position,
// so the position the sidebar reads (clientY) is set on the event directly.
const dragOver = (el: HTMLElement, clientY?: number): void => {
  const event = createEvent.dragOver(el, { dataTransfer });
  if (clientY !== undefined) Object.defineProperty(event, 'clientY', { value: clientY });
  fireEvent(el, event);
};
const upperHalf = (el: HTMLElement): number => el.getBoundingClientRect().top + 5;
const lowerHalf = (el: HTMLElement): number => el.getBoundingClientRect().top + 35;

describe('Sidebar live row drag', () => {
  it('moves the dragged row to where the pointer is, and the others slide aside', () => {
    renderNav();
    const before = topRows();
    expect(before.slice(0, 3)).toEqual(['Inbox', 'My tasks', 'Today']);

    fireEvent.dragStart(row('Inbox'), { dataTransfer });
    // Lower half of "My tasks": the dragged row goes after it.
    dragOver(row('My tasks'), lowerHalf(row('My tasks')));
    expect(topRows().slice(0, 3)).toEqual(['My tasks', 'Inbox', 'Today']);

    // Upper half of "Today" would put it before Today: that is where it already is.
    dragOver(row('Today'), upperHalf(row('Today')));
    expect(topRows().slice(0, 3)).toEqual(['My tasks', 'Inbox', 'Today']);
    // Lower half of Today: after it.
    dragOver(row('Today'), lowerHalf(row('Today')));
    expect(topRows().slice(0, 3)).toEqual(['My tasks', 'Today', 'Inbox']);
  });

  it('saves nothing until the drop, and saves the previewed order on it', () => {
    renderNav();
    fireEvent.dragStart(row('Inbox'), { dataTransfer });
    dragOver(row('My tasks'), lowerHalf(row('My tasks')));
    expect(update).not.toHaveBeenCalled();

    fireEvent.drop(row('Inbox'), { dataTransfer });
    expect(update).toHaveBeenCalledTimes(1);
    const next = (update.mock.calls[0][0] as (draft: SidebarPreferences) => SidebarPreferences)(
      preferences
    );
    expect(next.topLevelItemIds.slice(0, 3)).toEqual([
      'builtin:my-issues',
      'builtin:inbox',
      'builtin:today',
    ]);
  });

  it('puts everything back when the drag ends without a drop', () => {
    renderNav();
    const before = topRows();
    fireEvent.dragStart(row('Inbox'), { dataTransfer });
    dragOver(row('Today'), lowerHalf(row('Today')));
    expect(topRows()).not.toEqual(before);

    fireEvent.dragEnd(window);
    expect(topRows()).toEqual(before);
    expect(update).not.toHaveBeenCalled();
  });

  it('reorders whole groups live as well', () => {
    renderNav();
    const sections = (): string[] =>
      [...document.querySelectorAll('section[aria-label]')].map(
        (el) => el.getAttribute('aria-label') ?? ''
      );
    const header = (label: string): HTMLElement =>
      document.querySelector(
        `section[aria-label="${label}"] > div[draggable="true"]`
      ) as HTMLElement;
    const order = sections();
    expect(order.length).toBeGreaterThan(1);
    const [first, second] = order;

    fireEvent.dragStart(header(first), { dataTransfer });
    dragOver(header(second));
    expect(sections().slice(0, 2)).toEqual([second, first]);

    fireEvent.drop(header(second), { dataTransfer });
    const next = (update.mock.calls[0][0] as (draft: SidebarPreferences) => SidebarPreferences)(
      preferences
    );
    expect(next.groupOrder.indexOf(preferences.groupOrder[0])).toBe(1);
  });
});
