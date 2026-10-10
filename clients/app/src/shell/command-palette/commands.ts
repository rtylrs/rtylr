import { financeNavItems } from '../../modules/finance/navigation';
import { hrNavItems } from '../../modules/hr/navigation';
import { recruitNavItems } from '../../modules/recruit/navigation';
import { Webhook } from 'lucide-react';
import { AutomationsIcon } from '@voxire/ui';
import { CaretIcon } from '@voxire/caret';
import type { ComponentType } from 'react';
import { Bell, UserPlus, Plus } from 'lucide-react';
import { resolveAppUrl } from '@voxire/config';
import { canEffectiveAccess, hasEffectiveAppAccess } from '@voxire/permissions';
import { APP_ICONS, WORKSPACE_APP_IDS } from '../../shared/config/appIcons';
import { APP_NAMES } from '../../shared/config/appNames';
import type { StoredUser } from '../../shared/stores/userStore';
import { canAccessPortalPath } from '../../shared/utils/canAccessPortalPath';
import { erpNavItems } from '../../modules/erp/navigation';
import { crmNavItems } from '../../modules/crm/navigation';
import { flowNavItems } from '../../modules/flow/navigation';
import { insightsNavItems } from '../../modules/insights/navigation';
import { rawNavItems, rawMenuSubItems, settingsNavItems, type PortalNavItem } from '../navigation';

export const COMMAND_GROUPS = [
  'apps',
  'home',
  'insights',
  'flow',
  'crm',
  'erp',
  'finance',
  'hr',
  'recruit',
  'projects',
  'tasks',
  'chats',
  'settings',
  'actions',
] as const;
export type CommandGroup = (typeof COMMAND_GROUPS)[number];
export interface PortalCommand {
  id: string;
  label: string;
  description?: string;
  group: CommandGroup;
  icon: ComponentType<{ size?: number; className?: string }>;
  destination:
    | { path: string }
    | { href: string }
    | { newFlowTask: true }
    | { openFlowTask: string }
    | { openCaretChat: string; title?: string; projectId?: string };
}

export function getPortalCommands(
  user: StoredUser | null,
  t: (key: string) => string,
  flowPreview = false
): PortalCommand[] {
  const commands: PortalCommand[] = [];
  if (canAccessPortalPath(user, '/caret')) {
    commands.push({
      id: 'app-caret',
      label: 'Caret',
      group: 'apps',
      icon: CaretIcon,
      destination: { path: '/caret' },
    });
  }
  if (canAccessPortalPath(user, '/automations')) {
    commands.push({
      id: 'tool-automations',
      label: t('common.nav.automations'),
      group: 'actions',
      icon: AutomationsIcon,
      destination: { path: '/automations' },
    });
  }
  if (canAccessPortalPath(user, '/automations/webhooks')) {
    commands.push({
      id: 'tool-webhooks',
      label: t('common.nav.webhooks'),
      group: 'actions',
      icon: Webhook,
      destination: { path: '/automations/webhooks' },
    });
  }
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
        id === 'insights'
          ? { path: insightsPages[0].path }
          : id === 'erp'
            ? { path: '/erp' }
            : id === 'crm'
              ? { path: '/crm' }
              : id === 'flow'
                ? { path: '/flow' }
                : ['finance', 'hr', 'recruit'].includes(id)
                  ? { path: `/${id}` }
                  : { href: resolveAppUrl(id) },
    });
  }
  const addPages = (pages: PortalNavItem[], group: CommandGroup, namespace = 'common.nav') => {
    for (const page of pages) {
      if (!canAccessPortalPath(user, page.path)) continue;
      commands.push({
        id: `page-${page.path}`,
        label: t(
          page.key === 'flowSettings'
            ? 'settingsCenter.flowCommand'
            : page.key === 'erpSettings'
              ? 'settingsCenter.erpCommand'
              : page.key === 'crmSettings'
                ? 'settingsCenter.crmCommand'
                : page.key === 'insightsSettings'
                  ? 'settingsCenter.insightsCommand'
                  : `${namespace}.${page.key}`
        ),
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
  addPages([...financeNavItems], 'finance', 'financeNav');
  addPages([...hrNavItems], 'hr', 'hrNav');
  addPages([...recruitNavItems], 'recruit', 'recruitAdmin.nav');
  addPages([...erpNavItems], 'erp', 'erpNav');
  addPages([...crmNavItems], 'crm', 'crmNav');
  addPages(insightsPages, 'insights', 'insightsNav');
  for (const page of flowNavItems) {
    if (
      ('preview' in page && page.preview && !flowPreview) ||
      !canAccessPortalPath(user, page.path)
    )
      continue;
    commands.push({
      id: `page-${page.path}`,
      label: t(page.key),
      icon: page.icon,
      group: 'flow',
      destination: { path: page.path },
    });
  }
  if (canEffectiveAccess(user?.effectiveAccess, 'flow:tasks:create', 'flow')) {
    commands.push({
      id: 'flow-new-task',
      label: t('flow.shell.newTask'),
      group: 'actions',
      icon: Plus,
      destination: { newFlowTask: true },
    });
  }
  addPages(settingsNavItems, 'settings');
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
