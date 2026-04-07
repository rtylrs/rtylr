import { LucideIcon } from "lucide-react";

interface SidebarMenuItemProps {
  icon: LucideIcon;
  text: string;
  isActive: boolean;
  onClick: () => void;
}

const SidebarMenuItem = ({
  icon: Icon,
  text,
  isActive,
  onClick,
}: SidebarMenuItemProps) => {
  return (
    <div
      className={`vb-sidebar__link ${isActive ? "active" : ""}`}
      onClick={onClick}
    >
      <div className="vb-sidebar__left">
        <Icon size={20} className="vb-sidebar__icon" />
        <span className="vb-sidebar__text">{text}</span>
      </div>
    </div>
  );
};

export default SidebarMenuItem;