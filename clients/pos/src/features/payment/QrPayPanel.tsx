import React, { useEffect, useRef, useState } from 'react';
import QRCode from 'qrcode';
import { Check } from 'lucide-react';

interface QrPayPanelProps {
  orderId: string;
  amount: number;
  /** Fires when the operator attests they checked the payment elsewhere. */
  onAttested?: () => void;
  /** Fires when the operator revokes a previous attestation. */
  onReset?: () => void;
}

/**
 * Renders a QR code for this order's amount plus the operator's manual
 * attestation control.
 *
 * The encoded URL is not served by anything: there is no payment gateway
 * backend and no webhook, so the app never learns that the customer paid. The
 * "attest" button therefore records the operator's word, not a provider
 * confirmation, and the label says so. Wiring a real gateway means replacing
 * this control with the gateway's callback, not enabling it.
 */
const QrPayPanel: React.FC<QrPayPanelProps> = ({
  orderId,
  amount,
  onAttested,
  onReset,
}) => {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const [url, setUrl] = useState('');
  const [copied, setCopied] = useState(false);
  const [attested, setAttested] = useState(false);

  const paymentUrl = `https://pay.api.voxire.com/checkout?ref=${encodeURIComponent(orderId)}&amount=${amount.toFixed(2)}`;

  useEffect(() => {
    setUrl(paymentUrl);
    // The amount is part of the URL, so this also clears an attestation made
    // for a previous amount.
    setAttested(false);
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
        . Nothing on this page confirms the payment back to the register.
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
        {onAttested && (
          <button
            type="button"
            aria-pressed={attested}
            onClick={() => {
              setAttested((v) => {
                if (v) onReset?.();
                else onAttested();
                return !v;
              });
            }}
            className={`h-9 px-4 rounded-lg text-xs font-semibold transition-colors ${
              attested
                ? 'bg-pos-text text-pos-bg'
                : 'border border-pos-border text-pos-muted hover:border-white/20'
            }`}
          >
            {attested ? 'Attested paid' : 'I checked this payment elsewhere'}
          </button>
        )}
      </div>
    </div>
  );
};

export default QrPayPanel;
