import { FaInstagram } from 'react-icons/fa';
import { Link } from '@tanstack/react-router';
import { APP_URLS } from '@voxire/config';
import React from 'react';

const Footer: React.FC = () => {
  const landingUrl = (APP_URLS.landing || 'https://rtylr.com').replace(/\/$/, '');
  const contactEmail = 'info@rtylr.com';
  const appVersion = import.meta.env.VITE_APP_VERSION ?? '0.0.0';

  const footerSections = [
    {
      title: 'Company',
      links: [
        { to: `${landingUrl}/about`, label: 'About Us', external: true },
        { to: `${landingUrl}/contact`, label: 'Contact Us', external: true },
        { to: 'https://hire.voxire.com/jobs/', label: 'Careers', external: true },
      ],
    },
    {
      title: 'Quick Links',
      links: [
        { to: `${landingUrl}/support`, label: 'FAQs', external: true },
        { to: `${landingUrl}/pricing`, label: 'Pricing', external: true },
      ],
    },
    {
      title: 'Legal',
      links: [
        { to: `${landingUrl}/legal/privacy-policy`, label: 'Privacy Policy', external: true },
        { to: `${landingUrl}/legal/terms`, label: 'Terms and Conditions', external: true },
      ],
    },
  ];

  return (
    <footer className="bg-gray-900 text-white" role="contentinfo">
      <div className="max-w-6xl mx-auto px-6 py-12">
        <div className="mb-8">
          <img src="/logo.png" alt="Rtylr Logo" className="h-8 w-auto" />
        </div>

        <div className="grid grid-cols-2 md:grid-cols-4 gap-8 mb-12">
          {footerSections.map((section, index) => (
            <div key={index}>
              <h3 className="text-xs font-semibold uppercase tracking-wider text-white/50 mb-3">
                {section.title}
              </h3>
              <nav aria-label={`${section.title} navigation`}>
                <ul className="space-y-2">
                  {section.links.map((link, linkIndex) => (
                    <li key={linkIndex}>
                      {link.external ? (
                        <a
                          href={link.to}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="text-sm text-white/70 hover:text-white transition-colors"
                        >
                          {link.label}
                        </a>
                      ) : (
                        <Link
                          to={link.to}
                          className="text-sm text-white/70 hover:text-white transition-colors"
                        >
                          {link.label}
                        </Link>
                      )}
                    </li>
                  ))}
                </ul>
              </nav>
            </div>
          ))}

          <div>
            <h3 className="text-xs font-semibold uppercase tracking-wider text-white/50 mb-3">
              Contact
            </h3>
            <div className="space-y-2">
              <div>
                <span className="text-xs text-white/40 block">Email</span>
                <a
                  href={`mailto:${contactEmail}`}
                  className="text-sm text-white/70 hover:text-white transition-colors"
                >
                  {contactEmail}
                </a>
              </div>
              <div>
                <span className="text-xs text-white/40 block mb-1">Follow us</span>
                <a
                  href="https://www.instagram.com/voxire.tech"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-white/70 hover:text-white transition-colors"
                  aria-label="Follow us on Instagram"
                >
                  <FaInstagram size={18} />
                </a>
              </div>
            </div>
          </div>
        </div>

        <div className="border-t border-white/10 pt-6 flex flex-col sm:flex-row items-center justify-between gap-3">
          <div>
            <p className="text-sm text-white/50">
              &copy; {new Date().getFullYear()}{' '}
              <a
                href="https://voxire.com"
                target="_blank"
                rel="noopener noreferrer"
                className="hover:text-white transition-colors"
              >
                Voxire
              </a>{' '}
              All rights reserved.
            </p>
            <p className="text-xs text-white/30 mt-1">
              Unless otherwise noted, the use of third-party logos does not imply endorsement,
              sponsorship, or affiliation with Rtylr.
            </p>
          </div>
          <button
            type="button"
            className="text-xs text-white/30 hover:text-white/50 transition-colors"
            aria-label={`Current version: ${appVersion}`}
          >
            v{appVersion}
          </button>
        </div>
      </div>
    </footer>
  );
};

export default Footer;
