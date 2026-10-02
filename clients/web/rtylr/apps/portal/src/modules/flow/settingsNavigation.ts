import { Bell, Palette, Keyboard, Sparkles, Tag, Trash2, Clock, PanelLeft } from 'lucide-react';
import type { EffectiveAccess } from '@voxire/types';
import type { SettingsSectionId } from './shared/schemas/settings';
import { canFlowPermission } from './shared/lib/roleUtils';

export function getFlowSettingsSections(
  access: EffectiveAccess | undefined
): Array<{ id: SettingsSectionId; labelKey: string; icon: typeof Bell }> {
  return [
    { id: 'notifications', labelKey: 'flowCommon.settings.notifications', icon: Bell },
    { id: 'time', labelKey: 'flow.settings.time.section', icon: Clock },
    { id: 'appearance', labelKey: 'flowCommon.settings.appearance', icon: Palette },
    { id: 'sidebar', labelKey: 'flow.sidebar.settingsSection', icon: PanelLeft },
    { id: 'labels', labelKey: 'flowCommon.settings.labels', icon: Tag },
    { id: 'life', labelKey: 'flowCommon.settings.life', icon: Sparkles },
    { id: 'shortcuts', labelKey: 'flowCommon.settings.shortcuts', icon: Keyboard },
    ...(canFlowPermission('flow:projects:edit', access)
      ? [{ id: 'bin' as const, labelKey: 'flow.settings.bin.section', icon: Trash2 }]
      : []),
  ];
}
