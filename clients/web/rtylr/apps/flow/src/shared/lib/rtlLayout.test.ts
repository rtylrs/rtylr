import { readFileSync } from 'node:fs';
import { join } from 'node:path';

const sourceRoot = join(import.meta.dirname, '../..');

function source(path: string): string {
  return readFileSync(join(sourceRoot, path), 'utf8');
}

/**
 * Flow renders inside the shared `dir` attribute that `@voxire/i18n` sets on
 * `<html>`, so its own geometry has to be direction-agnostic. These assertions
 * mirror `apps/erp/src/shared/lib/rtlLayout.test.ts`: they pin the logical
 * utilities in place and fail if a physical `left-*`/`right-*`/`text-left`
 * class is reintroduced.
 */
describe('Flow RTL layout', () => {
  it('anchors the main nav sidebar to the inline start edge', () => {
    const layout = source('layouts/FlowLayout.tsx');

    expect(layout).toContain('fixed inset-y-0 start-0 z-50');
    expect(layout).toContain('border-e border-[var(--flow-border)]');
    // The slide-in transform is physical in CSS, so the closed state needs an
    // explicit RTL counterpart plus an explicit desktop reset for both
    // directions - otherwise variant ordering decides whether the desktop RTL
    // sidebar sits offscreen.
    expect(layout).toContain('-translate-x-full rtl:translate-x-full');
    expect(layout).toContain('lg:translate-x-0 lg:rtl:translate-x-0');

    expect(layout).not.toContain('left-0');
    expect(layout).not.toContain('border-r ');
  });

  it('keeps sidebar text and spacing logical', () => {
    const layout = source('layouts/FlowLayout.tsx');

    expect(layout).toContain('text-start');
    expect(layout).toContain('ms-auto');

    expect(layout).not.toContain('text-left');
    expect(layout).not.toContain('ml-auto');
  });

  it('mirrors directional icons when the document is RTL', () => {
    const stylesheet = source('globals.css');
    const layout = source('layouts/FlowLayout.tsx');
    const detail = source('features/tasks/TaskDetail.tsx');

    expect(stylesheet).toContain("[dir='rtl'] .flow-directional-icon");
    // ChevronRight in the user menu and the task back-arrow are the only two
    // horizontal glyphs in these files; ChevronDown never mirrors.
    expect(layout).toContain('className="flow-directional-icon shrink-0"');
    expect(detail).toContain('className="flow-directional-icon"');
  });

  it('anchors popovers that hang off a trigger to the inline start edge', () => {
    const comment = source('shared/components/CommentInput.tsx');
    const duedate = source('shared/components/DueDatePicker.tsx');
    const projectSelect = source('shared/components/ProjectSelect.tsx');

    expect(comment).toContain('absolute start-2 bottom-full');
    expect(duedate).toContain('absolute top-full start-0');
    expect(projectSelect).toContain('absolute top-full start-0');

    for (const file of [comment, duedate, projectSelect]) {
      expect(file).not.toContain('left-0');
      expect(file).not.toContain('left-2');
      expect(file).not.toContain('text-left');
    }
  });

  it('anchors the board column menu to the inline end edge', () => {
    const board = source('features/projects/ProjectBoard.tsx');

    expect(board).toContain('absolute end-0 top-full');

    expect(board).not.toContain('right-0');
    expect(board).not.toContain('text-left');

    // Was `expect(board).toContain('me-1')`, which pinned one utility on one
    // element. That element is gone: the empty-column button it sat in became a
    // sentence, and the add control beside it spaces with `gap`, which has no
    // direction to get wrong. Asserting the presence of a specific logical
    // utility fails when the code stops needing it, which is the wrong way
    // round.
    //
    // The property is that no *physical* inline spacing appears at all, which is
    // stronger than the old pair and cannot be satisfied by accident.
    for (const physical of ['mr-', 'ml-', 'pr-', 'pl-']) {
      expect(board, `physical spacing utility ${physical}`).not.toContain(physical);
    }
  });

  it('keeps the task detail meta rail and its popover logical', () => {
    const detail = source('features/tasks/TaskDetail.tsx');

    expect(detail).toContain('lg:border-s lg:border-t-0');
    expect(detail).toContain('absolute top-6 start-0');
    expect(detail).toContain('ms-1.5');
    expect(detail).toContain('ms-2');

    expect(detail).not.toContain('lg:border-l');
    expect(detail).not.toContain('left-0');
    expect(detail).not.toContain('text-left');
    expect(detail).not.toContain('ml-1.5');
    expect(detail).not.toContain('ml-2');
  });
});
