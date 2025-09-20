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
import { useRouter } from 'next/router';
import { useUserStore } from '../../stores/userStore';
import styles from "./ToolBar.module.scss";

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
  const animationId = useRef<number>();

  useEffect(() => {
    const calculateFPS = () => {
      frameCount.current++;
      const now = Date.now();
      const deltaTime = now - lastFrameTime.current;

      if (deltaTime >= 2000) {
        setFps(frameCount.current);
        frameCount.current = 0;
        lastFrameTime.current = now;
      }

      animationId.current = requestAnimationFrame(calculateFPS);
    };

    animationId.current = requestAnimationFrame(calculateFPS);

    return () => {
      if (animationId.current) {
        cancelAnimationFrame(animationId.current);
      }
    };
  }, []);

  return fps;
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

const ToolBarContent: React.FC = () => {
  const user = useUserStore((state) => state.user);
  const router = useRouter();
  const [mode, setMode] = useState<string>(
    localStorage.getItem("mode") || "prod"
  );
  const [menuOpen, setMenuOpen] = useState<boolean>(false);
  const fps = useFPS();

  const isAdmin = user?.role === "admin";
  const isLocalhost =
    window.location.hostname === "localhost" ||
    window.location.hostname === "127.0.0.1";
  const showAdminItems = isAdmin || !isLocalhost;

  const handleMenuToggle = () => {
    setMenuOpen((prev) => !prev);
  };

  useEffect(() => {
    localStorage.setItem("mode", mode);
  }, [mode]);

  const handleModeChange = () => {
    setMode((prevMode) => (prevMode === "dev" ? "prod" : "dev"));
  };

  const [memoryUsage, setMemoryUsage] = useState("N/A");

  useEffect(() => {
    const updateMemoryUsage = () => {
      if ((performance as any).memory) {
        const used = (performance as any).memory.usedJSHeapSize / 1024 / 1024;
        const total = (performance as any).memory.totalJSHeapSize / 1024 / 1024;
        setMemoryUsage(`${used.toFixed(1)}MB / ${total.toFixed(1)}MB`);
      }
    };

    const interval = setInterval(updateMemoryUsage, 5000);
    updateMemoryUsage();

    return () => clearInterval(interval);
  }, []);

  const toolbarNavigate = (path: string) => {
    setMenuOpen(false);
    router.push(path);
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

  return (
    <div className={styles['toolbar-container']}>
      <div className="toolbar">
        <div
          className="toolbar-item"
          onClick={() => {
            toolbarNavigate("/");
          }}
          title="Home"
        >
          <FontAwesomeIcon icon={faHome} />
        </div>
        {showAdminItems && (
          <div
            className="toolbar-item"
            onClick={() => {
              toolbarNavigate("/app/business/");
            }}
          >
            <FontAwesomeIcon icon={faUserTie} />
          </div>
        )}
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

        <div className="toolbar-item" onClick={handleMenuToggle}>
          <FontAwesomeIcon icon={menuOpen ? faTimes : faBars} />
        </div>

        {menuOpen && (
          <div className="dropdown-menu">
            <div className="dropdown-item" onClick={() => toolbarNavigate("/")}>
              Home
            </div>
            {showAdminItems && (
              <>
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
                  onClick={() => toolbarNavigate("/manage")}
                >
                  Settings
                </div>
              </>
            )}
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
          </div>
        )}
      </div>
    </div>
  );
};

const ToolBar: React.FC = () => {
  const [showToolbar, setShowToolbar] = useState<boolean>(
    localStorage.getItem("showToolbar") === "true"
  );

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
  }, []);

  const isProduction =
    window.location.hostname !== "localhost" &&
    window.location.hostname !== "127.0.0.1";

  if (!showToolbar || isProduction) {
    return null;
  }

  return <ToolBarContent />;
};

export default ToolBar;
