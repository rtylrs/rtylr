import {
  faBars,
  faChartLine,
  faHome,
  faPaperclip,
} from "@fortawesome/free-solid-svg-icons";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { Switch, styled } from "@mui/material";
import { useEffect, useState } from "react";
import { useSelector } from "react-redux";
import { useNavigate } from "react-router-dom";

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

function ToolBar() {
  const user = useSelector((state) => state.user.user);
  const navigate = useNavigate();

  const [mode, setMode] = useState(localStorage.getItem("mode") || "prod");

  useEffect(() => {
    localStorage.setItem("mode", mode);
  }, [mode]);
  const handleModeChange = () => {
    setMode((prevMode) => (prevMode === "dev" ? "prod" : "dev"));
  };

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
          navigate("/manage");
        }}
      >
        <FontAwesomeIcon icon={faPaperclip} />
      </div>
      <div className="toolbar-item">
        <FontAwesomeIcon icon={faChartLine} />
      </div>
      <div className="vertical-line"></div>
      <div className="toolbar-item">
        <CustomSwitch checked={mode === "dev"} onChange={handleModeChange} />
      </div>
      <div className="vertical-line"></div>
      <div className="toolbar-item profile-item">
        <img
          src={user?.image}
          alt="Profile 1"
          style={{
            width: "30px",
            height: "30px",
            borderRadius: "50%",
            border: "0.3px solid #fff",
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
}

export default ToolBar;
