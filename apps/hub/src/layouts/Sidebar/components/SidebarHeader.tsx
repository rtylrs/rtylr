import { useState } from 'react';
import { ASSETS } from '../../../lib/constants';

const SidebarHeader = () => {
  const [logoError, setLogoError] = useState(false);
  return (
    <div className="px-3 py-3.5 shrink-0">
      <img
        src={logoError ? ASSETS.logos.light : '/emblem.png'}
        alt="Logo"
        className="h-8 w-full object-contain object-left"
        onError={() => setLogoError(true)}
      />
    </div>
  );
};

export default SidebarHeader;
