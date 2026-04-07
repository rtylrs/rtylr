import { DotsThreeOutline, GearSix, SignOut } from "phosphor-react";
import { useState, useEffect, useRef } from "react";
import { useNavigate } from "react-router-dom";
import styles from "../Sidebar.module.scss";

const SidebarUser = () => {
  const [userMenuOpen, setUserMenuOpen] = useState(false);
  const navigate = useNavigate();
  const dropdownRef = useRef<HTMLDivElement>(null);

  const storedAuthUser = (() => {
    try {
      const raw = localStorage.getItem("authUser");
      return raw ? JSON.parse(raw) : null;
    } catch {
      return null;
    }
  })();

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (
        dropdownRef.current &&
        !dropdownRef.current.contains(event.target as Node)
      ) {
        setUserMenuOpen(false);
      }
    };

    if (userMenuOpen) {
      document.addEventListener("mousedown", handleClickOutside);
    } else {
      document.removeEventListener("mousedown", handleClickOutside);
    }

    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, [userMenuOpen]);

  const toggleUserMenu = () => {
    setUserMenuOpen(!userMenuOpen);
  };

  const handleLogout = () => {
    try {
      localStorage.removeItem("token");
      localStorage.removeItem("authUser");
    } catch {
      // Ignore errors
    }
    setUserMenuOpen(false);
    navigate("/auth/login", { replace: true });
  };

  const handleProfile = () => {
    navigate("/profile");
    setUserMenuOpen(false);
  };

  const handleSettings = () => {
    navigate("/settings");
    setUserMenuOpen(false);
  };

  const handleDownload = () => {
    window.open("https://github.com/your-repo/releases", "_blank");
    setUserMenuOpen(false);
  };

  return (
    <div className={styles.user}>
      <div className={styles.userAvatar}>
        {(() => {
          const nameOrEmail =
            storedAuthUser?.name || storedAuthUser?.email || "?";
          const initial =
            String(nameOrEmail).trim().charAt(0).toUpperCase() || "?";
          return (
            <div
              className={styles.avatarCircle}
              aria-label={`Avatar ${initial}`}
            >
              {initial}
            </div>
          );
        })()}
      </div>
      <div className={styles.userInfo}>
        <p className={styles.userName}>{storedAuthUser?.name || "User"}</p>
        <p className={styles.userEmail}>
          {storedAuthUser?.email || "user@rtylr.com"}
        </p>
      </div>
      <div
        className={styles.userMenu}
        onClick={toggleUserMenu}
        ref={dropdownRef}
      >
        <DotsThreeOutline size={16} className={styles.menuIcon} />

        {userMenuOpen && (
          <div className={styles.userDropdown}>
            <div className={styles.dropdownSection}>
              <div className={styles.platformOption} onClick={handleSettings}>
                <GearSix size={16} />
                <span>Settings</span>
              </div>
            </div>
            <div className={styles.dropdownSection}>
              <div
                className={styles.platformOption}
                onClick={handleLogout}
                style={{ color: "#ef4444", cursor: "pointer" }}
              >
                <SignOut size={16} />
                <span>Log out</span>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default SidebarUser;
