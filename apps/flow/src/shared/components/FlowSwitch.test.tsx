import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import FlowSwitch from './FlowSwitch';

describe('FlowSwitch', () => {
  it('paints the track from Flow tokens, because bg-accent resolves to nothing here', () => {
    render(<FlowSwitch checked aria-label="Billable" onCheckedChange={vi.fn()} />);

    const cls = screen.getByRole('switch').className;
    expect(cls).toContain('data-[state=checked]:bg-[var(--accent)]');
    expect(cls).toContain('data-[state=unchecked]:bg-[var(--flow-active)]');
  });

  it('keeps a visible border so the off state is not an invisible box', () => {
    render(<FlowSwitch checked={false} aria-label="Billable" onCheckedChange={vi.fn()} />);

    expect(screen.getByRole('switch').className).toContain('border-[var(--flow-border)]');
  });

  it('forwards the label, checked state and change handler', () => {
    const onCheckedChange = vi.fn();
    render(<FlowSwitch checked={false} aria-label="Billable" onCheckedChange={onCheckedChange} />);

    const control = screen.getByRole('switch', { name: 'Billable' });
    expect(control).toHaveAttribute('data-state', 'unchecked');

    fireEvent.click(control);
    expect(onCheckedChange).toHaveBeenCalledWith(true);
  });

  it('stays disabled when asked, and reports it', () => {
    const onCheckedChange = vi.fn();
    render(
      <FlowSwitch
        checked={false}
        disabled
        aria-label="Billable"
        onCheckedChange={onCheckedChange}
      />
    );

    const control = screen.getByRole('switch', { name: 'Billable' });
    expect(control).toBeDisabled();
    fireEvent.click(control);
    expect(onCheckedChange).not.toHaveBeenCalled();
  });

  it('keeps any caller className rather than replacing it', () => {
    render(
      <FlowSwitch checked aria-label="Billable" className="shrink-0" onCheckedChange={vi.fn()} />
    );

    expect(screen.getByRole('switch').className).toContain('shrink-0');
  });
});
