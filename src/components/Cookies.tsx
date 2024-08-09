import React, { useState, useEffect } from "react";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faCookieBite } from "@fortawesome/free-solid-svg-icons";
import VbitezButton from "./Buttons/Buttons";

const Cookies: React.FC = () => {
  const [showPopup, setShowPopup] = useState(true);

  useEffect(() => {
    const acceptedCookies = localStorage.getItem("acceptedCookies");
    if (acceptedCookies === "true") {
      setShowPopup(false);
    }
  }, []);

  const handleAccept = () => {
    localStorage.setItem("acceptedCookies", "true");
    setShowPopup(false);
  };
  const handleDecline = () => {
    localStorage.setItem("acceptedCookies", "false");
    setShowPopup(false);
  };

  if (!showPopup) {
    return null;
  }

  return (
    <React.Fragment>
      <div className="cookies-popup">
        <div className="cookies-content">
          <div className="cookies-header">
            <FontAwesomeIcon icon={faCookieBite} className="cookie-icon" />
            <h2>Cookie Consent</h2>
          </div>
          <p>
            We use cookies to make this website better and more functional for
            you. It's up to you to decide which cookies you want to accept on
            your device.
          </p>
          <div className="cookies-buttons">
            <VbitezButton
              onClick={handleAccept}
              styles={{
                height: "40px",
              }}
            >
              Accept
            </VbitezButton>
            <VbitezButton
              onClick={handleDecline}
              variant="secondary"
              styles={{
                height: "40px",
              }}
            >
              Decline
            </VbitezButton>
          </div>
        </div>
      </div>
    </React.Fragment>
  );
};

export default Cookies;
