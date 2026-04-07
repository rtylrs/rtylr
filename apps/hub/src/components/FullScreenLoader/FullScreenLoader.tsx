import React from "react";
import "./FullScreenLoader.scss";

export const FullScreenLoader: React.FC = () => {
  return (
    <div className="fullscreen-loader">
      <div className="loader-spinner">
        <div className="spinner"></div>
      </div>
    </div>
  );
};
