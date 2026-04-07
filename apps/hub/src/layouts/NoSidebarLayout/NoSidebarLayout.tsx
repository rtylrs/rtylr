import React from "react";
import Navbar from "../Navbar/Navbar";
import styles from "./NoSidebarLayout.module.scss";

interface NoSidebarLayoutProps {
  children?: React.ReactNode;
}

const NoSidebarLayout: React.FC<NoSidebarLayoutProps> = ({ children }) => {
  return (
    <div className={styles.appLayoutNoSidebar}>
      <Navbar />
      <div className={styles.appLayoutContent}>{children}</div>
    </div>
  );
};

export default NoSidebarLayout;


