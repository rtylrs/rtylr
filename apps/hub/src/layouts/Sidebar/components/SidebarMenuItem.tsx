import type { IconProps } from "phosphor-react";
import styles from "../Sidebar.module.scss";

interface SidebarMenuItemProps {
  icon?: React.ComponentType<IconProps>;
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
      className={`${styles.item} ${isActive ? styles.active : ""}`}
      onClick={onClick}
    >
      <div className={styles.left}>
        {Icon && <Icon size={16} className={styles.icon} />}
        <span className={styles.text}>{text}</span>
      </div>
    </div>
  );
};

export default SidebarMenuItem;
