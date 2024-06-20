import React from "react";
import Header from "./PublicHeader";
import Footer from "./PublicFooter";
import Loader from "../loader/Loader";

const Public: React.FC<{
  children: React.ReactNode;
}> = ({ children }) => {
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
            return React.isValidElement<{
              startLoading: () => void;
              endLoading: () => void;
            }>(child)
              ? React.cloneElement(child, {
                  startLoading: () => setLoading(true),
                  endLoading: () => setLoading(false),
                })
              : child;
          })}
        </div>
        <Footer />
      </div>
    </div>
  );
};

export default Public;
