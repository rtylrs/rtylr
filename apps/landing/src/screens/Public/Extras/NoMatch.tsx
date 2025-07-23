import { useNavigate } from "react-router-dom";
import { motion } from "framer-motion";
import VbitezButton from "../../../components/VbitezButton/VbitezButton";

const NoMatch = (): JSX.Element => {
  const navigate = useNavigate();

  const animation = {
    hidden: { opacity: 0.2, y: 100 },
    visible: { opacity: 1, y: 0 },
  };

  return (
    <motion.div
      variants={animation}
      initial="hidden"
      animate="visible"
      transition={{ duration: 0.5 }}
    >
      <div id="notfound">
        <div className="notfound">
          <div className="notfound-404">
            <h1>Oops!</h1>
            <h2>404 - The Page can't be found</h2>
          </div>
          <div
            style={{
              display: "flex",
              justifyContent: "center",
              alignItems: "center",
            }}
          >
            <VbitezButton
              onClick={() => {
                navigate("/");
              }}
            >
              Go Back
            </VbitezButton>
          </div>
        </div>
      </div>
    </motion.div>
  );
};

export default NoMatch;
