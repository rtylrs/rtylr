import React, { useEffect, useState } from 'react';
import dynamic from 'next/dynamic';

// Loaded only in the browser and only after mount, so the mascot adds
// nothing to the initial page and never runs during prerender.
const CaretMascot = dynamic(() => import('./CaretMascot'), { ssr: false });

/**
 * Renders the mascot only where it can't get in the way: a fine pointer
 * (mouse or trackpad), a wide screen, and no reduced-motion preference.
 */
const CaretMascotLoader: React.FC = () => {
  const [enabled, setEnabled] = useState(false);

  useEffect(() => {
    const query = window.matchMedia(
      '(pointer: fine) and (min-width: 1024px) and (prefers-reduced-motion: no-preference)',
    );
    const update = (): void => setEnabled(query.matches);
    update();
    query.addEventListener('change', update);
    return () => query.removeEventListener('change', update);
  }, []);

  return enabled ? <CaretMascot /> : null;
};

export default CaretMascotLoader;
