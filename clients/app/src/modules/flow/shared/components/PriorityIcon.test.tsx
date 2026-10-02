import fs from 'node:fs';
import path from 'node:path';
import React from 'react';
import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { Task } from '../stores/flowStore';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

import PriorityIcon from '../../shared/components/PriorityIcon';

const PRIORITIES: Task['priority'][] = ['urgent', 'high', 'medium', 'low', 'no_priority'];

describe('PriorityIcon', () => {
  it('draws a different shape for every priority', () => {
    // WCAG 1.4.1. The project list used one 6px circle for all five, so the
    // tint was the only signal and the board and list disagreed on encoding.
    const shapes = PRIORITIES.map((priority) => {
      const { container, unmount } = render(<PriorityIcon priority={priority} />);
      const svg = container.querySelector('svg');
      const paths = [...(svg?.querySelectorAll('path, line, polyline, circle') ?? [])]
        .map((node) => node.outerHTML)
        .join('');
      unmount();
      return paths;
    });

    expect(new Set(shapes).size).toBe(PRIORITIES.length);
    for (const shape of shapes) expect(shape.length).toBeGreaterThan(0);
  });

  it.each(PRIORITIES)('exposes %s to assistive technology, not only as a tint', (priority) => {
    render(<PriorityIcon priority={priority} />);
    const glyph = screen.getByRole('img');
    expect(glyph).toHaveAttribute('aria-label');
    expect(glyph.getAttribute('aria-label')).not.toBe('');
  });

  it('translates the label rather than hardcoding English', () => {
    // PRIORITY_LABELS in the store is an English map, and it was what this
    // component read, so the tooltip said "No priority" in Arabic and French.
    render(<PriorityIcon priority="no_priority" />);
    expect(screen.getByRole('img')).toHaveAttribute('aria-label', 'flowPriority.noPriority');
  });

  it('uses the irregular key for no_priority', () => {
    // The locale files call it noPriority, so interpolating the raw value
    // would not resolve.
    render(<PriorityIcon priority="no_priority" />);
    expect(screen.getByRole('img').getAttribute('aria-label')).not.toContain('no_priority');
  });

  it('drops the img role when it renders a visible label instead', () => {
    render(<PriorityIcon priority="urgent" showLabel />);
    expect(screen.queryByRole('img')).not.toBeInTheDocument();
    expect(screen.getByText('flowPriority.urgent')).toBeInTheDocument();
  });
});

describe('the project list uses the shared glyph', () => {
  // The same wiring hole a probe found on the 404 screen: a component can be
  // correct and tested while no screen renders it, leaving the reported
  // defect in place. The list's own suite mocks the row, so this is a source
  // assertion.
  const list = fs.readFileSync(
    path.resolve(__dirname, '../../features/projects/ProjectList.tsx'),
    'utf8'
  );

  it('renders PriorityIcon rather than a bare dot', () => {
    expect(list).toContain('<PriorityIcon priority={task.priority}');
    expect(list).toContain("import PriorityIcon from '../../shared/components/PriorityIcon'");
  });

  it('keeps no colour-only priority map of its own', () => {
    // PRIORITY_ICONS held a colour per priority and nothing else, which is
    // what made the tint the only signal.
    expect(list).not.toContain('PRIORITY_ICONS');
  });

  it('has no colour-only priority swatch left', () => {
    // Scoped to a dot painted from the task's priority. The status group
    // header keeps its dot, and should: it sits directly beside
    // {group.label}, so the colour is not carrying the meaning alone.
    const priorityDot = list.match(
      /rounded-full[^>]*background:\s*[^}]*task\.priority|task\.priority[^>]*rounded-full/
    );
    expect(priorityDot).toBeNull();
  });
});
