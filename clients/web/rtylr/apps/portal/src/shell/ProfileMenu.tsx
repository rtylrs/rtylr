import { useEffect, useRef } from 'react';
import { useTranslation } from 'react-i18next';
import { Sun, Moon, Monitor, LogOut, ExternalLink } from 'lucide-react';
import { signOut } from '@voxire/auth';
import { APP_URLS, getAvatarUrl } from '@voxire/config';
import { Avatar } from '../shared/components/ui';
import { useUserStore } from '../shared/stores/userStore';
import { useBusinessStore } from '../shared/stores/businessStore';
import { useConfigStore } from '../shared/stores/configStore';
import { useThemeStore } from '../shared/stores/themeStore';
import { selectLanguage } from '../shared/services/i18next';
import BusinessSwitcher from '../features/business-switcher/BusinessSwitcher';

export default function ProfileMenu() {
  const { user, logout } = useUserStore();
  const { theme, setTheme } = useThemeStore();
  const { t, i18n } = useTranslation();
  const ref = useRef<HTMLDetailsElement>(null);
  const close = () => {
    if (ref.current) ref.current.open = false;
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
    <details className="portal-profile" ref={ref}>
      <summary
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
      <div className="portal-profile-menu">
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
        <label className="portal-language">
          {t('common.nav.language')}
          <select
            value={i18n.language.split('-')[0]}
            onChange={(event) => selectLanguage(event.target.value)}
          >
            <option value="en">English</option>
            <option value="fr">Français</option>
            <option value="ar">العربية</option>
          </select>
        </label>
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
