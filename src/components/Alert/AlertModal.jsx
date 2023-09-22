import { useDispatch, useSelector } from "react-redux";
import { closeAlert } from "../../redux/reducers/alertSlice";

import { motion } from "framer-motion";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faTimes, faBell } from "@fortawesome/free-solid-svg-icons";

function AlertModal() {
  const dispatch = useDispatch();
  const open = useSelector((state) => state.customAlert.show);
  const message = useSelector((state) => state.customAlert.message);
  const type = useSelector((state) => state.customAlert.type);

  const animations = {
    initial: {
      opacity: 0,
      y: -1000,
    },
    animate: {
      opacity: 1,
      y: 0,
    },
    exit: {
      opacity: 0,
      y: -100,
    },
  };

  return (
    <motion.div
      variants={animations}
      initial="initial"
      animate={open ? "animate" : "exit"}
      exit="exit"
      transition={{ duration: 0.7 }}
      style={{
        display: open ? "flex" : "none",
        flexDirection: "column",
        position: "fixed",
        top: "5%",
        right: "0.6%",
        width: 350,
        height: 90,
        backgroundColor: type === "success" ? "#ac9888" : "#D80032",
        justifyContent: "center",
        alignItems: "center",
        zIndex: "100000",
        boxShadow: "0px 0px 10px 0px rgba(0,0,0,0.2)",
        borderRadius: 10,
      }}
    >
      <div className="alert-modal">
        <span>
          <FontAwesomeIcon
            icon={faBell}
            style={{
              color: "white",
              fontSize: 30,
            }}
          />
        </span>
        <p>{message}</p>
        <span
          onClick={() => {
            dispatch(closeAlert());
          }}
        >
          <FontAwesomeIcon
            icon={faTimes}
            style={{
              color: "white",
              fontSize: 20,
              cursor: "pointer",
            }}
          />
        </span>
      </div>
    </motion.div>
  );
}

export default AlertModal;
