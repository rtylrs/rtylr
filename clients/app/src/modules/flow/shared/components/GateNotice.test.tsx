import React from 'react';
import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import GateNotice from './GateNotice';

describe('GateNotice', () => {
  it('states the reason a control is unavailable', () => {
    render(<GateNotice reason="Logging for someone else needs manage team time." />);

    expect(
      screen.getByText('Logging for someone else needs manage team time.')
    ).toBeInTheDocument();
  });

  it('carries an id so the control it explains can describe itself with it', () => {
    render(
      <>
        <button type="button" disabled aria-describedby="gate-log-for">
          Log for Maya
        </button>
        <GateNotice id="gate-log-for" reason="You can only log your own time." />
      </>
    );

    const control = screen.getByRole('button', { name: 'Log for Maya' });
    const describedBy = control.getAttribute('aria-describedby');

    expect(describedBy).toBe('gate-log-for');
    expect(document.getElementById(describedBy as string)).toHaveTextContent(
      'You can only log your own time.'
    );
  });

  it('hides its icon from the accessibility tree, because it repeats the reason', () => {
    const { container } = render(<GateNotice reason="Needs manage project." />);

    const icon = container.querySelector('svg');
    expect(icon).toHaveAttribute('aria-hidden', 'true');
  });

  it('renders nothing when there is no reason, so callers need no conditional', () => {
    const { container } = render(<GateNotice reason="" />);

    expect(container).toBeEmptyDOMElement();
  });
});
