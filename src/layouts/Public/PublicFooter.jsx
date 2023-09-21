// Footer.js
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import {
  faFacebook,
  faTwitter,
  faInstagram,
} from "@fortawesome/free-brands-svg-icons";

const Footer = () => {
  return (
    <footer className="footer">
      <div className="footer-top">
        <img src="/path/to/your/logo.png" alt="Logo" className="footer-logo" />
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
          </ul>
        </div>
        <div className="footer-section">
          <h3>Screens</h3>
          <ul>
            <li>
              <a href="/home">Home</a>
            </li>
            <li>
              <a href="/products">Products</a>
            </li>
            <li>
              <a href="/about">About</a>
            </li>
            {/* Add more screen links as needed */}
          </ul>
        </div>
        <div className="footer-section">
          <h3>Contact</h3>
          <ul>
            <li>Email: contact@example.com</li>
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
