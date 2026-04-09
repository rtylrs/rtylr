import type { IconProps } from 'phosphor-react';
import { CaretDown, CaretUp } from 'phosphor-react';
import SidebarMenuItem from './SidebarMenuItem';
import { MenuItem } from './menuConfig';

interface SidebarDropdownProps {
  icon: React.ComponentType<IconProps>;
  title: string;
  sectionKey: string;
  items: MenuItem[];
  isVisible: boolean;
  currentPath: string;
  onToggle: (key: string) => void;
  onNavigate: (path: string) => void;
}

const SidebarDropdown = ({ icon: Icon, title, sectionKey, items, isVisible, currentPath, onToggle, onNavigate }: SidebarDropdownProps) => (
  <div>
    <div
      className="flex items-center justify-between px-2.5 py-2 rounded-lg cursor-pointer text-white/85 hover:bg-white/6 transition-colors select-none"
      onClick={() => onToggle(sectionKey)}
    >
      <div className="flex items-center gap-2">
        <Icon size={16} className="text-white/75 shrink-0" />
        <span className="text-[13px] font-medium">{title}</span>
      </div>
      {isVisible ? <CaretUp size={11} className="text-white/60" /> : <CaretDown size={11} className="text-white/60" />}
    </div>
    {isVisible && (
      <div className="ml-7 mr-2 mt-0.5 space-y-0.5">
        {items.map((item, i) => (
          <SidebarMenuItem
            key={i}
            text={item.text}
            isActive={currentPath === item.path}
            onClick={() => onNavigate(item.path)}
          />
        ))}
      </div>
    )}
  </div>
);

export default SidebarDropdown;
