import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import ThinkingBlock from './ThinkingBlock';

describe('ThinkingBlock', () => {
  it('shows the live orb while the model is still thinking', () => {
    render(<ThinkingBlock content="" isLive />);

    expect(screen.getByText('Thinking…')).not.toBeNull();
    expect(screen.getByRole('img', { name: 'Thinking…' })).not.toBeNull();
  });

  it('shows the finished thought and its duration', () => {
    render(<ThinkingBlock content="Checked stock levels" durationMs={2400} defaultOpen />);

    expect(screen.getByText('Thought process')).not.toBeNull();
    expect(screen.getByText('2.4s')).not.toBeNull();
    expect(screen.getByText('Checked stock levels')).not.toBeNull();
  });
});
