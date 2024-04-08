import { motion } from "framer-motion";
import PropTypes from "prop-types";

const animations = {
  initial: {
    opacity: 0,
    y: -100,
  },
  animate: {
    opacity: 1,
    y: 0,
  },
  exit: {
    opacity: 0,
    y: 100,
  },
};

const AnimatedRoutes = ({ children }) => {
  return (
    <motion.div
      variants={animations}
      initial="initial"
      animate="animate"
      exit="exit"
      transition={{ duration: 0.1 }}
    >
      {children}
    </motion.div>
  );
};

AnimatedRoutes.propTypes = {
  children: PropTypes.node.isRequired,
};

export default AnimatedRoutes;
