import { useState, useEffect } from "react";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import {
  faUser,
  faHeart,
  faBars,
  faTimes,
  faRightToBracket,
  faEnvelope,
  faMagnifyingGlass,
} from "@fortawesome/free-solid-svg-icons";
import { auth, db } from "../../firebase";
import { Link, useNavigate } from "react-router-dom";
import Logo from "../../assets/logo.png";

import useWindowDimensions from "../../utils/useWindowDimensions";
import { Badge } from "@mui/material";

const Header = () => {
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
  const [user, setUser] = useState(null);
  const [userData, setUserData] = useState(null);
  const navigate = useNavigate();
  const { width } = useWindowDimensions();
  const [notificationsCount, setNotificationsCount] = useState(0);

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

        db.collection("notifications")
          .where("uid", "==", authUser?.uid)
          .where("read", "==", false)
          .get()
          .then((snapshot) => {
            setNotificationsCount(snapshot.size);
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
      <div className="header-left">
        <Link to="/shop">Shop</Link>
        <Link to="/petitions">Petitions</Link>
        <Link to="/about">About</Link>
        <Link to="/contact">Contact Us</Link>
      </div>
      <div
        className="header-center"
        style={{
          cursor: "pointer",
        }}
        onClick={() => {
          navigate("/");
        }}
      >
        <img
          src={Logo}
          alt="Logo"
          style={{
            transform:
              width > 768
                ? user
                  ? document.location.pathname === "/"
                    ? "translateX(-135px)"
                    : "translateX(-120px)"
                  : document.location.pathname === "/"
                  ? "translateX(-183px)"
                  : "translateX(-150px)"
                : "translateX(-90%)",
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
      </div>
      <div className="header-right">
        {width > 768 && (
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
        )}

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
      </div>
      {isMobileMenuOpen && (
        <div className="mobile-menu">
          <Link to="/shop">Shop</Link>
          <Link to="/petitions">Petitions</Link>
          {user && !userData?.isVendor && (
            <Link to="/favorites">Favorites</Link>
          )}
          <Link to="/about">About</Link>
          <Link to="/contact">Contact Us</Link>

          <div>
            <input type="text" placeholder="Search" />
            <FontAwesomeIcon
              icon={faMagnifyingGlass}
              id="search-icon-menu"
              onClick={() => {
                navigate("/search");
              }}
            />
          </div>
          {user ? (
            <button
              onClick={() => {
                if (userData?.isVendor) {
                  navigate("/business");
                }

                navigate("/account");
              }}
            >
              My Account
            </button>
          ) : (
            <button
              onClick={() => {
                navigate("/login");
              }}
            >
              Login
            </button>
          )}
        </div>
      )}
    </header>
  );
};

export default Header;
