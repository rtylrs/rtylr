import { motion } from "framer-motion";

const animations = {
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
};

const AnimatedRoutes = ({ children }) => {
  return (
    <motion.div
      variants={animations}
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
  );
};

export default AnimatedRoutes;
