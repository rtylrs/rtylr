import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faInstagram } from '@fortawesome/free-brands-svg-icons';
import { Link } from 'react-router-dom';
import { useSelector } from 'react-redux';
import PackageJson from '@root/package.json';
import { RootState } from '../../../redux/store';
import React from 'react';
import './PublicFooter.scss';

const Footer: React.FC = () => {
  const config = useSelector((state: RootState) => state.config.contactInfo);

  const footerSections = [
    {
      title: 'Company',
      links: [
        { to: '/about', label: 'About Us' },
        { to: '/contact', label: 'Contact Us' },
        {
          to: 'https://voxire.com/careers',
          label: 'Careers',
          external: true,
        },
      ],
    },
    {
      title: 'Quick Links',
      links: [
        { to: '/faqs', label: 'FAQs' },
        { to: '/pricing', label: 'Pricing' },
        { to: '/feedback', label: 'Feedback' },
      ],
    },
    {
      title: 'Legal',
      links: [
        { to: '/', label: 'Privacy Policy' },
        { to: '/', label: 'Terms and Conditions' },
        { to: '/', label: 'Disclaimer' },
      ],
    },
  ];

  const socialLinks = [
    {
      href: 'https://www.instagram.com/voxire.tech',
      icon: faInstagram,
      label: 'Follow us on Instagram',
      external: true,
    },
  ];

  return (
    <footer className="footer" role="contentinfo">
      <div className="footer-container">
        <div className="footer-header">
          <img src="/logo.png" alt="Vbitez Logo" className="footer-logo" />
        </div>

        <div className="footer-content">
          <div className="footer-nav">
            {footerSections.map((section, index) => (
              <div key={index} className="footer-section">
                <h3 className="footer-section-title">{section.title}</h3>
                <nav aria-label={`${section.title} navigation`}>
                  <ul className="footer-links">
                    {section.links.map((link, linkIndex) => (
                      <li key={linkIndex}>
                        {link.external ? (
                          <a
                            href={link.to}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="footer-link"
                          >
                            {link.label}
                          </a>
                        ) : (
                          <Link to={link.to} className="footer-link">
                            {link.label}
                          </Link>
                        )}
                      </li>
                    ))}
                  </ul>
                </nav>
              </div>
            ))}

            <div className="footer-section">
              <h3 className="footer-section-title">Contact</h3>
              <div className="contact-info">
                <div className="contact-item">
                  <span className="contact-label">Email:</span>
                  <a
                    href={`mailto:${config?.contactEmail || 'info@vbitez.com'}`}
                    className="contact-link"
                    aria-label="Send us an email"
                  >
                    {config?.contactEmail || 'info@vbitez.com'}
                  </a>
                </div>

                <div className="contact-item">
                  <span className="contact-label">Follow us:</span>
                  <div
                    className="social-icons"
                    role="group"
                    aria-label="Social media links"
                  >
                    {socialLinks.map((social, index) => (
                      <a
                        key={index}
                        href={social.href}
                        target={social.external ? '_blank' : undefined}
                        rel={
                          social.external ? 'noopener noreferrer' : undefined
                        }
                        className="social-link"
                        aria-label={social.label}
                      >
                        <FontAwesomeIcon icon={social.icon} />
                      </a>
                    ))}
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div className="footer-bottom">
          <div className="footer-bottom-content">
            <p className="copyright">
              &copy; {new Date().getFullYear()}{' '}
              <a
                href="https://voxire.com"
                target="_blank"
                rel="noopener noreferrer"
                className="company-link"
              >
                Voxire
              </a>{' '}
              All rights reserved.
            </p>
            <p className="disclaimer">
              Unless otherwise noted, the use of third-party logos does not
              imply endorsement, sponsorship, or affiliation with Vbitez.
            </p>
          </div>
        </div>

        <div className="footer-version">
          <button
            className="version-button"
            aria-label={`Current version: ${PackageJson.version}`}
          >
            {PackageJson.version}
          </button>
        </div>
      </div>
    </footer>
  );
};

export default Footer;
