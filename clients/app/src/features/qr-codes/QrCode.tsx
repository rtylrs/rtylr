import React from 'react';
import { useState, useRef, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { cn } from '@voxire/ui';
import { HelmetContainer } from '@voxire/components';
import { toast } from 'sonner';
import { QRCode } from 'react-qrcode-logo';
import { saveAs } from 'file-saver';
import { useUserStore } from '../../shared/stores/userStore';
import { QrCodeTemplate, QrCodeTemplates, QrOptions } from './templates';

import { PORTAL_URL, buildPublicMenuUrl, buildPublicMenuUrlById } from '@voxire/config';

function QrCode() {
  const { t } = useTranslation();
  const { user } = useUserStore();
  const businessSlug = user?.businessSlug || user?.business_slug || '';
  const businessId = user?.businessId ?? user?.id;
  // A slug gives the canonical /<slug>/menu URL; without one the business is
  // addressed by id in the POS-style /b/<id>/menu form.
  const menuUrl = businessSlug
    ? buildPublicMenuUrl(businessSlug)
    : businessId != null
      ? buildPublicMenuUrlById(businessId)
      : PORTAL_URL;
  const initialUrl = menuUrl;

  const [qrOptions, setQrOptions] = useState<QrOptions>({
    url: initialUrl,
    size: 150,
    fgColor: '#000000',
    bgColor: '#ffffff',
    logoImage: false,
    logoWidth: 50,
    logoHeight: 50,
    eyeRadius: 10,
    qrStyle: 'squares',
  });
  const [isLogoLoaded, setIsLogoLoaded] = useState(false);

  const qrRef = useRef(null);

  useEffect(() => {
    setQrOptions((current) => ({
      ...current,
      url: menuUrl,
    }));
  }, [menuUrl]);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    const target = e.target as HTMLInputElement;
    const value =
      target.type === 'checkbox'
        ? target.checked
        : target.type === 'number'
          ? // <input type="number"> reports a string; react-qrcode-logo only
            // honors eyeRadius/logoWidth/logoHeight when they are numbers.
            Number(target.value) || 0
          : target.value;
    const name = target.name;
    setQrOptions((current) => ({ ...current, [name]: value }));
  };

  const handleDownload = async () => {
    if (qrOptions.logoImage && !isLogoLoaded) {
      toast.error('Please wait for the logo to load before downloading.');
      return;
    }

    if (!qrRef.current) return;
    try {
      const { default: html2canvas } = await import('html2canvas');
      const canvas = await html2canvas(qrRef.current);
      canvas.toBlob((blob) => {
        if (blob) {
          saveAs(blob, 'qr-code.png');
        } else {
          toast.error(t('qrCodes.exportFailed', 'Failed to export the QR code.'));
        }
      });
    } catch {
      toast.error(t('qrCodes.exportFailed', 'Failed to export the QR code.'));
    }
  };

  useEffect(() => {
    if (!qrOptions.logoImage) {
      setIsLogoLoaded(true);
      return;
    }
    const logoSrc = typeof user?.logo === 'string' ? user.logo : '';
    if (!logoSrc) {
      // No logo exists, so nothing is awaited and the download must not be blocked.
      setIsLogoLoaded(true);
      return;
    }
    let cancelled = false;
    setIsLogoLoaded(false);
    const img = new Image();
    img.onload = () => {
      if (!cancelled) setIsLogoLoaded(true);
    };
    img.onerror = () => {
      if (cancelled) return;
      toast.error('Failed to load logo image.');
      setIsLogoLoaded(false);
    };
    img.src = logoSrc;
    return () => {
      cancelled = true;
      img.onload = null;
      img.onerror = null;
    };
  }, [qrOptions.logoImage, user?.logo]);

  return (
    <>
      <HelmetContainer title="QR code generator · rtylr" />
      <div className="space-y-6">
        <h1 className="text-2xl font-bold text-primary">QR Code Generator</h1>
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <div
            className="rounded-xl p-5 border shadow-card space-y-4"
            style={{ background: 'var(--dash-card)', borderColor: 'var(--dash-border)' }}
          >
            <div className="flex flex-wrap gap-2">
              {QrCodeTemplates.map((template) => (
                <div
                  key={template.name}
                  className={cn(
                    'cursor-pointer rounded-lg border-2 p-1 transition-colors',
                    template.name === qrOptions.name ? 'border-accent' : 'border-black/10'
                  )}
                >
                  <QrCodeTemplate
                    {...template}
                    url={qrOptions.url}
                    logo={(user?.logo as string) ?? ''}
                    setQrOptions={() => {
                      setQrOptions({
                        ...qrOptions,
                        ...template,
                        size: 150,
                      });
                    }}
                  />
                </div>
              ))}
            </div>

            <div className="flex items-center gap-3">
              <label className="text-sm font-medium text-primary w-40">Background Color:</label>
              <input
                type="color"
                name="backgroundColor"
                value={qrOptions.backgroundColor}
                onChange={handleInputChange}
              />
            </div>

            <div className="flex items-center gap-3">
              <label className="text-sm font-medium text-primary w-40">QR Background Color:</label>
              <input
                type="color"
                name="bgColor"
                value={qrOptions.bgColor}
                onChange={handleInputChange}
              />
            </div>

            <div className="flex items-center gap-3">
              <label className="text-sm font-medium text-primary w-40">Color:</label>
              <input
                type="color"
                name="fgColor"
                value={qrOptions.fgColor}
                onChange={handleInputChange}
              />
            </div>

            <div className="flex items-center gap-3">
              <label className="text-sm font-medium text-primary w-40">Text Color:</label>
              <input
                type="color"
                name="textColor"
                value={qrOptions.textColor}
                onChange={handleInputChange}
              />
            </div>

            <div className="flex items-center gap-3">
              <label className="text-sm font-medium text-primary w-40">Logo Width:</label>
              <input
                className="border border-black/10 rounded px-2 py-1 text-sm w-20"
                type="number"
                name="logoWidth"
                value={qrOptions.logoWidth}
                onChange={handleInputChange}
              />
            </div>

            <div className="flex items-center gap-3">
              <label className="text-sm font-medium text-primary w-40">Logo Height:</label>
              <input
                className="border border-black/10 rounded px-2 py-1 text-sm w-20"
                type="number"
                name="logoHeight"
                value={qrOptions.logoHeight}
                onChange={handleInputChange}
              />
            </div>

            <div className="flex items-center gap-3">
              <label className="text-sm font-medium text-primary w-40">Eye Radius:</label>
              <input
                className="border border-black/10 rounded px-2 py-1 text-sm w-20"
                type="number"
                name="eyeRadius"
                value={qrOptions.eyeRadius}
                onChange={handleInputChange}
              />
            </div>

            <div className="flex items-center gap-3">
              <label className="text-sm font-medium text-primary w-40">Eye Color:</label>
              <input
                type="color"
                name="eyeColor"
                value={qrOptions.eyeColor}
                onChange={handleInputChange}
              />
            </div>

            <div className="flex items-center gap-3">
              <label className="text-sm font-medium text-primary w-40">QR Style:</label>
              <select
                className="border border-black/10 rounded px-2 py-1 text-sm"
                name="qrStyle"
                value={qrOptions.qrStyle}
                onChange={handleInputChange}
              >
                <option value="squares">Squares</option>
                <option value="dots">Dots</option>
                <option value="fluid">Fluid</option>
              </select>
            </div>

            <div className="flex items-center gap-3">
              <label className="text-sm font-medium text-primary w-40">Title:</label>
              <input
                className="border border-black/10 rounded px-2 py-1 text-sm flex-1"
                type="text"
                name="text"
                value={qrOptions.text}
                onChange={handleInputChange}
              />
            </div>

            <div className="flex items-center gap-3">
              <label className="text-sm font-medium text-primary w-40">Description:</label>
              <input
                className="border border-black/10 rounded px-2 py-1 text-sm flex-1"
                type="text"
                name="description"
                value={qrOptions.description}
                onChange={handleInputChange}
              />
            </div>

            <div className="flex items-center gap-3">
              <label className="text-sm font-medium text-primary w-40">Display Logo:</label>
              <input
                type="checkbox"
                name="logoImage"
                checked={qrOptions.logoImage}
                onChange={handleInputChange}
              />
            </div>
          </div>

          <div
            className="rounded-xl p-5 border shadow-card flex flex-col items-center gap-4"
            style={{ background: 'var(--dash-card)', borderColor: 'var(--dash-border)' }}
          >
            <div
              ref={qrRef}
              className="flex flex-col items-center justify-center rounded-lg overflow-hidden max-w-full"
              style={{
                width: qrOptions.size + 200,
                height: qrOptions.size + 200,
                maxWidth: '100%',
                backgroundColor: qrOptions.backgroundColor || qrOptions.bgColor,
              }}
            >
              {qrOptions.text && (
                <p style={{ color: qrOptions.textColor || qrOptions.fgColor }}>
                  <strong>{qrOptions.text}</strong>
                </p>
              )}
              <QRCode
                value={qrOptions.url}
                size={qrOptions.size + 100}
                fgColor={qrOptions.fgColor}
                bgColor={qrOptions.bgColor}
                logoImage={qrOptions.logoImage ? (user?.logo as string) : ''}
                logoWidth={qrOptions.logoWidth}
                logoHeight={qrOptions.logoHeight}
                eyeRadius={qrOptions.eyeRadius}
                qrStyle={qrOptions.qrStyle}
                eyeColor={qrOptions.eyeColor}
                logoPadding={5}
                removeQrCodeBehindLogo
              />
              {qrOptions.description && (
                <p style={{ color: qrOptions.textColor || qrOptions.fgColor }}>
                  {qrOptions.description}
                </p>
              )}
            </div>
            <button
              type="button"
              className="px-4 py-2 bg-[var(--dash-active)] text-[var(--dash-text)] rounded-lg text-sm font-medium hover:opacity-90 transition-opacity"
              onClick={handleDownload}
            >
              Download QR Code
            </button>
          </div>
        </div>
      </div>
    </>
  );
}

export default QrCode;
