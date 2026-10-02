import { render, screen, fireEvent } from '@testing-library/react';
import { beforeEach, describe, it, expect } from 'vitest';
import { createEffectiveAccess } from '../../../../test/effectiveAccess';
import { useUserStore } from '../../../shared/stores/userStore';
import { useFlowStore } from '../shared/stores/flowStore';
import { initTestI18n } from '../shared/testing/i18n';
import { getFlowHomePath } from '../navigation';
import { canAccessPortalPath } from '../../../shared/utils/canAccessPortalPath';
import Start from './Start';

beforeEach(async () => {
  await initTestI18n();
  useFlowStore.setState({ quickCreateOpen: false, projectCreateOpen: false });
});
describe('Flow access without read permission', () => {
  it('lets a write-only member create tasks without exposing task lists', () => {
    const user = {
      effectiveAccess: createEffectiveAccess('flow', { permissions: ['flow:tasks:create'] }),
    };
    useUserStore.setState({ user });
    expect(getFlowHomePath(user)).toBe('/flow/start');
    expect(canAccessPortalPath(user, '/flow/my-issues')).toBe(false);
    render(<Start />);
    fireEvent.click(screen.getByRole('button', { name: 'New task', exact: true }));
    expect(useFlowStore.getState().quickCreateOpen).toBe(true);
    expect(screen.queryByRole('button', { name: 'New project', exact: true })).toBeNull();
  });
  it('offers project creation independently from tasks', () => {
    useUserStore.setState({
      user: {
        effectiveAccess: createEffectiveAccess('flow', { permissions: ['flow:projects:create'] }),
      },
    });
    render(<Start />);
    fireEvent.click(screen.getByRole('button', { name: 'New project', exact: true }));
    expect(useFlowStore.getState().projectCreateOpen).toBe(true);
    expect(screen.queryByRole('button', { name: 'New task', exact: true })).toBeNull();
  });
  it('explains missing grants without a fake action', () => {
    useUserStore.setState({ user: { effectiveAccess: createEffectiveAccess('flow') } });
    render(<Start />);
    expect(screen.getByText('No actions assigned')).toBeVisible();
    expect(screen.queryByRole('button')).toBeNull();
  });
});
