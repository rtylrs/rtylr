import React from "react";
import Header from "./PublicHeader";
import Footer from "./PublicFooter";

import PropTypes from "prop-types";
import Loader from "../loader/Loader";
import Login from "../../components/Login";

function Public({ children }) {
  const [loading, setLoading] = React.useState(false);
  const [login, setLogin] = React.useState(false);
  const [register, setRegister] = React.useState(false);

  return (
    <div>
      {loading && <Loader />}
      <div
        style={{
          display: loading ? "none" : "",
        }}
      >
        <Header
          openLogin={() => setLogin(true)}
          openRegister={() => setRegister(true)}
        />
        <Login
          open={login}
          close={() => {
            setLogin(false);
          }}
          openRegister={() => setRegister(true)}
        />
        <div
          style={{
            marginTop: "90px",
          }}
        >
          {React.Children.map(children, (child) => {
            return React.cloneElement(child, {
              openLogin: () => setLogin(true),
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
