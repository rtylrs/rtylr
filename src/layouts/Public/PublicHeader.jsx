import { useState, useEffect } from "react";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import {
  faBars,
  faCartFlatbed,
  faCartShopping,
  faTimes,
} from "@fortawesome/free-solid-svg-icons";
import { auth, db } from "../../firebase";
import { Link, useNavigate } from "react-router-dom";
import Logo from "../../assets/logo.png";

import useWindowDimensions from "../../utils/useWindowDimensions";
import { useSelector } from "react-redux";

const Header = () => {
  const config = useSelector((state) => state.config.websiteData);
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
  const [user, setUser] = useState(null);
  const [userData, setUserData] = useState(null);
  const navigate = useNavigate();
  const { width } = useWindowDimensions();

  // Use useEffect to set up an authentication listener
  useEffect(() => {
    const unsubscribe = auth.onAuthStateChanged((authUser) => {
      if (authUser) {
        // User is logged in
        setUser(authUser);
        db.collection("Users")
          .doc(authUser?.uid)
          .get()
          .then((doc) => {
            if (doc.exists) {
              const user = Object.assign({}, doc.data(), {
                uid: doc.id,
              });
              setUserData(user);
            }
          });
      } else {
        // User is not logged in
        setUser(null);
      }
    });

    return () => {
      unsubscribe();
    };
  }, []);

  const toggleMobileMenu = () => {
    setIsMobileMenuOpen(!isMobileMenuOpen);
  };

  return (
    <header
      className={
        document.location.pathname === "/"
          ? `header ${isMobileMenuOpen ? "mobile-menu-open" : ""}`
          : `header-pages ${isMobileMenuOpen ? "mobile-menu-open" : ""} `
      }
    >
      <div
        className="header-center"
        style={{
          cursor: "pointer",
        }}
        onClick={() => {
          navigate("/");
        }}
      >
        {config && (
          <img
            src={config.icon}
            alt="Logo"
            style={{
              // transform:
              //   width > 768
              //     ? document.location.pathname === "/"
              //       ? "translateX(-110px)"
              //       : "translateX(-75px)"
              //     : "translateX(-90%)",
              filter:
                width > 768
                  ? document.location.pathname === "/"
                    ? "invert(1)"
                    : "none"
                  : isMobileMenuOpen
                  ? "none"
                  : document.location.pathname === "/"
                  ? "invert(1)"
                  : "none",
            }}
          />
        )}
      </div>
      <div
        className="header-left"
        style={{
          transform: auth.currentUser ? "translateX(13%)" : "translateX(0px)",
        }}
      >
        <Link to="/shop">Shop</Link>
        <Link to="/events">Events</Link>
        <Link to="/all-petitions">Petitions</Link>
        <Link to="/about">About</Link>
        <Link to="/contact">Contact Us</Link>
      </div>

      <div className="header-right">
        {/* {width > 768 && (
          <>
            {user ? ( // Check if user is logged in
              <Link to={userData?.isVendor ? "/business" : "/account"}>
                <FontAwesomeIcon icon={faUser} title="My Account" />
              </Link>
            ) : (
              // Display login icon if not logged in
              <div
                onClick={() => {
                  navigate("/login");
                }}
                className="LoginButton"
              >
                <FontAwesomeIcon icon={faRightToBracket} title="Login" />
              </div>
            )}
          </>
        )}

        {user && (
          <>
            {!userData?.isVendor && (
              <Link to="/manage">
                <FontAwesomeIcon icon={faHeart} title="favorites" />
              </Link>
            )}
            <Link to="/notifications">
              <Badge
                badgeContent={notificationsCount}
                color="error"
                style={{ marginRight: "10px" }}
                anchorOrigin={{
                  vertical: "top",
                  horizontal: "right",
                }}
              >
                <FontAwesomeIcon icon={faEnvelope} />
              </Badge>
            </Link>
          </>
        )} */}

        <button
          className="mobile-menu-toggle"
          onClick={toggleMobileMenu}
          style={{
            color:
              document.location.pathname === "/"
                ? isMobileMenuOpen
                  ? "#231f1b"
                  : "#f2ebe5"
                : "#231f1b",
          }}
        >
          {isMobileMenuOpen ? (
            <FontAwesomeIcon icon={faTimes} />
          ) : (
            <FontAwesomeIcon icon={faBars} />
          )}
        </button>

        <button
          className="browse-products"
          onClick={() => {
            navigate("/account/cart");
          }}
        >
          <FontAwesomeIcon icon={faCartShopping} />
        </button>

        <button
          onClick={() => {
            if (user) {
              if (userData?.isAdmin) {
                navigate("/manage");
                return;
              }
              navigate(userData?.isVendor ? "/business" : "/account");
            } else {
              navigate("/auth/login");
            }
          }}
          className="browse-products"
        >
          {user
            ? userData?.isAdmin
              ? "Admin Panel"
              : "My Account"
            : "Join Us"}
        </button>
      </div>
      {isMobileMenuOpen && (
        <div className="mobile-menu">
          <Link to="/">Shop</Link>
          <Link to="/brands">Brands</Link>
          <Link to="/events">Events</Link>
          <Link to="/petitions">Petitions</Link>
          <Link to="/about">About</Link>
          <Link to="/contact">Contact Us</Link>
          <button
            onClick={() => {
              navigate("/shop");
            }}
          >
            Explore Products
          </button>
        </div>
      )}
    </header>
  );
};

export default Header;
