import { AnimatePresence, motion } from "motion/react";
import React from "react";

type AnimationType =
  "slide" | "fade" | "scale" | "slideUp" | "slideDown" | "none";

interface AnimatedRoutesProps {
  children: React.ReactNode;
  animationType?: AnimationType;
  /** Changes to this value remount the route so exit animations can run. */
  animationKey?: React.Key;
}

const animations = {
  slide: {
    initial: {
      opacity: 0,
      x: -20,
    },
    animate: {
      opacity: 1,
      x: 0,
    },
    exit: {
      opacity: 0,
      x: 20,
    },
  },
  fade: {
    initial: {
      opacity: 0,
    },
    animate: {
      opacity: 1,
    },
    exit: {
      opacity: 0,
    },
  },
  scale: {
    initial: {
      opacity: 0,
      scale: 0.95,
    },
    animate: {
      opacity: 1,
      scale: 1,
    },
    exit: {
      opacity: 0,
      scale: 1.05,
    },
  },
  slideUp: {
    initial: {
      opacity: 0,
      y: 20,
    },
    animate: {
      opacity: 1,
      y: 0,
    },
    exit: {
      opacity: 0,
      y: -20,
    },
  },
  slideDown: {
    initial: {
      opacity: 0,
      y: -20,
    },
    animate: {
      opacity: 1,
      y: 0,
    },
    exit: {
      opacity: 0,
      y: 20,
    },
  },
  none: {
    initial: {},
    animate: {},
    exit: {},
  },
};

/**
 * @param children - React components/routes to animate
 * @param animationType - Type of animation: slide, fade, scale, slideUp, slideDown, or none (default: fade)
 * @param animationKey - Key identifying the current route; change it to run
 * exit/enter transitions.
 */
const AnimatedRoutes: React.FC<AnimatedRoutesProps> = ({
  children,
  animationType = "slideDown",
  animationKey,
}) => {
  const selectedAnimation = animations[animationType];

  return (
    <AnimatePresence mode="wait">
      <motion.div
        key={animationKey}
        variants={selectedAnimation}
        initial="initial"
        animate="animate"
        exit="exit"
        transition={{
          duration: 0.2,
          ease: "easeOut",
        }}
      >
        {children}
      </motion.div>
    </AnimatePresence>
  );
};

export default AnimatedRoutes;
