import React from 'react';
import { AnimatePresence, motion } from 'motion/react';
import { useRouterState } from '@tanstack/react-router';

interface AnimatedRoutesProps {
  children: React.ReactNode;
}

const AnimatedRoutes: React.FC<AnimatedRoutesProps> = ({ children }) => {
  const pathname = useRouterState({ select: (s) => s.location.pathname });
  return (
    <AnimatePresence mode="wait">
      <motion.div
        key={pathname}
        initial={{ opacity: 0, y: 6 }}
        animate={{ opacity: 1, y: 0 }}
        exit={{ opacity: 0, y: -6 }}
        transition={{ duration: 0.18 }}
        style={{ width: '100%', height: '100%' }}
      >
        {children}
      </motion.div>
    </AnimatePresence>
  );
};

export { AnimatedRoutes };
