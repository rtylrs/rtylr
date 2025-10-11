import { FaInstagram } from 'react-icons/fa';
import Link from 'next/link';
import PackageJson from '@root/package.json';
import { useConfigStore } from '../../../stores/configStore';
import { ASSETS } from '@voxire/shared';
import React from 'react';
import styles from './Footer.module.scss';

const Footer: React.FC = () => {
  const { contactInfo } = useConfigStore();

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
      links: [{ to: '/pricing', label: 'Pricing' }],
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
      icon: FaInstagram,
      label: 'Follow us on Instagram',
      external: true,
    },
  ];

  return (
    <footer className={styles.footer} role="contentinfo">
      <div className={styles['footer-container']}>
        <div className={styles['footer-header']}>
          <img
            src={ASSETS.logos.dark}
            alt="Rtylr Logo"
            className={styles['footer-logo']}
          />
        </div>

        <div className={styles['footer-content']}>
          <div className={styles['footer-nav']}>
            {footerSections.map((section, index) => (
              <div key={index} className={styles['footer-section']}>
                <h3 className={styles['footer-section-title']}>
                  {section.title}
                </h3>
                <nav aria-label={`${section.title} navigation`}>
                  <ul className={styles['footer-links']}>
                    {section.links.map((link, linkIndex) => (
                      <li key={linkIndex}>
                        {link.external ? (
                          <a
                            href={link.to}
                            target="_blank"
                            rel="noopener noreferrer"
                            className={styles['footer-link']}
                          >
                            {link.label}
                          </a>
                        ) : link.to ? (
                          <Link
                            href={link.to}
                            className={styles['footer-link']}
                          >
                            {link.label}
                          </Link>
                        ) : (
                          <span className={styles['footer-link']}>
                            {link.label}
                          </span>
                        )}
                      </li>
                    ))}
                  </ul>
                </nav>
              </div>
            ))}

            <div className={styles['footer-section']}>
              <h3 className={styles['footer-section-title']}>Contact</h3>
              <div className={styles['contact-info']}>
                <div className={styles['contact-item']}>
                  <span className={styles['contact-label']}>Email:</span>
                  <a
                    href={`mailto:${
                      contactInfo?.contactEmail || 'info@rtylr.com'
                    }`}
                    className={styles['contact-link']}
                    aria-label="Send us an email"
                  >
                    {contactInfo?.contactEmail || 'info@rtylr.com'}
                  </a>
                </div>

                <div className={styles['contact-item']}>
                  <span className={styles['contact-label']}>Follow us:</span>
                  <div
                    className={styles['social-icons']}
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
                        className={styles['social-link']}
                        aria-label={social.label}
                      >
                        <social.icon />
                      </a>
                    ))}
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div className={styles['footer-bottom']}>
          <div className={styles['footer-bottom-content']}>
            <p className={styles.copyright}>
              &copy; {new Date().getFullYear()}{' '}
              <a
                href="https://voxire.com"
                target="_blank"
                rel="noopener noreferrer"
                className={styles['company-link']}
              >
                Voxire
              </a>{' '}
              All rights reserved.
            </p>
            <p className={styles.disclaimer}>
              Unless otherwise noted, the use of third-party logos does not
              imply endorsement, sponsorship, or affiliation with Rtylr.
            </p>
          </div>
        </div>

        <div className={styles['footer-version']}>
          <button
            className={styles['version-button']}
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
