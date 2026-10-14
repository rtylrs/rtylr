import { FaInstagram } from 'react-icons/fa';
import { Link } from '@tanstack/react-router';
import { APP_URLS } from '@voxire/config';
import React from 'react';
import { useTranslation } from 'react-i18next';

const Footer: React.FC = () => {
  const { t } = useTranslation('footer');
  const landingUrl = (APP_URLS.landing || 'https://rtylr.com').replace(/\/$/, '');
  const contactEmail = 'info@rtylr.com';
  const appVersion = import.meta.env.VITE_APP_VERSION ?? '0.0.0';

  const footerSections = [
    {
      title: t('footer.company'),
      links: [
        { to: `${landingUrl}/about`, label: 'About Us', external: true },
        { to: `${landingUrl}/contact`, label: 'Contact Us', external: true },
        { to: 'https://hire.voxire.com/jobs/', label: 'Careers', external: true },
      ],
    },
    {
      title: t('footer.quickLinks'),
      links: [
        { to: `${landingUrl}/support`, label: 'FAQs', external: true },
        { to: `${landingUrl}/pricing`, label: 'Pricing', external: true },
      ],
    },
    {
      title: t('footer.legal'),
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
          <img src="/logo.png" alt={t('footer.logoAlt')} className="h-8 w-auto" />
        </div>

        <div className="grid grid-cols-2 md:grid-cols-4 gap-8 mb-12">
          {footerSections.map((section, index) => (
            <div key={index}>
              <h3 className="text-xs font-semibold uppercase tracking-wider text-white/50 mb-3">
                {section.title}
              </h3>
              <nav aria-label={t('footer.sectionNavigation', { section: section.title })}>
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
              {t('footer.contact')}
            </h3>
            <div className="space-y-2">
              <div>
                <span className="text-xs text-white/40 block">{t('footer.email')}</span>
                {/* dir="ltr" because an address is Latin script in every language, and
                    the bidi algorithm otherwise moves its trailing punctuation in ar. */}
                <a
                  href={`mailto:${contactEmail}`}
                  dir="ltr"
                  className="text-sm text-white/70 hover:text-white transition-colors"
                >
                  {contactEmail}
                </a>
              </div>
              <div>
                <span className="text-xs text-white/40 block mb-1">{t('footer.followUs')}</span>
                <a
                  href="https://www.instagram.com/voxire.tech"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-white/70 hover:text-white transition-colors"
                  aria-label={t('footer.instagram')}
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
              {t('footer.rightsReserved')}
            </p>
            <p className="text-xs text-white/30 mt-1">{t('footer.thirdPartyLogos')}</p>
          </div>
          <button
            type="button"
            dir="ltr"
            className="text-xs text-white/30 hover:text-white/50 transition-colors"
            aria-label={t('footer.currentVersion', { version: appVersion })}
          >
            v{appVersion}
          </button>
        </div>
      </div>
    </footer>
  );
};

export default Footer;
