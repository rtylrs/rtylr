// Footer.js
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import {
  faFacebook,
  faTwitter,
  faInstagram,
} from "@fortawesome/free-brands-svg-icons";
import Logo from "../../assets/logo.png";

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
              <a href="/faq">FAQ</a>
            </li>
            <li>
              <a href="/support">Support</a>
            </li>
            <li>
              <a href="/Privacy">Privacy</a>
            </li>
          </ul>
        </div>
        <div className="footer-section">
          <h3>Screens</h3>
          <ul>
            <li>
              <a href="/">Home</a>
            </li>
            <li>
              <a href="/shop">Products</a>
            </li>
            <li>
              <a href="/about">About</a>
            </li>{" "}
            <li>
              <a href="/contact">Contact Us</a>
            </li>
            {/* Add more screen links as needed */}
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
        <h3>Follow Us</h3>
        <div className="social-icons">
          <a href="https://www.facebook.com">
            <FontAwesomeIcon icon={faFacebook} />
          </a>
          <a href="https://www.twitter.com">
            <FontAwesomeIcon icon={faTwitter} />
          </a>
          <a href="https://www.instagram.com">
            <FontAwesomeIcon icon={faInstagram} />
          </a>
        </div>
      </div>
      <div className="footer-bottom">
        <p>&copy; {new Date().getFullYear()} Your Company Name</p>
      </div>
    </footer>
  );
};

export default Footer;
