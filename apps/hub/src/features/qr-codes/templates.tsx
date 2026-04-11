import React from 'react';
import { QRCode } from 'react-qrcode-logo';

export interface QrOptions {
  name?: string;
  text?: string;
  description?: string;
  size: number;
  fgColor: string;
  bgColor: string;
  logoImage?: boolean;
  logoWidth: number;
  logoHeight: number;
  eyeRadius: number;
  qrStyle: 'squares' | 'dots' | 'fluid' | undefined;
  backgroundColor?: string;
  url?: string;
  textColor?: string;
  eyeColor?: string;
}

export const QrCodeTemplates: QrOptions[] = [
  {
    name: 'template1',
    text: '',
    description: '',
    size: 50,
    fgColor: '#000000',
    bgColor: '#ffffff',
    logoWidth: 50,
    logoHeight: 50,
    eyeRadius: 10,
    qrStyle: 'squares',
    backgroundColor: '#ffffff',
    logoImage: false,
  },
  {
    name: 'template2',
    text: 'Table',
    description: 'Scan Me',
    size: 50,
    fgColor: '#000000',
    bgColor: '#ffffff',
    logoWidth: 50,
    logoHeight: 50,
    eyeRadius: 10,
    qrStyle: 'squares',
    logoImage: false,
  },
  {
    name: 'template3',
    text: 'Table',
    description: 'Scan Me',
    size: 50,
    fgColor: '#000000',
    bgColor: '#ffeb3b',
    logoWidth: 50,
    logoHeight: 50,
    eyeRadius: 10,
    qrStyle: 'squares',
    logoImage: false,
  },
  {
    name: 'template4',
    text: 'Table',
    description: 'Scan Me',
    size: 50,
    fgColor: '#000000',
    bgColor: '#ffeb3b',
    logoWidth: 50,
    logoHeight: 50,
    eyeRadius: 10,
    qrStyle: 'squares',
    logoImage: false,
  },
  {
    name: 'template5',
    text: 'Table',
    description: 'Scan Me',
    size: 50,
    fgColor: '#000000',
    bgColor: '#ffeb3b',
    logoWidth: 50,
    logoHeight: 50,
    eyeRadius: 10,
    qrStyle: 'squares',
    logoImage: false,
  },
  {
    name: 'template6',
    text: 'Table',
    description: 'Scan Me',
    size: 80,
    fgColor: '#000000',
    bgColor: '#ffeb3b',
    logoWidth: 50,
    logoHeight: 50,
    eyeRadius: 10,
    qrStyle: 'squares',
    logoImage: true,
  },
  {
    name: 'template7',
    text: 'Table',
    description: 'Scan Me',
    size: 80,
    fgColor: '#000000',
    bgColor: '#ffffff',
    logoWidth: 50,
    logoHeight: 50,
    eyeRadius: 10,
    qrStyle: 'squares',
    logoImage: true,
  },
];

interface QrCodeTemplateProps extends QrOptions {
  setQrOptions: () => void;
  logo: string;
}

export const QrCodeTemplate: React.FC<QrCodeTemplateProps> = ({
  name,
  url,
  // text,
  // description,
  fgColor,
  bgColor,
  logoImage,
  eyeRadius,
  size,
  qrStyle,
  backgroundColor,
  setQrOptions,
  logo,
}) => {
  return (
    <div key={name} className="qrOptions" onClick={() => setQrOptions()}>
      <div
        style={{
          width: 100,
          height: 100,
          backgroundColor: backgroundColor || bgColor,
          display: 'flex',
          justifyContent: 'center',
          alignItems: 'center',
        }}
      >
        <QRCode
          value={url}
          size={size}
          fgColor={fgColor}
          bgColor={bgColor}
          logoWidth={50}
          logoHeight={30}
          eyeRadius={eyeRadius}
          qrStyle={qrStyle}
          logoPadding={5}
          logoImage={logoImage ? logo : ''}
          removeQrCodeBehindLogo
        />
      </div>
    </div>
  );
};
