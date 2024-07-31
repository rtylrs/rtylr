import {
  faBars,
  faHome,
  faRotateRight,
  faTimes,
  faUserTie,
} from "@fortawesome/free-solid-svg-icons";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { Switch, styled } from "@mui/material";
import { useEffect, useState, useRef } from "react";
import { useSelector } from "react-redux";
import { useNavigate } from "react-router-dom";
import { RootState } from "../redux/store";

const CustomSwitch = styled(Switch)(({}) => ({
  width: 34,
  height: 20,
  padding: 0,
  "& .MuiSwitch-switchBase": {
    padding: 2,
    "&.Mui-checked": {
      transform: "translateX(14px)",
      color: "#fff",
      "& + .MuiSwitch-track": {
        backgroundColor: "#555",
        opacity: 1,
      },
      "&:hover": {
        backgroundColor: "rgba(255, 255, 255, 0.08)",
      },
    },
  },
  "& .MuiSwitch-thumb": {
    width: 16,
    height: 16,
  },
  "& .MuiSwitch-track": {
    borderRadius: 10,
    backgroundColor: "#555",
    opacity: 1,
  },
}));

const useFPS = (): number => {
  const [fps, setFps] = useState(0);
  const frameCount = useRef(0);
  const lastFrameTime = useRef(Date.now());

  useEffect(() => {
    const calculateFPS = () => {
      frameCount.current++;
      const now = Date.now();
      const deltaTime = now - lastFrameTime.current;

      if (deltaTime >= 1000) {
        setFps(frameCount.current);
        frameCount.current = 0;
        lastFrameTime.current = now;
      }

      requestAnimationFrame(calculateFPS);
    };

    requestAnimationFrame(calculateFPS);
  }, []);

  return fps;
};

const getUserInitials = (name: string): string => {
  const names = name.split(" ");
  const initials = names.map((n) => n[0]).join("");
  return initials.toUpperCase();
};

const getPlatform = () => {
  const userAgent = navigator.userAgent;
  if (/Windows/.test(userAgent)) {
    return "Windows";
  } else if (/Mac/.test(userAgent)) {
    return "Mac";
  } else if (/Linux/.test(userAgent)) {
    return "Linux";
  } else {
    return "Unknown";
  }
};

const ToolBar: React.FC = () => {
  const user = useSelector((state: RootState) => state.user.user);
  const navigate = useNavigate();
  const [mode, setMode] = useState<string>(
    localStorage.getItem("mode") || "prod"
  );
  const [showToolbar, setShowToolbar] = useState<boolean>(
    localStorage.getItem("showToolbar") === "true"
  );
  const [menuOpen, setMenuOpen] = useState<boolean>(false);
  const fps = useFPS();

  const handleMenuToggle = () => {
    setMenuOpen((prev) => !prev);
  };

  useEffect(() => {
    localStorage.setItem("mode", mode);
  }, [mode]);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      const os = getPlatform();
      const isMac = os === "Mac";

      const isCtrlAltTPressed =
        e.ctrlKey && e.altKey && e.key.toLowerCase() === "t";
      const isCmdOptTPressed =
        e.metaKey && e.ctrlKey && e.key.toLowerCase() === "t";

      if (isCtrlAltTPressed || (isMac && isCmdOptTPressed)) {
        setShowToolbar((prev) => {
          const newShowToolbar = !prev;
          localStorage.setItem("showToolbar", newShowToolbar.toString());
          return newShowToolbar;
        });
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => {
      window.removeEventListener("keydown", handleKeyDown);
    };
  }, [user]);

  const handleModeChange = () => {
    setMode((prevMode) => (prevMode === "dev" ? "prod" : "dev"));
  };

  const memoryUsage = (performance as any).memory
    ? ((performance as any).memory.usedJSHeapSize / 1024 / 1024).toFixed(2)
    : "Memory info not available";

  const toolbarNavigate = (path: string) => {
    setMenuOpen(false);
    navigate(path);
  };

  const handleClearStorage = () => {
    if (
      window.confirm(
        "Are you sure you want to clear the local storage? This action cannot be undone."
      )
    ) {
      localStorage.clear();
      window.location.reload();
    }
  };

  if (!showToolbar) {
    return null;
  }

  return (
    <div className="toolbar-container">
      <div className="toolbar">
        <div
          className="toolbar-item"
          onClick={() => {
            toolbarNavigate("/");
          }}
        >
          <FontAwesomeIcon icon={faHome} />
        </div>
        <div
          className="toolbar-item"
          onClick={() => {
            toolbarNavigate("/app/business/");
          }}
        >
          <FontAwesomeIcon icon={faUserTie} />
        </div>
        <div
          className="toolbar-item"
          onClick={() => {
            window.location.reload();
          }}
        >
          <FontAwesomeIcon icon={faRotateRight} />
        </div>
        <div className="vertical-line"></div>
        <div className="toolbar-item">
          <CustomSwitch checked={mode === "dev"} onChange={handleModeChange} />
        </div>
        <div className="vertical-line"></div>
        <div className="toolbar-item">
          <span>FPS: {fps}</span>
        </div>
        <div className="vertical-line"></div>
        <div className="toolbar-item">
          <span>RAM: {memoryUsage}</span>
        </div>
        <div className="vertical-line"></div>
        <div className="toolbar-item profile-item">
          {user?.name ? (
            <div className="profile-initials" title={user.name}>
              {getUserInitials(user.name)}
            </div>
          ) : (
            <div className="profile-initials"></div>
          )}
        </div>
        <div className="vertical-line"></div>
        <div className="toolbar-item" onClick={handleMenuToggle}>
          <FontAwesomeIcon icon={menuOpen ? faTimes : faBars} />
        </div>

        {menuOpen && (
          <div className="dropdown-menu">
            <div className="dropdown-item" onClick={() => toolbarNavigate("/")}>
              Home
            </div>
            <div
              className="dropdown-item"
              onClick={() => toolbarNavigate("/app/business/")}
            >
              Statistics
            </div>
            <div
              className="dropdown-item"
              onClick={() => toolbarNavigate("/manage")}
            >
              Restaurants
            </div>
            <div
              className="dropdown-item"
              onClick={() => toolbarNavigate("/manage")}
            >
              Contacts
            </div>
            <div
              className="dropdown-item"
              onClick={handleClearStorage}
              style={{
                color: "#ff0000",
                fontWeight: "bold",
              }}
            >
              Clear Storage
            </div>
            <div
              className="dropdown-item"
              onClick={() => toolbarNavigate("/manage")}
            >
              Settings
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default ToolBar;
