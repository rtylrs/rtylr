import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

vi.mock('@voxire/components', () => ({ HelmetContainer: () => null }));
vi.mock('../../shared/hooks/queries/useCrmSettings', () => ({
  useCrmSettings: () => ({ data: undefined }),
  useUpdateCrmSettings: () => ({ mutate: vi.fn(), isPending: false }),
}));

import SettingsScreen from './Settings';

describe('CRM team settings', () => {
  it('routes member, group, license, and permission changes to Dash', () => {
    render(<SettingsScreen />);
    fireEvent.click(screen.getByRole('button', { name: 'crm.settings.tabTeam' }));

    expect(screen.getByRole('link', { name: 'crm.settings.teamManagementAction' })).toHaveAttribute(
      'href',
      expect.stringMatching(/\/team$/)
    );
  });
});
