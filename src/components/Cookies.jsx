/*
  File Created By: Mohamad Homsi
  Last Edited: Sunday 12 May 2024
  Last Edited by: Mohamad Homsi
*/

import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faCookieBite } from "@fortawesome/free-solid-svg-icons";
import { useState, useEffect } from "react";
import VbitezButton from "./Buttons/Buttons";

function Cookies() {
  const [showPopup, setShowPopup] = useState(true);

  // Check if user has accepted cookies from local storage
  useEffect(() => {
    const acceptedCookies = localStorage.getItem("acceptedCookies");
    if (acceptedCookies === "true") {
      setShowPopup(false);
    }
  }, []);

  const handleAccept = () => {
    // Set accepted cookies in local storage
    localStorage.setItem("acceptedCookies", true);
    setShowPopup(false);
  };
  const handleDecline = () => {
    // Set declined cookies in local storage
    localStorage.setItem("acceptedCookies", false);
    setShowPopup(false);
  };

  if (!showPopup) {
    // If user has accepted cookies, don't show popup
    return null;
  }

  return (
    <>
      <div className="cookies-popup">
        <div className="cookies-content">
          <div className="cookies-header">
            <FontAwesomeIcon icon={faCookieBite} className="cookie-icon" />
            <h2>Cookie Consent</h2>
          </div>
          <p>
            This website uses cookies to ensure you get the best experience on
            our website.
          </p>
          <div className="cookies-buttons">
            <VbitezButton onClick={handleAccept}>Accept</VbitezButton>
            <VbitezButton onClick={handleDecline}>Decline</VbitezButton>
          </div>
        </div>
      </div>
    </>
  );
}

export default Cookies;
