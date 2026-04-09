import type { IconProps } from 'phosphor-react';
import { cn } from '@voxire/ui';

interface SidebarMenuItemProps {
  icon?: React.ComponentType<IconProps>;
  text: string;
  isActive: boolean;
  onClick: () => void;
}

const SidebarMenuItem = ({ icon: Icon, text, isActive, onClick }: SidebarMenuItemProps) => (
  <div
    className={cn(
      'flex items-center gap-2 px-2.5 py-2 rounded-lg cursor-pointer transition-colors text-[13px] font-medium select-none',
      isActive
        ? 'bg-white/13 text-white'
        : 'text-white/85 hover:bg-white/6 hover:text-white'
    )}
    onClick={onClick}
  >
    {Icon && <Icon size={16} className={cn('shrink-0', isActive ? 'text-white/95' : 'text-white/75')} />}
    <span>{text}</span>
  </div>
);

export default SidebarMenuItem;
