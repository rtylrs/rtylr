import { useEffect } from 'react';
import { useRouterState } from '@tanstack/react-router';

const ScrollToTop: React.FC = () => {
  const pathname = useRouterState({ select: (s) => s.location.pathname });

  useEffect(() => {
    const scrollToTop = () => {
      window.scrollTo({ top: 0, behavior: 'smooth' });
    };

    scrollToTop();

    return () => {
      window.removeEventListener('scroll', scrollToTop);
    };
  }, [pathname]);

  return null;
};

export default ScrollToTop;
