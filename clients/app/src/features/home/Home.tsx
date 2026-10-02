import { Link } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { Search, Users, Folder, ChevronRight, Settings } from 'lucide-react';
import { useUserStore } from '../../shared/stores/userStore';
import { useCommandPalette } from '../../shared/stores/commandPaletteStore';
import { canAccessPortalPath } from '../../shared/utils/canAccessPortalPath';
import { APP_NAMES } from '../../shared/config/appNames';
import { APP_ICONS, WORKSPACE_APP_IDS } from '../../shared/config/appIcons';
import { insightsNavItems } from '../../modules/insights/navigation';
import './home.css';

const shortcuts = [
  { path: '/team', key: 'team', description: 'teamDescription', icon: Users },
  { path: '/team/groups', key: 'groups', description: 'groupsDescription', icon: Folder },
  { path: '/settings/account', key: 'account', description: 'accountDescription', icon: Settings },
] as const;

export default function Home() {
  const { t } = useTranslation();
  const { user } = useUserStore();
  const setPaletteOpen = useCommandPalette((state) => state.setOpen);
  const visibleApps = WORKSPACE_APP_IDS.filter((id) =>
    user?.effectiveAccess?.apps.some((app) => app.appId === id && app.licensed && app.hasAccess)
  );
  return (
    <div className="portal-home">
      <div className="portal-home-intro">
        <p>{user?.businessName}</p>
        <h1>
          {t('common.portal.welcome', {
            name: user?.fullName?.split(' ')[0] || t('common.portal.home'),
          })}
        </h1>
        <span>{t('common.portal.welcomeDescription')}</span>
      </div>
      <button type="button" className="portal-home-search" onClick={() => setPaletteOpen(true)}>
        <Search size={19} />
        <span>{t('common.portal.search')}</span>
        <kbd>⌘K</kbd>
      </button>
      <section className="portal-home-shortcuts" aria-label={t('common.portal.shortcuts')}>
        {shortcuts
          .filter((shortcut) => canAccessPortalPath(user, shortcut.path))
          .map(({ path, key, description, icon: Icon }) => (
            <Link key={path} to={path}>
              <Icon size={20} />
              <div>
                <strong>{t(`common.nav.${key}`)}</strong>
                <p>{t(`common.portal.${description}`)}</p>
              </div>
              <ChevronRight size={16} className="rtl:rotate-180" />
            </Link>
          ))}
      </section>
      <section className="portal-home-apps">
        <div className="portal-home-section-heading">
          <h2>{t('common.nav.apps')}</h2>
        </div>
        {visibleApps.length ? (
          <div className="portal-home-app-grid">
            {visibleApps.map((id) => {
              const Icon = APP_ICONS[id];
              const path =
                id === 'insights'
                  ? insightsNavItems.find((page) => canAccessPortalPath(user, page.path))?.path ||
                    '/'
                  : `/${id}`;
              return (
                <Link
                  key={id}
                  to={path}
                  aria-label={t('common.portal.openWorkspaceApp', { app: t(APP_NAMES[id].full) })}
                >
                  <Icon size={20} />
                  <span>{t(APP_NAMES[id].short)}</span>
                  <ChevronRight size={14} className="rtl:rotate-180" />
                </Link>
              );
            })}
          </div>
        ) : (
          <p className="portal-home-no-apps">{t('common.portal.noApps')}</p>
        )}
      </section>
    </div>
  );
}
