import path from 'node:path';
import { cleanup, render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, beforeAll, describe, expect, it, vi } from 'vitest';
import { initTestI18nFrom } from '../../../test/i18nTestInstance';
import PortalPageCard from './PortalPageCard';

const access = vi.hoisted(() => ({ owner: true }));
vi.mock('../../shared/stores/userStore', async () => {
  const { createEffectiveAccess } = await import('../../../test/effectiveAccess');
  return {
    useUserStore: (selector: (state: unknown) => unknown) =>
      selector({
        user: {
          id: 1,
          effectiveAccess: createEffectiveAccess(undefined, {
            appIds: ['crm'],
            isOwner: access.owner,
          }),
        },
      }),
  };
});
vi.mock('@tanstack/react-router', () => ({
  useRouter: () => ({
    getMatchedRoutes: (pathname: string) => ({
      foundRoute: { path: ['/team', '/crm/customers'].includes(pathname) ? pathname : '*' },
    }),
  }),
}));

const Icon = () => null;
const rail = {
  top: [
    { id: 'caret', icon: Icon },
    { id: 'home', icon: Icon },
  ],
  bottom: [],
};

beforeAll(() =>
  initTestI18nFrom(path.resolve(import.meta.dirname, '../../locales'), 'en', [
    path.resolve(import.meta.dirname, '../../modules/crm/locales'),
  ])
);
afterEach(() => {
  cleanup();
  access.owner = true;
});

describe('PortalPageCard', () => {
  it('shows where the page is and opens it with the button to press', async () => {
    const onOpen = vi.fn();
    render(
      <PortalPageCard spec={{ path: '/crm/customers', action: 'Add customer' }} rail={rail} onOpen={onOpen} />
    );
    const card = screen.getByRole('button', { name: 'Open Customers in CRM' });
    expect(card).toHaveTextContent('CRM · press Add customer');
    // The sketch marks the page in CRM's own sidebar.
    expect(card.querySelector('.ppc-nav.is-active')).toHaveTextContent('Customers');
    await userEvent.click(card);
    expect(onOpen).toHaveBeenCalledWith({ path: '/crm/customers', action: 'Add customer' });
  });

  it('cannot open a page the person has no access to', () => {
    access.owner = false;
    render(<PortalPageCard spec={{ path: '/team' }} rail={rail} onOpen={vi.fn()} />);
    expect(screen.getByRole('button', { name: 'Team: No access' })).toBeDisabled();
  });

  it('says so when the page is not in the Portal', () => {
    render(
      <PortalPageCard spec={{ path: '/made-up', title: 'Made up' }} rail={rail} onOpen={vi.fn()} />
    );
    expect(screen.getByRole('button', { name: 'Made up: Not in rtylr' })).toBeDisabled();
  });
});
