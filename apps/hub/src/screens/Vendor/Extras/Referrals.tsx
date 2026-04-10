import React from 'react';
import { Gift, Copy, Users, DollarSign } from 'lucide-react';
import { toast } from 'sonner';
import { Button } from '../../../components/ui';
import { Breadcrumbs, SettingSection, StatCard } from '../../../components/hub';

const Referrals: React.FC = () => {
  const referralCode = 'RTYLR-2K7X';
  const referralLink = `https://rtylr.com/invite?code=${referralCode}`;

  const copyLink = () => {
    navigator.clipboard.writeText(referralLink);
    toast.success('Referral link copied');
  };

  return (
    <div className="space-y-5 max-w-3xl">
      
      <div className="flex items-center gap-3">
        <div className="h-9 w-9 rounded-xl flex items-center justify-center" style={{ background: 'rgba(236,72,153,0.15)' }}>
          <Gift size={16} style={{ color: '#ec4899' }} />
        </div>
        <div>
          <h1 className="text-lg font-semibold" style={{ color: 'var(--hub-text)' }}>Referrals</h1>
          <p className="text-xs mt-0.5" style={{ color: 'var(--hub-text-dim)' }}>
            Invite other businesses to rtylr and earn rewards.
          </p>
        </div>
      </div>

      <div className="grid grid-cols-3 gap-3">
        <StatCard label="Invited" value="12" icon={Users} accent="#4583C4" />
        <StatCard label="Signed up" value="4" icon={Gift} accent="#ec4899" />
        <StatCard label="Earned" value="$120" icon={DollarSign} accent="#10b981" />
      </div>

      <SettingSection title="Your referral link" description="Share with other business owners. You get $30 credit per successful signup.">
        <div
          className="flex items-center gap-2 p-3 rounded-lg font-mono text-xs"
          style={{ background: 'var(--hub-hover)', border: '1px solid var(--hub-border)' }}
        >
          <span className="flex-1 truncate" style={{ color: 'var(--hub-text)' }}>{referralLink}</span>
          <Button size="sm" variant="secondary" onClick={copyLink}>
            <Copy size={12} />
            Copy
          </Button>
        </div>
      </SettingSection>
    </div>
  );
};

export default Referrals;
