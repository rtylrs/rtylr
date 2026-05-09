import React from 'react';
import { Gift } from 'lucide-react';

const Referrals: React.FC = () => {
  return (
    <div className="space-y-5 max-w-3xl">
      <div className="flex items-center gap-3">
        <div
          className="h-9 w-9 rounded-xl flex items-center justify-center"
          style={{ background: 'rgba(236,72,153,0.15)' }}
        >
          <Gift size={16} style={{ color: '#ec4899' }} />
        </div>
        <div>
          <h1 className="text-lg font-semibold" style={{ color: 'var(--dash-text)' }}>
            Referrals
          </h1>
          <p className="text-xs mt-0.5" style={{ color: 'var(--dash-text-dim)' }}>
            Invite other businesses to rtylr and earn rewards.
          </p>
        </div>
      </div>

      <div
        className="rounded-xl p-8 flex flex-col items-center justify-center text-center"
        style={{ background: 'var(--dash-card)', border: '1px solid var(--dash-border)' }}
      >
        <Gift size={28} className="mb-3" style={{ color: 'var(--dash-text-dim)' }} />
        <p className="text-sm font-semibold" style={{ color: 'var(--dash-text)' }}>
          Referral program coming soon
        </p>
        <p className="text-xs mt-1" style={{ color: 'var(--dash-text-muted)' }}>
          Earn $30 credit for every business you refer. Your unique link will appear here when the
          program launches.
        </p>
      </div>
    </div>
  );
};

export default Referrals;
