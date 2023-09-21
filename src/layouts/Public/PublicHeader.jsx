import { useState, useEffect } from "react";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import {
  faUser,
  faHeart,
  faBars,
  faTimes,
  faBasketShopping,
  faRightToBracket,
} from "@fortawesome/free-solid-svg-icons";
import { auth } from "../../firebase";
import { Link } from "react-router-dom";

import Logo from "../../assets/images/logo.png";
import PropTypes from "prop-types";

const Header = ({ openLogin }) => {
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
  const [user, setUser] = useState(null);

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
        <a href="/">Shop</a>
        <a href="/">About</a>
        <a href="/">Contact Us</a>
      </div>
      <div className="header-center">
        <img src={Logo} alt="Logo" />
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
        <Link to="/wishlist">
          <FontAwesomeIcon icon={faHeart} />
        </Link>
        <Link to="/cart">
          <FontAwesomeIcon icon={faBasketShopping} />
        </Link>
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
          <a href="/">Mobile Link 1</a>
          <a href="/">Mobile Link 2</a>
          <a href="/">Mobile Link 3</a>
        </div>
      )}
    </header>
  );
};

Header.propTypes = {
  openLogin: PropTypes.func.isRequired,
};

export default Header;
