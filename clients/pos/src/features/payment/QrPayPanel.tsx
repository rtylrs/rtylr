import React, { useEffect, useRef, useState } from 'react';
import QRCode from 'qrcode';
import { Check } from 'lucide-react';

interface QrPayPanelProps {
  orderId: string;
  amount: number;
  /** Fires when the QR code has been scanned and payment confirmed */
  onConfirmed?: () => void;
}

/**
 * Generates a QR code encoding a payment URL for this order.
 * The URL format is: https://pay.api.voxire.com/checkout?ref={orderId}&amount={amount}
 * A real payment gateway backend would serve that URL and POST back a webhook on completion.
 */
const QrPayPanel: React.FC<QrPayPanelProps> = ({ orderId, amount, onConfirmed }) => {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const [url, setUrl] = useState('');
  const [copied, setCopied] = useState(false);

  const paymentUrl = `https://pay.api.voxire.com/checkout?ref=${encodeURIComponent(orderId)}&amount=${amount.toFixed(2)}`;

  useEffect(() => {
    setUrl(paymentUrl);
    if (canvasRef.current) {
      QRCode.toCanvas(canvasRef.current, paymentUrl, {
        width: 220,
        margin: 2,
        color: { dark: '#000000', light: '#ffffff' },
        errorCorrectionLevel: 'M',
      }).catch(console.error);
    }
  }, [paymentUrl]);

  const copyTimer = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(
    () => () => {
      if (copyTimer.current !== null) clearTimeout(copyTimer.current);
    },
    []
  );

  const copyLink = async () => {
    await navigator.clipboard.writeText(url).catch(() => undefined);
    setCopied(true);
    if (copyTimer.current !== null) clearTimeout(copyTimer.current);
    copyTimer.current = setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="flex flex-col items-center gap-4">
      <div className="rounded-2xl bg-white p-3 shadow-inner">
        <canvas ref={canvasRef} className="block" />
      </div>
      <p className="text-xs text-pos-muted text-center max-w-[220px]">
        Customer scans this code to pay{' '}
        <span className="font-semibold text-pos-text">
          ${amount.toFixed(2)}
        </span>
        . Confirm once payment notification arrives.
      </p>
      <div className="flex gap-2">
        <button
          type="button"
          onClick={copyLink}
          className="inline-flex h-9 items-center justify-center gap-1.5 px-3 rounded-lg border border-pos-border text-pos-muted text-xs hover:border-white/20 transition-colors"
        >
          {copied && <Check className="h-3.5 w-3.5" aria-hidden="true" />}
          {copied ? 'Copied' : 'Copy link'}
        </button>
        {onConfirmed && (
          <button
            type="button"
            onClick={onConfirmed}
            className="h-9 px-4 rounded-lg bg-emerald-600 text-white text-xs font-semibold hover:bg-emerald-700 transition-colors"
          >
            Confirm payment
          </button>
        )}
      </div>
    </div>
  );
};

export default QrPayPanel;
