import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('@voxire/api', () => ({
  unwrap: (value: unknown) => value,
  getOneKitErrorMessage: (_error: unknown, fallback: string) => fallback,
}));

vi.mock('@tanstack/react-query', () => ({
  useMutation: () => ({ mutate: vi.fn(), isPending: false }),
}));

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

const updateBusinessState = vi.hoisted(() => ({
  mutate: vi.fn(),
  isPending: false,
}));

vi.mock('../../../../../shared/hooks/queries/useBusiness', () => ({
  useUpdateBusiness: () => updateBusinessState,
}));

vi.mock('../../../../../shared/lib/clients', () => ({
  rest: { post: vi.fn() },
}));

import Email from './Email';

beforeEach(() => {
  vi.clearAllMocks();
});

describe('Email: notification preferences', () => {
  it('disables the switches and says the preference is not saved yet', () => {
    render(<Email />);

    expect(screen.getByText('settings.email.notificationsUnavailable')).toBeInTheDocument();
    expect(screen.getAllByRole('switch')).toHaveLength(5);
    for (const id of ['notif-new-order', 'notif-status', 'notif-weekly', 'notif-feedback']) {
      expect(document.getElementById(id)).toBeDisabled();
    }
    // The SMTP toggle only reveals the form; it is unrelated to this gap.
    expect(document.getElementById('smtp-toggle')).not.toBeDisabled();
  });

  it('does not fire a mutation that cannot carry the preference', () => {
    render(<Email />);

    // Pre-fix the switch fired updateBusiness.mutate({notificationPreferences}),
    // which useUpdateBusiness dropped before the request.
    fireEvent.click(document.getElementById('notif-weekly')!);

    expect(updateBusinessState.mutate).not.toHaveBeenCalled();
  });
});

describe('Email: sender identity', () => {
  it('still saves the sender fields through the config contract', () => {
    render(<Email />);

    fireEvent.change(screen.getByLabelText('settings.email.fromEmail'), {
      target: { value: 'orders@example.com' },
    });
    fireEvent.click(screen.getByRole('button', { name: 'Save changes' }));

    expect(updateBusinessState.mutate).toHaveBeenCalledWith({
      data: { senderName: '', senderEmail: 'orders@example.com', replyTo: '' },
    });
  });
});
