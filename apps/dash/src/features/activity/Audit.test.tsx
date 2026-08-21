import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

const auditData = vi.hoisted(() => ({
  data: { entries: [] as { id: string }[], total: 0 } as
    | { entries: { id: string }[]; total: number }
    | undefined,
  isLoading: false,
  isFetching: false,
  refetch: vi.fn(),
}));

vi.mock('../../shared/hooks/queries/useActivity', () => ({
  useAuditLog: () => auditData,
}));

import Audit from './Audit';

beforeEach(() => {
  vi.clearAllMocks();
  auditData.data = { entries: [], total: 0 };
  auditData.isLoading = false;
  auditData.isFetching = false;
});

describe('Audit log empty state', () => {
  it('shows the illustration, title and a refresh call-to-action when there are no entries', () => {
    render(<Audit />);

    expect(screen.getByText('activity.audit.emptyTitle')).toBeInTheDocument();
    expect(screen.getByText('activity.audit.emptyDescription')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'common.refresh' })).toBeInTheDocument();
  });

  it('refetches the audit log when the refresh call-to-action is clicked', () => {
    render(<Audit />);

    fireEvent.click(screen.getByRole('button', { name: 'common.refresh' }));

    expect(auditData.refetch).toHaveBeenCalledTimes(1);
  });

  it('does not show the empty state when there are entries', () => {
    auditData.data = {
      entries: [
        {
          id: '1',
          actor: 'jane@rtylr.com',
          action: 'login',
          target: 'session',
          app: 'dash',
          createdAt: new Date().toISOString(),
        } as unknown as { id: string },
      ],
      total: 1,
    };
    render(<Audit />);

    expect(screen.queryByText('activity.audit.emptyTitle')).not.toBeInTheDocument();
  });
});
