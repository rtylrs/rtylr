import React from "react";
import Header from "./PublicHeader";
import Footer from "./PublicFooter";

import PropTypes from "prop-types";
import Loader from "../loader/Loader";

function Public({ children }) {
  const [loading, setLoading] = React.useState(false);

  return (
    <div>
      {loading && <Loader />}
      <div
        style={{
          display: loading ? "none" : "",
        }}
      >
        <Header />
        <div
          style={{
            marginTop: document.location.pathname === "/" ? "0px" : "80px",
          }}
        >
          {React.Children.map(children, (child) => {
            return React.cloneElement(child, {
              startLoading: () => setLoading(true),
              endLoading: () => setLoading(false),
            });
          })}
        </div>
        <Footer />
      </div>
    </div>
  );
}

Public.propTypes = {
  children: PropTypes.node.isRequired,
};

export default Public;
