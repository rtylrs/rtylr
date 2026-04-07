import React from "react";
import Navbar from "../Navbar/Navbar";
import Sidebar from "../Sidebar/Sidebar";
import styles from "./DefaultLayout.module.scss";

interface DefaultLayoutProps {
  children?: React.ReactNode;
}

const DefaultLayout: React.FC<DefaultLayoutProps> = ({ children }) => {
  return (
    <div className={styles.appLayout}>
      <Sidebar />
      <Navbar />
      <div className={styles.appLayoutContent}>{children}</div>
    </div>
  );
};

export default DefaultLayout;
