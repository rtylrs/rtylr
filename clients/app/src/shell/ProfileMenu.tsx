import { useEffect, useId, useRef, useState, type MouseEvent as ReactMouseEvent } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Sun,
  Moon,
  Monitor,
  LogOut,
  ExternalLink,
  Check,
  ChevronDown,
  ChevronRight,
  Pencil,
  UserRound,
} from 'lucide-react';
import { signOut } from '@voxire/auth';
import { APP_URLS } from '@voxire/config';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@voxire/ui';
import { Avatar } from '../shared/components/ui';
import { useUserStore } from '../shared/stores/userStore';
import { AvatarPicker, initialsOf } from '../shared/avatar/AvatarPicker';
import { useSettingsModalStore } from '../features/settings/settingsModalStore';
import { useBusinessStore } from '../shared/stores/businessStore';
import { useConfigStore } from '../shared/stores/configStore';
import { useThemeStore } from '../shared/stores/themeStore';
import { selectLanguage } from '../shared/services/i18next';
import BusinessSwitcher from '../features/business-switcher/BusinessSwitcher';

/** Account settings: avatar, name and email. The portal's profile page. */
const PROFILE_PATH = '/settings/account';

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
  const [pickerOpen, setPickerOpen] = useState(false);
  // A real link, so a middle click still opens it in a new tab; a plain click
  // opens the settings modal over the current page like every settings link.
  const openProfile = (event: ReactMouseEvent<HTMLAnchorElement>) => {
    if (event.button !== 0 || event.metaKey || event.ctrlKey || event.shiftKey || event.altKey)
      return;
    event.preventDefault();
    close();
    useSettingsModalStore.getState().open(PROFILE_PATH);
  };
  const changeAvatar = () => {
    close();
    setPickerOpen(true);
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
      <summary role="button" className="portal-rail-button" aria-label={t('common.userMenu')}>
        <Avatar
          seed={user?.id !== undefined ? String(user.id) : undefined}
          initials={initialsOf(user?.fullName)}
          name={user?.fullName || undefined}
          size="sm"
          aria-hidden
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
          <button
            type="button"
            className="portal-profile-avatar"
            onClick={changeAvatar}
            aria-label={t('common.changeYourAvatar')}
          >
            <Avatar
              seed={user?.id !== undefined ? String(user.id) : undefined}
              initials={initialsOf(user?.fullName)}
              size="md"
              aria-hidden
            />
            <span className="portal-profile-avatar-edit" aria-hidden="true">
              <Pencil size={10} />
            </span>
          </button>
          <a href={PROFILE_PATH} className="portal-profile-identity-link" onClick={openProfile}>
            <strong>{user?.fullName}</strong>
            <span>{user?.email}</span>
            <ChevronRight
              className="portal-profile-identity-chevron"
              size={14}
              aria-hidden="true"
            />
          </a>
        </div>
        <a href={PROFILE_PATH} className="portal-profile-link" onClick={openProfile}>
          <UserRound size={16} />
          {t('common.yourProfile')}
        </a>
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
      {pickerOpen && (
        <AvatarPicker
          open={pickerOpen}
          onOpenChange={setPickerOpen}
          userId={String(user?.id ?? '')}
          name={user?.fullName}
          current={user?.avatar}
        />
      )}
    </details>
  );
}
