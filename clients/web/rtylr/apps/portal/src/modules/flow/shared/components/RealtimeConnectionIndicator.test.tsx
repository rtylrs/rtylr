import React from 'react';
import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import RealtimeConnectionIndicator from './RealtimeConnectionIndicator';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
}));

describe('RealtimeConnectionIndicator', () => {
  it('renders nothing while connecting, so a normal boot never flashes it', () => {
    const { container } = render(<RealtimeConnectionIndicator status="connecting" />);
    expect(container).toBeEmptyDOMElement();
  });

  it('renders nothing while connected, so the healthy path stays silent', () => {
    const { container } = render(<RealtimeConnectionIndicator status="connected" />);
    expect(container).toBeEmptyDOMElement();
  });

  it('announces reconnecting as a status region', () => {
    render(<RealtimeConnectionIndicator status="reconnecting" />);

    const status = screen.getByRole('status');
    expect(status).toHaveTextContent('flow.shell.realtimeReconnecting');
    expect(status).toHaveAttribute('aria-live', 'polite');
  });

  it('announces a fully dropped connection with its own copy', () => {
    render(<RealtimeConnectionIndicator status="disconnected" />);

    expect(screen.getByRole('status')).toHaveTextContent('flow.shell.realtimeDisconnected');
  });

  it('falls back to an aria-label in compact mode, since there is no visible text', () => {
    render(<RealtimeConnectionIndicator status="reconnecting" compact />);

    expect(
      screen.getByRole('status', { name: 'flow.shell.realtimeReconnecting' })
    ).toBeInTheDocument();
  });
});
