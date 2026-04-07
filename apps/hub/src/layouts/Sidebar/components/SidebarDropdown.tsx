import type { IconProps } from "phosphor-react";
import { CaretDown, CaretUp } from "phosphor-react";
import SidebarMenuItem from "./SidebarMenuItem";
import { MenuItem } from "./menuConfig";
import styles from "../Sidebar.module.scss";

interface SidebarDropdownProps {
  icon: React.ComponentType<IconProps>;
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
    <div>
      <div
        className={`${styles.item} ${styles.header}`}
        onClick={() => onToggle(sectionKey)}
      >
        <div className={styles.left}>
          <Icon size={16} className={styles.icon} />
          <span className={styles.text}>{title}</span>
        </div>
        {isVisible ? (
          <CaretUp size={12} className={styles.chevron} />
        ) : (
          <CaretDown size={12} className={styles.chevron} />
        )}
      </div>
      {isVisible && (
        <div className={styles.submenuItems}>
          {items.map((item, index) => (
            <SidebarMenuItem
              key={index}
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