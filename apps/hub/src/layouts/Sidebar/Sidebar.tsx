import {
  SidebarHeader,
  SidebarUser,
  SidebarMenu,
} from "./components";
import styles from "./Sidebar.module.scss";

const Sidebar = () => {
  return (
    <div className={styles.sidebar}>
      <SidebarHeader />
      <SidebarMenu />
      <SidebarUser />
    </div>
  );
};

export default Sidebar;