import { useTranslation } from 'react-i18next';
import { Search } from 'lucide-react';
import { Link, useRouterState } from '@tanstack/react-router';
import { getFlowSettingsSections } from '../modules/flow/settingsNavigation';
import { useUserStore } from '../shared/stores/userStore';
import { PortalNavLink, type PortalNavItem } from './navigation';

export default function SettingsNavigation({
  items,
  query,
  onQueryChange,
  onNavigate,
}: {
  items: PortalNavItem[];
  query: string;
  onQueryChange: (query: string) => void;
  onNavigate: () => void;
}) {
  const { t } = useTranslation();
  const access = useUserStore((state) => state.user?.effectiveAccess);
  const location = useRouterState({ select: (state) => state.location });
  const flowSections = getFlowSettingsSections(access);
  const selectedSection = (location.search as { section?: string })?.section ?? 'notifications';
  const groups = (['account', 'business', 'apps'] as const).map((section) => ({
    section,
    items: items.filter(
      (item) =>
        item.section === section &&
        `${t(`common.nav.${item.key}`)} ${t(`settingsCenter.groups.${section}.title`)} ${item.key === 'flowSettings' ? flowSections.map((entry) => t(entry.labelKey)).join(' ') : ''}`
          .toLocaleLowerCase()
          .includes(query.trim().toLocaleLowerCase())
    ),
  }));
  return (
    <>
      <div className="portal-settings-search">
        <Search size={15} aria-hidden />
        <input
          type="search"
          value={query}
          onChange={(event) => onQueryChange(event.target.value)}
          placeholder={t('settingsCenter.search')}
          aria-label={t('settingsCenter.search')}
        />
      </div>
      {groups.map(
        ({ section, items }) =>
          items.length > 0 && (
            <section key={section} className="portal-settings-nav-group">
              <p className="portal-section-label">{t(`settingsCenter.groups.${section}.title`)}</p>
              {items.map((item) => (
                <div key={item.path}>
                  <PortalNavLink
                    {...item}
                    label={t(`common.nav.${item.key}`)}
                    onNavigate={onNavigate}
                  />
                  {item.key === 'flowSettings' &&
                    (location.pathname === '/settings/flow' || query.trim()) && (
                      <div className="portal-settings-subnav">
                        {flowSections
                          .filter(
                            (entry) =>
                              !query.trim() ||
                              `${t('common.nav.flowSettings')} ${t(entry.labelKey)}`
                                .toLocaleLowerCase()
                                .includes(query.trim().toLocaleLowerCase())
                          )
                          .map((entry) => (
                            <Link
                              key={entry.id}
                              to="/settings/flow"
                              search={{ section: entry.id }}
                              onClick={onNavigate}
                              className="portal-nav-link"
                              aria-current={
                                location.pathname === '/settings/flow' &&
                                selectedSection === entry.id
                                  ? 'page'
                                  : undefined
                              }
                            >
                              {t(entry.labelKey)}
                            </Link>
                          ))}
                      </div>
                    )}
                </div>
              ))}
            </section>
          )
      )}
      {groups.every((group) => group.items.length === 0) && (
        <p role="status" className="portal-section-label">
          {t('settingsCenter.noResults')}
        </p>
      )}
    </>
  );
}
