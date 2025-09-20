import { faCaretDown, faCaretUp } from "@fortawesome/free-solid-svg-icons";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import React from "react";
import { useRouter } from "next/router";

const FreeAccount: React.FC<{}> = () => {
  const [open, setOpen] = React.useState<boolean>(true);
  const router = useRouter();

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
          onClick={() => router.push("/app/business/billing")}
        >
          Select a Plan
        </button>
      </div>
    </div>
  );
};

export default FreeAccount;
