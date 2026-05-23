import React, { useEffect, useState } from 'react';
import { useDesktopStore } from '../../stores/useDesktopStore';
import { Desktop } from './index';
import { OsModeBar } from './OsModeBar';

interface DesktopOrWebsiteProps {
  websiteContent: React.ReactNode;
}

export const DesktopOrWebsite: React.FC<DesktopOrWebsiteProps> = ({ websiteContent }) => {
  const experience = useDesktopStore((s) => s.experience);
  const setExperience = useDesktopStore((s) => s.setExperience);
  const [isMobile, setIsMobile] = useState(false);
  const [ready, setReady] = useState(false);

  useEffect(() => {
    const mobile = window.innerWidth < 768;
    setIsMobile(mobile);
    if (mobile) setExperience('website');
    setReady(true);
  }, [setExperience]);

  if (!ready) return null;

  if (experience === 'os' && !isMobile) {
    return <Desktop />;
  }

  return (
    <>
      {websiteContent}
      {!isMobile && <OsModeBar />}
    </>
  );
};
