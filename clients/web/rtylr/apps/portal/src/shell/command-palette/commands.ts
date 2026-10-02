import type { ComponentType } from 'react';
import { Bell, UserPlus } from 'lucide-react';
import { resolveAppUrl } from '@voxire/config';
import { canEffectiveAccess, hasEffectiveAppAccess } from '@voxire/permissions';
import { APP_ICONS, WORKSPACE_APP_IDS } from '../../shared/config/appIcons';
import { APP_NAMES } from '../../shared/config/appNames';
import type { StoredUser } from '../../shared/stores/userStore';
import { canAccessPortalPath } from '../../shared/utils/canAccessPortalPath';
import { insightsNavItems } from '../../modules/insights/navigation';
import { rawNavItems, rawMenuSubItems, settingsNavItems, type PortalNavItem } from '../navigation';

export const COMMAND_GROUPS = ['apps', 'home', 'insights', 'settings', 'actions'] as const;
export type CommandGroup = (typeof COMMAND_GROUPS)[number];
export interface PortalCommand {
  id: string;
  label: string;
  description?: string;
  group: CommandGroup;
  icon: ComponentType<{ size?: number; className?: string }>;
  destination: { path: string } | { href: string };
}

export function getPortalCommands(
  user: StoredUser | null,
  t: (key: string) => string
): PortalCommand[] {
  const commands: PortalCommand[] = [];
  const insightsPages = insightsNavItems.filter((page) => canAccessPortalPath(user, page.path));
  for (const id of WORKSPACE_APP_IDS) {
    if (!hasEffectiveAppAccess(user?.effectiveAccess, id)) continue;
    if (id === 'insights' && !insightsPages.length) continue;
    commands.push({
      id: `app-${id}`,
      label: t(APP_NAMES[id].short),
      description: t(`appNames.desc.${id}`),
      group: 'apps',
      icon: APP_ICONS[id],
      destination:
        id === 'insights' ? { path: insightsPages[0].path } : { href: resolveAppUrl(id) },
    });
  }
  const addPages = (pages: PortalNavItem[], group: CommandGroup, namespace = 'common.nav') => {
    for (const page of pages) {
      if (!canAccessPortalPath(user, page.path)) continue;
      commands.push({
        id: `page-${page.path}`,
        label: t(`${namespace}.${page.key}`),
        group,
        icon: page.icon,
        destination: { path: page.path },
      });
    }
  };
  addPages(
    [
      ...rawNavItems,
      ...rawMenuSubItems,
      { key: 'notifications', path: '/notifications', icon: Bell },
    ],
    'home'
  );
  addPages(insightsPages, 'insights', 'insightsNav');
  addPages(
    settingsNavItems.filter((page) => page.key !== 'integrations'),
    'settings'
  );
  if (canEffectiveAccess(user?.effectiveAccess, 'hub:team:invite')) {
    commands.push({
      id: 'invite-member',
      label: t('common.nav.addMember'),
      group: 'actions',
      icon: UserPlus,
      destination: { path: '/team/invite' },
    });
  }
  return commands;
}

const normalize = (text: string) =>
  text
    .normalize('NFD')
    .replace(/\p{Diacritic}/gu, '')
    .toLowerCase();

export function searchPortalCommands(
  commands: PortalCommand[],
  query: string,
  groupLabel: (group: CommandGroup) => string
): PortalCommand[] {
  const words = normalize(query).trim().split(/\s+/).filter(Boolean);
  return commands.filter((command) => {
    const path =
      'path' in command.destination ? command.destination.path.replace(/[/_-]/g, ' ') : '';
    const searchable = normalize(
      [command.label, command.description, groupLabel(command.group), path].join(' ')
    );
    return words.every((word) => searchable.includes(word));
  });
}
