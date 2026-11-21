import React, { useState } from 'react';

interface VenueLogoProps {
  src?: string | null;
  className?: string;
}

/**
 * The business's own logo, beside its name wherever Rennili shows the venue.
 * Decorative (the name sits next to it), and gone rather than a broken image
 * when the business has none or the file does not load.
 */
export function VenueLogo({ src, className }: VenueLogoProps): React.JSX.Element | null {
  const [failedSrc, setFailedSrc] = useState<string | null>(null);
  const url = src?.trim() ?? '';
  if (!url || failedSrc === url) return null;
  return (
    <img
      src={url}
      alt=""
      className={className}
      decoding="async"
      onError={() => setFailedSrc(url)}
    />
  );
}
