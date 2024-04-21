import { useState, useEffect } from "react";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faBars, faTimes } from "@fortawesome/free-solid-svg-icons";
import { auth, db } from "../../firebase";
import { Link, useNavigate } from "react-router-dom";

import { useSelector } from "react-redux";
import VbitezButton from "../../components/Buttons/Buttons";
import useColors from "../../hooks/useColors";

const Header = () => {
  const config = useSelector((state) => state.config.websiteData);
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
  const [user, setUser] = useState(null);
  const [userData, setUserData] = useState(null);
  const navigate = useNavigate();
  const { primary } = useColors();

  const LinkWithColor = (props) => {
    // eslint-disable-next-line react/prop-types
    const { color, ...rest } = props;
    return (
      <Link
        style={{
          color: color || primary,
        }}
        {...rest}
      />
    );
  };

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

  const [scrolled, setScrolled] = useState(false);

  useEffect(() => {
    window.addEventListener("scroll", () => {
      if (window.scrollY > 150) {
        setScrolled(true);
      } else {
        setScrolled(false);
      }
    });
  }, []);

  return (
    <header
      className={
        document.location.pathname === "/"
          ? `header ${isMobileMenuOpen ? "mobile-menu-open" : ""}`
          : `header-pages ${isMobileMenuOpen ? "mobile-menu-open" : ""}
          ${scrolled ? "scrolled" : ""}`
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
            style={
              {
                // transform:
                //   width > 768
                //     ? document.location.pathname === "/"
                //       ? "translateX(-110px)"
                //       : "translateX(-75px)"
                //     : "translateX(-90%)",
                // filter:
                //   width > 768
                //     ? document.location.pathname === "/"
                //       ? "invert(1)"
                //       : "none"
                //     : isMobileMenuOpen
                //     ? "none"
                //     : document.location.pathname === "/"
                //     ? "invert(1)"
                //     : "none",
              }
            }
          />
        )}
      </div>
      <div
        className="header-left"
        style={{
          transform: "translateX(5%)",
        }}
      >
        <LinkWithColor to="/restaurant/demo">Demo</LinkWithColor>
        <LinkWithColor to="/pricing">Product</LinkWithColor>
        <LinkWithColor to="/blog">Demo</LinkWithColor>
        <LinkWithColor to="/contact">Testimonials</LinkWithColor>
        <LinkWithColor to="/about">About Us</LinkWithColor>
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

        <VbitezButton
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
          text={
            user
              ? userData?.isAdmin
                ? "Admin Panel"
                : "My Account"
              : "Sign Up Now"
          }
        />
      </div>
      {isMobileMenuOpen && (
        <div className="mobile-menu">
          <LinkWithColor to="/menu">Menu</LinkWithColor>
          <LinkWithColor to="/deals">Deals</LinkWithColor>
          <LinkWithColor to="/blog">blog</LinkWithColor>
          <LinkWithColor to="/about">About</LinkWithColor>
          <LinkWithColor to="/contact">Contact Us</LinkWithColor>
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
