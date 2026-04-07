import { LucideIcon, ChevronDown, ChevronUp } from "lucide-react";
import SidebarMenuItem from "./SidebarMenuItem";
import { MenuItem } from "./menuConfig";

interface SidebarDropdownProps {
  icon: LucideIcon;
  title: string;
  sectionKey: string;
  items: MenuItem[];
  isVisible: boolean;
  currentPath: string;
  onToggle: (sectionKey: string) => void;
  onNavigate: (path: string) => void;
}

const SidebarDropdown = ({
  icon: Icon,
  title,
  sectionKey,
  items,
  isVisible,
  currentPath,
  onToggle,
  onNavigate,
}: SidebarDropdownProps) => {
  return (
    <div className="vb-sidebar__section">
      <div
        className="vb-sidebar__link vb-sidebar__link--header"
        onClick={() => onToggle(sectionKey)}
      >
        <div className="vb-sidebar__left">
          <Icon size={20} className="vb-sidebar__icon" />
          <span className="vb-sidebar__text">{title}</span>
        </div>
        {isVisible ? (
          <ChevronUp size={14} className="vb-sidebar__chevron" />
        ) : (
          <ChevronDown size={14} className="vb-sidebar__chevron" />
        )}
      </div>
      {isVisible && (
        <div className="vb-sidebar__submenu-items">
          {items.map((item, index) => (
            <SidebarMenuItem
              key={index}
              icon={item.icon}
              text={item.text}
              isActive={currentPath === item.path}
              onClick={() => onNavigate(item.path)}
            />
          ))}
        </div>
      )}
    </div>
  );
};

export default SidebarDropdown;