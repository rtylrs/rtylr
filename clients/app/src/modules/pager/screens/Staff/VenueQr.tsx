import React from 'react';
import { QRCode } from 'react-qrcode-logo';
import { QrCodeTemplates } from '../../../../features/qr-codes/templates';

// The qr-codes feature's plain black-on-white template: it scans best from
// across a counter.
const QR_STYLE = QrCodeTemplates[0];

/** The venue's one QR, the same on the board, the counter display and the poster. */
export function VenueQr({ url, size }: { url: string; size: number }): React.JSX.Element {
  return (
    <QRCode
      value={url}
      size={size}
      fgColor={QR_STYLE.fgColor}
      bgColor={QR_STYLE.bgColor}
      eyeRadius={QR_STYLE.eyeRadius}
      qrStyle={QR_STYLE.qrStyle}
    />
  );
}
