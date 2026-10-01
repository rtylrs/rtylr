import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

const permissionsState = vi.hoisted(() => ({
  data: undefined as
    | {
        permissions: {
          key: string;
          appId: string;
          resource: string;
          action: string;
          isSensitive: boolean;
          isHighRisk: boolean;
        }[];
      }
    | undefined,
  isLoading: false,
  isError: false,
  refetch: vi.fn(),
}));

vi.mock('../../shared/hooks/queries/usePermissions', () => ({
  usePermissionCatalog: () => permissionsState,
}));

import Permissions from './Permissions';

beforeEach(() => {
  vi.clearAllMocks();
  permissionsState.data = undefined;
  permissionsState.isLoading = false;
  permissionsState.isError = false;
});

describe('Permissions: failed catalog request', () => {
  it('shows a retry instead of leaving the skeleton on screen', () => {
    permissionsState.isError = true;
    render(<Permissions />);

    // Pre-fix the `isLoading || !data` guard matched a failed query forever,
    // so the skeleton was the only thing that ever rendered.
    expect(screen.getByText('common.permissionCatalog.loadErrorTitle')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'common.refresh' }));
    expect(permissionsState.refetch).toHaveBeenCalledTimes(1);
  });
});

describe('Permissions: catalog render', () => {
  it('renders app resources and actions from the permission catalog', () => {
    permissionsState.data = {
      permissions: [
        {
          key: 'flow:projects:read',
          appId: 'flow',
          resource: 'projects',
          action: 'read',
          isSensitive: false,
          isHighRisk: false,
        },
      ],
    };
    render(<Permissions />);

    expect(screen.getByText('flow')).toBeInTheDocument();
    expect(screen.getByText('projects · read')).toBeInTheDocument();
    expect(screen.getByText('flow:projects:read')).toBeInTheDocument();
  });
});
