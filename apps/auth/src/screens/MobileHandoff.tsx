import React from 'react';
import { Smartphone, RefreshCcw } from 'lucide-react';

import AuthShell from '../components/AuthShell';
import AuthButton from '../components/AuthButton';
import BackLink from '../components/BackLink';

const MobileHandoff: React.FC = () => {
  const token = `rtylr-handoff-${Date.now()}`;
  const qrUrl = `https://api.qrserver.com/v1/create-qr-code/?size=220x220&data=${encodeURIComponent(
    `https://auth.rtylr.com/magic-link/confirm?token=${token}`
  )}`;

  return (
    <AuthShell
      title="Scan to sign in"
      subtitle="Already signed in on your phone? Scan this QR code to continue on your desktop."
      footer={
        <div className="flex justify-center">
          <BackLink label="Back to sign in" />
        </div>
      }
    >
      <div className="flex flex-col items-center gap-5">
        <div className="flex h-14 w-14 items-center justify-center rounded-2xl bg-[#64C5B9]/15 border border-[#64C5B9]/25">
          <Smartphone className="h-6 w-6 text-[#64C5B9]" />
        </div>

        <div className="rounded-2xl border border-white/10 bg-white/[0.03] p-5">
          <img
            src={qrUrl}
            alt="Mobile handoff QR code"
            className="h-[220px] w-[220px] rounded-xl bg-white p-2"
          />
        </div>

        <ol className="w-full flex flex-col gap-2 text-sm text-white/65">
          <li className="flex gap-2">
            <span className="font-semibold text-[#64C5B9]">1.</span>
            Open the rtylr app on your phone
          </li>
          <li className="flex gap-2">
            <span className="font-semibold text-[#64C5B9]">2.</span>
            Tap the scan icon in the top right
          </li>
          <li className="flex gap-2">
            <span className="font-semibold text-[#64C5B9]">3.</span>
            Point your camera at the QR code above
          </li>
        </ol>

        <AuthButton variant="outline">
          <span className="inline-flex items-center justify-center gap-2">
            <RefreshCcw className="h-[18px] w-[18px]" />
            Refresh code
          </span>
        </AuthButton>
      </div>
    </AuthShell>
  );
};

export default MobileHandoff;
