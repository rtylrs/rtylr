import React from 'react';
import { cn } from '@voxire/ui';
import { cdnImageUrl } from '@voxire/config';
import type { Customization } from '../../shared/lib/types';

import { cleanMojibake } from '../../shared/lib/sanitize';

interface HeaderProps {
  customization: Customization;
  compact?: boolean;
}

// Shared header surface. Each template can wrap or replace it, but
// defaulting here keeps branding consistent across templates and lets
// the user-supplied header_html inject raw markup (e.g. a nav back to
// the main site) without the template having to know about it.
const Header: React.FC<HeaderProps> = ({ customization, compact = false }) => {
  const logoUrl = cdnImageUrl(customization.logoUrl);
  const brandName = cleanMojibake(customization.brandName);
  const tagline = cleanMojibake(customization.tagline);

  return (
    <header
      className="border-b"
      style={{
        borderColor: 'var(--recruit-border)',
        background: 'var(--recruit-surface)',
      }}
    >
      {customization.headerHtml && (
        <div
          className="px-6 py-2 text-xs"
          style={{ color: 'var(--recruit-muted)' }}
          // Businesses supply the header HTML themselves; rendering raw
          // lets them link back to their main site and drop a light
          // navbar without us shipping a full editor. Sanitisation is
          // their responsibility since only authenticated admins can
          // set this field.
          dangerouslySetInnerHTML={{ __html: customization.headerHtml }}
        />
      )}
      <div
        className={cn('mx-auto flex max-w-5xl items-center gap-3 px-6', compact ? 'py-4' : 'py-6')}
      >
        {logoUrl ? (
          <img src={logoUrl} alt={`${brandName} logo`} className="h-8 w-auto" />
        ) : (
          <div
            className="flex h-8 w-8 items-center justify-center rounded-md text-sm font-semibold text-white"
            style={{ background: 'var(--recruit-accent)' }}
          >
            {brandName.slice(0, 1).toUpperCase()}
          </div>
        )}
        <div className="min-w-0">
          <div className="truncate text-sm font-semibold">{brandName}</div>
          {!compact && tagline && (
            <div className="truncate text-xs" style={{ color: 'var(--recruit-muted)' }}>
              {tagline}
            </div>
          )}
        </div>
      </div>
    </header>
  );
};

export default Header;
