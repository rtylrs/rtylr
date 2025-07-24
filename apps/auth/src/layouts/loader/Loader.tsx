import React from "react";
import "./loader.scss";
import { CircularProgress } from "@mui/material";

const Loader: React.FC<{}> = () => (
  <div className="overlay">
    <div className="fallback-spinner">
      <div className="loading">
        <CircularProgress size={60} color="primary" />
      </div>
    </div>
  </div>
);
export default Loader;
