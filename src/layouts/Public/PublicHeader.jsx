import { useState, useEffect } from "react";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import {
  faUser,
  faHeart,
  faBars,
  faTimes,
  faRightToBracket,
} from "@fortawesome/free-solid-svg-icons";
import { auth } from "../../firebase";
import { Link, useNavigate } from "react-router-dom";
import Logo from "../../assets/logo.png";

import PropTypes from "prop-types";

const Header = ({ openLogin }) => {
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
  const [user, setUser] = useState(null);
  const navigate = useNavigate();
  // Use useEffect to set up an authentication listener
  useEffect(() => {
    const unsubscribe = auth.onAuthStateChanged((authUser) => {
      if (authUser) {
        // User is logged in
        setUser(authUser);
      } else {
        // User is not logged in
        setUser(null);
      }
    });

    return () => {
      // Unsubscribe from the listener when the component unmounts
      unsubscribe();
    };
  }, []);

  const toggleMobileMenu = () => {
    setIsMobileMenuOpen(!isMobileMenuOpen);
  };

  return (
    <header className={`header ${isMobileMenuOpen ? "mobile-menu-open" : ""}`}>
      <div className="header-left">
        <Link to="/shop">Shop</Link>
        <Link to="/pricing">Pricing</Link>
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
            transform: user ? "translateX(-125px)" : "translateX(-145px)",
          }}
        />
      </div>
      <div className="header-right">
        {user ? ( // Check if user is logged in
          <Link to="/account">
            <FontAwesomeIcon icon={faUser} />
          </Link>
        ) : (
          // Display login icon if not logged in
          <div
            onClick={() => {
              openLogin();
            }}
          >
            <FontAwesomeIcon icon={faRightToBracket} />
          </div>
        )}

        {user && (
          <>
            <Link to="/admin">
              <FontAwesomeIcon icon={faHeart} />
            </Link>
          </>
        )}
        <button className="mobile-menu-toggle" onClick={toggleMobileMenu}>
          {isMobileMenuOpen ? (
            <FontAwesomeIcon icon={faTimes} />
          ) : (
            <FontAwesomeIcon icon={faBars} />
          )}
        </button>
      </div>
      {isMobileMenuOpen && (
        <div className="mobile-menu">
          <Link to="/">Shop</Link>
          <Link to="/">Account</Link>
          <Link to="/">Wishlist</Link>
          <Link to="/">About</Link>
          <Link to="/">Contact Us</Link>
        </div>
      )}
    </header>
  );
};

Header.propTypes = {
  openLogin: PropTypes.func.isRequired,
};

export default Header;
