import { Link } from '@tanstack/react-router';
import { ArrowRight } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { HelmetContainer } from '@voxire/components';
import { useUserStore } from '../../shared/stores/userStore';
import { useAddons, isAddonEnabled } from '../../shared/hooks/queries/useAddons';
import { canAccessPortalPath } from '../../shared/utils/canAccessPortalPath';
import { addonSettingsItems, settingsNavItems } from '../../shell/navigation';

export default function SettingsOverview() {
  const { t } = useTranslation();
  const user = useUserStore((state) => state.user);
  const { data: addons = [] } = useAddons();

  const allItems = [
    ...settingsNavItems,
    ...addonSettingsItems.filter((item) => isAddonEnabled(addons, item.addonId)),
  ];

  return (
    <div className="portal-settings-overview">
      <HelmetContainer title={t('settingsCenter.title')} />
      <h1>{t('settingsCenter.title')}</h1>
      <p className="portal-settings-description">{t('settingsCenter.description')}</p>
      {(['account', 'business', 'apps'] as const).map((section) => {
        const items = allItems.filter(
          (item) => item.section === section && canAccessPortalPath(user, item.path)
        );
        if (!items.length) return null;
        return (
          <section key={section} className="portal-settings-category">
            <div>
              <h2>{t(`settingsCenter.groups.${section}.title`)}</h2>
              <p>{t(`settingsCenter.groups.${section}.description`)}</p>
            </div>
            <div className="portal-settings-links">
              {items.map(({ path, key, icon: Icon }) => (
                <Link key={path} to={path} className="portal-settings-card">
                  <Icon size={19} aria-hidden />
                  <span>{t(`common.nav.${key}`)}</span>
                  <ArrowRight size={16} className="rtl:rotate-180" aria-hidden />
                </Link>
              ))}
            </div>
          </section>
        );
      })}
    </div>
  );
}
