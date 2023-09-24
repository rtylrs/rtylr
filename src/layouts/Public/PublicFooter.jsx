// Footer.js
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import {
  faInstagram,
  faWhatsappSquare,
} from "@fortawesome/free-brands-svg-icons";
import Logo from "../../assets/logo.png";
import { Link } from "react-router-dom";

const Footer = () => {
  return (
    <footer className="footer">
      <div className="footer-top">
        <img src={Logo} alt="Logo" className="footer-logo" />
      </div>
      <div className="footer-sections">
        <div className="footer-section">
          <h3>Help</h3>
          <ul>
            <li>
              <Link to="/Privacy">Privacy</Link>
            </li>
            <li>
              <Link to="/feedback">Feedback</Link>
            </li>
            <li>
              <Link to="/pricing">Pricing</Link>
            </li>
          </ul>
        </div>
        <div className="footer-section">
          <h3>Support</h3>
          <ul>
            <li>
              <Link to="/petition">Petition</Link>
            </li>
            <li>
              <Link to="/donate">Donate</Link>
            </li>
            <li>
              <Link to="/blog/public/add ">Publish Link Blog</Link>
            </li>
          </ul>
        </div>
        <div className="footer-section">
          <h3>Navigate</h3>
          <ul>
            <li>
              <Link to="/">Home</Link>
            </li>
            <li>
              <Link to="/shop">Shop</Link>
            </li>
            <li>
              <Link to="/about">About</Link>
            </li>
            <li>
              <Link to="/contact">Contact Us</Link>
            </li>
          </ul>
        </div>
        <div className="footer-section">
          <h3>Contact</h3>
          <ul>
            <li>
              Email:{" "}
              <span
                style={{
                  fontWeight: "bold",
                  cursor: "pointer",
                }}
                onClick={() => {
                  window.open("mailto:hello@sanitykuwait.com");
                }}
              >
                hello@sanitykuwait.com
              </span>
            </li>
            <li>Phone: +1234567890</li>
          </ul>
        </div>
      </div>
      <div className="footer-social">
        <h3>Follow or Contact Us</h3>
        <div className="social-icons">
          <a href="https://api.whatsapp.com/send?phone=96598895586">
            <FontAwesomeIcon icon={faWhatsappSquare} />
          </a>

          <a href="https://www.instagram.com/TheSanityStore">
            <FontAwesomeIcon icon={faInstagram} />
          </a>
        </div>
      </div>
      <div className="footer-bottom">
        <p>&copy; {new Date().getFullYear()} Sanity</p>
      </div>
    </footer>
  );
};

export default Footer;
