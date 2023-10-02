import "./loader.scss";
import { CircularProgress } from "@mui/material";

const Loader = () => (
  <div className="fallback-spinner">
    <div className="loading">
      <CircularProgress size={60} color="primary" />
    </div>
  </div>
);
export default Loader;
