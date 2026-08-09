import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import EmptyState from './EmptyState';

describe('EmptyState', () => {
  it('keeps the starting canvas visually quiet while retaining an accessible label', () => {
    const { container } = render(
      <EmptyState label="No messages yet. Ask a question about this page." />
    );

    expect(screen.getByText('No messages yet. Ask a question about this page.')).toBeTruthy();
    expect(container.querySelectorAll('button')).toHaveLength(0);
    expect(container.querySelector('.caret-empty-state')).toBeTruthy();
  });
});
