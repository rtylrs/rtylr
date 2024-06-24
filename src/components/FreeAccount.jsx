import { faCaretDown, faCaretUp } from "@fortawesome/free-solid-svg-icons";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import React from "react";
import { useNavigate } from "react-router-dom";

function FreeAccount() {
  const [open, setOpen] = React.useState(false);
  const navigate = useNavigate();

  return (
    <div
      className="free-account-popup"
      style={{
        height: open ? "auto" : "50px",
      }}
    >
      <div className="free-account-header">
        <button className="dropdown-button" onClick={() => setOpen(!open)}>
          <FontAwesomeIcon icon={open ? faCaretDown : faCaretUp} />
        </button>
      </div>
      <div
        style={{
          display: open ? "block" : "none",
        }}
      >
        <div className="free-account-content">
          <p>You are currently using a free tier.</p>
        </div>
        <button
          className="upgrade-button"
          onClick={() => navigate("/app/business/billing")}
        >
          Upgrade Now
        </button>
      </div>
    </div>
  );
}

export default FreeAccount;
