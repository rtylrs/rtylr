import React, { useState, useEffect } from "react";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faCookieBite } from "@fortawesome/free-solid-svg-icons";
import { Button } from '@voxire/shared';
import "./Cookies.scss";

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
    <div className="cookies-overlay">
      <div className="cookies-popup">
        <div className="cookies-header">
          <div className="cookies-icon">
            <FontAwesomeIcon icon={faCookieBite} />
          </div>
          <h3>We value your privacy</h3>
        </div>

        <div className="cookies-content">
          <p>
            We use cookies to enhance your browsing experience, analyze site
            traffic, and personalize content. By continuing to use our site, you
            consent to our use of cookies.
          </p>
        </div>

        <div className="cookies-actions">
          <Button variant="primary" size="sm" onClick={handleAccept}>
            Accept All
          </Button>
          <Button variant="outline" size="sm" onClick={handleDecline}>
            Decline
          </Button>
        </div>

        <div className="cookies-footer">
          <a
            href="/privacy"
            target="_blank"
            rel="noopener noreferrer"
            className="privacy-link"
          >
            Learn more about our privacy policy
          </a>
        </div>
      </div>
    </div>
  );
};

export default Cookies;
