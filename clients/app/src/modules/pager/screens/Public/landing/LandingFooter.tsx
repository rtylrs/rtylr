import React from 'react';
import { useTranslation } from 'react-i18next';
import { APP_URLS, RTYLR_CONTACT_CHANNELS, type RtylrContactChannelId } from '@voxire/config';
import { Mail, MessageCircle, Phone } from 'lucide-react';
import type { LucideIcon } from 'lucide-react';
import LanguageSwitch from '../../../shared/components/LanguageSwitch';
import { ByRtylr, RenniliLogo, RtylrBrandText } from '../../../shared/components/RenniliLogo';
import {
  PAGER_FAQ_PATH,
  PAGER_FEATURES_PATH,
  PAGER_HOW_PATH,
  PAGER_LANDING_PATH,
  PAGER_RTYLR_PATH,
  PAGER_TRY_PATH,
} from '../../../pagerPaths';

const CHANNEL_ICONS: Record<RtylrContactChannelId, LucideIcon> = {
  whatsapp: MessageCircle,
  phone: Phone,
  email: Mail,
};

interface LandingFooterProps {
  onLanding?: boolean;
}

export const LandingFooter = (_props: LandingFooterProps): React.JSX.Element => {
  const { t } = useTranslation('pager');
  const year = new Date().getFullYear();

  const navLinks = [
    { href: PAGER_LANDING_PATH, label: t('pager.landing.nav.home') },
    { href: PAGER_HOW_PATH, label: t('pager.landing.nav.how') },
    { href: PAGER_FEATURES_PATH, label: t('pager.landing.nav.features') },
    { href: PAGER_RTYLR_PATH, label: t('pager.landing.nav.rtylr') },
    { href: PAGER_FAQ_PATH, label: t('pager.landing.nav.faq') },
    { href: PAGER_TRY_PATH, label: t('pager.landing.nav.tryCta') },
  ];

  const legalLinks = [
    { href: `${APP_URLS.landing}/legal/terms/`, label: t('pager.landing.footer.terms') },
    { href: `${APP_URLS.landing}/legal/privacy-policy/`, label: t('pager.landing.footer.privacy') },
    { href: `${APP_URLS.landing}/legal/security/`, label: t('pager.landing.footer.security') },
    { href: 'https://voxire.com', label: t('pager.landing.footer.builtBy') },
  ];

  return (
    <footer className="rl-footer">
      <div className="rl-wrap">
        <div className="rl-footer-contact">
          <h2 className="rl-footer-heading">
            <RtylrBrandText text={t('pager.landing.footer.contactTitle')} />
          </h2>
          <div className="rl-footer-channels">
            {RTYLR_CONTACT_CHANNELS.map((channel, index) => {
              const Icon = CHANNEL_ICONS[channel.id];
              const featured = index === 0;
              return (
                <a
                  key={channel.id}
                  href={channel.href}
                  target={channel.href.startsWith('http') ? '_blank' : undefined}
                  rel={channel.href.startsWith('http') ? 'noreferrer' : undefined}
                  className={`rl-channel-card ${featured ? 'rl-channel-card--featured' : ''}`}
                >
                  <span className="rl-channel-icon" aria-hidden="true">
                    <Icon size={20} strokeWidth={2} />
                  </span>
                  <span className="rl-channel-body">
                    <span className="rl-channel-label">
                      {t(`pager.landing.footer.channels.${channel.id}.label`)}
                    </span>
                    <span className="rl-channel-value" dir="ltr">
                      {channel.value}
                    </span>
                    <span className="rl-channel-hint">
                      {t(`pager.landing.footer.channels.${channel.id}.hint`)}
                    </span>
                  </span>
                </a>
              );
            })}
          </div>
        </div>

        <div className="rl-footer-bottom">
          <div className="rl-footer-brand">
            <div className="rl-footer-brand-row">
              <RenniliLogo size={26} />
              <ByRtylr />
            </div>
            <p className="rl-footer-tagline">
              <RtylrBrandText text={t('pager.landing.footer.tagline')} />
            </p>
          </div>

          <div className="rl-footer-links-col">
            <ul className="rl-footer-links">
              {navLinks.map((link) => (
                <li key={link.href}>
                  <a href={link.href} className="rl-footer-link">
                    <RtylrBrandText text={link.label} />
                  </a>
                </li>
              ))}
            </ul>
            <ul className="rl-footer-legal">
              {legalLinks.map((link) => (
                <li key={link.href}>
                  <a
                    href={link.href}
                    target="_blank"
                    rel="noreferrer"
                    className="rl-footer-legal-link"
                  >
                    {link.label}
                  </a>
                </li>
              ))}
            </ul>
          </div>

          <div className="rl-footer-end">
            <LanguageSwitch />
            <p className="rl-footer-copy" dir="ltr">
              <RtylrBrandText text={t('pager.landing.footer.copyright', { year })} />
            </p>
          </div>
        </div>
      </div>
    </footer>
  );
};
