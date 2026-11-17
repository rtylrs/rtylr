import type { ReactNode } from 'react';
import { Link, useRouterState } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import { ChevronDown, LayoutGrid } from 'lucide-react';
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
} from '@voxire/ui';
import { useUserStore } from '../../shared/stores/userStore';
import { canAccessPortalPath } from '../../shared/utils/canAccessPortalPath';
import { insightsNavItems } from './navigation';
import './insights.css';
import RtylrLogo from '../../shared/components/RtylrLogo';

export default function InsightsHeader({ actions }: { actions: ReactNode }) {
  const { t } = useTranslation();
  const user = useUserStore((state) => state.user);
  const pathname = useRouterState({ select: (state) => state.location.pathname });
  const pages = insightsNavItems.filter((page) => canAccessPortalPath(user, page.path));
  // Nested pages such as /insights/projects/4 take their parent's name; the
  // Overview root only matches itself.
  const currentPage =
    insightsNavItems.find((page) => page.path === pathname) ??
    insightsNavItems.find(
      (page) => page.path !== '/insights' && pathname.startsWith(`${page.path}/`)
    );
  return (
    <header className="insights-header">
      <div className="insights-header-start">
        <Link
          to={pages[0]?.path || '/'}
          className="insights-brand"
          aria-label={t('appNames.short.insights')}
        >
          <RtylrLogo />
          <span> / {t(`insightsNav.${currentPage?.key || 'overview'}`)}</span>
        </Link>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <button type="button" className="insights-page-picker">
              <LayoutGrid size={16} />
              <span>{t('insightsNav.pages')}</span>
              <ChevronDown size={13} />
            </button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="start" className="insights-pages-menu">
            {pages.map(({ path, key, icon: Icon }) => (
              <DropdownMenuItem key={path} asChild>
                <Link to={path} aria-current={pathname === path ? 'page' : undefined}>
                  <Icon size={16} />
                  {t(`insightsNav.${key}`)}
                </Link>
              </DropdownMenuItem>
            ))}
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
      {actions}
    </header>
  );
}
