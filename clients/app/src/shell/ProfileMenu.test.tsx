import { fireEvent, render } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import ProfileMenu from './ProfileMenu';

vi.mock('../features/business-switcher/BusinessSwitcher', () => ({ default: () => null }));
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
  Object.defineProperty(menu, 'offsetHeight', { value: height });
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

  it('opens upwards from the bottom of the app bar', () => {
    expect(open(false, window.innerHeight - 60, 400)).not.toHaveClass('opens-down');
  });
});
