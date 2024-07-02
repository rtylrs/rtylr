import {
  faBars,
  faHome,
  faPaperclip,
  faUserTie,
} from "@fortawesome/free-solid-svg-icons";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { Switch, styled } from "@mui/material";
import { useEffect, useState, useRef } from "react";
import { useSelector } from "react-redux";
import { useNavigate } from "react-router-dom";
import { RootState } from "../redux/store";

const CustomSwitch = styled(Switch)(({ theme }) => ({
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

const ToolBar: React.FC = () => {
  const user = useSelector((state: RootState) => state.user.user);
  const navigate = useNavigate();
  const [mode, setMode] = useState<string>(
    localStorage.getItem("mode") || "prod"
  );
  const [showToolbar, setShowToolbar] = useState<boolean>(
    localStorage.getItem("showToolbar") === "true"
  );
  const fps = useFPS();

  useEffect(() => {
    localStorage.setItem("mode", mode);
  }, [mode]);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      const isMac = navigator.platform.toUpperCase().indexOf("MAC") >= 0;
      const isCtrlAltTPressed = e.ctrlKey && e.altKey && e.key === "t";
      const isCmdOptTPressed = e.metaKey && e.altKey && e.key === "t";

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

  if (!showToolbar) {
    return null;
  }

  return (
    <div className="toolbar">
      <div
        className="toolbar-item"
        onClick={() => {
          navigate("/");
        }}
      >
        <FontAwesomeIcon icon={faHome} />
      </div>
      <div
        className="toolbar-item"
        onClick={() => {
          navigate("/app/business/");
        }}
      >
        <FontAwesomeIcon icon={faUserTie} />
      </div>
      <div
        className="toolbar-item"
        onClick={() => {
          navigate("/manage");
        }}
      >
        <FontAwesomeIcon icon={faPaperclip} />
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
        <img
          src={user?.image}
          alt="Profile"
          style={{
            width: "30px",
            height: "30px",
            borderRadius: "50%",
            border: "0.02px solid #fff",
          }}
        />
      </div>
      <div className="vertical-line"></div>
      <div
        className="toolbar-item"
        onClick={() => {
          navigate("/manage");
        }}
      >
        <FontAwesomeIcon icon={faBars} />
      </div>
    </div>
  );
};

export default ToolBar;
