import { useState } from "react";

const SidebarHeader = () => {
  const [logoError, setLogoError] = useState(false);

  const customLogo = "/emblem.png";
  const fallbackLogo = "https://cdn.voxire.dev/web_content/logo.webp";

  const handleLogoError = () => {
    setLogoError(true);
  };

  return (
    <div className="vb-sidebar__header">
      <div className="vb-sidebar__logo">
        <img
          src={logoError ? fallbackLogo : customLogo}
          alt="Logo"
          className="logo__image"
          onError={handleLogoError}
        />
      </div>
    </div>
  );
};

export default SidebarHeader;