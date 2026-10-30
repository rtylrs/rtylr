import { useEffect, useId, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Sun, Moon, Monitor, LogOut, ExternalLink, Check, ChevronDown } from 'lucide-react';
import { signOut } from '@voxire/auth';
import { APP_URLS, getAvatarUrl } from '@voxire/config';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@voxire/ui';
import { Avatar } from '../shared/components/ui';
import { useUserStore } from '../shared/stores/userStore';
import { useBusinessStore } from '../shared/stores/businessStore';
import { useConfigStore } from '../shared/stores/configStore';
import { useThemeStore } from '../shared/stores/themeStore';
import { selectLanguage } from '../shared/services/i18next';
import BusinessSwitcher from '../features/business-switcher/BusinessSwitcher';

const languages = [
  { code: 'en', label: 'English' },
  { code: 'fr', label: 'Français' },
  { code: 'ar', label: 'العربية' },
] as const;

/**
 * The avatar menu. In the top of the app bar it opens downwards, unless the
 * person moved it so low there's no room below.
 */
export default function ProfileMenu({ opensDown = false }: { opensDown?: boolean }) {
  const { user, logout } = useUserStore();
  const { theme, setTheme } = useThemeStore();
  const { t, i18n } = useTranslation();
  const languageLabelId = useId();
  const language = i18n.language.split('-')[0];
  const selectedLanguage = languages.find(({ code }) => code === language) ?? languages[0];
  const ref = useRef<HTMLDetailsElement>(null);
  const close = () => {
    if (ref.current) ref.current.open = false;
  };
  const [placement, setPlacement] = useState<{ down: boolean; room?: number }>({
    down: opensDown,
  });
  // Open towards the side with room for the whole menu, so Sign out is never
  // cut off. On a window too short for either side, take the roomier one and
  // scroll inside the menu.
  const place = () => {
    const details = ref.current;
    const menu = details?.querySelector<HTMLElement>('.portal-profile-menu');
    if (!details?.open || !menu) return;
    const { top, bottom } = details.getBoundingClientRect();
    const below = window.innerHeight - top;
    const height = menu.scrollHeight;
    const fitsBelow = below >= height;
    const fitsAbove = bottom >= height;
    const down =
      fitsBelow || fitsAbove
        ? opensDown
          ? fitsBelow || !fitsAbove
          : fitsBelow && !fitsAbove
        : below > bottom;
    const room = down ? below : bottom;
    setPlacement({ down, room: height > room ? room - 8 : undefined });
  };
  useEffect(() => {
    const outside = (event: MouseEvent) => {
      if (event.target instanceof Element && event.target.closest('[data-portal-profile-menu]'))
        return;
      if (!ref.current?.contains(event.target as Node)) close();
    };
    const escape = (event: KeyboardEvent) => {
      if (event.defaultPrevented) return;
      if (event.key === 'Escape' && ref.current?.open) {
        close();
        ref.current.querySelector('summary')?.focus();
      }
    };
    document.addEventListener('mousedown', outside);
    document.addEventListener('keydown', escape);
    return () => {
      document.removeEventListener('mousedown', outside);
      document.removeEventListener('keydown', escape);
    };
  }, []);
  return (
    <details
      className={`portal-profile${placement.down ? ' opens-down' : ''}`}
      ref={ref}
      onToggle={place}
    >
      <summary
        role="button"
        className="portal-rail-button"
        aria-label={t('common.userMenu')}
        title={t('common.userMenu')}
      >
        <Avatar
          src={getAvatarUrl(user?.email || user?.fullName)}
          alt=""
          initials={user?.fullName?.slice(0, 2)}
          size="sm"
        />
      </summary>
      <div
        className="portal-profile-menu"
        style={
          placement.room === undefined
            ? undefined
            : { maxHeight: placement.room, overflowY: 'auto' }
        }
      >
        <div className="portal-profile-identity">
          <strong>{user?.fullName}</strong>
          <span>{user?.email}</span>
        </div>
        <BusinessSwitcher onNavigate={close} />
        <fieldset>
          <legend>{t('common.appearance.title')}</legend>
          <div className="portal-theme-options">
            {(
              [
                { value: 'light', icon: Sun },
                { value: 'dark', icon: Moon },
                { value: 'system', icon: Monitor },
              ] as const
            ).map(({ value, icon: Icon }) => (
              <button
                key={value}
                type="button"
                aria-pressed={theme === value}
                onClick={() => setTheme(value)}
              >
                <Icon size={16} />
                {t(`common.appearance.${value}`)}
              </button>
            ))}
          </div>
        </fieldset>
        <div className="portal-language">
          <span id={languageLabelId}>{t('common.nav.language')}</span>
          <DropdownMenu modal={false} dir={i18n.dir()}>
            <DropdownMenuTrigger asChild>
              <button
                type="button"
                className="portal-language-trigger"
                aria-labelledby={`${languageLabelId} ${languageLabelId}-value`}
              >
                <span id={`${languageLabelId}-value`} lang={selectedLanguage.code} dir="auto">
                  {selectedLanguage.label}
                </span>
                <ChevronDown size={14} aria-hidden="true" />
              </button>
            </DropdownMenuTrigger>
            <DropdownMenuContent
              data-portal-profile-menu
              aria-labelledby={languageLabelId}
              align="end"
              side="top"
              sideOffset={6}
              collisionPadding={12}
              className="portal-language-menu z-[70] min-w-[148px]"
            >
              {languages.map(({ code, label }) => (
                <DropdownMenuItem
                  key={code}
                  role="menuitemradio"
                  aria-checked={language === code}
                  className="portal-language-option"
                  onSelect={() => selectLanguage(code)}
                >
                  <span lang={code} dir="auto">
                    {label}
                  </span>
                  {language === code && <Check size={14} aria-hidden="true" />}
                </DropdownMenuItem>
              ))}
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
        {user?.isConsoleAdmin && (
          <a href={APP_URLS.console}>
            <ExternalLink size={16} />
            {t('common.openConsole')}
          </a>
        )}
        <button
          className="portal-signout"
          type="button"
          onClick={() => {
            useBusinessStore.getState().reset();
            useConfigStore.getState().reset();
            logout();
            signOut('dash');
          }}
        >
          <LogOut size={16} />
          {t('common.signOut')}
        </button>
      </div>
    </details>
  );
}
