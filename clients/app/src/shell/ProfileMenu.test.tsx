import { act, fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import ProfileMenu from './ProfileMenu';
import { useSettingsModalStore } from '../features/settings/settingsModalStore';

vi.mock('../features/business-switcher/BusinessSwitcher', () => ({ default: () => null }));
vi.mock('../shared/avatar/AvatarPicker', () => ({
  initialsOf: (name?: string) => (name ?? '?').slice(0, 2).toUpperCase(),
  AvatarPicker: ({ open }: { open: boolean }) =>
    open ? <div role="dialog" aria-label="avatar picker" /> : null,
}));
vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key, i18n: { language: 'en', dir: () => 'ltr' } }),
}));

// Opens the menu with the avatar at `top` and the menu `height` px tall.
function open(opensDown: boolean, top: number, height: number) {
  const { container } = render(<ProfileMenu opensDown={opensDown} />);
  const details = container.querySelector('details')!;
  const menu = container.querySelector<HTMLElement>('.portal-profile-menu')!;
  vi.spyOn(details, 'getBoundingClientRect').mockReturnValue({
    top,
    bottom: top + 40,
  } as DOMRect);
  Object.defineProperty(menu, 'scrollHeight', { value: height });
  details.open = true;
  fireEvent(details, new Event('toggle'));
  return details;
}

describe('ProfileMenu', () => {
  it('opens downwards at the top of the app bar', () => {
    expect(open(true, 20, 400)).toHaveClass('opens-down');
  });

  it('opens upwards when moved too low for the menu to fit below', () => {
    expect(open(true, window.innerHeight - 60, 400)).not.toHaveClass('opens-down');
  });

  it('scrolls inside the roomier side when neither side fits', () => {
    const details = open(true, 20, window.innerHeight + 100);
    expect(details).toHaveClass('opens-down');
    expect(details.querySelector('.portal-profile-menu')).toHaveStyle({
      maxHeight: `${window.innerHeight - 28}px`,
      overflowY: 'auto',
    });
  });

  it('opens upwards from the bottom of the app bar', () => {
    expect(open(false, window.innerHeight - 60, 400)).not.toHaveClass('opens-down');
  });
});

describe('ProfileMenu profile access', () => {
  beforeEach(() => {
    useSettingsModalStore.getState().close();
  });

  // Opening <details> fires its toggle event asynchronously; let it settle.
  async function openMenu() {
    const view = render(<ProfileMenu />);
    const details = view.container.querySelector('details')!;
    await act(async () => {
      details.open = true;
      await new Promise((resolve) => setTimeout(resolve, 0));
    });
    return details;
  }

  it('opens Account settings from the name and email', async () => {
    const details = await openMenu();
    const link = details.querySelector<HTMLAnchorElement>('.portal-profile-identity-link')!;
    expect(link).toHaveAttribute('href', '/settings/account');
    fireEvent.click(link);
    expect(useSettingsModalStore.getState().path).toBe('/settings/account');
    expect(details.open).toBe(false);
  });

  it('opens Account settings from the Your profile row', async () => {
    const details = await openMenu();
    fireEvent.click(screen.getByRole('link', { name: 'common.yourProfile' }));
    expect(useSettingsModalStore.getState().path).toBe('/settings/account');
    expect(details.open).toBe(false);
  });

  it('leaves a modified click to the browser, so it can open a new tab', async () => {
    await openMenu();
    fireEvent.click(screen.getByRole('link', { name: 'common.yourProfile' }), { metaKey: true });
    expect(useSettingsModalStore.getState().path).toBeNull();
  });

  it('opens the avatar picker from your avatar', async () => {
    const details = await openMenu();
    expect(screen.queryByRole('dialog', { name: 'avatar picker' })).toBeNull();
    fireEvent.click(screen.getByRole('button', { name: 'common.changeYourAvatar' }));
    expect(screen.getByRole('dialog', { name: 'avatar picker' })).toBeInTheDocument();
    expect(details.open).toBe(false);
  });
});
