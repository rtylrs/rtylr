import React, { useState, useEffect } from "react";
import { Cookie } from "phosphor-react";
import { Button } from "../ui/button";

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

  if (!showPopup) return null;

  return (
    <div className="fixed inset-0 z-50 bg-black/30 flex items-end justify-center p-4 sm:items-center">
      <div className="w-full max-w-md bg-white rounded-2xl shadow-xl p-6 space-y-4">
        <div className="flex items-center gap-3">
          <div className="w-9 h-9 rounded-xl bg-accent/10 flex items-center justify-center text-accent">
            <Cookie size={20} />
          </div>
          <h3 className="text-base font-semibold text-primary">We value your privacy</h3>
        </div>

        <p className="text-sm text-muted leading-relaxed">
          We use cookies to enhance your browsing experience, analyze site traffic, and personalize content.
          By continuing to use our site, you consent to our use of cookies.
        </p>

        <div className="flex gap-3">
          <Button variant="default" size="sm" onClick={handleAccept}>Accept All</Button>
          <Button variant="outline" size="sm" onClick={handleDecline}>Decline</Button>
        </div>

        <a href="/privacy" target="_blank" rel="noopener noreferrer" className="block text-xs text-accent hover:underline">
          Learn more about our privacy policy
        </a>
      </div>
    </div>
  );
};

export default Cookies;
