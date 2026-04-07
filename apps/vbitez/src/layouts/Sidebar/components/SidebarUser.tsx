import { MoreHorizontal, User, Settings, LogOut, Download } from "lucide-react";
import { useState, useEffect, useRef } from "react";
import { useNavigate } from "react-router-dom";
import { usePlatform } from "../../../hooks/usePlatform";

const SidebarUser = () => {
  const [userMenuOpen, setUserMenuOpen] = useState(false);
  const navigate = useNavigate();
  const dropdownRef = useRef<HTMLDivElement>(null);
  const { isWeb } = usePlatform();

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
    <div className="vb-sidebar__user">
      <div className="user__avatar">
        {(() => {
          const nameOrEmail =
            storedAuthUser?.name || storedAuthUser?.email || "?";
          const initial =
            String(nameOrEmail).trim().charAt(0).toUpperCase() || "?";
          return (
            <div className="avatar__circle" aria-label={`Avatar ${initial}`}>
              {initial}
            </div>
          );
        })()}
      </div>
      <div className="user__info">
        <p className="user__name">{storedAuthUser?.name || "User"}</p>
        <p className="user__email">
          {storedAuthUser?.email || "user@vbitez.com"}
        </p>
      </div>
      <div className="user__menu" onClick={toggleUserMenu} ref={dropdownRef}>
        <MoreHorizontal size={16} className="menu__icon" />

        {userMenuOpen && (
          <div className="user__dropdown">
            <div className="dropdown__section">
              <div className="platform__option" onClick={handleProfile}>
                <User size={16} />
                <span>Profile</span>
              </div>
              <div className="platform__option" onClick={handleSettings}>
                <Settings size={16} />
                <span>Settings</span>
              </div>
              {isWeb && (
                <div className="platform__option" onClick={handleDownload}>
                  <Download size={16} />
                  <span>Download App</span>
                </div>
              )}
            </div>
            <div className="dropdown__section">
              <div
                className="platform__option"
                onClick={handleLogout}
                style={{ color: "#ef4444", cursor: "pointer" }}
              >
                <LogOut size={16} />
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
