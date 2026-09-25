import React from 'react';

import { useConsent } from './useConsent';

/**
 * Footer link that reopens the consent banner. GDPR Art. 7(3) requires
 * withdrawing consent to be as easy as giving it; this is the surface
 * we point customers at when they want to revisit their choice.
 *
 * Renders nothing while the banner itself is showing (`undecided`) to
 * avoid a duplicate UI affordance. Otherwise renders an anchor that
 * consumers wrap in their footer's link styling - pass `className` to
 * inherit whatever footer-link class the host site uses.
 *
 * Behavior: clicking calls `clearConsent()` which deletes the cookie
 * on the same Domain it was set with, and dispatches the change event
 * so the CookieBanner remounts.
 */
interface ManageConsentLinkProps {
  /** Class applied to the rendered anchor. Use the same class your
   *  footer uses for its <Link>/<a> elements so styling matches. */
  className?: string;
  /** Override the visible label. Defaults to "Cookie preferences". */
  children?: React.ReactNode;
}

export function ManageConsentLink({
  className,
  children,
}: ManageConsentLinkProps): React.JSX.Element | null {
  const { state, reopen } = useConsent();
  if (state === 'undecided') return null;

  return (
    <a
      href="#cookie-preferences"
      onClick={(event) => {
        event.preventDefault();
        reopen();
      }}
      className={className}
      style={{
        cursor: 'pointer',
        ...(className
          ? {}
          : {
              color: 'inherit',
              font: 'inherit',
              textDecoration: 'underline',
              textUnderlineOffset: 2,
            }),
      }}
    >
      {children ?? 'Cookie preferences'}
    </a>
  );
}
