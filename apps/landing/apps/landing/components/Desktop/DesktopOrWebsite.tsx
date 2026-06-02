import React, { useEffect, useState } from 'react';
import { AnimatePresence } from 'motion/react';
import { useDesktopStore } from '../../stores/useDesktopStore';
import { Desktop } from './index';
import { OsModeBar } from './OsModeBar';
import { BootScreen } from './BootScreen';

interface DesktopOrWebsiteProps {
  websiteContent: React.ReactNode;
}

export const DesktopOrWebsite: React.FC<DesktopOrWebsiteProps> = ({ websiteContent }) => {
  const experience = useDesktopStore((s) => s.experience);
  const setExperience = useDesktopStore((s) => s.setExperience);
  const [isMobile, setIsMobile] = useState(false);
  const [ready, setReady] = useState(false);
  const [booting, setBooting] = useState(true);

  useEffect(() => {
    const mobile = window.innerWidth < 768;
    setIsMobile(mobile);
    if (mobile) setExperience('website');
    setReady(true);
  }, [setExperience]);

  // Keep boot screen up briefly after desktop mounts so the fade-out is visible
  useEffect(() => {
    if (ready && experience === 'os' && !isMobile) {
      const t = setTimeout(() => setBooting(false), 600);
      return () => clearTimeout(t);
    } else if (ready) {
      setBooting(false);
    }
  }, [ready, experience, isMobile]);

  if (experience === 'os' && !isMobile) {
    return (
      <>
        {ready && <Desktop />}
        <AnimatePresence>{booting && <BootScreen key="boot" />}</AnimatePresence>
      </>
    );
  }

  if (!ready) return null;

  return (
    <>
      {websiteContent}
      {!isMobile && <OsModeBar />}
    </>
  );
};
