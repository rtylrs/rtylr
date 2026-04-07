import { SearchBar } from "./components";
import styles from "./Navbar.module.scss";

const Navbar = () => {
  return (
    <div className={styles.navbar}>
      <div className={styles.navbarLeft} />

      {/* Centered Global Search */}
      <SearchBar />

      <div className={styles.navbarRight} />
    </div>
  );
};

export default Navbar;
