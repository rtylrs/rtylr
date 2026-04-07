import React from "react";
import styles from "./CustomSidebarLayout.module.scss";

interface CustomSidebarLayoutProps {
  children?: React.ReactNode;
  sidebar?: React.ReactNode;
  navbar?: React.ReactNode;
  showNavbar?: boolean;
}

const CustomSidebarLayout: React.FC<CustomSidebarLayoutProps> = ({
  children,
  sidebar,
  navbar,
  showNavbar = true,
}) => {
  return (
    <div className={styles.appLayoutCustomSidebar}>
      {showNavbar && navbar && (
        <div className={styles.appLayoutCustomNavbar}>{navbar}</div>
      )}
      {sidebar && (
        <div className={styles.appLayoutCustomSidebarInner}>{sidebar}</div>
      )}
      <div className={styles.appLayoutCustomContentInner}>{children}</div>
    </div>
  );
};

export default CustomSidebarLayout;


