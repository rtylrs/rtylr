import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import { ASSETS } from "../../../../../lib/constants";

const SettingsNavbar: React.FC = () => {
  const navigate = useNavigate();
  const [logoError, setLogoError] = useState(false);

  return (
    <div className="h-14 bg-white border-b border-black/8 flex items-center px-6">
      <div
        className="cursor-pointer"
        onClick={() => navigate("/")}
      >
        <img
          src={logoError ? ASSETS.logos.light : "/emblem.png"}
          alt="Logo"
          className="h-7 w-auto object-contain"
          onError={() => setLogoError(true)}
        />
      </div>
    </div>
  );
};

export default SettingsNavbar;
