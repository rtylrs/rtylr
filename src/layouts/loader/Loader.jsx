import { Spinner } from "reactstrap";
import "./loader.scss";

const Loader = () => (
  <div className="fallback-spinner">
    <div className="loading">
      <Spinner color="primary" />
      <h1>Loading ..</h1>
    </div>
  </div>
);
export default Loader;
