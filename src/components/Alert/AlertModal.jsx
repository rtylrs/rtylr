import { useDispatch, useSelector } from "react-redux";
import { closeAlert } from "../../Redux/reducers/alertSlice";
import "./AlertModal.css";

import { motion } from "framer-motion";

function AlertModal() {
  const dispatch = useDispatch();
  const open = useSelector((state) => state.customAlert.show);
  const message = useSelector((state) => state.customAlert.message);
  const type = useSelector((state) => state.customAlert.type);

  const animations = {
    initial: {
      opacity: 0,
      x: -100,
    },
    animate: {
      opacity: 1,
      x: 0,
    },
    exit: {
      opacity: 0,
      x: -100,
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
        top: "10%",
        right: "0.6%",
        width: 350,
        height: 90,
        backgroundColor: type === "success" ? "#a1ccd8fa" : "#E8745D",
        justifyContent: "center",
        alignItems: "center",
        zIndex: "100000",
        boxShadow: "0px 0px 10px 0px rgba(0,0,0,0.2)",
        borderRadius: 10,
      }}
    >
      <div className="alert-modal">
        <span>
          <i className="fa-solid fa-bell"></i>
        </span>
        <p>{message}</p>
        <span
          onClick={() => {
            dispatch(closeAlert());
          }}
        >
          <i className="fa-solid fa-times" style={{ cursor: "pointer" }}></i>
        </span>
      </div>
    </motion.div>
  );
}

export default AlertModal;
